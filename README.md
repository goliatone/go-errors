# Go Errors

A comprehensive error handling package that provides structured errors with rich context, location tracking, severity levels, error collection, validation support, retryable errors, and JSON serialization.

## Features

- **Structured Error Types**: Categorized errors with consistent structure across packages
- **Rich Context**: Metadata, stack traces, request IDs, and timestamps
- **🎯 Location Tracking**: Automatic file/line location capture with runtime.Caller
- **📊 Severity Levels**: Hierarchical error classification (Debug, Info, Warning, Error, Critical, Fatal)
- **📦 Error Collection**: Thread-safe aggregation and batch error handling with analysis
- **Validation Support**: Built-in handling for field-level validation errors with ozzo-validation integration
- **Retryable Errors**: Support for retryable errors with exponential backoff
- **JSON Serialization**: Full JSON marshaling/unmarshaling support
- **Error Wrapping**: Compatible with Go's `errors.Is` and `errors.As`
- **Fluent Interface**: Chainable methods for building complex errors
- **Code Support**: Both numeric and text codes for easy error identification
- **HTTP Integration**: Built-in HTTP status code mapping and error response structures
- **Logging Support**: Enhanced slog integration with severity-based logging and collector attributes
- **Production Ready**: Configurable features with build tags for performance optimization

## Installation

```bash
go get github.com/goliatone/go-errors
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/goliatone/go-errors"
)

func main() {
    // Create a simple error - note the parameter order change
    err := errors.New("user not found", errors.CategoryNotFound)

    // Add context using fluent interface
    enrichedErr := err.
        WithMetadata(map[string]any{"user_id": 123}).
        WithRequestID("req-456").
        WithStackTrace().
        WithCode(404).
        WithTextCode("USER_NOT_FOUND")

    fmt.Println(enrichedErr.Error())

    // Create a validation error
    validationErr := errors.NewValidation("invalid input",
        errors.FieldError{Field: "email", Message: "required"},
        errors.FieldError{Field: "age", Message: "must be positive"},
    )

    // Wrap an existing error
    wrappedErr := errors.Wrap(err, errors.CategoryInternal, "database query failed")
}
```

## Error Categories

The package provides a set of predefined error categories:

- `CategoryValidation` - Input validation failures
- `CategoryAuth` - Authentication errors
- `CategoryAuthz` - Authorization errors
- `CategoryOperation` - Operation failures
- `CategoryNotFound` - Resource not found
- `CategoryConflict` - Resource conflicts
- `CategoryRateLimit` - Rate limiting errors
- `CategoryBadInput` - Malformed input
- `CategoryInternal` - Internal system errors
- `CategoryExternal` - External service errors
- `CategoryMiddleware` - Middleware errors
- `CategoryRouting` - Routing errors
- `CategoryHandler` - Handler errors
- `CategoryMethodNotAllowed` - HTTP method not allowed
- `CategoryCommand` - Command execution errors

## Enhanced Features

### 🎯 Location Tracking

Automatically capture the file, line, and function where errors are created:

```go
// Automatic location capture (enabled by default)
err := errors.New("database connection failed", errors.CategoryExternal)
fmt.Printf("Error occurred at: %s\n", err.GetLocation().String())
// Output: Error occurred at: main.go:25

// Explicit location capture
err = errors.New("critical error").WithLocation(errors.Here())

// Check if error has location
if err.HasLocation() {
    loc := err.GetLocation()
    fmt.Printf("File: %s, Line: %d, Function: %s\n", 
        loc.File, loc.Line, loc.Function)
}

// Configure location capture globally
errors.SetLocationCapture(false) // Disable for production
errors.SetLocationCapture(true)  // Re-enable

// Environment variable control
// GO_ERRORS_DISABLE_LOCATION=true disables location capture
```

### 📊 Severity Levels

Classify errors by severity with built-in level hierarchy:

