package main

import (
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
	sessionName string
	client      *twilio.RestClient
	store       *sessions.CookieStore
}
}

type ValidationCodeRequest struct {
	Username, Password, Number string
	Errors                     map[string]string
}
	files := []string{
		"./ui/templates/base.tmpl",
		"./ui/templates/code-request-form.tmpl",
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
	if passwordLength < 5 || usernameLength > 255 {
		v.Errors["password"] = "Please enter a password between 5 and 255 characters"
	}

	match := rxPhone.Match([]byte(v.Number))
	if match == false {
		v.Errors["Number"] = "Please enter a phone number in E.164 format"
	}

	return len(v.Errors) == 0
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

	if v.Validate() == false {
		session, err := a.store.Get(r, a.sessionName)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.AddFlash(v)
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
	return
}

// renderCodeVerificationForm renders a form where the user can validate a
// verification code which they received via SMS. The form can display messages
// to the user indicating if there were errors submitting the form and if form
// submission was successful.
func (a App) renderCodeVerificationForm(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/templates/base.tmpl",
		"./ui/templates/code-verification-form.tmpl",
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

// processCodeVerificationForm processes submission of the code verification
// form. If the code is valid, then the user is redirected to the faux
// post-login page with a message saying that they are now logged in. If the
// code is not valid, the reader is redirected back to the code verification
// form.
func (a App) processCodeVerificationForm(w http.ResponseWriter, r *http.Request) {

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

	app := App{
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: os.Getenv("TWILIO_ACCOUNT_SID"),
			Password: os.Getenv("TWILIO_AUTH_TOKEN"),
		}),
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
