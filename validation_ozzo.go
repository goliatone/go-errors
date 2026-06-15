package errors

import (
	"fmt"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func FromOzzoValidation(err error, message string) *Error {
	if err == nil {
		return nil
	}

	var validationErrors validation.Errors
	if As(err, &validationErrors) {
		return fromOzzoValidationErrors(validationErrors, message, captureLocation(1))
	}

	// other types create a general validation error
	return &Error{
		Category:  CategoryValidation,
		Message:   message,
		Source:    err,
		Timestamp: time.Now(),
		Location:  captureLocation(1),
		Severity:  SeverityError,
	}
}

func fromOzzoValidationErrors(validationErrors validation.Errors, message string, location *ErrorLocation) *Error {
	var fieldErrors ValidationErrors

	for field, fieldErr := range validationErrors {
		if nestedErrors, ok := fieldErr.(validation.Errors); ok {
			for nestedField, nestedErr := range nestedErrors {
				fieldName := fmt.Sprintf("%s.%s", field, nestedField)
				fieldErrors = append(fieldErrors, FieldError{
					Field:   fieldName,
					Message: strings.TrimSpace(nestedErr.Error()),
				})
			}
		} else {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   field,
				Message: strings.TrimSpace(fieldErr.Error()),
			})
		}
	}

	return &Error{
		Category:         CategoryValidation,
		Message:          message,
		ValidationErrors: fieldErrors,
		Timestamp:        time.Now(),
		Location:         location,
		Severity:         SeverityError,
	}
}

func ValidateWithOzzo(validateFunc func() error, message string) *Error {
	if err := validateFunc(); err != nil {
		return FromOzzoValidation(err, message)
	}
	return nil
}
