package main

import (
	"net/http"
)

func (app *application) render(w http.ResponseWriter, filename string, data interface{}) {

	if app.tp == nil {
		http.Error(w, "template renderer is not initialized", http.StatusInternalServerError)
		return
	}

	app.tp.Render(w, filename, data)
}
