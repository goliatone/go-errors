package errors

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityDebug, "DEBUG"},
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
		{SeverityFatal, "FATAL"},
		{Severity(99), "UNKNOWN"}, // Invalid severity
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverity_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"debug", SeverityDebug, `"DEBUG"`},
		{"info", SeverityInfo, `"INFO"`},
		{"warning", SeverityWarning, `"WARNING"`},
		{"error", SeverityError, `"ERROR"`},
		{"critical", SeverityCritical, `"CRITICAL"`},
		{"fatal", SeverityFatal, `"FATAL"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.severity)
			if err != nil {
				t.Errorf("Severity.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("Severity.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestSeverity_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Severity
		wantErr bool
	}{
		{"debug", `"DEBUG"`, SeverityDebug, false},
		{"info", `"INFO"`, SeverityInfo, false},
		{"warning", `"WARNING"`, SeverityWarning, false},
		{"error", `"ERROR"`, SeverityError, false},
		{"critical", `"CRITICAL"`, SeverityCritical, false},
		{"fatal", `"FATAL"`, SeverityFatal, false},
		{"invalid", `"INVALID"`, Severity(0), true},
		{"malformed", `invalid`, Severity(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Severity
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Severity.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Severity.UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewWithSeverity(t *testing.T) {
	// Test default severity
	err := New("test error", CategoryInternal)
	if err.Severity != SeverityError {
		t.Errorf("Expected default severity to be SeverityError, got %v", err.Severity)
	}

	// Test WithSeverity method
	err.WithSeverity(SeverityCritical)
	if err.Severity != SeverityCritical {
		t.Errorf("Expected severity to be SeverityCritical, got %v", err.Severity)
	}
}

func TestSeverityConstructors(t *testing.T) {
	tests := []struct {
		name        string
		constructor func(string, Category) *Error
		want        Severity
	}{
		{"NewDebug", NewDebug, SeverityDebug},
		{"NewInfo", NewInfo, SeverityInfo},
		{"NewWarning", NewWarning, SeverityWarning},
		{"NewCritical", NewCritical, SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test message", CategoryInternal)
			if err.Severity != tt.want {
				t.Errorf("%s() severity = %v, want %v", tt.name, err.Severity, tt.want)
			}
			if err.Message != "test message" {
				t.Errorf("%s() message = %v, want %v", tt.name, err.Message, "test message")
			}
		})
	}
}

func TestSeverityMethods(t *testing.T) {
	err := NewCritical("critical error", CategoryExternal)

	// Test GetSeverity
	if got := err.GetSeverity(); got != SeverityCritical {
		t.Errorf("GetSeverity() = %v, want %v", got, SeverityCritical)
	}

	// Test HasSeverity
	if !err.HasSeverity(SeverityCritical) {
		t.Error("HasSeverity(SeverityCritical) should return true")
	}
	if err.HasSeverity(SeverityWarning) {
		t.Error("HasSeverity(SeverityWarning) should return false")
	}

	// Test IsAboveSeverity
	if !err.IsAboveSeverity(SeverityError) {
		t.Error("IsAboveSeverity(SeverityError) should return true for Critical error")
	}
	if !err.IsAboveSeverity(SeverityCritical) {
		t.Error("IsAboveSeverity(SeverityCritical) should return true for Critical error")
	}
	if err.IsAboveSeverity(SeverityFatal) {
		t.Error("IsAboveSeverity(SeverityFatal) should return false for Critical error")
	}
}

func TestSeverityJSONSerialization(t *testing.T) {
	err := NewWarning("warning message", CategoryValidation)

	// Test JSON marshaling
	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("JSON marshal failed: %v", jsonErr)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"severity":"WARNING"`) {
		t.Error("JSON should contain severity field as string")
	}

	// Test JSON round-trip
	var unmarshaled Error
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if unmarshaled.Severity != SeverityWarning {
		t.Errorf("Unmarshaled severity = %v, want %v", unmarshaled.Severity, SeverityWarning)
	}
}

