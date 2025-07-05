package errors

import "time"

type BaseError = Error

// RetryableError extends Error with retry functionality
type RetryableError struct {
	*BaseError
	retryable bool
	baseDelay time.Duration
}

func (r *RetryableError) Error() string {
	if r.BaseError != nil {
		return r.BaseError.Error()
	}
	return "retryable error: <nil>"
}

// IsRetryable returns whether this error should trigger a retry
func (r *RetryableError) IsRetryable() bool {
	return r.retryable
}

// RetryDelay calculates the delay before the next retry attempt
// Uses exponential backoff: baseDelay * (2^(attempt-1))
func (r *RetryableError) RetryDealy(attempt int) time.Duration {
	if attempt <= 0 {
		return r.baseDelay
	}
	delay := r.baseDelay
	for i := 1; i < attempt && delay < 30*time.Second; i++ {
		delay *= 2
	}

	if delay > 30*time.Second {
		delay = 30 * time.Second
	}

	return delay
}

// WithRetryable sets whether this error should be retryable
func (r *RetryableError) WithRetryable(retryable bool) *RetryableError {
	r.retryable = retryable
	return r
}

// WithRetryDelay sets the base delay for retry attempts
func (r *RetryableError) WithRetryDelay(delay time.Duration) *RetryableError {
	r.baseDelay = delay
	return r
}

func (r *RetryableError) WithMetadata(metas ...map[string]any) *RetryableError {
	r.BaseError.WithMetadata(metas...)
	return r
}

func (r *RetryableError) WithStackTrace() *RetryableError {
	r.BaseError.WithStackTrace()
	return r
}

func (r *RetryableError) WithCode(code int) *RetryableError {
	r.BaseError.WithCode(code)
	return r
}

func (r *RetryableError) WithTextCode(code string) *RetryableError {
	r.BaseError.TextCode = code
	return r
}

func NewRetryable(message string, category Category) *RetryableError {
	return &RetryableError{
		BaseError: New(message, category),
		retryable: true,
		baseDelay: 1 * time.Second,
	}
}

func WrapRetryable(source error, category Category, message string) *RetryableError {
	if source == nil {
		return nil
	}

	return &RetryableError{
		BaseError: Wrap(source, category, message),
		retryable: true,
		baseDelay: 1 * time.Second,
	}
}

// NewNonRetryable creates a non-retryable error (useful for validation, auth, etc.)
func NewNonRetryable(message string, category Category) *RetryableError {
	return NewRetryable(message, category).
		WithRetryable(false)
}

// NewRetryableOperation creates a retryable error for operation failures
// with a short delay of 500 millis by default
func NewRetryableOperation(message string, delay ...time.Duration) *RetryableError {
	var del time.Duration = 500 * time.Millisecond
	if len(delay) > 0 {
		del = delay[0]
	}

	return NewRetryable(message, CategoryOperation).
		WithRetryDelay(del)
}

// NewRetryableExternal creates a retryable error for an external service
func NewRetryableExternal(message string) *RetryableError {
	return NewRetryable(message, CategoryExternal).
		WithRetryDelay(2 * time.Second).
		WithCode(502).
		WithTextCode("EXTERNAL_SERVICE_ERROR")
}

// IsRetryableError checks if an error implements the IsRetryable interface
// and returns true
func IsRetryableError(err error) bool {
	if retryable, ok := err.(interface{ IsRetryable() bool }); ok {
		return retryable.IsRetryable()
	}
	return false
}
