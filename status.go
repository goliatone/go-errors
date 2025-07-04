package errors

import "net/http"

const (
	CodeNotFound        = http.StatusNotFound
	CodeConflict        = http.StatusConflict
	CodeBadRequest      = http.StatusBadRequest
	CodeForbidden       = http.StatusForbidden
	CodeInternal        = http.StatusInternalServerError
	CodeUnauthorized    = http.StatusUnauthorized
	CodeRequestTimeout  = http.StatusRequestTimeout
	CodeTooManyRequests = http.StatusTooManyRequests
)
