package errors

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	tests := []struct {
		name string
		opts []CollectorOption
		want func(*ErrorCollector) bool
	}{
		{
			name: "default collector",
			opts: nil,
			want: func(c *ErrorCollector) bool {
				return c.maxErrors == 100 && !c.strictMode && c.context != nil
			},
		},
		{
			name: "with max errors",
			opts: []CollectorOption{WithMaxErrors(50)},
			want: func(c *ErrorCollector) bool {
				return c.maxErrors == 50
			},
		},
		{
			name: "with strict mode",
			opts: []CollectorOption{WithStrictMode(true)},
			want: func(c *ErrorCollector) bool {
				return c.strictMode
			},
		},
		{
			name: "with context",
			opts: []CollectorOption{WithContext(context.Background())},
			want: func(c *ErrorCollector) bool {
				return c.context == context.Background()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector(tt.opts...)
			if !tt.want(c) {
				t.Errorf("NewCollector() failed validation for %s", tt.name)
			}
		})
	}
}

func TestErrorCollector_Add(t *testing.T) {
	tests := []struct {
		name       string
		collector  *ErrorCollector
		errors     []error
		wantCounts []int
		wantResult []bool
	}{
		{
			name:       "add nil error",
			collector:  NewCollector(),
			errors:     []error{nil},
			wantCounts: []int{0},
			wantResult: []bool{true},
		},
		{
			name:       "add single error",
			collector:  NewCollector(),
			errors:     []error{New("test error")},
			wantCounts: []int{1},
			wantResult: []bool{true},
		},
		{
			name:       "add multiple errors",
			collector:  NewCollector(),
			errors:     []error{New("error 1"), New("error 2"), fmt.Errorf("std error")},
			wantCounts: []int{1, 2, 3},
			wantResult: []bool{true, true, true},
		},
		{
			name:       "strict mode full collector",
			collector:  NewCollector(WithMaxErrors(2), WithStrictMode(true)),
			errors:     []error{New("error 1"), New("error 2"), New("error 3")},
			wantCounts: []int{1, 2, 2},
			wantResult: []bool{true, true, false},
		},
		{
			name:       "non-strict mode full collector",
			collector:  NewCollector(WithMaxErrors(2), WithStrictMode(false)),
			errors:     []error{New("error 1"), New("error 2"), New("error 3")},
			wantCounts: []int{1, 2, 2},
			wantResult: []bool{true, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, err := range tt.errors {
				result := tt.collector.Add(err)
				if result != tt.wantResult[i] {
					t.Errorf("Add() result = %v, want %v for error %d", result, tt.wantResult[i], i)
				}
				if tt.collector.Count() != tt.wantCounts[i] {
					t.Errorf("Count() = %v, want %v after adding error %d", tt.collector.Count(), tt.wantCounts[i], i)
				}
			}
		})
	}
}

func TestErrorCollector_BasicOperations(t *testing.T) {
	c := NewCollector()

	// Test empty collector
	if c.HasErrors() {
		t.Error("HasErrors() should return false for empty collector")
	}
	if c.Count() != 0 {
		t.Error("Count() should return 0 for empty collector")
	}
	if len(c.Errors()) != 0 {
		t.Error("Errors() should return empty slice for empty collector")
	}

	// Add some errors
	err1 := New("error 1", CategoryValidation)
	err2 := NewWarning("warning", CategoryExternal)

	c.Add(err1)
	c.Add(err2)

	// Test with errors
	if !c.HasErrors() {
		t.Error("HasErrors() should return true after adding errors")
	}
	if c.Count() != 2 {
		t.Errorf("Count() = %d, want 2", c.Count())
	}

	errors := c.Errors()
	if len(errors) != 2 {
		t.Errorf("Errors() length = %d, want 2", len(errors))
	}

	// Test that returned slice is a copy
	errors[0] = nil
	if c.Errors()[0] == nil {
		t.Error("Errors() should return a copy, not the original slice")
	}

	// Test reset
	c.Reset()
	if c.HasErrors() {
		t.Error("HasErrors() should return false after reset")
	}
	if c.Count() != 0 {
		t.Error("Count() should return 0 after reset")
	}
}