```go
// Severity levels (from lowest to highest)
// SeverityDebug, SeverityInfo, SeverityWarning, 
// SeverityError, SeverityCritical, SeverityFatal

// Create errors with specific severity
debugErr := errors.NewDebug("Cache miss", errors.CategoryInternal)
infoErr := errors.NewInfo("User logged in", errors.CategoryAuth)
warningErr := errors.NewWarning("Rate limit approaching", errors.CategoryRateLimit)
errorErr := errors.New("Validation failed", errors.CategoryValidation) // Default: Error
criticalErr := errors.NewCritical("Database down", errors.CategoryExternal)

// Set severity on existing errors
err := errors.New("system issue").WithSeverity(errors.SeverityCritical)

// Check severity levels
if err.IsAboveSeverity(errors.SeverityWarning) {
    // Handle serious errors
}

if err.HasSeverity(errors.SeverityCritical) {
    // Handle critical errors specifically
}

// Get severity level
severity := err.GetSeverity()
fmt.Printf("Error severity: %s\n", severity) // Output: CRITICAL

// Severity-based logging (automatically maps to slog levels)
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
errors.LogBySeverity(logger, err) // Logs with appropriate level
```

### 📦 Error Collection

Collect and aggregate multiple errors with thread-safe operations:

```go
// Create error collector with options
collector := errors.NewCollector(
    errors.WithMaxErrors(50),
    errors.WithStrictMode(false), // false = FIFO when full
    errors.WithContext(context.Background()),
)

// Add various types of errors
collector.Add(errors.NewValidation("Email invalid", 
    errors.FieldError{Field: "email", Message: "required"}))
collector.AddValidation("age", "must be positive")
collector.AddRetryable(nil, errors.CategoryExternal, "Service timeout")

// Basic collector operations
fmt.Printf("Total errors: %d\n", collector.Count())
fmt.Printf("Has errors: %t\n", collector.HasErrors())
fmt.Printf("Has retryable: %t\n", collector.HasRetryableErrors())

// Error analysis and filtering
categoryStats := collector.CategoryStats()
// Returns: map[Category]int{CategoryValidation: 2, CategoryExternal: 1}

severityStats := collector.SeverityDistribution()
// Returns: map[Severity]int{SeverityError: 2, SeverityWarning: 1}

mostCommon := collector.MostCommonCategory()
criticalErrors := collector.FilterBySeverity(errors.SeverityCritical)
validationErrors := collector.FilterByCategory(errors.CategoryValidation)

// Aggregate errors into a single error
merged := collector.Merge()
// Creates single error with metadata containing statistics

// Convert to HTTP response
httpResponse := collector.ToErrorResponse(false)
jsonData, _ := json.Marshal(httpResponse)

// Structured logging for all collected errors
collector.LogErrors(logger) // Logs each error with collector context

// Reset collector for reuse
collector.Reset()
```

### Advanced Usage Patterns

#### Batch Processing with Error Collection

```go
func processBatch(items []Item) error {
    collector := errors.NewCollector(errors.WithMaxErrors(100))
    
    for _, item := range items {
        if err := processItem(item); err != nil {
            collector.Add(err)
        }
    }
    
    if !collector.HasErrors() {
        return nil // All items processed successfully
    }
    
    // Analyze errors to determine strategy
    retryableCount := len(collector.GetRetryableErrors())
    criticalCount := len(collector.FilterBySeverity(errors.SeverityCritical))
    
    if criticalCount > 0 {
        // Abort batch due to critical errors
        return collector.Merge()
    } else if retryableCount > collector.Count()/2 {
        // Retry entire batch if mostly retryable errors
        return errors.NewRetryable("Batch has retryable errors", 
            errors.CategoryOperation).BaseError
    }
    
    // Return aggregated error for individual item handling
    return collector.Merge()
}
```

#### Web API Error Handling

```go
func registerUser(req UserRequest) (*UserResponse, error) {
    collector := errors.NewCollector()
    
    // Validate input
    if req.Email == "" {
        collector.AddValidation("email", "email is required")
    }
    if req.Age < 18 {
        collector.AddValidation("age", "must be 18 or older")
    }
    
    // Check business rules
    if isReservedUsername(req.Username) {
        collector.Add(errors.New("Username reserved", errors.CategoryConflict).
            WithCode(409).WithTextCode("USERNAME_RESERVED"))
    }
    
    // External service checks
    if err := validateEmailWithService(req.Email); err != nil {
        collector.AddRetryable(err, errors.CategoryExternal, 
            "Email validation service failed")
    }
    
    if collector.HasErrors() {
        // Return HTTP-ready error response
        return nil, collector.Merge()
    }
    
    // Proceed with user creation...
    return &UserResponse{}, nil
}
```

