package errors

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"maps"
	"strings"
	"time"
)

// Global behavior
var (
	Verbose       = false
	IsDevelopment = false
)

var (
	As     = goerrors.As
	Is     = goerrors.Is
	Unwrap = goerrors.Unwrap
	Join   = goerrors.Join
)

type Error struct {
	Category         Category         `json:"category"`
	Code             int              `json:"code,omitempty"`
	TextCode         string           `json:"text_code,omitempty"`
	Message          string           `json:"message"`
	Source           error            `json:"-"`
	ValidationErrors ValidationErrors `json:"validation_errors,omitempty"`
	Metadata         map[string]any   `json:"metadata,omitempty"`
	RequestID        string           `json:"request_id,omitempty"`
	Timestamp        time.Time        `json:"timestamp"`
	StackTrace       StackTrace       `json:"stack_trace,omitempty"`
}

func (e *Error) Error() string {
	var parts []string

	if e.TextCode != "" {
		parts = append(parts, fmt.Sprintf("[%s:%s] %s", e.Category, e.TextCode, e.Message))
	} else {
		parts = append(parts, fmt.Sprintf("[%s] %s", e.Category, e.Message))
	}

	if len(e.ValidationErrors) > 0 {
		parts = append(parts, fmt.Sprintf("validation: %s", e.ValidationErrors.Error()))
	}

	if e.Source != nil {
		parts = append(parts, fmt.Sprintf("source: %v", e.Source))
	}

	if len(e.Metadata) > 0 {
		parts = append(parts, fmt.Sprintf("metadata: %d items", len(e.Metadata)))
	}

	return strings.Join(parts, "; ")
}

func (e *Error) ErrorWithStack() string {
	base := e.Error()
	if len(e.StackTrace) > 0 {
		return base + "\n\nStack Trace:\n" + e.StackTrace.String()
	}
	return base
}

func (e *Error) Unwrap() error {
	return e.Source
}

func (e *Error) WithMetadata(metas ...map[string]any) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}

	for _, meta := range metas {
		maps.Copy(e.Metadata, meta)
	}

	return e
}

// TODO: either remove or rename to WithTraceID
func (e *Error) WithRequestID(id string) *Error {
	e.RequestID = id
	return e
}

func (e *Error) WithStackTrace() *Error {
	e.StackTrace = CaptureStackTrace(1)
	return e
}

func (e *Error) WithCode(code int) *Error {
	e.Code = code
	return e
}

func (e *Error) WithTextCode(code string) *Error {
	e.TextCode = code
	return e
}

// ValidationMap returns validation errors as a map
// for easy template usage
func (e *Error) ValiationMap() map[string]string {
	result := make(map[string]string)
	for _, fieldErr := range e.ValidationErrors {
		result[fieldErr.Field] = fieldErr.Message
	}
	return result
}

func (e *Error) MarshalJSON() ([]byte, error) {
	type alias struct {
		Category         Category         `json:"category"`
		Code             int              `json:"code,omitempty"`
		TextCode         string           `json:"text_code,omitempty"`
		Message          string           `json:"message"`
		Source           string           `json:"source,omitempty"`
		ValidationErrors ValidationErrors `json:"validation_errors,omitempty"`
		Metadata         map[string]any   `json:"metadata,omitempty"`
		RequestID        string           `json:"request_id,omitempty"`
		Timestamp        string           `json:"timestamp"`
		StackTrace       StackTrace       `json:"stack_trace,omitempty"`
	}

	aux := alias{
		Category:         e.Category,
		Code:             e.Code,
		TextCode:         e.TextCode,
		Message:          e.Message,
		ValidationErrors: e.ValidationErrors,
		Metadata:         e.Metadata,
		RequestID:        e.RequestID,
		Timestamp:        e.Timestamp.Format(time.RFC3339),
		StackTrace:       e.StackTrace,
	}

	if e.Source != nil {
		aux.Source = e.Source.Error()
	}

	return json.Marshal(aux)
}

// New creates a new Error with the specified category and message
func New(message string, category ...Category) *Error {
	cat := CategoryInternal
	if len(category) > 0 {
		cat = category[0]
	}
	return &Error{
		Category:  cat,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// Wrap creates a new Error that wraps an existing error
func Wrap(source error, category Category, message string) *Error {
	return &Error{
		Category:  category,
		Message:   message,
		Source:    source,
		Timestamp: time.Now(),
	}
}

// IsWrapped checks if an error is already wrapped by our custom error types
func IsWrapped(err error) bool {
	if err == nil {
		return false
	}

	var customErr *Error
	var retryableErr *RetryableError

	return As(err, &customErr) || As(err, &retryableErr)
}
