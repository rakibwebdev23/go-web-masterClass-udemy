package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./public/"))
	mux.Handle("/public/", http.StripPrefix("/public", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/register", app.register)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/about", app.about)
	mux.HandleFunc("/contact", app.contact)

	return mux
}
