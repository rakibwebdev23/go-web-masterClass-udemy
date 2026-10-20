package main

import (
	"database/sql"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB
var testApp *application

// TestMain(m *testing.M) — সব test শুরু হওয়ার আগে database ও application setup করে, test চালায় এবং শেষে database বন্ধ করে।
func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	if err := testDB.Ping(); err != nil {
		panic(err)
	}
	testApp = setupApp(testDB)
	if err = setupTestSchema(testDB); err != nil {
		panic(err)
	}

	code := m.Run()
	testDB.Close()
	os.Exit(code)

}

// setupApp(db *sql.DB) — Test-এর জন্য application তৈরি করে; যেমন session, logger, repository ও template renderer setup করে।
func setupApp(db *sql.DB) *application {
	session := sessions.New([]byte("u46IpCV9y5VlurXXXODJEhgOY8m9JVE4"))
	session.Lifetime = 24 * time.Hour
	session.Secure = true

	app := &application{
		errorLog:    log.New(io.Discard, "", 0),
		infoLog:     log.New(io.Discard, "", 0),
		userRepo:    NewSQLUserRepository(db),
		postRepo:    NewSQLPostRepository(db),
		templateDir: "./templates",
		publicPath:  "./public",
		session:     session,
	}
	app.tp = NewTemplateRenderer(app.templateDir, false)
	return app
}

// setupTestSchema(db *sql.DB) — Test database-এ প্রয়োজনীয় SQL table তৈরি করে, যাতে test data insert ও query করা যায়।
func setupTestSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE users (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
   name TEXT NOT NULL,
   email TEXT NOT NULL UNIQUE,
   hashed_password TEXT NOT NULL,
   created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE profile (
     user_id INTEGER PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
     avatar TEXT NOT NULL,
     created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE posts (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
   url TEXT NOT NULL,
   title TEXT NOT NULL UNIQUE,
   user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
   created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  body TEXT NOT NULL,
  user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
  post_id INTEGER REFERENCES posts(post_id) ON DELETE CASCADE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE votes (
   user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
   post_id INTEGER REFERENCES posts(post_id) ON DELETE CASCADE,
   created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
   PRIMARY KEY (user_id, post_id)
);

	`
	_, err := db.Exec(schema)
	return err
}

// cleanupTestData(t *testing.T) — প্রতিটি test-এর পর database-এর test data মুছে দেয়, যাতে আগের test-এর data পরের test-কে প্রভাবিত না করে।
func cleanupTestData(t *testing.T) {
	tables := []string{
		"profile",
		"votes",
		"comments",
		"posts",
		"users",
	}
	for _, table := range tables {
		_, err := testDB.Exec("DELETE FROM " + table)
		assert.NoError(t, err)
	}

}

// মনে রাখার সহজ উপায়:
// - TestMain = Test শুরু ও শেষ করা।
// - setupApp = Application প্রস্তুত করা।
// - setupTestSchema = Database table তৈরি করা।
// - cleanupTestData = Test data পরিষ্কার করা।