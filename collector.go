package errors

import (
	"context"
	"log/slog"
	"sync"
)

// ErrorCollector provides thread-safe collection and aggregation of errors
// for batch operations and complex error handling scenarios
type ErrorCollector struct {
	mu         sync.RWMutex
	errors     []*Error
	maxErrors  int
	strictMode bool
	context    context.Context
}

// CollectorOption defines functional options for ErrorCollector configuration
type CollectorOption func(*ErrorCollector)

// NewCollector creates a new ErrorCollector with the provided options
func NewCollector(opts ...CollectorOption) *ErrorCollector {
	c := &ErrorCollector{
		errors:     make([]*Error, 0),
		maxErrors:  100, // Default maximum
		strictMode: false,
		context:    context.Background(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithMaxErrors sets the maximum number of errors to collect
func WithMaxErrors(max int) CollectorOption {
	return func(c *ErrorCollector) {
		c.maxErrors = max
	}
}

// WithStrictMode enables strict mode where the collector returns false
// when maximum errors are reached (instead of silently dropping errors)
func WithStrictMode(strict bool) CollectorOption {
	return func(c *ErrorCollector) {
		c.strictMode = strict
	}
}

// WithContext sets the context for the collector
func WithContext(ctx context.Context) CollectorOption {
	return func(c *ErrorCollector) {
		c.context = ctx
	}
}

// Add adds an error to the collector in a thread-safe manner
// Returns true if the error was added, false if the collector is full
// and operating in strict mode
func (c *ErrorCollector) Add(err error) bool {
	if err == nil {
		return true
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we've reached the maximum
	if len(c.errors) >= c.maxErrors {
		if c.strictMode {
			return false
		}
		// In non-strict mode, remove the oldest error to make room
		c.errors = c.errors[1:]
	}

	// Convert to our Error type if needed
	var customErr *Error
	if As(err, &customErr) {
		c.errors = append(c.errors, customErr)
	} else {
		// Wrap foreign errors
		wrappedErr := Wrap(err, CategoryInternal, err.Error())
		c.errors = append(c.errors, wrappedErr)
	}

	return true
}

// HasErrors returns true if the collector contains any errors
func (c *ErrorCollector) HasErrors() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.errors) > 0
}

// Count returns the number of errors currently in the collector
func (c *ErrorCollector) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.errors)
}

// Errors returns a copy of all errors in the collector
func (c *ErrorCollector) Errors() []*Error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*Error, len(c.errors))
	copy(result, c.errors)
	return result
}

// Reset clears all errors from the collector
func (c *ErrorCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors = c.errors[:0] // Clear slice but keep capacity
}

// Merge creates a single aggregate error from all collected errors
// Returns nil if no errors have been collected
func (c *ErrorCollector) Merge() *Error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.errors) == 0 {
		return nil
	}

	if len(c.errors) == 1 {
		return c.errors[0].Clone()
	}

	// Create aggregate error with metadata about collected errors
	categoryStats := c.categoryStatsUnsafe()
	severityStats := c.severityDistributionUnsafe()

	// Find the highest severity level
	highestSeverity := SeverityDebug
	for severity := range severityStats {
		if severity > highestSeverity {
			highestSeverity = severity
		}
	}

	// Use the most common category, or CategoryInternal if tied
	mostCommonCategory := c.mostCommonCategoryUnsafe()

	// Create the aggregate error
	aggregate := New("Multiple errors occurred", mostCommonCategory).
		WithSeverity(highestSeverity).
		WithMetadata(map[string]any{
			"error_count":    len(c.errors),
			"category_stats": categoryStats,
			"severity_stats": severityStats,
			"aggregated_at":  c.errors[0].Timestamp, // Use first error's timestamp
		})

	// Collect all validation errors
	var allValidationErrors ValidationErrors
	for _, err := range c.errors {
		allValidationErrors = append(allValidationErrors, err.ValidationErrors...)
	}
	if len(allValidationErrors) > 0 {
		aggregate.ValidationErrors = allValidationErrors
	}

	return aggregate
}

