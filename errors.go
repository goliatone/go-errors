package errors

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"maps"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	Location         *ErrorLocation   `json:"location,omitempty"`
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

	if Verbose && e.Location != nil {
		parts = append(parts, fmt.Sprintf("location: %s", e.Location.String()))
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

// WithLocation sets the location where the error occurred
func (e *Error) WithLocation(loc *ErrorLocation) *Error {
	e.Location = loc
	return e
}

// GetLocation returns the location where the error occurred
func (e *Error) GetLocation() *ErrorLocation {
	return e.Location
}

// HasLocation returns true if the error has location information
func (e *Error) HasLocation() bool {
	return e.Location != nil
}

// ValidationMap returns validation errors as a map
// for easy template usage
func (e *Error) ValidationMap() map[string]string {
	return e.allValidationMapWithPath("")
}

func (e *Error) AllValidationErrors() ValidationErrors {
	var allErrors ValidationErrors
	allErrors = append(allErrors, e.ValidationErrors...)
	if len(e.ValidationErrors) == 0 && e.Source != nil {
		if validationErrors, ok := e.Source.(validation.Errors); ok {
			for field, fieldErr := range validationErrors {
				allErrors = append(allErrors, FieldError{
					Field:   field,
					Message: strings.TrimSpace(fieldErr.Error()),
				})
			}
		}
	}

	if e.Source != nil {
		if soureErr, ok := e.Source.(*Error); ok {
			allErrors = append(allErrors, soureErr.AllValidationErrors()...)
		}
	}

	return allErrors
}

func (e *Error) allValidationMapWithPath(prefix string) map[string]string {
	result := make(map[string]string)

	for _, fieldErr := range e.ValidationErrors {
		key := fieldErr.Field
		if prefix != "" {
			key = prefix + "." + fieldErr.Field
		}
		result[key] = fieldErr.Message
	}

	if len(e.ValidationErrors) == 0 && e.Source != nil {
		if validationErrors, ok := e.Source.(validation.Errors); ok {
			for field, fieldErr := range validationErrors {
				key := field
				if prefix != "" {
					key = prefix + "." + field
				}
				result[key] = strings.TrimSpace(fieldErr.Error())
			}
		}
	}

	if e.Source != nil {
		if sourceErr, ok := e.Source.(*Error); ok {
			newPrefix := prefix
			if newPrefix == "" {
				newPrefix = "source"
			} else {
				newPrefix = prefix + ".source"
			}

			for k, v := range sourceErr.allValidationMapWithPath(newPrefix) {
				result[k] = v
			}
		}
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
		Location         *ErrorLocation   `json:"location,omitempty"`
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
		Location:         e.Location,
	}

	if e.Source != nil {
		aux.Source = e.Source.Error()
	}

	return json.Marshal(aux)
}

func (e *Error) Clone() *Error {
	if e == nil {
		return nil
	}

	clone := *e // shallow copy

	if e.ValidationErrors != nil {
		clone.ValidationErrors = make(ValidationErrors, len(e.ValidationErrors))
		copy(clone.ValidationErrors, e.ValidationErrors)
	}

	if e.Metadata != nil {
		clone.Metadata = make(map[string]any, len(e.Metadata))
		maps.Copy(clone.Metadata, e.Metadata)
	}

	return &clone
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
		Location:  captureLocation(1), // Capture caller's location
	}
}

// Wrap creates a new Error that wraps an existing error
func Wrap(source error, category Category, message string) *Error {
	if source == nil {
		return nil
	}

	var e *Error
	if As(source, &e) {
		nerr := e.Clone()
		nerr.Message = fmt.Sprintf("%s: %s", message, e.Message)
		// Keep original location when wrapping existing Error
		return nerr
	}

	return &Error{
		Category:  category,
		Message:   message,
		Source:    source,
		Timestamp: time.Now(),
		Location:  captureLocation(1), // Capture new location for non-Error sources
	}
}

// NewWithLocation creates a new Error with explicit location setting
func NewWithLocation(message string, category Category, location *ErrorLocation) *Error {
	return &Error{
		Category:  category,
		Message:   message,
		Timestamp: time.Now(),
		Location:  location,
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

func RootCause(err error) error {
	for {
		unwrapped := Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

func RootCategory(err error) Category {
	if rootErr, ok := RootCause(err).(*Error); ok {
		return rootErr.Category
	}
	return CategoryInternal
}
