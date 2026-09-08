package main

import "net/http"

var htmlContent = `...`

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.infoLogger.Printf("session data: %v", app.session.GetString(r, "userId"))
	app.render(w, "index.html", nil)
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	app.render(w, "register.html", nil)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := NewForm(r.PostForm)
		form.Required("email", "password").MaxLength("password", 100).MinLength("password", 6).IsEmail("email")

		if !form.Valid() {
			app.errorLogger.Printf("form is not valid", form.Errors)
			app.render(w, "login.html", nil)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		app.infoLogger.Printf("Email: %s, Password: %s", email, password)

	}

	app.render(w, "login.html", nil)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, "about.html", nil)
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	app.render(w, "contact.html", nil)
}

func (app *application) submit(w http.ResponseWriter, r *http.Request) {
	app.render(w, "submit.html", nil)
}
