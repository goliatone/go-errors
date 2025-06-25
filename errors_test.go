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

	err = err.WithMetadata(meta1)
	if err.Metadata["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", err.Metadata["key1"])
	}

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

	err := errors.New(message, category)

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

	errNoStack := &errors.Error{
		Category: errors.CategoryInternal,
		Message:  "internal error",
	}

	resultNoStack := errNoStack.ErrorWithStack()
	if resultNoStack != "[internal] internal error" {
		t.Errorf("Expected simple error message, got %q", resultNoStack)
	}
}

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

func TestError_ValidationMap(t *testing.T) {
	tests := []struct {
		name     string
		err      *errors.Error
		expected map[string]string
	}{
		{
			name: "empty validation errors",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				Message:  "validation failed",
			},
			expected: map[string]string{},
		},
		{
			name: "single validation error",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				Message:  "validation failed",
				ValidationErrors: errors.ValidationErrors{
					{Field: "email", Message: "invalid format"},
				},
			},
			expected: map[string]string{
				"email": "invalid format",
			},
		},
		{
			name: "multiple validation errors",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				Message:  "validation failed",
				ValidationErrors: errors.ValidationErrors{
					{Field: "email", Message: "invalid format"},
					{Field: "name", Message: "required"},
					{Field: "age", Message: "must be positive"},
				},
			},
			expected: map[string]string{
				"email": "invalid format",
				"name":  "required",
				"age":   "must be positive",
			},
		},
		{
			name: "duplicate field names (last one wins)",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				Message:  "validation failed",
				ValidationErrors: errors.ValidationErrors{
					{Field: "email", Message: "invalid format"},
					{Field: "email", Message: "required"},
				},
			},
			expected: map[string]string{
				"email": "required",
			},
		},
		{
			name: "validation errors with empty field or message",
			err: &errors.Error{
				Category: errors.CategoryValidation,
				Message:  "validation failed",
				ValidationErrors: errors.ValidationErrors{
					{Field: "", Message: "empty field"},
					{Field: "name", Message: ""},
					{Field: "email", Message: "valid message"},
				},
			},
			expected: map[string]string{
				"":      "empty field",
				"name":  "",
				"email": "valid message",
			},
		},
		{
			name: "nil validation errors slice",
			err: &errors.Error{
				Category:         errors.CategoryValidation,
				Message:          "validation failed",
				ValidationErrors: nil,
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.ValidationMap()

			if len(got) != len(tt.expected) {
				t.Errorf("ValidationMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for expectedKey, expectedValue := range tt.expected {
				if gotValue, exists := got[expectedKey]; !exists {
					t.Errorf("ValidationMap() missing key %q", expectedKey)
				} else if gotValue != expectedValue {
					t.Errorf("ValidationMap()[%q] = %q, want %q", expectedKey, gotValue, expectedValue)
				}
			}

			for gotKey := range got {
				if _, exists := tt.expected[gotKey]; !exists {
					t.Errorf("ValidationMap() has unexpected key %q with value %q", gotKey, got[gotKey])
				}
			}
		})
	}
}

func TestError_ValidationMap_Immutable(t *testing.T) {
	err := &errors.Error{
		Category: errors.CategoryValidation,
		Message:  "validation failed",
		ValidationErrors: errors.ValidationErrors{
			{Field: "email", Message: "invalid format"},
			{Field: "name", Message: "required"},
		},
	}

	validationMap1 := err.ValidationMap()

	validationMap1["new_field"] = "new_message"
	delete(validationMap1, "email")

	validationMap2 := err.ValidationMap()

	expected := map[string]string{
		"email": "invalid format",
		"name":  "required",
	}

	if len(validationMap2) != len(expected) {
		t.Errorf("Second ValidationMap() call returned map with %d entries, want %d", len(validationMap2), len(expected))
	}

	for expectedKey, expectedValue := range expected {
		if gotValue, exists := validationMap2[expectedKey]; !exists {
			t.Errorf("Second ValidationMap() call missing key %q", expectedKey)
		} else if gotValue != expectedValue {
			t.Errorf("Second ValidationMap() call [%q] = %q, want %q", expectedKey, gotValue, expectedValue)
		}
	}

	if _, exists := validationMap2["new_field"]; exists {
		t.Error("Second ValidationMap() call should not contain 'new_field' added to first map")
	}
}

func BenchmarkError_ValidationMap(b *testing.B) {
	err := &errors.Error{
		Category: errors.CategoryValidation,
		Message:  "validation failed",
		ValidationErrors: errors.ValidationErrors{
			{Field: "email", Message: "invalid format"},
			{Field: "name", Message: "required"},
			{Field: "age", Message: "must be positive"},
			{Field: "phone", Message: "invalid format"},
			{Field: "address", Message: "too long"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.ValidationMap()
	}
}
