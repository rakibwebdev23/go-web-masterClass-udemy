package main

import (
	"html/template"
	"net/http"
	"path"
)

func (app *application) render(w http.ResponseWriter, filename string, data interface{}) {

	fulPath := path.Join(app.templateDir, filename)
	tmpl, err := template.ParseFiles(fulPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