// FilterBySeverity returns all errors with severity at or above the specified minimum
func (c *ErrorCollector) FilterBySeverity(min Severity) []*Error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var filtered []*Error
	for _, err := range c.errors {
		if err.GetSeverity() >= min {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// FilterByCategory returns all errors matching the specified category
func (c *ErrorCollector) FilterByCategory(cat Category) []*Error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var filtered []*Error
	for _, err := range c.errors {
		if err.Category == cat {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// GetValidationErrors aggregates all validation errors from collected errors
func (c *ErrorCollector) GetValidationErrors() ValidationErrors {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allValidationErrors ValidationErrors
	for _, err := range c.errors {
		allValidationErrors = append(allValidationErrors, err.ValidationErrors...)
	}
	return allValidationErrors
}

// CategoryStats returns the count of errors by category
func (c *ErrorCollector) CategoryStats() map[Category]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.categoryStatsUnsafe()
}

// categoryStatsUnsafe returns category statistics without acquiring locks
// Must be called while holding at least a read lock
func (c *ErrorCollector) categoryStatsUnsafe() map[Category]int {
	stats := make(map[Category]int)
	for _, err := range c.errors {
		stats[err.Category]++
	}
	return stats
}

// MostCommonCategory returns the category with the highest error count
func (c *ErrorCollector) MostCommonCategory() Category {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mostCommonCategoryUnsafe()
}

// mostCommonCategoryUnsafe returns the most common category without acquiring locks
// Must be called while holding at least a read lock
func (c *ErrorCollector) mostCommonCategoryUnsafe() Category {
	stats := c.categoryStatsUnsafe()

	var mostCommon Category = CategoryInternal
	maxCount := 0

	for category, count := range stats {
		if count > maxCount {
			maxCount = count
			mostCommon = category
		}
	}

	return mostCommon
}

// SeverityDistribution returns the count of errors by severity level
func (c *ErrorCollector) SeverityDistribution() map[Severity]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.severityDistributionUnsafe()
}

// severityDistributionUnsafe returns severity distribution without acquiring locks
// Must be called while holding at least a read lock
func (c *ErrorCollector) severityDistributionUnsafe() map[Severity]int {
	stats := make(map[Severity]int)
	for _, err := range c.errors {
		stats[err.GetSeverity()]++
	}
	return stats
}

// AddValidation adds a validation error for a specific field
func (c *ErrorCollector) AddValidation(field, message string) {
	validationErr := NewValidation("Validation failed", FieldError{
		Field:   field,
		Message: message,
	})
	c.Add(validationErr)
}

// AddFieldErrors adds multiple field errors as a single validation error
func (c *ErrorCollector) AddFieldErrors(errors ...FieldError) {
	if len(errors) == 0 {
		return
	}

	validationErr := NewValidation("Validation failed", errors...)
	c.Add(validationErr)
}

// GetAllValidationErrors aggregates all validation errors from all collected errors
// This includes both direct validation errors and validation errors from wrapped errors
func (c *ErrorCollector) GetAllValidationErrors() ValidationErrors {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allValidationErrors ValidationErrors
	for _, err := range c.errors {
		// Get all validation errors (including from wrapped errors)
		allValidationErrors = append(allValidationErrors, err.AllValidationErrors()...)
	}
	return allValidationErrors
}

// AddRetryable adds a retryable error to the collector
func (c *ErrorCollector) AddRetryable(err error, category Category, message string) {
	var retryableErr *RetryableError
	if err != nil {
		retryableErr = WrapRetryable(err, category, message)
	} else {
		retryableErr = NewRetryable(message, category)
	}
	c.Add(retryableErr.BaseError)
}

// HasRetryableErrors returns true if any collected errors are retryable
func (c *ErrorCollector) HasRetryableErrors() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, err := range c.errors {
		// Check if this error would be retryable by wrapping it in RetryableError
		// and checking its retryability based on severity
		if err.GetSeverity() < SeverityCritical {
			return true
		}
	}
	return false
}

// GetRetryableErrors returns all errors that could be considered retryable
// based on their severity level (errors below Critical severity)
func (c *ErrorCollector) GetRetryableErrors() []*Error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var retryableErrors []*Error
	for _, err := range c.errors {
		// Errors with severity below Critical are potentially retryable
		if err.GetSeverity() < SeverityCritical {
			retryableErrors = append(retryableErrors, err)
		}
	}
	return retryableErrors
}

// ToErrorResponse converts the collector's state to an HTTP error response
// If the collector has no errors, returns nil
// If the collector has exactly one error, returns that error's response
// If the collector has multiple errors, returns a merged error response
func (c *ErrorCollector) ToErrorResponse(includeStack bool) *ErrorResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.errors) == 0 {
		return nil
	}

	if len(c.errors) == 1 {
		// For single error, use its existing ToErrorResponse method
		response := c.errors[0].ToErrorResponse(includeStack, c.errors[0].StackTrace)
		return &response
	}

	// For multiple errors, create a merged error
	merged := c.mergeUnsafe()
	if merged == nil {
		return nil
	}

	response := merged.ToErrorResponse(includeStack, merged.StackTrace)
	return &response
}

