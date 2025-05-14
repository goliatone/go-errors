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
		MapHTTPErrors,
		MapAuthErrors,
	}
}

func MapHTTPErrors(err error) *Error {
	var httpErr interface{ StatusCode() int }
	if As(err, &httpErr) {
		code := httpErr.StatusCode()
		category := HTTPStatusToCategory(code)

		result := New(category, err.Error()).
			WithCode(code).
			WithTextCode(HTTPStatusToTextCode(code))

		return result
	}

	return nil
}

func MapAuthErrors(err error) *Error {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "authentication"):
		return New(CategoryAuth, err.Error()).
			WithCode(http.StatusUnauthorized).
			WithTextCode("UNAUTHORIZED")
	case strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "authorization"):
		return New(CategoryAuthz, err.Error()).
			WithCode(http.StatusForbidden).
			WithTextCode("FORBIDDEN")
	case strings.Contains(errMsg, "token expired"):
		return New(CategoryAuth, err.Error()).
			WithCode(http.StatusUnauthorized).
			WithTextCode("TOKEN_EXPIRED")
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
