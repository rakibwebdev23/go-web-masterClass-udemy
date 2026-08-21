package main

import (
	"net/http"
	"time"
)

func (app *application) serve() error {
	serv := http.Server{
		Addr:    ":8080",
		ReadTimeout: 2 * time.Second,
		Handler: app.routes(),
	}
	return serv.ListenAndServe()
}