func TestRetryableErrorWithSeverity(t *testing.T) {
	// Test WithSeverity on RetryableError
	retryErr := NewRetryable("retryable error", CategoryOperation).
		WithSeverity(SeverityWarning)

	if retryErr.BaseError.Severity != SeverityWarning {
		t.Errorf("RetryableError severity = %v, want %v", retryErr.BaseError.Severity, SeverityWarning)
	}

	// Test that critical errors are not retryable
	criticalRetryErr := NewRetryable("critical error", CategoryExternal).
		WithSeverity(SeverityCritical)

	if criticalRetryErr.IsRetryable() {
		t.Error("Critical severity errors should not be retryable")
	}

	// Test that fatal errors are not retryable
	fatalRetryErr := NewRetryable("fatal error", CategoryExternal).
		WithSeverity(SeverityFatal)

	if fatalRetryErr.IsRetryable() {
		t.Error("Fatal severity errors should not be retryable")
	}

	// Test that warning errors are still retryable
	warningRetryErr := NewRetryable("warning error", CategoryExternal).
		WithSeverity(SeverityWarning)

	if !warningRetryErr.IsRetryable() {
		t.Error("Warning severity errors should be retryable")
	}
}

func TestSeverityLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	tests := []struct {
		name     string
		severity Severity
		wantLog  string
	}{
		{"debug", SeverityDebug, "level=DEBUG"},
		{"info", SeverityInfo, "level=INFO"},
		{"warning", SeverityWarning, "level=WARN"},
		{"error", SeverityError, "level=ERROR"},
		{"critical", SeverityCritical, "level=ERROR"},
		{"fatal", SeverityFatal, "level=ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := New("test error", CategoryInternal).WithSeverity(tt.severity)
			LogBySeverity(logger, err)

			logOutput := buf.String()
			if !strings.Contains(logOutput, tt.wantLog) {
				t.Errorf("Expected log output to contain %s, got: %s", tt.wantLog, logOutput)
			}

			// Check that severity is included in attributes
			if !strings.Contains(logOutput, "severity="+tt.severity.String()) {
				t.Errorf("Expected log output to contain severity=%s, got: %s", tt.severity.String(), logOutput)
			}
		})
	}
}

func TestSlogAttributesWithSeverity(t *testing.T) {
	err := NewCritical("critical error", CategoryExternal).
		WithCode(500).
		WithTextCode("CRITICAL_ERROR")

	attrs := ToSlogAttributes(err)

	// Check that severity is included
	var severityFound bool
	for _, attr := range attrs {
		if attr.Key == "severity" {
			severityFound = true
			if attr.Value.String() != "CRITICAL" {
				t.Errorf("Expected severity attribute to be CRITICAL, got %s", attr.Value.String())
			}
			break
		}
	}

	if !severityFound {
		t.Error("Severity attribute not found in slog attributes")
	}
}

func TestLogBySeverityNilHandling(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	// Test with nil logger
	LogBySeverity(nil, New("test", CategoryInternal))
	// Should not panic

	// Test with nil error
	LogBySeverity(logger, nil)
	// Should not panic

	// Ensure no output was generated
	if buf.Len() > 0 {
		t.Error("Expected no log output for nil inputs")
	}
}

func TestSeverityWithStackTrace(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	// Test that critical errors include stack trace
	criticalErr := NewCritical("critical error", CategoryInternal).WithStackTrace()
	LogBySeverity(logger, criticalErr)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Stack Trace:") {
		t.Error("Expected critical error log to contain stack trace")
	}

	buf.Reset()

	// Test that regular errors don't include stack trace
	regularErr := New("regular error", CategoryInternal)
	LogBySeverity(logger, regularErr)

	logOutput = buf.String()
	if strings.Contains(logOutput, "Stack Trace:") {
		t.Error("Expected regular error log to not contain stack trace")
	}
}

// Benchmark severity operations
func BenchmarkSeverityString(b *testing.B) {
	severity := SeverityCritical
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = severity.String()
	}
}

func BenchmarkSeverityMarshalJSON(b *testing.B) {
	severity := SeverityCritical
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(severity)
	}
}

func BenchmarkNewWithSeverity(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = New("benchmark error", CategoryInternal).WithSeverity(SeverityCritical)
	}
}

func BenchmarkSeverityConstructors(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewCritical("benchmark error", CategoryInternal)
	}
}
