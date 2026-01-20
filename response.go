package errors

import (
	"net/http"
	"strings"
)

// ErrorMapper is a function that can map specific error types to our Error type
type ErrorMapper func(error) *Error

// ErrorResponse represents the standard structure for API error responses
type ErrorResponse struct {
	Error *Error `json:"error"`
}

func (e *Error) ToErrorResponse(includeStack bool, stackTrace StackTrace) ErrorResponse {
	response := ErrorResponse{
		Error: e,
	}

	if includeStack {
		response.Error.StackTrace = stackTrace
	} else {
		response.Error.StackTrace = nil
	}

	return response
}

// MapToError converts any error to our Error type using provided mappers
func MapToError(err error, mappers []ErrorMapper) *Error {
	var customErr *Error
	if As(err, &customErr) {
		return customErr
	}

	for _, mapper := range mappers {
		if mappedErr := mapper(err); mappedErr != nil {
			return mappedErr
		}
	}

	customErr = Wrap(err, CategoryInternal, "An unexpected error occurred")
	customErr.Code = 500
	customErr.TextCode = "INTERNAL_ERROR"

	return customErr
}

func DefaultErrorMappers() []ErrorMapper {
	return []ErrorMapper{
		MapOnboardingErrors,
		MapAuthErrors,
		MapHTTPErrors,
	}
}

func MapHTTPErrors(err error) *Error {
	var httpErr interface{ StatusCode() int }
	if As(err, &httpErr) {
		code := httpErr.StatusCode()
		category := HTTPStatusToCategory(code)

		result := New(err.Error(), category).
			WithCode(code).
			WithTextCode(HTTPStatusToTextCode(code))

		return result
	}

	return nil
}

func MapAuthErrors(err error) *Error {
	msg := normalizeErrorMessage(err)
	switch {
	case containsAny(msg, "too many attempts", "too many login attempts"):
		return New(err.Error(), CategoryRateLimit).
			WithCode(http.StatusTooManyRequests).
			WithTextCode(TextCodeTooManyAttempts)
	case containsAny(msg, "token expired", "token is expired"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusUnauthorized).
			WithTextCode(TextCodeTokenExpired)
	case containsAny(msg, "token malformed", "token is malformed", "malformed token", "missing or malformed jwt"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusBadRequest).
			WithTextCode(TextCodeTokenMalformed)
	case containsAny(msg, "account is suspended", "account suspended", "user account is suspended"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeAccountSuspended)
	case containsAny(msg, "account is disabled", "account disabled", "user account is disabled"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeAccountDisabled)
	case containsAny(msg, "account is archived", "account archived", "user account is archived"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeAccountArchived)
	case containsAny(msg, "account is pending", "account pending", "user account is pending"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusForbidden).
			WithTextCode(TextCodeAccountPending)
	case containsAny(msg, "unauthorized", "authentication"):
		return New(err.Error(), CategoryAuth).
			WithCode(http.StatusUnauthorized).
			WithTextCode("UNAUTHORIZED")
	case containsAny(msg, "forbidden", "authorization"):
		return New(err.Error(), CategoryAuthz).
			WithCode(http.StatusForbidden).
			WithTextCode("FORBIDDEN")
	}
	return nil
}

// HTTPStatusToCategory maps HTTP status codes to error categories
func HTTPStatusToCategory(code int) Category {
	switch {
	case code == http.StatusNotFound:
		return CategoryNotFound
	case code == http.StatusUnauthorized:
		return CategoryAuth
	case code == http.StatusForbidden:
		return CategoryAuthz
	case code == http.StatusConflict:
		return CategoryConflict
	case code == http.StatusTooManyRequests:
		return CategoryRateLimit
	case code == http.StatusMethodNotAllowed:
		return CategoryMethodNotAllowed
	case code >= 400 && code < 500:
		return CategoryBadInput
	case code >= 500:
		return CategoryInternal
	default:
		return CategoryInternal
	}
}

// HTTPStatusToTextCode generates text codes from HTTP status codes
func HTTPStatusToTextCode(code int) string {
	return strings.ToUpper(strings.ReplaceAll(http.StatusText(code), " ", "_"))
}
