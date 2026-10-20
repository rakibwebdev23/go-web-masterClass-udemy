
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLogin_GET_NotAuthenticated পরীক্ষা করে,
// login না করা user GET request দিয়ে login page দেখতে পারে কি না।
func TestLogin_GET_NotAuthenticated(t *testing.T) {
	// Test শেষে database-এর test data পরিষ্কার করা হবে।
	defer cleanupTestData(t)

	// Session middleware এবং authentication middleware যুক্ত করা।
	// এরপর login handler execute হবে।
	handler := testApp.session.Enable(
		testApp.authenticate(http.HandlerFunc(testApp.login)),
	)

	// /login URL-এ GET request তৈরি করা।
	req := httptest.NewRequest(http.MethodGet, "/login", nil)

	// HTTP response record করার জন্য ResponseRecorder তৈরি করা।
	w := httptest.NewRecorder()

	// Request-টি handler-এ পাঠিয়ে response তৈরি করা।
	handler.ServeHTTP(w, req)

	// Login page সফলভাবে render হলে HTTP status 200 হবে।
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLogin_GET_AlreadyAuthenticated পরীক্ষা করে,
// আগে থেকেই login করা user আবার login page-এ গেলে
// তাকে home page-এ redirect করা হয় কি না।
func TestLogin_GET_AlreadyAuthenticated(t *testing.T) {
	// Test শেষে database-এর test data পরিষ্কার করা হবে।
	defer cleanupTestData(t)

	// Test database-এ একটি user তৈরি করা।
	_, err := testApp.userRepo.CreateUser(
		"login",
		"login@test.com",
		"goodpassword",
		"avatar",
	)

	// User তৈরি করার সময় কোনো error হয়নি তা নিশ্চিত করা।
	assert.NoError(t, err)

	// একটি temporary handler তৈরি করা,
	// যা session-এ user-এর login information সংরক্ষণ করে।
	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Session-এ user-এর email রাখা হচ্ছে,
		// যাতে পরের request-এ user-কে authenticated হিসেবে চেনা যায়।
		testApp.session.Put(r, loggedInUserKey, "login@test.com")

		// Setup request-এর status code 200 দেওয়া।
		w.WriteHeader(http.StatusOK)
	})

	// Session middleware দিয়ে temporary handler wrap করা।
	setupChain := testApp.session.Enable(setupHandler)

	// Session cookie তৈরি করার জন্য প্রথম request পাঠানো।
	req1 := httptest.NewRequest(http.MethodGet, "/setup", nil)
	w1 := httptest.NewRecorder()
	setupChain.ServeHTTP(w1, req1)

	// প্রকৃত login handler-এর জন্য middleware তৈরি করা।
	handler := testApp.session.Enable(
		testApp.authenticate(http.HandlerFunc(testApp.login)),
	)

	// Login করা user-এর দ্বিতীয় GET request তৈরি করা।
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)

	// প্রথম response-এ session cookie থাকলে
	// সেই cookie দ্বিতীয় request-এ পাঠানো।
	if cookies := w1.Result().Cookies(); len(cookies) > 0 {
		for _, cookie := range cookies {
			req2.AddCookie(cookie)
		}
	}

	// দ্বিতীয় request execute করে response সংগ্রহ করা।
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	// আগে থেকেই authenticated user-কে redirect করা উচিত।
	assert.Equal(t, http.StatusSeeOther, w2.Code)

	// Redirect-এর destination home page (/) কি না পরীক্ষা করা।
	assert.Equal(t, "/", w2.Header().Get("Location"))
}

