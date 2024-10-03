package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
)

type App struct {
	client *twilio.RestClient
}

// renderCodeRequestForm renders a form where the user can enter the details
// required to request a verification code. The form can display messages to the
// user indicating if there were errors submitting the form and if form
// submission was successful.
func (a App) renderCodeRequestForm(w http.ResponseWriter, r *http.Request) {

}

// processCodeRequestForm processes submission of the code request form. If the
// submitted details are valid, then a request is made to Twilio for a
// verification code to be sent to the user via SMS. If the submitted form
// details are not valid, the reader is redirected back to the code request
// form, where any supplied form details that were correct will be pre-filled in
// the form.
func (a App) processCodeRequestForm(w http.ResponseWriter, r *http.Request) {

}

// renderCodeVerificationForm renders a form where the user can validate a
// verification code which they received via SMS. The form can display messages
// to the user indicating if there were errors submitting the form and if form
// submission was successful.
func (a App) renderCodeVerificationForm(w http.ResponseWriter, r *http.Request) {

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
	app := App{client: twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})}

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