## Core Types

### Error

The main error type that carries all error information:

```go
type Error struct {
    Category         Category           `json:"category"`
    Code             int                `json:"code,omitempty"`
    TextCode         string             `json:"text_code,omitempty"`
    Message          string             `json:"message"`
    Source           error              `json:"-"`
    ValidationErrors ValidationErrors   `json:"validation_errors,omitempty"`
    Metadata         map[string]any     `json:"metadata,omitempty"`
    RequestID        string             `json:"request_id,omitempty"`
    Timestamp        time.Time          `json:"timestamp"`
    StackTrace       StackTrace         `json:"stack_trace,omitempty"`
    Location         *ErrorLocation     `json:"location,omitempty"`      // NEW: Location tracking
    Severity         Severity           `json:"severity"`                // NEW: Severity level
}
```

### ValidationErrors

For handling multiple field-level validation errors:

```go
type ValidationErrors []FieldError

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   any    `json:"value,omitempty"`
}
```

### ErrorLocation

For capturing error location information:

```go
type ErrorLocation struct {
    File     string `json:"file"`     // Full file path
    Line     int    `json:"line"`     // Line number
    Function string `json:"function"` // Function name
}
```

### Severity

For error classification by severity level:

```go
type Severity int

const (
    SeverityDebug Severity = iota    // Debug information
    SeverityInfo                     // Informational messages  
    SeverityWarning                  // Warning conditions
    SeverityError                    // Error conditions (default)
    SeverityCritical                 // Critical conditions
    SeverityFatal                    // Fatal conditions
)
```

### ErrorCollector

For aggregating multiple errors:

```go
type ErrorCollector struct {
    // Internal fields (thread-safe)
}

// Functional options for configuration
type CollectorOption func(*ErrorCollector)

// Option functions:
// WithMaxErrors(max int) - Set maximum errors (default: 100)
// WithStrictMode(strict bool) - Enable strict mode (default: false)  
// WithContext(ctx context.Context) - Set context
```

## Constructor Functions

### Basic Constructors

```go
// New creates an error with message and optional category (defaults to CategoryInternal)
err := errors.New("resource not found", errors.CategoryNotFound)

// Wrap an existing error with additional context
wrappedErr := errors.Wrap(sourceErr, errors.CategoryInternal, "operation failed")
```

### Validation Constructors

```go
err := errors.NewValidation("validation failed",
    errors.FieldError{Field: "email", Message: "invalid format"},
    errors.FieldError{Field: "age", Message: "must be positive"},
)

// From a map
fieldMap := map[string]string{
    "name": "required",
    "email": "invalid format",
}
err := errors.NewValidationFromMap("validation failed", fieldMap)

// From grouped errors (useful for complex validations)
groups := map[string][]string{
    "user": {"name required", "email invalid"},
    "address": {"city required"},
}
err := errors.NewValidationFromGroups("validation failed", groups)
```

### Severity-Based Constructors

```go
// Create errors with specific severity levels
debugErr := errors.NewDebug("Debug info", errors.CategoryInternal)
infoErr := errors.NewInfo("User action", errors.CategoryAuth)  
warningErr := errors.NewWarning("Resource low", errors.CategoryOperation)
criticalErr := errors.NewCritical("System failing", errors.CategoryInternal)

// Default constructor creates SeverityError
errorErr := errors.New("Something failed", errors.CategoryValidation)
```

### Collector Constructors

```go
// Create error collector with default settings
collector := errors.NewCollector()

// Create collector with custom configuration
collector := errors.NewCollector(
    errors.WithMaxErrors(200),
    errors.WithStrictMode(true),
    errors.WithContext(ctx),
)
```

### Location-Aware Constructors

```go
// Automatic location capture (default behavior)
err := errors.New("error message", errors.CategoryInternal)

// Explicit location setting  
err := errors.NewWithLocation("error message", 
    errors.CategoryInternal, errors.Here())

// Disable location for this error only
errors.SetLocationCapture(false)
err := errors.New("no location", errors.CategoryInternal)
errors.SetLocationCapture(true)
```