// TestLogin_POST_ValidCredentials পরীক্ষা করে,
// সঠিক email ও password দিলে login সফল হয় কি না।
func TestLogin_POST_ValidCredentials(t *testing.T) {
	// Test শেষে database-এর test data পরিষ্কার করা হবে।
	defer cleanupTestData(t)

	// Test database-এ login করার জন্য user তৈরি করা।
	_, err := testApp.userRepo.CreateUser(
		"login",
		"login@test.com",
		"goodpassword",
		"avatar",
	)

	// User তৈরি সফল হয়েছে কি না পরীক্ষা করা।
	assert.NoError(t, err)

	// Session ও authentication middleware-সহ login handler তৈরি করা।
	handler := testApp.session.Enable(
		testApp.authenticate(http.HandlerFunc(testApp.login)),
	)

	// POST request-এর form data তৈরি করা।
	formData := "email=login@test.com&password=goodpassword"

	// Form data-সহ POST request তৈরি করা।
	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(formData),
	)

	// Request body URL-encoded form data হিসেবে চিহ্নিত করা।
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Response সংগ্রহ করার জন্য recorder তৈরি করা।
	w := httptest.NewRecorder()

	// Login request execute করা।
	handler.ServeHTTP(w, req)

	// Login সফল হলে redirect status 303 হওয়া উচিত।
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Login সফল হওয়ার পর /submit URL-এ redirect হচ্ছে কি না পরীক্ষা করা।
	assert.Equal(t, "/submit", w.Header().Get("Location"))
}

// TestLogin_POST_InvalidFormData পরীক্ষা করে,
// email ও password ফাঁকা থাকলে validation error দেখায় কি না।
func TestLogin_POST_InvalidFormData(t *testing.T) {
	// Test শেষে database-এর test data পরিষ্কার করা হবে।
	defer cleanupTestData(t)

	// Login handler-এর জন্য middleware তৈরি করা।
	handler := testApp.session.Enable(
		testApp.authenticate(http.HandlerFunc(testApp.login)),
	)

	// Email ও password ফাঁকা রেখে form data তৈরি করা।
	formData := "email=&password="

	// ফাঁকা form data-সহ POST request তৈরি করা।
	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(formData),
	)

	// Request body-এর content type নির্ধারণ করা।
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Response record করার জন্য recorder তৈরি করা।
	w := httptest.NewRecorder()

	// Request execute করা।
	handler.ServeHTTP(w, req)

	// Response body-এর HTML/text সংগ্রহ করা।
	body := w.Body.String()

	// Validation error হলেও login page render হলে status 200 হবে।
	assert.Equal(t, http.StatusOK, w.Code)

	// সাধারণ validation error message আছে কি না পরীক্ষা করা।
	assert.Contains(t, body, "The data you submitted was not valid")

	// Email field-এর required error আছে কি না পরীক্ষা করা।
	assert.Contains(t, body, "This field email is required")

	// Password field-এর required error আছে কি না পরীক্ষা করা।
	assert.Contains(t, body, "This field password is required")
}

// TestLogin_POST_InvalidAuthenticationData পরীক্ষা করে,
// user database-এ না থাকলে login ব্যর্থ হয় কি না।
func TestLogin_POST_InvalidAuthenticationData(t *testing.T) {
	// Test শেষে database-এর test data পরিষ্কার করা হবে।
	defer cleanupTestData(t)

	// Login handler-এর জন্য middleware তৈরি করা।
	handler := testApp.session.Enable(
		testApp.authenticate(http.HandlerFunc(testApp.login)),
	)

	// এমন email দেওয়া হচ্ছে যা database-এ নেই।
	formData := "email=test@test.com&password=goodpassword"

	// Form data-সহ POST request তৈরি করা।
	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(formData),
	)

	// Request-টিকে URL-encoded form হিসেবে চিহ্নিত করা।
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Response সংগ্রহ করার জন্য recorder তৈরি করা।
	w := httptest.NewRecorder()

	// Login request execute করা।
	handler.ServeHTTP(w, req)

	// Response body সংগ্রহ করা।
	body := w.Body.String()

	// Authentication ব্যর্থ হলেও login page render হলে status 200 হবে।
	assert.Equal(t, http.StatusOK, w.Code)

	// Database-এ user না পাওয়ার error message আছে কি না পরীক্ষা করা।
	assert.Contains(t, body, "sql: no rows in result set")
}
