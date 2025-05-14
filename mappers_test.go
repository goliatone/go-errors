package errors_test

import (
	stdErrors "errors"
	"testing"

	"github.com/goliatone/go-errors"
)

type customError struct {
	code    int
	message string
}

func (e customError) Error() string   { return e.message }
func (e customError) StatusCode() int { return e.code }

func TestErrorMappers(t *testing.T) {

	mappers := []errors.ErrorMapper{
		func(err error) *errors.Error {
			var ce customError
			if stdErrors.As(err, &ce) {
				return errors.New(errors.CategoryExternal, ce.message).
					WithCode(ce.code).
					WithTextCode("CUSTOM_ERROR")
			}
			return nil
		},
	}

	// Test the mapper
	originalErr := customError{code: 502, message: "external service error"}
	mappedErr := errors.MapToError(originalErr, mappers)

	if mappedErr.Category != errors.CategoryExternal {
		t.Errorf("expected category %s, got %s", errors.CategoryExternal, mappedErr.Category)
	}

	if mappedErr.Code != 502 {
		t.Errorf("expected code 502, got %d", mappedErr.Code)
	}

	if mappedErr.TextCode != "CUSTOM_ERROR" {
		t.Errorf("expected text code 'CUSTOM_ERROR', got '%s'", mappedErr.TextCode)
	}
}

func TestErrorChaining(t *testing.T) {
	originalErr := stdErrors.New("original error")
	wrappedErr := errors.Wrap(originalErr, errors.CategoryInternal, "wrapped error")

	if unwrapped := errors.Unwrap(wrappedErr); unwrapped != originalErr {
		t.Errorf("expected unwrapped error to be original error")
	}

	if !errors.Is(wrappedErr, originalErr) {
		t.Error("expected errors.Is to return true for wrapped error")
	}

	var targetErr *errors.Error
	if !errors.As(wrappedErr, &targetErr) {
		t.Error("expected errors.As to successfully unwrap to Error type")
	}

	if targetErr.Category != errors.CategoryInternal {
		t.Errorf("expected category %s, got %s", errors.CategoryInternal, targetErr.Category)
	}
}
