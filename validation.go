package errors

import (
	"fmt"
	"strings"
	"time"
)

// FieldError reprents a single validation error for a given field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of field level validation errors
type ValidationErrors []FieldError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}

	parts := make([]string, len(e))
	for i, err := range e {
		parts[i] = err.Error()
	}
	return strings.Join(parts, "; ")
}

func NewValidation(message string, fieldErrors ...FieldError) *Error {
	return &Error{
		Category:         CategoryValidation,
		Message:          message,
		ValidationErrors: fieldErrors,
		Timestamp:        time.Now(),
	}
}

func NewValidationFromMap(message string, fieldMap map[string]string) *Error {
	var fieldErrors ValidationErrors
	for field, msg := range fieldMap {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   field,
			Message: msg,
		})
	}
	return &Error{
		Category:         CategoryValidation,
		Message:          message,
		ValidationErrors: fieldErrors,
		Timestamp:        time.Now(),
	}
}

func NewValidationFromGroups(message string, groups map[string][]string) *Error {
	var fieldErrors ValidationErrors
	for group, messages := range groups {
		for _, msg := range messages {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   group,
				Message: msg,
			})
		}
	}
	return &Error{
		Category:         CategoryValidation,
		Message:          message,
		ValidationErrors: fieldErrors,
		Timestamp:        time.Now(),
	}
}

func GetValidationErrors(err error) (ValidationErrors, bool) {
	var allErrors ValidationErrors
	found := false
	currentErr := err
	for currentErr != nil {
		var e *Error
		if As(currentErr, &e) && len(e.ValidationErrors) > 0 {
			allErrors = append(allErrors, e.ValidationErrors...)
			found = true
		}
		currentErr = Unwrap(currentErr)
	}
	return allErrors, found
}
