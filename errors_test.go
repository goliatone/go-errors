package errors_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/goliatone/go-errors"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *errors.Error
		expected string
	}{
		{
			name: "basic error",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				Message:  "invalid input",
			},
			expected: "[validation] invalid input",
		},
		{
			name: "error with text code",
			err: &errors.Error{
				Category: errors.CategoryAuth,
				TextCode: "TOKEN_EXPIRED",
				Message:  "token has expired",
			},
			expected: "[authentication:TOKEN_EXPIRED] token has expired",
		},
		{
			name: "error with source",
			err: &errors.Error{
				Category: errors.CategoryInternal,
				Message:  "database error",
				Source:   fmt.Errorf("connection failed"),
			},
			expected: "[internal] database error; source: connection failed",
		},
		{
			name: "error with validation errors",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				Message:  "validation failed",
				ValidationErrors: errors.ValidationErrors{
					{Field: "email", Message: "invalid format"},
					{Field: "age", Message: "must be positive"},
				},
			},
			expected: "[validation] validation failed; validation: email: invalid format; age: must be positive",
		},
		{
			name: "error with metadata",
			err: &errors.Error{
				Category: errors.CategoryNotFound,
				Message:  "user not found",
				Metadata: map[string]any{
					"user_id": 123,
					"table":   "users",
				},
			},
			expected: "[not_found] user not found; metadata: 2 items",
		},
		{
			name: "complex error",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				TextCode: "VALIDATION_ERROR",
				Message:  "multiple validation failures",
				Source:   fmt.Errorf("upstream validation failed"),
				ValidationErrors: errors.ValidationErrors{
					{Field: "name", Message: "required"},
				},
				Metadata: map[string]any{
					"request_id": "req-123",
				},
			},
			expected: "[validation:VALIDATION_ERROR] multiple validation failures; validation: name: required; source: upstream validation failed; metadata: 1 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	sourceErr := fmt.Errorf("source error")
	err := &errors.Error{
		Category: errors.CategoryInternal,
		Message:  "wrapped error",
		Source:   sourceErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != sourceErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, sourceErr)
	}

	// Test with no source error
	errNoSource := &errors.Error{
		Category: errors.CategoryNotFound,
		Message:  "not found",
	}
	if errNoSource.Unwrap() != nil {
		t.Errorf("Unwrap() on error with no source should return nil")
	}
}

func TestError_WithMetadata(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryInternal,
		Message:  "test error",
	}

	meta1 := map[string]any{"key1": "value1"}
	meta2 := map[string]any{"key2": "value2", "key1": "overwritten"}

	// Test first metadata addition
	err = err.WithMetadata(meta1)
	if err.Metadata["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", err.Metadata["key1"])
	}

	// Test metadata merging and overwriting
	err = err.WithMetadata(meta2)
	if err.Metadata["key1"] != "overwritten" {
		t.Errorf("Expected key1=overwritten, got %v", err.Metadata["key1"])
	}
	if err.Metadata["key2"] != "value2" {
		t.Errorf("Expected key2=value2, got %v", err.Metadata["key2"])
	}
}

func TestError_WithRequestID(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryNotFound,
		Message:  "test error",
	}

	requestID := "req-12345"
	err = err.WithRequestID(requestID)

	if err.RequestID != requestID {
		t.Errorf("Expected RequestID=%s, got %s", requestID, err.RequestID)
	}
}

func TestError_WithStackTrace(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryInternal,
		Message:  "test error",
	}

	err = err.WithStackTrace()

	if len(err.StackTrace) == 0 {
		t.Error("Expected stack trace to be captured")
	}

	// Verify that the stack trace contains this test function
	found := false
	for _, frame := range err.StackTrace {
		if strings.Contains(frame.Function, "TestError_WithStackTrace") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected stack trace to contain test function name")
	}
}

func TestError_WithCode(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryValidation,
		Message:  "test error",
	}

	code := 400
	err = err.WithCode(code)

	if err.Code != code {
		t.Errorf("Expected Code=%d, got %d", code, err.Code)
	}
}

func TestError_WithTextCode(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryAuth,
		Message:  "test error",
	}

	textCode := "AUTH_FAILED"
	err = err.WithTextCode(textCode)

	if err.TextCode != textCode {
		t.Errorf("Expected TextCode=%s, got %s", textCode, err.TextCode)
	}
}

