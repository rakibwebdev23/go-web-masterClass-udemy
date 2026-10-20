package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewForm: NewForm() সঠিকভাবে form তৈরি করে কি না পরীক্ষা করে।
func TestNewForm(t *testing.T) {
	// Form-এর input data রাখার জন্য url.Values তৈরি করা হয়েছে।
	values := url.Values{}

	// Form-এর মধ্যে email field ও তার value যোগ করা হয়েছে।
	values.Add("email", "test@test.com")

	// Input values ব্যবহার করে নতুন form তৈরি করা হচ্ছে।
	form := NewForm(values)

	// Form তৈরি হয়েছে এবং nil নয় কি না পরীক্ষা করে।
	assert.NotNil(t, form)

	// Form থেকে email-এর value সঠিকভাবে পাওয়া যাচ্ছে কি না পরীক্ষা করে।
	assert.Equal(t, "test@test.com", form.Get("email"))

	// Form-এর Errors object তৈরি হয়েছে কি না পরীক্ষা করে।
	assert.NotNil(t, form.Errors)

	// নতুন form-এ শুরুতে কোনো validation error নেই কি না পরীক্ষা করে।
	assert.Len(t, form.Errors, 0)
}

// TestForm_Required: Required() method প্রয়োজনীয় field খালি থাকলে
// validation error তৈরি করে কি না পরীক্ষা করে।
func TestForm_Required(t *testing.T) {
	// Form-এর input data রাখার জন্য url.Values তৈরি করা হয়েছে।
	values := url.Values{}

	// Email field-এ একটি valid value দেওয়া হয়েছে।
	values.Add("email", "test@test.com")

	// Empty field-এর ক্ষেত্রে শুধু whitespace দেওয়া হয়েছে।
	values.Add("empty", "   ")

	// Input values দিয়ে নতুন form তৈরি করা হচ্ছে।
	form := NewForm(values)

	// Email, password ও empty field-গুলো required কি না যাচাই করা হচ্ছে।
	form.Required("email", "password", "empty")

	// Form object ঠিক আছে কি না পরীক্ষা করে।
	assert.NotNil(t, form)

	// Email দেওয়া আছে, তাই email-এর validation error খালি কি না পরীক্ষা করে।
	assert.Equal(t, "", form.Errors.Get("email"))

	// Password দেওয়া হয়নি, তাই required validation error এসেছে কি না পরীক্ষা করে।
	assert.Contains(t, form.Errors.Get("password"), "password is required")

	// Empty field-এ শুধু whitespace থাকায় required validation error এসেছে কি না পরীক্ষা করে।
	assert.Contains(t, form.Errors.Get("empty"), "empty is required")
}