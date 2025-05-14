# Go Errors

A simple error handling package that provides structured errors with rich context, validation support, and JSON serialization.

## Features

- **Structured Error Types**: Categorized errors with consistent structure across packages
- **Rich Context**: Metadata, stack traces, request IDs, and timestamps
- **Validation Support**: Built-in handling for field-level validation errors
- **JSON Serialization**: Full JSON marshaling/unmarshaling support
- **Error Wrapping**: Compatible with Go's `errors.Is` and `errors.As`
- **Fluent Interface**: Chainable methods for building complex errors
- **Code Support**: Both numeric and text codes for easy error identification
- **Domain Agnostic**: Core package has no HTTP or domain-specific dependencies

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
    // a simple error
    err := errors.New(errors.CategoryNotFound, "user not found")

    // Add context
    enrichedErr := err.
        WithMetadata(map[string]any{"user_id": 123}).
        WithRequestID("req-456").
        WithStackTrace().
        WithCode(404).
        WithTextCode("USER_NOT_FOUND")

    fmt.Println(enrichedErr.Error())

    // create a validation error
    validationErr := errors.NewValidation("invalid input",
        errors.FieldError{Field: "email", Message: "required"},
        errors.FieldError{Field: "age", Message: "must be positive"},
    )

    // wrapping an existing error
    wrappedErr := errors.Wrap(err, errors.CategoryInternal, "database query failed")
}
```

## Error Categories

The package provides a set of pre defined error categories:

- `CategoryValidation` - Input validation failures
- `CategoryAuth` - Authentication errors
- `CategoryAuthz` - Authorization errors
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

## Constructor Functions

### Basic Constructors

```go
err := errors.New(errors.CategoryNotFound, "resource not found")

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

## Fluent Interface

Chain methods to build rich error context:

```go
err := errors.New(errors.CategoryAuth, "authentication failed").
    WithCode(401).
    WithTextCode("AUTH_FAILED").
    WithMetadata(map[string]any{
        "user_id": 123,
        "attempt": 3,
    }).
    WithRequestID("req-789").
    WithStackTrace()
```

## Error Checking

The package provides utility functions to check an error's category:

```go
// Check error category
if errors.IsValidation(err) {
    // Handle validation error
}

// Check specific category
if errors.IsCategory(err, errors.CategoryNotFound) {
    // Handle not found
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
```

## JSON Serialization

Errors are fully JSON serializable:

```go
err := errors.New(errors.CategoryValidation, "validation failed").
    WithCode(400).
    WithMetadata(map[string]any{"field": "email"})

data, _ := json.Marshal(err)
fmt.Println(string(data))
// Output: {"category":"validation","code":400,"message":"validation failed","metadata":{"field":"email"},"timestamp":"2023-01-01T12:00:00Z"}
```

## Stack Traces

Capture stack traces for debugging:

```go
err := errors.New(errors.CategoryInternal, "system error").WithStackTrace()

// Print error with stack trace
fmt.Println(err.ErrorWithStack())
```

## License

MIT License, see LICENSE file for details.
