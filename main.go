package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/golangcollege/sessions"
	_ "github.com/mattn/go-sqlite3"
)

// routing - mux
// routing - handler - controller - handler
// Get - homeage
// Post - create user

type application struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
	userRepo    UserRepository
	mux         *http.ServeMux
	templateDir string
	tp          *TemplateRenderer
	session     *sessions.Session
}

func main() {
	mux := http.NewServeMux()

	db, err := connectToDatabase("users_database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//session setup for user login
	session := sessions.New([]byte("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4"))
	session.Lifetime = 24 * time.Hour
	session.Secure = false // ⚠️ local env false, production true
	session.SameSite = http.SameSiteLaxMode

	app := &application{
		errorLogger: log.New(log.Writer(), "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(log.Writer(), "INFO\t", log.Ldate|log.Ltime),
		userRepo:    NewSQLUserRepository(db),
		mux:         mux,
		templateDir: "./templates",
		session:     session,
	}

	app.tp = NewTemplateRenderer(app.templateDir, true)

	log.Println("Server is running on http://localhost:8080")

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}

func connectToDatabase(name string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
