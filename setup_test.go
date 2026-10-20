
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

// testDB: সব test চালানোর জন্য ব্যবহৃত test database connection।
var testDB *sql.DB

// testApp: test করার জন্য তৈরি application-এর instance।
var testApp *application

// TestMain: সব test function চালানোর আগে একবার execute হয়।
// এখানে database ও application setup করা হয়।
// সব test শেষ হলে database connection বন্ধ করা হয়।
func TestMain(m *testing.M) {
	var err error

	// SQLite in-memory database তৈরি করা হচ্ছে।
	// :memory: ব্যবহার করায় database-টি memory-তে থাকবে।
	testDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	// Database-এর সঙ্গে connection ঠিক আছে কি না পরীক্ষা করা।
	if err := testDB.Ping(); err != nil {
		panic(err)
	}

	// Test-এর জন্য application setup করা।
	testApp = setupApp(testDB)

	// Test-এর জন্য প্রয়োজনীয় database table তৈরি করা।
	if err = setupTestSchema(testDB); err != nil {
		panic(err)
	}

	// সব test function চালানো।
	// m.Run() test-এর ফলাফল হিসেবে exit code ফেরত দেয়।
	code := m.Run()

	// সব test শেষ হওয়ার পর database connection বন্ধ করা।
	testDB.Close()

	// Test-এর exit code দিয়ে program বন্ধ করা।
	// 0 মানে সফল, non-zero মানে কোনো সমস্যা হয়েছে।
	os.Exit(code)
}

// setupApp: test-এর জন্য application-এর প্রয়োজনীয় dependency তৈরি করে।
// যেমন session, logger, repository এবং template renderer।
func setupApp(db *sql.DB) *application {

	// Session management-এর জন্য session object তৈরি করা।
	session := sessions.New(
		[]byte("u46IpCV9y5VlurXXXODJEhgOY8m9JVE4"),
	)

	// Session-এর মেয়াদ ২৪ ঘণ্টা নির্ধারণ করা।
	session.Lifetime = 24 * time.Hour

	// Secure cookie শুধু HTTPS connection-এ পাঠানোর নির্দেশ দেয়।
	session.Secure = true

	// Application-এর প্রয়োজনীয় dependency initialize করা।
	app := &application{
		// io.Discard ব্যবহার করায় test চলাকালে error log print হবে না।
		errorLog: log.New(io.Discard, "", 0),

		// Info log-ও print হবে না।
		infoLog: log.New(io.Discard, "", 0),

		// User-related database operation-এর repository।
		userRepo: NewSQLUserRepository(db),

		// Post-related database operation-এর repository।
		postRepo: NewSQLPostRepository(db),

		// HTML template-এর directory।
		templateDir: "./templates",

		// Static file-এর directory।
		publicPath: "./public",

		// Application-এ session management যুক্ত করা।
		session: session,
	}

	// Application-এর HTML template renderer তৈরি করা।
	// false দিয়ে template development/debug mode বন্ধ রাখা হয়েছে।
	app.tp = NewTemplateRenderer(app.templateDir, false)

	// সম্পূর্ণ প্রস্তুত application ফেরত দেওয়া।
	return app
}

// setupTestSchema: test database-এ প্রয়োজনীয় table তৈরি করে।
// ফলে repository test-গুলো database-এ data insert ও query করতে পারে।
func setupTestSchema(db *sql.DB) error {

	// একাধিক SQL CREATE TABLE statement একসঙ্গে রাখা হয়েছে।
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			hashed_password TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE profile (
			user_id INTEGER PRIMARY KEY REFERENCES users(user_id)
				ON DELETE CASCADE,
			avatar TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			title TEXT NOT NULL UNIQUE,
			user_id INTEGER REFERENCES users(user_id)
				ON DELETE CASCADE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			body TEXT NOT NULL,
			user_id INTEGER REFERENCES users(user_id)
				ON DELETE CASCADE,
			post_id INTEGER REFERENCES posts(post_id)
				ON DELETE CASCADE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE votes (
			user_id INTEGER REFERENCES users(user_id)
				ON DELETE CASCADE,
			post_id INTEGER REFERENCES posts(post_id)
				ON DELETE CASCADE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, post_id)
		);
	`

	// SQL statement execute করে সব table তৈরি করা।
	// কোনো সমস্যা হলে error ফেরত দেওয়া হবে।
	_, err := db.Exec(schema)
	return err
}

// cleanupTestData: প্রতিটি test-এর পরে table-এর data মুছে দেয়।
// এতে একটি test-এর data অন্য test-কে প্রভাবিত করার ঝুঁকি কমে।
func cleanupTestData(t *testing.T) {

	// Foreign key dependency মাথায় রেখে child table আগে পরিষ্কার করা।
	tables := []string{
		"profile",
		"votes",
		"comments",
		"posts",
		"users",
	}

	// প্রতিটি table-এর সব row delete করা।
	for _, table := range tables {
		_, err := testDB.Exec("DELETE FROM " + table)

		// DELETE operation-এ error হলে test-কে fail হিসেবে চিহ্নিত করা।
		assert.NoError(t, err)
	}
}

// মনে রাখার সহজ উপায়:
//
// TestMain:
// সব test-এর আগে setup এবং সব test-এর পরে cleanup/exit পরিচালনা করে।
//
// setupApp:
// Application-এর প্রয়োজনীয় dependency প্রস্তুত করে।
//
// setupTestSchema:
// Test database-এর table তৈরি করে।
//
// cleanupTestData:
// Table-এর test data মুছে দিয়ে পরবর্তী test-এর জন্য database পরিষ্কার করে।
