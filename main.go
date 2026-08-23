package main

import (
	"database/sql"
	"log"
	"net/http"

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
}

func main() {
	mux := http.NewServeMux()

	db, err := connectToDatabase("users_database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &application{
		errorLogger: log.New(log.Writer(), "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(log.Writer(), "INFO\t", log.Ldate|log.Ltime),
		userRepo:    NewSQLUserRepository(db),
		mux:         mux,
		templateDir: "./templates",
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
