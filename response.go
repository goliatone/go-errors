package errors

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

	return Wrap(err, CategoryInternal, "An unexpected error occurred")
}
