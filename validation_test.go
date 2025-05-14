package errors_test

import (
	"fmt"
	"testing"

	"github.com/goliatone/go-errors"
)

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errs     errors.ValidationErrors
		expected string
	}{
		{
			name:     "empty validation errors",
			errs:     errors.ValidationErrors{},
			expected: "validation failed",
		},
		{
			name: "single validation error",
			errs: errors.ValidationErrors{
				{Field: "email", Message: "required"},
			},
			expected: "email: required",
		},
		{
			name: "multiple validation errors",
			errs: errors.ValidationErrors{
				{Field: "email", Message: "required"},
				{Field: "age", Message: "must be positive"},
			},
			expected: "email: required; age: must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.errs.Error()
			if got != tt.expected {
				t.Errorf("ValidationErrors.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFieldError_Error(t *testing.T) {
	fe := errors.FieldError{
		Field:   "username",
		Message: "must be at least 3 characters",
		Value:   "ab",
	}

	expected := "username: must be at least 3 characters"
	got := fe.Error()
	if got != expected {
		t.Errorf("FieldError.Error() = %q, want %q", got, expected)
	}
}

func TestNewValidation(t *testing.T) {
	message := "validation failed"
	fieldErrors := []errors.FieldError{
		{Field: "email", Message: "required"},
		{Field: "age", Message: "must be positive", Value: -5},
	}

	err := errors.NewValidation(message, fieldErrors...)

	if err.Category != errors.CategoryValidation {
		t.Errorf("Expected Category=%v, got %v", errors.CategoryValidation, err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected Message=%s, got %s", message, err.Message)
	}
	if len(err.ValidationErrors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(err.ValidationErrors))
	}
	if err.ValidationErrors[0].Field != "email" {
		t.Errorf("Expected first error field=email, got %s", err.ValidationErrors[0].Field)
	}
}

func TestNewValidationFromMap(t *testing.T) {
	message := "validation failed"
	fieldMap := map[string]string{
		"email": "invalid format",
		"name":  "required",
	}

	err := errors.NewValidationFromMap(message, fieldMap)

	if err.Category != errors.CategoryValidation {
		t.Errorf("Expected Category=%v, got %v", errors.CategoryValidation, err.Category)
	}
	if len(err.ValidationErrors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(err.ValidationErrors))
	}

	// Check that both fields are present (order may vary due to map iteration)
	fields := make(map[string]string)
	for _, fieldErr := range err.ValidationErrors {
		fields[fieldErr.Field] = fieldErr.Message
	}
	if fields["email"] != "invalid format" {
		t.Errorf("Expected email error='invalid format', got %s", fields["email"])
	}
	if fields["name"] != "required" {
		t.Errorf("Expected name error='required', got %s", fields["name"])
	}
}

func TestNewValidationFromGroups(t *testing.T) {
	message := "route validation failed"
	groups := map[string][]string{
		"api":   {"missing routes: /health", "missing routes: /metrics"},
		"admin": {"missing group"},
	}

	err := errors.NewValidationFromGroups(message, groups)

	if err.Category != errors.CategoryValidation {
		t.Errorf("Expected Category=%v, got %v", errors.CategoryValidation, err.Category)
	}
	if len(err.ValidationErrors) != 3 {
		t.Errorf("Expected 3 validation errors, got %d", len(err.ValidationErrors))
	}

	// Verify all error messages are present
	errorCount := make(map[string]int)
	for _, fieldErr := range err.ValidationErrors {
		errorCount[fieldErr.Field]++
	}
	if errorCount["api"] != 2 {
		t.Errorf("Expected 2 api errors, got %d", errorCount["api"])
	}
	if errorCount["admin"] != 1 {
		t.Errorf("Expected 1 admin error, got %d", errorCount["admin"])
	}
}

func TestGetValidationErrors(t *testing.T) {
	// Test with validation error
	validationErr := &errors.Error{
		Category: errors.CategoryValidation,
		Message:  "validation failed",
		ValidationErrors: errors.ValidationErrors{
			{Field: "email", Message: "required"},
			{Field: "age", Message: "invalid"},
		},
	}

	errs, found := errors.GetValidationErrors(validationErr)
	if !found {
		t.Error("Expected to find validation errors")
	}
	if len(errs) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(errs))
	}

	// Test with non-validation error
	nonValidationErr := &errors.Error{
		Category: errors.CategoryAuth,
		Message:  "auth failed",
	}

	_, found = errors.GetValidationErrors(nonValidationErr)
	if found {
		t.Error("Expected not to find validation errors in non-validation error")
	}

	// Test with regular error
	regularErr := fmt.Errorf("regular error")
	_, found = errors.GetValidationErrors(regularErr)
	if found {
		t.Error("Expected not to find validation errors in regular error")
	}
}
