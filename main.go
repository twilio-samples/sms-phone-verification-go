package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

var rxPhone = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

type App struct {
	sessionName, flashKey string
	client                *twilio.RestClient
	store                 *sessions.CookieStore
}

type VerificationResponse struct {
	Message string
	Error   bool
}

type ValidationCodeRequest struct {
	Username, Password, Number string
	Errors                     map[string]string
}

// Validate validates if the username, password, and phone number supplied pass
// the validation criteria.
func (v *ValidationCodeRequest) Validate() bool {
	v.Errors = make(map[string]string)

	if strings.TrimSpace(v.Username) == "" {
		v.Errors["username"] = "Please enter a username"
	}

	usernameLength := len(strings.TrimSpace(v.Username))
	if usernameLength < 5 || usernameLength > 255 {
		v.Errors["username"] = "Please enter a username"
	}

	if strings.TrimSpace(v.Password) == "" {
		v.Errors["password"] = "Please enter a password"
	}

	passwordLength := len(strings.TrimSpace(v.Password))
	if passwordLength < 10 {
		v.Errors["password"] = "Please enter a password at least 10 characters long"
	}

	match := rxPhone.Match([]byte(v.Number))
	if !match {
		v.Errors["Number"] = "Please enter a phone number in E.164 format"
	}

	return len(v.Errors) == 0
}

func render(w http.ResponseWriter, filename string, data interface{}) {
	files := []string{
		"./ui/templates/base.tmpl",
		filename,
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

// renderCodeRequestForm renders a form where the user can enter the details
// required to request a verification code. The form can display messages to the
// user indicating if there were errors submitting the form and if form
// submission was successful.
func (a App) renderCodeRequestForm(w http.ResponseWriter, r *http.Request) {
	template := "./ui/templates/code-request-form.tmpl"

	session, err := a.store.Get(r, a.sessionName)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashes := session.Flashes(a.flashKey)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(flashes) > 0 {
		flashMessage := flashes[0]
		if data, ok := flashMessage.(*ValidationCodeRequest); ok {
			render(w, template, data)
			return
		}
	}

	render(w, template, nil)
}

// processCodeRequestForm processes submission of the code request form. If the
// submitted details are valid, then a request is made to Twilio for a
// verification code to be sent to the user via SMS. If the submitted form
// details are not valid, the reader is redirected back to the code request
// form, where any supplied form details that were correct will be pre-filled in
// the form.
func (a App) processCodeRequestForm(w http.ResponseWriter, r *http.Request) {
	v := &ValidationCodeRequest{
		Number:   r.PostFormValue("number"),
		Password: r.PostFormValue("password"),
		Username: r.PostFormValue("username"),
	}

	if !v.Validate() {
		session, err := a.store.Get(r, a.sessionName)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.AddFlash(v, a.flashKey)
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	params := &verify.CreateVerificationParams{}
	params.SetChannel("sms")
	params.SetTo(v.Number)

	resp, err := a.client.VerifyV2.CreateVerification(os.Getenv("TWILIO_VERIFICATION_SID"), params)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.Status == nil {
		log.Println("response status was not set")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := a.store.Get(r, a.sessionName)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("storing phone number (%s) in session", v.Number)
	session.Values["number"] = v.Number
	err = session.Save(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/verify", http.StatusSeeOther)
}

// renderCodeVerificationForm renders a form where the user can validate a
// verification code which they received via SMS. The form can display messages
// to the user indicating if there were errors submitting the form and if form
// submission was successful.
func (a App) renderCodeVerificationForm(w http.ResponseWriter, r *http.Request) {
	template := "./ui/templates/code-verification-form.tmpl"

	session, err := a.store.Get(r, a.sessionName)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !a.phoneNumberSet(session) {
		log.Println("Phone number not available in session or not set")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	flashes := session.Flashes(a.flashKey)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(flashes) > 0 {
		if data, ok := flashes[0].(*VerificationResponse); ok {
			render(w, template, data)
			return
		}
	}

	render(w, template, nil)
}

func (a App) phoneNumberSet(session *sessions.Session) bool {
	val := session.Values["number"]
	number, ok := val.(string)
	return ok && len(number) > 0
}

// processCodeVerificationForm processes submission of the code verification
// form. If the code is valid, then the user is redirected to the faux
// post-login page with a message saying that they are now logged in. If the
// code is not valid, the reader is redirected back to the code verification
// form.
func (a App) processCodeVerificationForm(w http.ResponseWriter, r *http.Request) {
	code := r.PostFormValue("code")

	session, err := a.store.Get(r, a.sessionName)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	val := session.Values["number"]
	number, ok := val.(string)
	if !ok || len(number) == 0 {
		log.Println("phone number not available in session or not set")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(number)
	params.SetCode(code)
	resp, err := a.client.VerifyV2.CreateVerificationCheck(os.Getenv("TWILIO_VERIFICATION_SID"), params)

	if err != nil {
		log.Println(err.Error())

		session.AddFlash(
			&VerificationResponse{
				Message: "The verification code was not valid",
				Error:   true,
			},
			a.flashKey,
		)
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	log.Printf("validation status was: %s", *resp.Status)

	// Remove the phone number from the session
	delete(session.Values, "number")
	err = session.Save(r, w)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the logged/in route as the user is now authenticated
	http.Redirect(w, r, "/logged-in", http.StatusSeeOther)
}

// renderLoginPage renders a static HTML template telling the user that they are
// now logged in.
func (a App) renderLoginPage(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/templates/base.tmpl",
		"./ui/templates/logged-in.tmpl",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	store.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8,
		HttpOnly: true,
	}

	// Register the type so that it can be flashed to the session
	gob.Register(&ValidationCodeRequest{})
	gob.Register(&VerificationResponse{})

	app := App{
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: os.Getenv("TWILIO_ACCOUNT_SID"),
			Password: os.Getenv("TWILIO_AUTH_TOKEN"),
		}),
		flashKey:    os.Getenv("FLASH_KEY"),
		sessionName: os.Getenv("SESSION_NAME"),
		store:       store,
	}

	fileServer := http.FileServer(http.Dir("./static/"))

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /", app.renderCodeRequestForm)
	mux.HandleFunc("POST /", app.processCodeRequestForm)
	mux.HandleFunc("GET /verify", app.renderCodeVerificationForm)
	mux.HandleFunc("POST /verify", app.processCodeVerificationForm)
	mux.HandleFunc("GET /logged-in", app.renderLoginPage)

	fmt.Println("Server is running on port :8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
