package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	defaultMiddleware := alice.New(app.recover, app.logger)
	secureMiddleware := alice.New(app.session.Enable)

	fileServer := http.FileServer(http.Dir("./public/"))
	mux.Handle("/public/", http.StripPrefix("/public", fileServer))
	mux.Handle("/", secureMiddleware.ThenFunc(app.home))
	mux.Handle("/register", secureMiddleware.ThenFunc(app.register))
	mux.Handle("/submit", secureMiddleware.Append(app.requireAuth).ThenFunc(app.submit))
	mux.Handle("/login", secureMiddleware.ThenFunc(app.login))
	mux.Handle("/about", secureMiddleware.ThenFunc(app.about))
	mux.Handle("/contact", secureMiddleware.ThenFunc(app.contact))

	handler := defaultMiddleware.Then(mux)
	return handler
}