func TestErrorCollector_Merge(t *testing.T) {
	tests := []struct {
		name   string
		errors []*Error
		want   func(*Error) bool
	}{
		{
			name:   "no errors",
			errors: nil,
			want: func(e *Error) bool {
				return e == nil
			},
		},
		{
			name:   "single error",
			errors: []*Error{New("single error")},
			want: func(e *Error) bool {
				return e != nil && e.Message == "single error"
			},
		},
		{
			name: "multiple errors",
			errors: []*Error{
				New("error 1", CategoryValidation).WithSeverity(SeverityError),
				NewWarning("warning", CategoryExternal),
				NewCritical("critical", CategoryInternal),
			},
			want: func(e *Error) bool {
				return e != nil &&
					e.Message == "Multiple errors occurred" &&
					e.Severity == SeverityCritical &&
					e.Metadata != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			for _, err := range tt.errors {
				c.Add(err)
			}

			merged := c.Merge()
			if !tt.want(merged) {
				t.Errorf("Merge() failed validation for %s", tt.name)
			}

			if merged != nil && len(tt.errors) > 1 {
				// Check metadata for multiple errors
				if merged.Metadata["error_count"] != len(tt.errors) {
					t.Errorf("Metadata error_count = %v, want %d", merged.Metadata["error_count"], len(tt.errors))
				}
			}
		})
	}
}

func TestErrorCollector_Filtering(t *testing.T) {
	c := NewCollector()

	// Add errors with different severities and categories
	c.Add(NewDebug("debug", CategoryInternal))
	c.Add(NewWarning("warning", CategoryValidation))
	c.Add(NewCritical("critical", CategoryExternal))
	c.Add(New("error", CategoryValidation))

	// Test FilterBySeverity
	criticalAndAbove := c.FilterBySeverity(SeverityCritical)
	if len(criticalAndAbove) != 1 {
		t.Errorf("FilterBySeverity(Critical) = %d errors, want 1", len(criticalAndAbove))
	}

	warningAndAbove := c.FilterBySeverity(SeverityWarning)
	if len(warningAndAbove) != 3 {
		t.Errorf("FilterBySeverity(Warning) = %d errors, want 3", len(warningAndAbove))
	}

	// Test FilterByCategory
	validationErrors := c.FilterByCategory(CategoryValidation)
	if len(validationErrors) != 2 {
		t.Errorf("FilterByCategory(Validation) = %d errors, want 2", len(validationErrors))
	}

	internalErrors := c.FilterByCategory(CategoryInternal)
	if len(internalErrors) != 1 {
		t.Errorf("FilterByCategory(Internal) = %d errors, want 1", len(internalErrors))
	}
}

func TestErrorCollector_CategoryAnalysis(t *testing.T) {
	c := NewCollector()

	// Add errors with different categories
	c.Add(New("error 1", CategoryValidation))
	c.Add(New("error 2", CategoryValidation))
	c.Add(New("error 3", CategoryExternal))
	c.Add(New("error 4", CategoryInternal))

	// Test CategoryStats
	stats := c.CategoryStats()
	if stats[CategoryValidation] != 2 {
		t.Errorf("CategoryStats[Validation] = %d, want 2", stats[CategoryValidation])
	}
	if stats[CategoryExternal] != 1 {
		t.Errorf("CategoryStats[External] = %d, want 1", stats[CategoryExternal])
	}

	// Test MostCommonCategory
	mostCommon := c.MostCommonCategory()
	if mostCommon != CategoryValidation {
		t.Errorf("MostCommonCategory() = %v, want %v", mostCommon, CategoryValidation)
	}

	// Test SeverityDistribution
	severityDist := c.SeverityDistribution()
	if severityDist[SeverityError] != 4 { // All errors default to Error severity
		t.Errorf("SeverityDistribution[Error] = %d, want 4", severityDist[SeverityError])
	}
}

func TestErrorCollector_ValidationErrors(t *testing.T) {
	c := NewCollector()

	// Add validation error using helper method
	c.AddValidation("email", "invalid email format")
	c.AddFieldErrors(
		FieldError{Field: "name", Message: "required"},
		FieldError{Field: "age", Message: "must be positive"},
	)

	// Add regular error with validation errors
	validationErr := NewValidation("validation failed",
		FieldError{Field: "phone", Message: "invalid format"})
	c.Add(validationErr)

	// Test GetValidationErrors
	validationErrors := c.GetValidationErrors()
	if len(validationErrors) != 4 {
		t.Errorf("GetValidationErrors() = %d errors, want 4", len(validationErrors))
	}

	// Test GetAllValidationErrors
	allValidationErrors := c.GetAllValidationErrors()
	if len(allValidationErrors) != 4 {
		t.Errorf("GetAllValidationErrors() = %d errors, want 4", len(allValidationErrors))
	}
}