func TestError_MarshalJSON(t *testing.T) {
	err := &errors.Error{
		Category:  errors.CategoryValidation,
		Code:      400,
		TextCode:  "VALIDATION_ERROR",
		Message:   "validation failed",
		Source:    fmt.Errorf("source error"),
		RequestID: "req-123",
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		ValidationErrors: errors.ValidationErrors{
			{Field: "email", Message: "required"},
		},
		Metadata: map[string]any{
			"key": "value",
		},
	}

	data, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal error: %v", marshalErr)
	}

	var unmarshaled map[string]interface{}
	if unmarshalErr := json.Unmarshal(data, &unmarshaled); unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal error: %v", unmarshalErr)
	}

	// Verify key fields
	if unmarshaled["category"] != "validation" {
		t.Errorf("Expected category=validation, got %v", unmarshaled["category"])
	}
	if unmarshaled["message"] != "validation failed" {
		t.Errorf("Expected message=validation failed, got %v", unmarshaled["message"])
	}
	if unmarshaled["source"] != "source error" {
		t.Errorf("Expected source=source error, got %v", unmarshaled["source"])
	}
	if unmarshaled["request_id"] != "req-123" {
		t.Errorf("Expected request_id=req-123, got %v", unmarshaled["request_id"])
	}
}

func TestNew(t *testing.T) {
	category := errors.CategoryNotFound
	message := "resource not found"

	err := errors.New(category, message)

	if err.Category != category {
		t.Errorf("Expected Category=%v, got %v", category, err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected Message=%s, got %s", message, err.Message)
	}
	if err.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

func TestWrap(t *testing.T) {
	sourceErr := fmt.Errorf("original error")
	category := errors.CategoryInternal
	message := "wrapped error"

	err := errors.Wrap(sourceErr, category, message)

	if err.Category != category {
		t.Errorf("Expected Category=%v, got %v", category, err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected Message=%s, got %s", message, err.Message)
	}
	if err.Source != sourceErr {
		t.Errorf("Expected Source=%v, got %v", sourceErr, err.Source)
	}
	if err.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

func TestStackTrace_String(t *testing.T) {
	st := errors.StackTrace{
		{Function: "main.main", File: "/app/main.go", Line: 10},
		{Function: "github.com/example/pkg.Function", File: "/app/pkg/file.go", Line: 25},
	}

	result := st.String()
	expected := "main.main\n\t/app/main.go:10\ngithub.com/example/pkg.Function\n\t/app/pkg/file.go:25"

	if result != expected {
		t.Errorf("StackTrace.String() = %q, want %q", result, expected)
	}
}

func TestError_ErrorWithStack(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryInternal,
		Message:  "internal error",
		StackTrace: errors.StackTrace{
			{Function: "main.main", File: "/app/main.go", Line: 10},
		},
	}

	result := err.ErrorWithStack()
	if !strings.Contains(result, "[internal] internal error") {
		t.Error("Expected error message in ErrorWithStack output")
	}
	if !strings.Contains(result, "Stack Trace:") {
		t.Error("Expected stack trace header in ErrorWithStack output")
	}
	if !strings.Contains(result, "main.main") {
		t.Error("Expected stack trace content in ErrorWithStack output")
	}

	// Test without stack trace
	errNoStack := &errors.Error{
		Category: errors.CategoryInternal,
		Message:  "internal error",
	}

	resultNoStack := errNoStack.ErrorWithStack()
	if resultNoStack != "[internal] internal error" {
		t.Errorf("Expected simple error message, got %q", resultNoStack)
	}
}

// Benchmark tests
func BenchmarkError_Error(b *testing.B) {
	err := &errors.Error{
		Category: errors.CategoryValidation,
		TextCode: "VALIDATION_ERROR",
		Message:  "validation failed",
		ValidationErrors: errors.ValidationErrors{
			{Field: "email", Message: "required"},
			{Field: "name", Message: "too short"},
		},
		Source: fmt.Errorf("source error"),
		Metadata: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkNewValidation(b *testing.B) {
	fieldErrors := []errors.FieldError{
		{Field: "email", Message: "required"},
		{Field: "name", Message: "too short"},
		{Field: "age", Message: "must be positive"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.NewValidation("validation failed", fieldErrors...)
	}
}

func BenchmarkError_MarshalJSON(b *testing.B) {
	err := &errors.Error{
		Category: errors.CategoryValidation,
		Code:     400,
		TextCode: "VALIDATION_ERROR",
		Message:  "validation failed",
		ValidationErrors: errors.ValidationErrors{
			{Field: "email", Message: "required"},
		},
		Metadata: map[string]any{
			"request_id": "req-123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(err)
	}
}
