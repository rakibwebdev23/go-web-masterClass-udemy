package main

import (
	"net/http"
)

func (app *application) render(w http.ResponseWriter, r *http.Request, filename string, data *templateData) {

	if app.tp == nil {
		http.Error(w, "template renderer is not initialized", http.StatusInternalServerError)
		return
	}

	app.tp.Render(w, filename, app.defaultTemplateData(data, r))
}

func (app *application) defaultTemplateData(data *templateData, r *http.Request) *templateData {
	if data == nil {
		data = &templateData{}
	}

	if app.session != nil {
		data.Flash = app.session.PopString(r, "flash")
	}
	return data
}
