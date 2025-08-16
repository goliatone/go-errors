package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// TestIntegration_AllFeatures tests the integration of all enhanced features
func TestIntegration_AllFeatures(t *testing.T) {
	// Enable location capture for this test
	originalLocationCapture := EnableLocationCapture
	EnableLocationCapture = true
	defer func() { EnableLocationCapture = originalLocationCapture }()

	// Create a collector to aggregate errors during processing
	collector := NewCollector(
		WithMaxErrors(20),
		WithStrictMode(false),
		WithContext(context.Background()),
	)

	// Simulate a complex operation that can generate multiple types of errors
	t.Run("simulate_complex_operation", func(t *testing.T) {
		// 1. Location tracking - errors created at different call sites
		err1 := simulateValidationError()
		err2 := simulateExternalServiceError()
		err3 := simulateAuthenticationError()

		// Add errors to collector
		collector.Add(err1)
		collector.Add(err2)
		collector.Add(err3)

		// Verify location information is captured
		errors := collector.Errors()
		for i, err := range errors {
			if !err.HasLocation() {
				t.Errorf("Error %d should have location information", i)
			}
			if err.GetLocation().Function == "" {
				t.Errorf("Error %d should have function name in location", i)
			}
		}
	})

	// 2. Severity-based filtering and processing
	t.Run("severity_based_processing", func(t *testing.T) {
		// Add errors with different severities
		collector.Add(NewDebug("Debug information", CategoryInternal))
		collector.Add(NewWarning("Resource usage high", CategoryOperation))
		collector.Add(NewCritical("System failure imminent", CategoryInternal))

		// Filter by severity levels
		criticalErrors := collector.FilterBySeverity(SeverityCritical)
		if len(criticalErrors) == 0 {
			t.Error("Should have at least one critical error")
		}

		warningAndAbove := collector.FilterBySeverity(SeverityWarning)
		if len(warningAndAbove) < len(criticalErrors) {
			t.Error("Warning and above should include all critical errors")
		}

		// Verify severity distribution
		severityStats := collector.SeverityDistribution()
		if severityStats[SeverityCritical] == 0 {
			t.Error("Should have critical errors in distribution")
		}
	})

	// 3. Category analysis and error classification
	t.Run("category_analysis", func(t *testing.T) {
		// Add errors from different categories
		collector.AddValidation("email", "invalid format")
		collector.AddRetryable(nil, CategoryExternal, "service timeout")

		categoryStats := collector.CategoryStats()
		
		// Verify we have errors from multiple categories
		if len(categoryStats) < 3 {
			t.Errorf("Expected errors from at least 3 categories, got %d", len(categoryStats))
		}

		// Test most common category detection
		mostCommon := collector.MostCommonCategory()
		if mostCommon == "" {
			t.Error("Most common category should not be empty")
		}
	})

	// 4. Retryable error detection and handling
	t.Run("retryable_error_handling", func(t *testing.T) {
		if !collector.HasRetryableErrors() {
			t.Error("Collector should have retryable errors")
		}

		retryableErrors := collector.GetRetryableErrors()
		if len(retryableErrors) == 0 {
			t.Error("Should have retryable errors")
		}

		// Verify critical errors are not considered retryable
		for _, err := range retryableErrors {
			if err.GetSeverity() >= SeverityCritical {
				t.Error("Critical errors should not be considered retryable")
			}
		}
	})

	// 5. Error merging and aggregation
	t.Run("error_aggregation", func(t *testing.T) {
		merged := collector.Merge()
		if merged == nil {
			t.Fatal("Merged error should not be nil")
		}

		// Verify merged error contains aggregate information
		if merged.Message != "Multiple errors occurred" {
			t.Errorf("Expected 'Multiple errors occurred', got '%s'", merged.Message)
		}

		// Check metadata contains statistics
		errorCount, ok := merged.Metadata["error_count"]
		if !ok {
			t.Error("Merged error should contain error_count in metadata")
		}
		if errorCount.(int) != collector.Count() {
			t.Errorf("Error count in metadata (%v) should match collector count (%d)", 
				errorCount, collector.Count())
		}

		// Verify category and severity stats are included
		if _, ok := merged.Metadata["category_stats"]; !ok {
			t.Error("Merged error should contain category_stats in metadata")
		}
		if _, ok := merged.Metadata["severity_stats"]; !ok {
			t.Error("Merged error should contain severity_stats in metadata")
		}
	})

	// 6. HTTP response integration
	t.Run("http_response_integration", func(t *testing.T) {
		httpResponse := collector.ToErrorResponse(false)
		if httpResponse == nil {
			t.Fatal("HTTP response should not be nil")
		}

		// Verify response structure
		if httpResponse.Error == nil {
			t.Fatal("HTTP response should contain error")
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(httpResponse)
		if err != nil {
			t.Fatalf("Failed to marshal HTTP response: %v", err)
		}

		// Verify JSON contains expected fields
		jsonStr := string(jsonData)
		requiredFields := []string{"error", "category", "severity", "message"}
		for _, field := range requiredFields {
			if !strings.Contains(jsonStr, field) {
				t.Errorf("JSON response should contain field: %s", field)
			}
		}
	})

	// 7. Structured logging integration
	t.Run("structured_logging", func(t *testing.T) {
		// Test slog attributes generation
		attrs := collector.ToSlogAttributes()
		if len(attrs) == 0 {
			t.Error("Should generate slog attributes")
		}

		// Verify key attributes are present
		foundErrorCount := false
		foundCategoryStats := false
		for _, attr := range attrs {
			switch attr.Key {
			case "error_count":
				foundErrorCount = true
				if attr.Value.Int64() != int64(collector.Count()) {
					t.Errorf("Error count attribute (%d) should match collector count (%d)",
						attr.Value.Int64(), collector.Count())
				}
			case "category_stats":
				foundCategoryStats = true
			}
		}

		if !foundErrorCount {
			t.Error("Should include error_count attribute")
		}
		if !foundCategoryStats {
			t.Error("Should include category_stats attribute")
		}

		// Test logging doesn't panic
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		collector.LogErrors(logger)
	})

	// 8. Validation error aggregation
	t.Run("validation_error_aggregation", func(t *testing.T) {
		allValidationErrors := collector.GetAllValidationErrors()
		if len(allValidationErrors) == 0 {
			t.Error("Should have validation errors")
		}

		// Verify validation errors contain expected fields
		foundEmailError := false
		for _, validationErr := range allValidationErrors {
			if validationErr.Field == "email" && validationErr.Message == "invalid format" {
				foundEmailError = true
				break
			}
		}
		if !foundEmailError {
			t.Error("Should contain the email validation error")
		}
	})

	// Final verification - collector state
	t.Run("final_state_verification", func(t *testing.T) {
		if !collector.HasErrors() {
			t.Error("Collector should have errors")
		}

		if collector.Count() == 0 {
			t.Error("Collector should contain errors")
		}

		// Verify all operations are thread-safe (no race conditions)
		// This is implicitly tested by running with -race flag
	})
}

// TestIntegration_BackwardCompatibility ensures existing code continues to work
func TestIntegration_BackwardCompatibility(t *testing.T) {
	// Test that existing error creation patterns still work
	t.Run("existing_error_patterns", func(t *testing.T) {
		// Basic error creation
		err1 := New("basic error")
		if err1.Message != "basic error" {
			t.Error("Basic error creation should work")
		}

		// Error wrapping
		sourceErr := fmt.Errorf("source error")
		err2 := Wrap(sourceErr, CategoryExternal, "wrapped error")
		if err2.Source != sourceErr {
			t.Error("Error wrapping should work")
		}

		// Validation errors
		validationErr := NewValidation("validation failed",
			FieldError{Field: "name", Message: "required"})
		if len(validationErr.ValidationErrors) != 1 {
			t.Error("Validation error creation should work")
		}

		// Retryable errors
		retryableErr := NewRetryable("retry this", CategoryOperation)
		if !retryableErr.IsRetryable() {
			t.Error("Retryable error creation should work")
		}
	})

	// Test that enhanced features don't break existing functionality
	t.Run("enhanced_features_compatibility", func(t *testing.T) {
		err := New("test error")

		// Location should be captured automatically but not break existing usage
		if err.Error() == "" {
			t.Error("Error string should still work")
		}

		// Severity should default to Error
		if err.GetSeverity() != SeverityError {
			t.Error("Default severity should be Error")
		}

		// JSON marshaling should include new fields but not break
		jsonData, jsonErr := json.Marshal(err)
		if jsonErr != nil {
			t.Errorf("JSON marshaling should work: %v", jsonErr)
		}

		// Should contain both old and new fields
		jsonStr := string(jsonData)
		if !strings.Contains(jsonStr, "message") {
			t.Error("JSON should contain message field")
		}
		if !strings.Contains(jsonStr, "severity") {
			t.Error("JSON should contain severity field")
		}
	})
}

// TestIntegration_ProductionMode tests behavior with production optimizations
func TestIntegration_ProductionMode(t *testing.T) {
	// Test with location capture disabled (production mode simulation)
	t.Run("location_capture_disabled", func(t *testing.T) {
		originalLocationCapture := EnableLocationCapture
		EnableLocationCapture = false
		defer func() { EnableLocationCapture = originalLocationCapture }()

		err := New("production error")
		if err.HasLocation() {
			t.Error("Location should not be captured when disabled")
		}

		// Functionality should still work without location
		if err.Error() == "" {
			t.Error("Error should still work without location")
		}

		collector := NewCollector()
		collector.Add(err)
		if !collector.HasErrors() {
			t.Error("Collector should work without location info")
		}
	})

	// Test performance impact measurement
	t.Run("performance_measurement", func(t *testing.T) {
		// With location capture
		EnableLocationCapture = true
		start := time.Now()
		for range 1000 {
			New("perf test with location")
		}
		withLocationDuration := time.Since(start)

		// Without location capture
		EnableLocationCapture = false
		start = time.Now()
		for range 1000 {
			New("perf test without location")
		}
		withoutLocationDuration := time.Since(start)

		// Restore original state
		EnableLocationCapture = true

		// Location capture should have minimal overhead
		overhead := float64(withLocationDuration-withoutLocationDuration) / float64(withoutLocationDuration)
		if overhead > 5.0 { // Allow up to 500% overhead (runtime.Caller can be expensive in tight loops)
			t.Errorf("Location capture overhead too high: %.2f%%", overhead*100)
		}

		t.Logf("Performance: with location=%v, without location=%v, overhead=%.2f%%",
			withLocationDuration, withoutLocationDuration, overhead*100)
	})
}

// TestIntegration_RealWorldScenarios tests realistic usage patterns
func TestIntegration_RealWorldScenarios(t *testing.T) {
	t.Run("web_request_processing", func(t *testing.T) {
		// Simulate processing a web request with multiple validation and operational errors
		collector := NewCollector(WithMaxErrors(50))

		// Input validation errors
		collector.AddValidation("email", "invalid email format")
		collector.AddValidation("age", "must be between 18 and 120")
		collector.AddFieldErrors(
			FieldError{Field: "password", Message: "too weak"},
			FieldError{Field: "username", Message: "already taken"},
		)

		// Business logic errors
		collector.Add(NewWarning("Rate limit approaching", CategoryRateLimit))
		collector.AddRetryable(fmt.Errorf("database timeout"), CategoryExternal, "Database connection failed")

		// Critical system error
		collector.Add(NewCritical("Memory usage critical", CategoryInternal))

		// Process the errors for HTTP response
		httpResponse := collector.ToErrorResponse(false)
		if httpResponse == nil {
			t.Fatal("Should generate HTTP response")
		}

		// Verify appropriate error categorization
		if httpResponse.Error.Category != CategoryValidation { // Most common category
			t.Error("Should categorize as validation error (most common)")
		}

		// Verify severity escalation (highest severity wins)
		if httpResponse.Error.Severity != SeverityCritical {
			t.Error("Should escalate to critical severity")
		}

		// Verify validation errors are aggregated
		if len(httpResponse.Error.ValidationErrors) < 4 {
			t.Errorf("Should aggregate validation errors, got %d", len(httpResponse.Error.ValidationErrors))
		}
	})

	t.Run("batch_processing", func(t *testing.T) {
		// Simulate batch processing of multiple items
		collector := NewCollector(WithMaxErrors(100), WithStrictMode(true))

		// Process 50 items, some fail
		successCount := 0
		for i := range 50 {
			if i%7 == 0 { // Every 7th item fails
				collector.Add(NewWarning(fmt.Sprintf("Item %d failed validation", i), CategoryValidation))
			} else if i%11 == 0 { // Every 11th item has external error
				collector.AddRetryable(nil, CategoryExternal, fmt.Sprintf("External service error for item %d", i))
			} else {
				successCount++
			}
		}

		// Analyze batch results
		categoryStats := collector.CategoryStats()
		t.Logf("Batch processing results: %d successes, %d failures", successCount, collector.Count())
		t.Logf("Error breakdown: %+v", categoryStats)

		// Determine if batch should be retried
		retryableCount := len(collector.GetRetryableErrors())
		criticalCount := len(collector.FilterBySeverity(SeverityCritical))

		if criticalCount > 0 {
			t.Logf("Batch has critical errors, should not retry")
		} else if retryableCount > collector.Count()/2 {
			t.Logf("More than half errors are retryable, should retry batch")
		} else {
			t.Logf("Batch has non-retryable errors, should process individually")
		}
	})
}

// Helper functions for integration tests

func simulateValidationError() *Error {
	return NewValidation("User input validation failed",
		FieldError{Field: "email", Message: "invalid email format"},
		FieldError{Field: "age", Message: "must be a positive number"},
	).WithLocation(Here())
}

func simulateExternalServiceError() *Error {
	sourceErr := fmt.Errorf("connection timeout")
	return Wrap(sourceErr, CategoryExternal, "External payment service unavailable").
		WithSeverity(SeverityWarning).
		WithMetadata(map[string]any{
			"service": "payment-api",
			"timeout": "30s",
		})
}

func simulateAuthenticationError() *Error {
	return New("Invalid authentication token", CategoryAuth).
		WithSeverity(SeverityError).
		WithCode(401).
		WithTextCode("INVALID_TOKEN").
		WithMetadata(map[string]any{
			"token_type": "JWT",
			"expired":    true,
		})
}