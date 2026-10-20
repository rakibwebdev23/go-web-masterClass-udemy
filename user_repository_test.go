package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSQLUserRepository_CreateUser:
// User তৈরি, তৈরি হওয়া user খুঁজে পাওয়া এবং invalid password length দিলে error আসে কি না পরীক্ষা করে।
func TestSQLUserRepository_CreateUser(t *testing.T) {
	// Test শেষ হলে database-এর test data মুছে ফেলা হবে।
	defer cleanupTestData(t)

	// Database ব্যবহার করে user operation করার জন্য repository তৈরি করা হয়েছে।
	repo := NewSQLUserRepository(testDB)

	// নতুন user তৈরি করা হচ্ছে।
	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)

	// User তৈরি করতে error হলে test-এর বাকি অংশ চালানো হবে না।
	if !assert.NoError(t, err) {
		return
	}

	// তৈরি হওয়া user-এর ID শূন্যের চেয়ে বড় কি না পরীক্ষা করে।
	assert.Greater(t, userID, int64(0))

	// Email দিয়ে database থেকে user খুঁজে বের করা হচ্ছে।
	user, err := repo.GetUserByEmail("john@doe.com")

	// Error নেই এবং user পাওয়া গেছে কি না পরীক্ষা করে।
	if !assert.NoError(t, err) || !assert.NotNil(t, user) {
		return
	}

	// পাওয়া user-এর ID তৈরি করার সময় পাওয়া ID-এর সমান কি না পরীক্ষা করে।
	assert.Equal(t, userID, int64(user.ID))

	// অতিরিক্ত লম্বা password দিয়ে user তৈরি করার চেষ্টা করা হচ্ছে।
	userID, err = repo.CreateUser(
		"John Doe",
		"john2@doe.com",
		generateString(73),
		"avatar",
	)

	// Invalid password length-এর কারণে error আসছে কি না পরীক্ষা করে।
	assert.Error(t, err)

	// ব্যর্থ operation-এর ক্ষেত্রে user ID শূন্য কি না পরীক্ষা করে।
	assert.Equal(t, int64(0), userID)
}

// TestSQLUserRepository_CreateUser_DuplicateEmail:
// একই email দিয়ে দ্বিতীয়বার user তৈরি করতে গেলে error আসে কি না পরীক্ষা করে।
func TestSQLUserRepository_CreateUser_DuplicateEmail(t *testing.T) {
	// Test শেষে test data মুছে ফেলা হবে।
	defer cleanupTestData(t)

	// User repository তৈরি করা হয়েছে।
	repo := NewSQLUserRepository(testDB)

	// প্রথম user তৈরি করা হচ্ছে।
	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)

	// প্রথম user তৈরি সফল হয়েছে কি না পরীক্ষা করে।
	if !assert.NoError(t, err) {
		return
	}

	// প্রথম user-এর ID বৈধ কি না পরীক্ষা করে।
	assert.Greater(t, userID, int64(0))

	// একই email দিয়ে দ্বিতীয় user তৈরির চেষ্টা করা হচ্ছে।
	_, err = repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)

	// Duplicate email গ্রহণ না করে error এসেছে কি না পরীক্ষা করে।
	assert.Error(t, err)
}

// TestSQLUserRepository_Authenticate_Success:
// সঠিক email ও password দিয়ে user authenticate করা যায় কি না পরীক্ষা করে।
func TestSQLUserRepository_Authenticate_Success(t *testing.T) {
	// Test শেষে database পরিষ্কার করা হবে।
	defer cleanupTestData(t)

	// User repository তৈরি করা হয়েছে।
	repo := NewSQLUserRepository(testDB)

	// Authentication test-এর জন্য user তৈরি করা হচ্ছে।
	currentUserID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)

	// User তৈরি সফল হয়েছে কি না পরীক্ষা করে।
	if !assert.NoError(t, err) {
		return
	}

	// তৈরি হওয়া user-এর ID বৈধ কি না পরীক্ষা করে।
	assert.Greater(t, currentUserID, int64(0))

	// সঠিক email ও password দিয়ে login যাচাই করা হচ্ছে।
	authUser, err := repo.Authenticate("john@doe.com", "testpassword")

	// Authentication-এ error হয়নি এবং user পাওয়া গেছে কি না পরীক্ষা করে।
	if !assert.NoError(t, err) || !assert.NotNil(t, authUser) {
		return
	}

	// Authenticated user-এর ID সঠিক কি না পরীক্ষা করে।
	assert.Equal(t, currentUserID, int64(authUser.ID))
}

// TestSQLUserRepository_Authenticate_WrongPassword:
// ভুল password দিলে authentication ব্যর্থ হয় কি না পরীক্ষা করে।
func TestSQLUserRepository_Authenticate_WrongPassword(t *testing.T) {
	// Test শেষে database পরিষ্কার করা হবে।
	defer cleanupTestData(t)

	// User repository তৈরি করা হয়েছে।
	repo := NewSQLUserRepository(testDB)

	// Authentication test-এর জন্য user তৈরি করা হচ্ছে।
	currentUserID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)

	// User তৈরি সফল হয়েছে কি না পরীক্ষা করে।
	if !assert.NoError(t, err) {
		return
	}

	// User ID বৈধ কি না পরীক্ষা করে।
	assert.Greater(t, currentUserID, int64(0))

	// সঠিক email কিন্তু ভুল password দিয়ে authentication চেষ্টা করা হচ্ছে।
	_, err = repo.Authenticate("john@doe.com", "wrongpassword")

	// Authentication ব্যর্থ হয়ে error এসেছে কি না পরীক্ষা করে।
	assert.Error(t, err)

	// Error-টি expected invalid credentials error কি না পরীক্ষা করে।
	assert.Equal(t, errInvalidCredentials, err)
}

// generateString:
// নির্দিষ্ট দৈর্ঘ্যের একটি string তৈরি করে।
// এখানে প্রতিটি character হিসেবে 'a' ব্যবহার করা হয়েছে।
// যেমন generateString(5) দিলে "aaaaa" পাওয়া যাবে।
func generateString(n int) string {
	// n সংখ্যক byte রাখার জন্য একটি slice তৈরি করা হচ্ছে।
	buf := make([]byte, n)

	// Slice-এর প্রতিটি position-এ 'a' বসানো হচ্ছে।
	for i := 0; i < n; i++ {
		buf[i] = 'a'
	}

	// Byte slice-কে string-এ রূপান্তর করে return করা হচ্ছে।
	return string(buf)
}