## Fluent Interface

Chain methods to build rich error context:

```go
err := errors.New("authentication failed", errors.CategoryAuth).
    WithCode(401).
    WithTextCode("AUTH_FAILED").
    WithSeverity(errors.SeverityCritical).           // NEW: Set severity
    WithLocation(errors.Here()).                     // NEW: Set location
    WithMetadata(map[string]any{
        "user_id": 123,
        "attempt": 3,
    }).
    WithRequestID("req-789").
    WithStackTrace()

// Additional fluent methods for enhanced features:
err = err.WithSeverity(errors.SeverityWarning)      // Change severity
if err.HasLocation() {                               // Check location
    fmt.Printf("Location: %s\n", err.GetLocation().String())
}
if err.IsAboveSeverity(errors.SeverityInfo) {       // Check severity threshold
    // Handle serious errors
}
```

## Error Checking

The package provides utility functions to check error types and categories:

```go
// Check specific error categories
if errors.IsValidation(err) {
    // Handle validation error
}

if errors.IsAuth(err) {
    // Handle authentication error
}

if errors.IsNotFound(err) {
    // Handle not found error
}

if errors.IsInternal(err) {
    // Handle internal error
}

if errors.IsCommand(err) {
    // Handle command error
}

// Check any category
if errors.IsCategory(err, errors.CategoryRateLimit) {
    // Handle rate limit error
}

// Check if category exists anywhere in error chain
if errors.HasCategory(err, errors.CategoryExternal) {
    // Handle external service error (even if wrapped)
}

// Extract validation errors
if validationErrs, ok := errors.GetValidationErrors(err); ok {
    for _, fieldErr := range validationErrs {
        fmt.Printf("%s: %s\n", fieldErr.Field, fieldErr.Message)
    }
}

// Use standard library error functions
if errors.Is(err, originalErr) {
    // Handle specific error
}

var myErr *errors.Error
if errors.As(err, &myErr) {
    // Access structured error fields
    fmt.Println(myErr.Category)
}

// Get root cause of error chain
rootErr := errors.RootCause(err)

// Get root category from error chain
rootCategory := errors.RootCategory(err)

// NEW: Enhanced error checking with severity and location
if err.HasSeverity(errors.SeverityCritical) {
    // Handle critical errors
}

if err.IsAboveSeverity(errors.SeverityWarning) {
    // Handle serious errors (Warning, Error, Critical, Fatal)
}

severity := err.GetSeverity()
fmt.Printf("Error severity: %s\n", severity)

if err.HasLocation() {
    loc := err.GetLocation()
    fmt.Printf("Error at %s:%d in %s\n", loc.File, loc.Line, loc.Function)
}
```

## JSON Serialization

Errors implement JSON marshaling for API responses:

```go
err := errors.New("validation failed", errors.CategoryValidation).
    WithCode(400).
    WithSeverity(errors.SeverityWarning).
    WithMetadata(map[string]any{"field": "email"})

data, _ := json.Marshal(err)
fmt.Println(string(data))
// Output: {"category":"validation","code":400,"message":"validation failed","metadata":{"field":"email"},"timestamp":"2023-01-01T12:00:00Z","location":{"file":"main.go","line":42,"function":"main.main"},"severity":"WARNING"}

// Create an error response for APIs
errResp := err.ToErrorResponse(false, nil) // excludes stack trace
responseData, _ := json.Marshal(errResp)

// Error collector JSON serialization
collector := errors.NewCollector()
collector.Add(err)
collector.AddValidation("name", "required")

collectorResponse := collector.ToErrorResponse(false)
// Returns aggregated error with metadata about collected errors
```

## Stack Traces

Capture stack traces for debugging:

```go
err := errors.New("system error", errors.CategoryInternal).WithStackTrace()

// Print error with stack trace
fmt.Println(err.ErrorWithStack())
```

## Retryable Errors

The package provides support for retryable errors with exponential backoff:

```go
// Create a retryable error
retryErr := errors.NewRetryable("service unavailable", errors.CategoryExternal).
    WithRetryDelay(2 * time.Second).
    WithCode(503)

// Create a retryable operation error with default 500ms delay
opErr := errors.NewRetryableOperation("operation failed")

// Create a retryable external service error
extErr := errors.NewRetryableExternal("external API error")

// Create a non-retryable error
nonRetryErr := errors.NewNonRetryable("invalid credentials", errors.CategoryAuth)

// Check if error is retryable
if errors.IsRetryableError(err) {
    // Calculate delay for attempt 3
    if retryableErr, ok := err.(*errors.RetryableError); ok {
        delay := retryableErr.RetryDelay(3) // exponential backoff
    }
}
```

## HTTP Integration

The package includes HTTP error mapping and response utilities:

```go
// Map HTTP errors to structured errors
httpErr := errors.MapHTTPErrors(err)

// Convert category to HTTP status code
status := errors.HTTPStatusToCategory(404) // returns CategoryNotFound

// Generate text code from HTTP status
textCode := errors.HTTPStatusToTextCode(404) // returns "NOT_FOUND"

// Use predefined HTTP status constants
err := errors.New("not found", errors.CategoryNotFound).
    WithCode(errors.CodeNotFound) // 404

// Map errors with custom mappers
mappedErr := errors.MapToError(err, errors.DefaultErrorMappers())
```

## Enhanced Logging Integration

Integrate with structured logging using slog with enhanced features:

```go
import "log/slog"

// Convert error to slog attributes (now includes severity and location)
attrs := errors.ToSlogAttributes(err)
slog.Error("operation failed", attrs...)

// Attributes include: error_code, text_code, category, severity, request_id,
// validation_errors, metadata, and location information

// NEW: Severity-based logging (automatically uses appropriate slog level)
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
errors.LogBySeverity(logger, err)
// Maps SeverityDebug->Debug, SeverityInfo->Info, SeverityWarning->Warn,
// SeverityError->Error, SeverityCritical->Error+Stack, SeverityFatal->Error+Stack

// NEW: Error collector logging
collector := errors.NewCollector()
collector.Add(errors.NewWarning("Warning 1", errors.CategoryOperation))
collector.Add(errors.NewCritical("Critical issue", errors.CategoryExternal))

// Log all collected errors with collector context
collector.LogErrors(logger)

// Get collector-specific attributes
collectorAttrs := collector.ToSlogAttributes()
logger.Info("Error summary", collectorAttrs...)
// Includes: error_count, category_stats, severity_stats, most_common_category,
// validation_error_count, retryable_error_count
```

## Validation Methods

The Error type provides additional validation helper methods:

```go
// Get all validation errors as a map (including nested)
validationMap := err.ValidationMap()
// Returns: map[string]string{"email": "required", "age": "must be positive"}

// Get all validation errors including wrapped errors
allErrors := err.AllValidationErrors()
// Returns: []FieldError with all validation errors in the chain

// Clone an error for modification
clonedErr := err.Clone()
clonedErr.WithMetadata(map[string]any{"new_field": "value"})
```

## Global Configuration

### Enhanced Configuration Options

```go
// Enable verbose error output (includes location in Error() string)
errors.Verbose = true

// Set development mode for detailed debugging
errors.IsDevelopment = true

// NEW: Location tracking configuration
errors.SetLocationCapture(true)   // Enable location capture (default)
errors.SetLocationCapture(false)  // Disable for production performance

// Check current location capture state
enabled := errors.IsLocationCaptureEnabled()

// Environment variable configuration
// Set GO_ERRORS_DISABLE_LOCATION=true to disable location capture
```

### Performance Considerations

The enhanced features are designed with performance in mind:

- **Location Capture**: ~400ns overhead per error (can be disabled in production)
- **Severity Levels**: Zero overhead - stored as integers
- **Error Collection**: ~90ns per Add operation with thread-safe mutex
- **Memory Usage**: ~1KB per 10 collected errors

```go
// Production configuration example
func init() {
    if os.Getenv("ENV") == "production" {
        errors.SetLocationCapture(false)  // Disable location for performance
        errors.Verbose = false            // Disable verbose output
    }
}
```

### Build Tag Support

For advanced production optimization, you can use build tags:

```bash
# Build with production optimizations (disables expensive features)
go build -tags prod

# Development build (all features enabled)
go build
```

## License

MIT License, see LICENSE file for details.
