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
	app.render(w, "login.html", nil)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, "about.html", nil)
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	app.render(w, "contact.html", nil)
}
