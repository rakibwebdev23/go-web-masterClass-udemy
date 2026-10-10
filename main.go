package main

import (
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"time"

	"github.com/golangcollege/sessions"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed script/table.sql
var databaseSchema string

// routing - mux
// routing - handler - controller - handler
// Get - homeage
// Post - create user

type application struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
	userRepo    UserRepository
	postRepo    PostRepository
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

	if _, err := db.Exec(databaseSchema); err != nil {
		log.Fatal(err)
	}

	//session setup for user login
	session := sessions.New([]byte("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4"))
	session.Lifetime = 24 * time.Hour
	session.Secure = false
	session.SameSite = http.SameSiteLaxMode

	app := &application{
		errorLogger: log.New(log.Writer(), "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(log.Writer(), "INFO\t", log.Ldate|log.Ltime),
		userRepo:    NewSQLUserRepository(db),
		postRepo:    NewSQLPostRepository(db),
		mux:         mux,
		templateDir: "./templates",
		session:     session,
	}

	app.tp = NewTemplateRenderer(app.templateDir, false)

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
