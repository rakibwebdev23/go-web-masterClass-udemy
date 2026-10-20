package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLogger: logger middleware HTTP request ও response-এর তথ্য log করে কি না পরীক্ষা করে।
func TestLogger(t *testing.T) {
	// Test করার জন্য একটি dummy handler তৈরি করা হয়েছে, যা 200 status ও OK response দেয়।
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// bytes.Buffer log message memory-তে ধরে রাখে, যাতে পরে log পরীক্ষা করা যায়।
	buf := new(bytes.Buffer)
	testApp.infoLog = log.New(buf, "", 0)

	// Logger middleware-কে test handler-এর চারপাশে যুক্ত করা হয়েছে।
	handler := testApp.logger(testHandler)

	// একটি fake GET request তৈরি করা হয়েছে।
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// ResponseWriter-এর পরিবর্তে recorder ব্যবহার করে response পরীক্ষা করা হয়।
	w := httptest.NewRecorder()

	// Request-টি middleware-এর মধ্য দিয়ে চালানো হয়েছে।
	handler.ServeHTTP(w, req)

	// Response status 200 কি না পরীক্ষা করে।
	assert.Equal(t, http.StatusOK, w.Code)

	// Response body ঠিক "OK" কি না পরীক্ষা করে।
	assert.Equal(t, "OK", w.Body.String())

	// Log-এর মধ্যে HTTP request-এর তথ্য এসেছে কি না পরীক্ষা করে।
	assert.Contains(t, buf.String(), "HTTP/1.1 GET /test")
}

// TestRecover: handler-এর মধ্যে panic হলে recovery middleware server crash না করে
// Internal Server Error response দেয় কি না পরীক্ষা করে।
func TestRecover(t *testing.T) {
	// এমন একটি handler তৈরি করা হয়েছে, যা ইচ্ছাকৃতভাবে panic করবে।
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Recover middleware-কে panic handler-এর সঙ্গে যুক্ত করা হয়েছে।
	handler := testApp.recover(panicHandler)

	// Fake request এবং response recorder তৈরি করা হয়েছে।
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Request চালিয়ে দেখা হচ্ছে panic সামলানো যায় কি না।
	handler.ServeHTTP(w, req)

	// Panic হওয়ার পর HTTP status 500 এসেছে কি না পরীক্ষা করে।
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Error response-এর body সঠিক কি না পরীক্ষা করে।
	assert.Equal(t, "Internal Server Error\n", w.Body.String())

	// Connection header close করা হয়েছে কি না পরীক্ষা করে।
	assert.Equal(t, "close", w.Header().Get("Connection"))
}

// TestRequireAuth_Authenticated: authenticated user protected page access করতে পারে কি না পরীক্ষা করে।
func TestRequireAuth_Authenticated(t *testing.T) {
	// Protected resource-এর জন্য dummy handler তৈরি করা হয়েছে।
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("protected"))
	})

	// Fake GET request তৈরি করা হয়েছে।
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Request context-এ authenticated status true সেট করা হয়েছে।
	req = req.WithContext(contextWithAuth(req.Context(), true))
	w := httptest.NewRecorder()

	// Authentication middleware-কে handler-এর সঙ্গে যুক্ত করা হয়েছে।
	handler := testApp.requireAuth(testHandler)

	// Request middleware-এর মধ্য দিয়ে চালানো হয়েছে।
	handler.ServeHTTP(w, req)

	// Authenticated user সফলভাবে access পেয়েছে কি না পরীক্ষা করে।
	assert.Equal(t, http.StatusOK, w.Code)

	// Protected content পাওয়া গেছে কি না পরীক্ষা করে।
	assert.Equal(t, "protected", w.Body.String())

	// Response-এ no-cache header আছে কি না পরীক্ষা করে।
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
}

