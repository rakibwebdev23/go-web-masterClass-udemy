package main

import "net/http"

var htmlContent = `...`

const (
	loggedInUserKey = "logged_in_user_id"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// app.infoLogger.Printf("session data: %v", app.session.GetString(r, "userId"))
	app.render(w, r, "index.html", nil)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := NewForm(r.PostForm)
		form.Required("email", "password").
			MaxLength("email", 255).
			MaxLength("password", 255).
			MinLength("password", 3).
			MinLength("email", 3).
			IsEmail("email")

		if !form.Valid() {
			form.Errors.Add("generic", "The data you submitted was not valid")
			app.render(w, r, "login.html", &templateData{
				Form: *form,
			})
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		_, err := app.userRepo.Authenticate(email, password)
		if err != nil {
			form.Errors.Add("generic", err.Error())
			app.render(w, r, "login.html", &templateData{
				Form: *form,
			})
			return
		}

		app.session.Put(r, loggedInUserKey, email)
		app.session.Put(r, "flash", "You are logged in")

		app.infoLogger.Println("Logged in")

		http.Redirect(w, r, "/submit", http.StatusSeeOther)
		return
	}

	app.render(w, r, "login.html", &templateData{
		Form: *NewForm(r.PostForm),
	})
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {

	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := NewForm(r.PostForm)
		form.Required("email", "password", "name").
			MaxLength("email", 255).
			MaxLength("password", 255).
			MinLength("password", 3).
			MinLength("email", 3).
			MinLength("name", 3).
			IsEmail("email")

		if !form.Valid() {
			form.Errors.Add("generic", "The data you submitted was not valid")
			app.render(w, r, "register.html", &templateData{
				Form: *form,
			})
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		name := r.FormValue("name")
		avatar := r.FormValue("avatar")

		_, err := app.userRepo.CreateUser(name, email, password, avatar)
		if err != nil {
			form.Errors.Add("generic", err.Error())
			app.render(w, r, "register.html", &templateData{
				Form: *form,
			})
			return
		}

		app.session.Put(r, "flash", "You are registered")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.render(w, r, "register.html", &templateData{
		Form: *NewForm(r.PostForm),
	})
}
func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "about.html", nil)
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "contact.html", nil)
}

func (app *application) submit(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "submit.html", nil)
}

// package main

// import "net/http"

// var htmlContent = `...`

// const (
// 	loggedInUserKey = "logged_in_user_id"
// )

// func (app *application) home(w http.ResponseWriter, r *http.Request) {
// 	// app.infoLogger.Printf("session data: %v", app.session.GetString(r, "userId"))
// 	app.render(w, r, "index.html", nil)
// }

// func (app *application) register(w http.ResponseWriter, r *http.Request) {
// 	app.render(w, r, "register.html", nil)
// }

// func (app *application) login(w http.ResponseWriter, r *http.Request) {

// 	if r.Method == http.MethodPost {
// 		if err := r.ParseForm(); err != nil {
// 			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
// 			return
// 		}

// 		form := NewForm(r.PostForm)
// 		form.Required("email", "password").MaxLength("password", 100).MinLength("password", 6).IsEmail("email")

// 		if !form.Valid() {
// 			app.errorLogger.Printf("form is not valid", form.Errors)
// 			app.render(w, r, "login.html", nil)
// 			return
// 		}

// 		email := r.FormValue("email")
// 		password := r.FormValue("password")
// 		name := r.FormValue("name")
// 		avatar := r.FormValue("avatar")

// 		app.infoLogger.Printf("Email: %s, Password: %s", email, password)

// 	}

// 	app.render(w, r, "login.html", nil)
// }

// func (app *application) about(w http.ResponseWriter, r *http.Request) {
// 	app.render(w, r, "about.html", nil)
// }

// func (app *application) contact(w http.ResponseWriter, r *http.Request) {
// 	app.render(w, r, "contact.html", nil)
// }

// func (app *application) submit(w http.ResponseWriter, r *http.Request) {
// 	app.render(w, r, "submit.html", nil)
// }
