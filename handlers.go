package main

import "net/http"

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, "index.html", nil)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, "about.html", nil)
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	app.render(w, "contact.html", nil)
}