// mergeUnsafe is an internal version of Merge that doesn't acquire locks
// Must be called while holding at least a read lock
func (c *ErrorCollector) mergeUnsafe() *Error {
	if len(c.errors) == 0 {
		return nil
	}

	if len(c.errors) == 1 {
		return c.errors[0].Clone()
	}

	// Create aggregate error with metadata about collected errors
	categoryStats := c.categoryStatsUnsafe()
	severityStats := c.severityDistributionUnsafe()

	// Find the highest severity level
	highestSeverity := SeverityDebug
	for severity := range severityStats {
		if severity > highestSeverity {
			highestSeverity = severity
		}
	}

	// Use the most common category, or CategoryInternal if tied
	mostCommonCategory := c.mostCommonCategoryUnsafe()

	// Create the aggregate error
	aggregate := New("Multiple errors occurred", mostCommonCategory).
		WithSeverity(highestSeverity).
		WithMetadata(map[string]any{
			"error_count":    len(c.errors),
			"category_stats": categoryStats,
			"severity_stats": severityStats,
			"aggregated_at":  c.errors[0].Timestamp, // Use first error's timestamp
		})

	// Collect all validation errors
	var allValidationErrors ValidationErrors
	for _, err := range c.errors {
		allValidationErrors = append(allValidationErrors, err.ValidationErrors...)
	}
	if len(allValidationErrors) > 0 {
		aggregate.ValidationErrors = allValidationErrors
	}

	return aggregate
}

// ToSlogAttributes creates slog attributes for the collector's current state
// Includes error count, context information, and category/severity statistics
func (c *ErrorCollector) ToSlogAttributes() []slog.Attr {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var attrs []slog.Attr

	// Basic collector information
	attrs = append(attrs, slog.Int("error_count", len(c.errors)))
	attrs = append(attrs, slog.Int("max_errors", c.maxErrors))
	attrs = append(attrs, slog.Bool("strict_mode", c.strictMode))

	if len(c.errors) > 0 {
		// Category statistics
		categoryStats := c.categoryStatsUnsafe()
		attrs = append(attrs, slog.Any("category_stats", categoryStats))

		// Severity distribution
		severityStats := c.severityDistributionUnsafe()
		attrs = append(attrs, slog.Any("severity_stats", severityStats))

		// Most common category
		mostCommon := c.mostCommonCategoryUnsafe()
		attrs = append(attrs, slog.String("most_common_category", mostCommon.String()))

		// Validation error count
		var allValidationErrors ValidationErrors
		for _, err := range c.errors {
			allValidationErrors = append(allValidationErrors, err.ValidationErrors...)
		}
		if len(allValidationErrors) > 0 {
			attrs = append(attrs, slog.Int("validation_error_count", len(allValidationErrors)))
		}

		// Retryable error count
		retryableCount := len(c.getRetryableErrorsUnsafe())
		if retryableCount > 0 {
			attrs = append(attrs, slog.Int("retryable_error_count", retryableCount))
		}
	}

	return attrs
}

// LogErrors logs all collected errors using the provided logger
// Each error is logged individually with its appropriate severity level
func (c *ErrorCollector) LogErrors(logger *slog.Logger) {
	if logger == nil {
		return
	}

	c.mu.RLock()
	errors := make([]*Error, len(c.errors))
	copy(errors, c.errors)
	c.mu.RUnlock()

	// Add collector context to each log entry
	collectorAttrs := c.ToSlogAttributes()

	for i, err := range errors {
		// Create combined attributes with both error and collector information
		errorAttrs := ToSlogAttributes(err)
		allAttrs := make([]slog.Attr, 0, len(errorAttrs)+len(collectorAttrs)+1)
		allAttrs = append(allAttrs, errorAttrs...)
		allAttrs = append(allAttrs, collectorAttrs...)
		allAttrs = append(allAttrs, slog.Int("error_index", i))

		// Convert to []any for logging
		anyAttrs := make([]any, len(allAttrs))
		for j, attr := range allAttrs {
			anyAttrs[j] = attr
		}

		// Log with appropriate severity level
		switch err.Severity {
		case SeverityDebug:
			logger.Debug(err.Error(), anyAttrs...)
		case SeverityInfo:
			logger.Info(err.Error(), anyAttrs...)
		case SeverityWarning:
			logger.Warn(err.Error(), anyAttrs...)
		case SeverityError:
			logger.Error(err.Error(), anyAttrs...)
		case SeverityCritical:
			logger.Error(err.ErrorWithStack(), anyAttrs...)
		case SeverityFatal:
			logger.Error(err.ErrorWithStack(), anyAttrs...)
		default:
			logger.Error(err.Error(), anyAttrs...)
		}
	}
}

// getRetryableErrorsUnsafe returns retryable errors without acquiring locks
// Must be called while holding at least a read lock
func (c *ErrorCollector) getRetryableErrorsUnsafe() []*Error {
	var retryableErrors []*Error
	for _, err := range c.errors {
		if err.GetSeverity() < SeverityCritical {
			retryableErrors = append(retryableErrors, err)
		}
	}
	return retryableErrors
}