// TestRequireAuth_NotAuthenticated: login না করা user protected page access করতে গেলে
// login page-এ redirect হয় কি না পরীক্ষা করে।
func TestRequireAuth_NotAuthenticated(t *testing.T) {
	// Protected resource-এর dummy handler তৈরি করা হয়েছে।
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("protected"))
	})

	// Context-এ authentication দেওয়া হয়নি এমন একটি request তৈরি করা হয়েছে।
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Authentication middleware যুক্ত করা হয়েছে।
	handler := testApp.requireAuth(testHandler)

	// Request চালানো হয়েছে।
	handler.ServeHTTP(w, req)

	// Unauthenticated user-কে redirect করা হয়েছে কি না পরীক্ষা করে।
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Redirect URL সঠিক কি না পরীক্ষা করে।
	assert.Equal(t, "/login?redirectTo=/test", w.Header().Get("Location"))
}

// TestAuthenticate_ValidSession: valid session থেকে user শনাক্ত করে
// request context-এ user information রাখা হয় কি না পরীক্ষা করে।
func TestAuthenticate_ValidSession(t *testing.T) {
	// Test শেষে database-এর test data মুছে ফেলা হবে।
	defer cleanupTestData(t)

	// Test করার জন্য database-এ একজন user তৈরি করা হয়েছে।
	userID, err := testApp.userRepo.CreateUser(
		"session user",
		"session@test.com",
		"passwordgood",
		"avatar",
	)

	// User তৈরিতে error হয়নি কি না পরীক্ষা করে।
	assert.NoError(t, err)

	// User ID শূন্যের চেয়ে বড় কি না পরীক্ষা করে।
	assert.Greater(t, userID, 0)

	// প্রথম handler session-এ login করা user-এর email সংরক্ষণ করে।
	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testApp.session.Put(r, loggedInUserKey, "session@test.com")
		w.WriteHeader(http.StatusOK)
	})

	// দ্বিতীয় handler পরীক্ষা করে authentication middleware user-কে শনাক্ত করেছে কি না।
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Request authenticated কি না পরীক্ষা করে।
		assert.True(t, testApp.isAuthenticated(r))

		// Request context থেকে authenticated user বের করা হয়।
		user := testApp.getUserFromContext(r.Context())

		// Context-এর user ID database-এর user ID-এর সঙ্গে মেলে কি না পরীক্ষা করে।
		assert.Equal(t, userID, user.ID)

		// Context-এর email সঠিক কি না পরীক্ষা করে।
		assert.Equal(t, "session@test.com", user.Email)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Session middleware চালু করে প্রথম request দিয়ে session cookie তৈরি করা হয়।
	setupChain := testApp.session.Enable(setupHandler)
	req1 := httptest.NewRequest(http.MethodGet, "/setup", nil)
	w1 := httptest.NewRecorder()
	setupChain.ServeHTTP(w1, req1)

	// Session middleware ও authentication middleware একসঙ্গে যুক্ত করা হয়েছে।
	testChain := testApp.session.Enable(testApp.authenticate(testHandler))

	// দ্বিতীয় request তৈরি করা হয়েছে।
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)

	// প্রথম response-এর session cookie দ্বিতীয় request-এ যোগ করা হয়েছে।
	if cookies := w1.Result().Cookies(); len(cookies) > 0 {
		for _, cookie := range cookies {
			req2.AddCookie(cookie)
		}
	}

	w2 := httptest.NewRecorder()

	// Session cookie-সহ request চালিয়ে user authentication পরীক্ষা করা হয়।
	testChain.ServeHTTP(w2, req2)

	// Request সফল হয়েছে কি না পরীক্ষা করে।
	assert.Equal(t, http.StatusOK, w2.Code)

	// Response body সঠিক কি না পরীক্ষা করে।
	assert.Equal(t, "OK", w2.Body.String())
}

// contextWithAuth: request context-এ authentication-এর status সংরক্ষণ করে,
// যাতে middleware বা handler সেই status পরীক্ষা করতে পারে।
func contextWithAuth(ctx context.Context, isAuth interface{}) context.Context {
	return context.WithValue(ctx, contextAuthKey, isAuth)
}