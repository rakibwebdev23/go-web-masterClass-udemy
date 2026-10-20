package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLUserRepository_CreateUser(t *testing.T) {
	defer cleanupTestData(t)

	repo := NewSQLUserRepository(testDB)

	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, userID, int64(0))

	user, err := repo.GetUserByEmail("john@doe.com")
	if !assert.NoError(t, err) || !assert.NotNil(t, user) {
		return
	}
	assert.Equal(t, userID, int64(user.ID))

	userID, err = repo.CreateUser(
		"John Doe",
		"john2@doe.com",
		generateString(73),
		"avatar",
	)
	assert.Error(t, err)
	assert.Equal(t, int64(0), userID)
}

func TestSQLUserRepository_CreateUser_DuplicateEmail(t *testing.T) {
	defer cleanupTestData(t)
	repo := NewSQLUserRepository(testDB)

	userID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, userID, int64(0))

	_, err = repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)
	assert.Error(t, err)
}

func TestSQLUserRepository_Authenticate_Success(t *testing.T) {
	defer cleanupTestData(t)

	repo := NewSQLUserRepository(testDB)

	currentUserID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, currentUserID, int64(0))

	authUser, err := repo.Authenticate("john@doe.com", "testpassword")
	if !assert.NoError(t, err) || !assert.NotNil(t, authUser) {
		return
	}
	assert.Equal(t, currentUserID, int64(authUser.ID))
}

func TestSQLUserRepository_Authenticate_WrongPassword(t *testing.T) {
	defer cleanupTestData(t)

	repo := NewSQLUserRepository(testDB)

	currentUserID, err := repo.CreateUser(
		"John Doe",
		"john@doe.com",
		"testpassword",
		"avatar",
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, currentUserID, int64(0))

	_, err = repo.Authenticate("john@doe.com", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, errInvalidCredentials, err)
}

func generateString(n int) string {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = 'a'
	}
	return string(buf)
}
