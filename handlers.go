package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

var htmlContent = `...`

const (
	loggedInUserKey = "logged_in_user_id"
)

func (app *application) readIntWithDefault(r *http.Request, key string, dvalue int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return dvalue
	}
	return v
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	filter := Filter{
		Query:    r.URL.Query().Get("q"),
		OrderBy:  r.URL.Query().Get("order_by"),
		Page:     app.readIntWithDefault(r, "page", 1),
		PageSize: app.readIntWithDefault(r, "page_size", 50),
	}

	posts, metadata, err := app.postRepo.GetAll(filter)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	app.infoLogger.Printf("\nMetadata: %+v\n", metadata)

	pageLink := func(page int) string {
		query := url.Values{}
		query.Set("q", filter.Query)
		query.Set("order_by", filter.OrderBy)
		query.Set("page", strconv.Itoa(page))
		query.Set("page_size", strconv.Itoa(filter.PageSize))
		return "/?" + query.Encode()
	}

	app.render(w, r, "index.html", &templateData{
		Posts:    posts,
		Metadata: metadata,
		NextLink: pageLink(metadata.NextPage),
		PrevLink: pageLink(metadata.PreviousPage),
	})
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

func (app *application) vote(w http.ResponseWriter, r *http.Request) {
	postID := app.readIntWithDefault(r, "post_id", 0)
	u := app.getUserFromContext(r.Context())

	err := app.postRepo.AddVote(u.ID, postID)
	if err != nil {
		app.errorLogger.Printf("error adding vote: %s\n", err.Error())
		app.session.Put(r, "flash", "voting failed")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	app.session.Put(r, "flash", fmt.Sprintf("You voted for post with id #%d", postID))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "contact.html", nil)
}

func (app *application) submit(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost{
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		form := NewForm(r.PostForm)
		form.Required("title", "url").
		MaxLength("title", 255).
		MaxLength("url", 255).
		MinLength("url", 3)

		if !form.Valid(){
		form.Errors.Add("generic", "This submited data was not valid")
		app.render(w, r, "submit.html", &templateData{
			Form: *form,
		})
		return
	}

	title := r.FormValue("title")
	url := r.FormValue("url")

	user := app.getUserFromContext(r.Context())
	id, err := app.postRepo.CreatePost(title, url, user.ID)
	if err != nil{
		app.errorLogger.Printf("error creating post: %s\n", err.Error())
		form.Errors.Add("generic", err.Error())
		app.render(w, r, "submit.html", &templateData{
			Form: *form,
		})
		return
	}

	app.session.Put(r, "flash", "Your link has been submitted")
	app.infoLogger.Printf("post created with %d", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
		 
	}

	app.render(w, r, "submit.html", &templateData{
		Form: *NewForm(r.PostForm),
	})
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, loggedInUserKey)
	app.session.Put(r, "flash", "You have been logged out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