func TestErrorCollector_RetryableErrors(t *testing.T) {
	c := NewCollector()

	// Add retryable error
	c.AddRetryable(nil, CategoryExternal, "service unavailable")

	// Add non-retryable (critical) error
	c.Add(NewCritical("critical error", CategoryInternal))

	// Add regular error (potentially retryable)
	c.Add(New("regular error", CategoryOperation))

	// Test HasRetryableErrors
	if !c.HasRetryableErrors() {
		t.Error("HasRetryableErrors() should return true")
	}

	// Test GetRetryableErrors
	retryableErrors := c.GetRetryableErrors()
	if len(retryableErrors) != 2 { // external and operation errors are retryable
		t.Errorf("GetRetryableErrors() = %d errors, want 2", len(retryableErrors))
	}
}

func TestErrorCollector_HttpResponse(t *testing.T) {
	tests := []struct {
		name       string
		errors     []*Error
		wantNil    bool
		wantSingle bool
	}{
		{
			name:    "no errors",
			errors:  nil,
			wantNil: true,
		},
		{
			name:       "single error",
			errors:     []*Error{New("single error")},
			wantSingle: true,
		},
		{
			name: "multiple errors",
			errors: []*Error{
				New("error 1"),
				New("error 2"),
			},
			wantSingle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollector()
			for _, err := range tt.errors {
				c.Add(err)
			}

			response := c.ToErrorResponse(false)

			if tt.wantNil && response != nil {
				t.Error("ToErrorResponse() should return nil for empty collector")
			}
			if !tt.wantNil && response == nil {
				t.Error("ToErrorResponse() should not return nil when collector has errors")
			}
			if response != nil && tt.wantSingle && response.Error.Message != "single error" {
				t.Error("ToErrorResponse() should return single error for single-error collector")
			}
			if response != nil && !tt.wantSingle && len(tt.errors) > 1 && response.Error.Message != "Multiple errors occurred" {
				t.Error("ToErrorResponse() should return merged error for multi-error collector")
			}
		})
	}
}

func TestErrorCollector_Logging(t *testing.T) {
	c := NewCollector()
	c.Add(New("test error", CategoryValidation))
	c.Add(NewWarning("warning", CategoryExternal))

	// Test ToSlogAttributes
	attrs := c.ToSlogAttributes()
	if len(attrs) == 0 {
		t.Error("ToSlogAttributes() should return non-empty attributes")
	}

	// Verify basic attributes exist
	foundErrorCount := false
	for _, attr := range attrs {
		if attr.Key == "error_count" && attr.Value.Int64() == 2 {
			foundErrorCount = true
		}
	}
	if !foundErrorCount {
		t.Error("ToSlogAttributes() should include error_count attribute")
	}

	// Test LogErrors (basic test - just ensure it doesn't panic)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	c.LogErrors(logger)

	// Test with nil logger
	c.LogErrors(nil) // Should not panic
}

func TestErrorCollector_ThreadSafety(t *testing.T) {
	c := NewCollector(WithMaxErrors(1000))

	const goroutines = 10
	const errorsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Start multiple goroutines adding errors concurrently
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < errorsPerGoroutine; j++ {
				err := New(fmt.Sprintf("error %d-%d", id, j), CategoryInternal)
				c.Add(err)
			}
		}(i)
	}

	// Start goroutines reading concurrently
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				c.Count()
				c.HasErrors()
				c.Errors()
				c.CategoryStats()
				c.SeverityDistribution()
				time.Sleep(time.Millisecond)
			}
		}()
		wg.Add(1)
	}

	wg.Wait()

	// Verify final state
	totalExpected := goroutines * errorsPerGoroutine
	if c.Count() != totalExpected {
		t.Errorf("Count() = %d, want %d after concurrent operations", c.Count(), totalExpected)
	}
}

func BenchmarkErrorCollector_Add(b *testing.B) {
	c := NewCollector()
	err := New("benchmark error", CategoryInternal)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Add(err)
	}
}

func BenchmarkErrorCollector_AddConcurrent(b *testing.B) {
	c := NewCollector()
	err := New("benchmark error", CategoryInternal)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(err)
		}
	})
}

func BenchmarkErrorCollector_Merge(b *testing.B) {
	c := NewCollector()

	// Pre-populate with errors
	for i := 0; i < 100; i++ {
		c.Add(New(fmt.Sprintf("error %d", i), CategoryInternal))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Merge()
	}
}

func BenchmarkErrorCollector_CategoryStats(b *testing.B) {
	c := NewCollector()

	// Pre-populate with errors from different categories
	categories := []Category{CategoryValidation, CategoryExternal, CategoryInternal, CategoryAuth}
	for i := 0; i < 100; i++ {
		c.Add(New(fmt.Sprintf("error %d", i), categories[i%len(categories)]))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CategoryStats()
	}
}
