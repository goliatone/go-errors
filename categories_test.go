package errors_test

import (
	"fmt"
	"testing"

	"github.com/goliatone/go-errors"
)

func TestIsCategory(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryNotFound,
		Message:  "not found",
	}

	if !errors.IsCategory(err, errors.CategoryNotFound) {
		t.Error("Expected IsCategory to return true for matching category")
	}
	if errors.IsCategory(err, errors.CategoryValidation) {
		t.Error("Expected IsCategory to return false for non-matching category")
	}

	// Test with regular error
	regularErr := fmt.Errorf("regular error")
	if errors.IsCategory(regularErr, errors.CategoryNotFound) {
		t.Error("Expected IsCategory to return false for regular error")
	}
}

func TestCategoryCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{
			name:     "validation error",
			err:      &errors.Error{Category: errors.CategoryValidation, Message: "test"},
			checker:  errors.IsValidation,
			expected: true,
		},
		{
			name:     "auth error",
			err:      &errors.Error{Category: errors.CategoryAuth, Message: "test"},
			checker:  errors.IsAuth,
			expected: true,
		},
		{
			name:     "not found error",
			err:      &errors.Error{Category: errors.CategoryNotFound, Message: "test"},
			checker:  errors.IsNotFound,
			expected: true,
		},
		{
			name:     "internal error",
			err:      &errors.Error{Category: errors.CategoryInternal, Message: "test"},
			checker:  errors.IsInternal,
			expected: true,
		},
		{
			name:     "wrong category",
			err:      &errors.Error{Category: errors.CategoryAuth, Message: "test"},
			checker:  errors.IsValidation,
			expected: false,
		},
		{
			name:     "regular error",
			err:      fmt.Errorf("regular error"),
			checker:  errors.IsValidation,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checker(tt.err)
			if got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}
