package errors

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestErrorLocation_String(t *testing.T) {
	tests := []struct {
		name     string
		location *ErrorLocation
		want     string
	}{
		{
			name:     "nil location",
			location: nil,
			want:     "",
		},
		{
			name: "normal location",
			location: &ErrorLocation{
				File:     "/path/to/file.go",
				Line:     42,
				Function: "main.TestFunction",
			},
			want: "file.go:42",
		},
		{
			name: "location with no path separator",
			location: &ErrorLocation{
				File:     "file.go",
				Line:     10,
				Function: "test",
			},
			want: "file.go:10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.location.String(); got != tt.want {
				t.Errorf("ErrorLocation.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaptureLocation(t *testing.T) {
	// Ensure location capture is enabled for this test
	originalEnabled := EnableLocationCapture
	EnableLocationCapture = true
	defer func() {
		EnableLocationCapture = originalEnabled
	}()

	loc := captureLocation(0)
	if loc == nil {
		t.Fatal("captureLocation(0) returned nil")
	}

	// Check that we captured this file
	if !strings.Contains(loc.File, "location_test.go") {
		t.Errorf("Expected file to contain 'location_test.go', got %s", loc.File)
	}

	// Check that we have a reasonable line number
	if loc.Line <= 0 {
		t.Errorf("Expected positive line number, got %d", loc.Line)
	}

	// Check that we have a function name
	if loc.Function == "" {
		t.Error("Expected non-empty function name")
	}
}

func TestCaptureLocationDisabled(t *testing.T) {
	// Disable location capture
	originalEnabled := EnableLocationCapture
	EnableLocationCapture = false
	defer func() {
		EnableLocationCapture = originalEnabled
	}()

	loc := captureLocation(0)
	if loc != nil {
		t.Errorf("Expected nil when location capture disabled, got %+v", loc)
	}
}

func TestHere(t *testing.T) {
	// Ensure location capture is enabled for this test
	originalEnabled := EnableLocationCapture
	EnableLocationCapture = true
	defer func() {
		EnableLocationCapture = originalEnabled
	}()

	loc := Here()
	if loc == nil {
		t.Fatal("Here() returned nil")
	}

	// Check that we captured this file
	if !strings.Contains(loc.File, "location_test.go") {
		t.Errorf("Expected file to contain 'location_test.go', got %s", loc.File)
	}

	// The line should be around where Here() was called
	if loc.Line <= 0 {
		t.Errorf("Expected positive line number, got %d", loc.Line)
	}
}

func TestNewWithLocation(t *testing.T) {
	location := &ErrorLocation{
		File:     "test.go",
		Line:     100,
		Function: "TestFunction",
	}

	err := New("test error", CategoryInternal)
	if err.Location == nil {
		t.Error("Expected New() to capture location automatically")
	}

	// Test explicit location setting
	err2 := NewWithLocation("explicit location error", CategoryValidation, location)
	if err2.Location != location {
		t.Error("Expected NewWithLocation to use provided location")
	}

	if err2.Category != CategoryValidation {
		t.Errorf("Expected category %v, got %v", CategoryValidation, err2.Category)
	}
}

func TestErrorWithLocation(t *testing.T) {
	// Ensure location capture is enabled for this test
	originalEnabled := EnableLocationCapture
	EnableLocationCapture = true
	defer func() {
		EnableLocationCapture = originalEnabled
	}()

	err := New("test error", CategoryInternal)
	if err.Location == nil {
		t.Fatal("Expected error to have location")
	}

	// Test that location appears in error string when Verbose is true
	originalVerbose := Verbose
	Verbose = true
	defer func() {
		Verbose = originalVerbose
	}()

	errStr := err.Error()
	if !strings.Contains(errStr, "location:") {
		t.Errorf("Expected error string to contain location info, got: %s", errStr)
	}

	// Test that location doesn't appear when Verbose is false
	Verbose = false
	errStr = err.Error()
	if strings.Contains(errStr, "location:") {
		t.Errorf("Expected error string to not contain location info when Verbose=false, got: %s", errStr)
	}
}

func TestErrorLocationMethods(t *testing.T) {
	location := &ErrorLocation{
		File:     "test.go",
		Line:     50,
		Function: "TestMethod",
	}

	err := NewWithLocation("test", CategoryInternal, location)

	// Test WithLocation
	newLocation := &ErrorLocation{
		File:     "new.go",
		Line:     75,
		Function: "NewMethod",
	}
	err.WithLocation(newLocation)
	if err.GetLocation() != newLocation {
		t.Error("WithLocation did not update location")
	}

	// Test GetLocation
	if got := err.GetLocation(); got != newLocation {
		t.Errorf("GetLocation() = %+v, want %+v", got, newLocation)
	}

	// Test HasLocation
	if !err.HasLocation() {
		t.Error("HasLocation() should return true when location is set")
	}

	// Test with nil location
	err.WithLocation(nil)
	if err.HasLocation() {
		t.Error("HasLocation() should return false when location is nil")
	}

	if err.GetLocation() != nil {
		t.Error("GetLocation() should return nil when no location is set")
	}
}

func TestWrapWithLocation(t *testing.T) {
	// Ensure location capture is enabled for this test
	originalEnabled := EnableLocationCapture
	EnableLocationCapture = true
	defer func() {
		EnableLocationCapture = originalEnabled
	}()

	// Test wrapping a standard error
	baseErr := New("base error", CategoryInternal)
	wrappedErr := Wrap(baseErr, CategoryExternal, "wrapped")

	// When wrapping an existing Error, it should preserve the original location
	if wrappedErr.Location != baseErr.Location {
		t.Error("Wrap should preserve original location when wrapping Error")
	}

	// Test wrapping a non-Error
	stdErr := &standardError{msg: "standard error"}
	wrappedStd := Wrap(stdErr, CategoryExternal, "wrapped standard")

	// When wrapping non-Error, it should capture new location
	if wrappedStd.Location == nil {
		t.Error("Wrap should capture location when wrapping non-Error")
	}
}

// Helper type for testing
type standardError struct {
	msg string
}

func (e *standardError) Error() string {
	return e.msg
}

func TestLocationJSONSerialization(t *testing.T) {
	location := &ErrorLocation{
		File:     "/path/to/test.go",
		Line:     123,
		Function: "main.TestFunction",
	}

	err := NewWithLocation("test error", CategoryInternal, location)

	// Test JSON marshaling
	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("JSON marshal failed: %v", jsonErr)
	}

	// Check that location is included in JSON
	jsonStr := string(data)
	if !strings.Contains(jsonStr, "location") {
		t.Error("JSON should contain location field")
	}
	if !strings.Contains(jsonStr, "test.go") {
		t.Error("JSON should contain file name")
	}
	if !strings.Contains(jsonStr, "123") {
		t.Error("JSON should contain line number")
	}

	// Test with nil location
	err2 := NewWithLocation("test error 2", CategoryInternal, nil)
	data2, jsonErr2 := json.Marshal(err2)
	if jsonErr2 != nil {
		t.Fatalf("JSON marshal failed for nil location: %v", jsonErr2)
	}

	// Location field should be omitted when nil (due to omitempty tag)
	jsonStr2 := string(data2)
	if strings.Contains(jsonStr2, "location") {
		t.Error("JSON should not contain location field when nil")
	}
}

func TestConfigurationFromEnvironment(t *testing.T) {
	// Save original environment
	originalValue := os.Getenv("GO_ERRORS_DISABLE_LOCATION")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("GO_ERRORS_DISABLE_LOCATION")
		} else {
			os.Setenv("GO_ERRORS_DISABLE_LOCATION", originalValue)
		}
	}()

	// Test with environment variable set to disable
	os.Setenv("GO_ERRORS_DISABLE_LOCATION", "true")

	// We can't easily test the init() function, but we can test the behavior
	// by temporarily setting the global variable
	originalEnabled := EnableLocationCapture
	EnableLocationCapture = false
	defer func() {
		EnableLocationCapture = originalEnabled
	}()

	err := New("test", CategoryInternal)
	if err.Location != nil {
		t.Error("Expected no location when capture is disabled")
	}
}

func TestSetLocationCapture(t *testing.T) {
	// Save original state
	originalEnabled := EnableLocationCapture
	defer func() {
		EnableLocationCapture = originalEnabled
	}()

	// Test enabling
	SetLocationCapture(true)
	if !IsLocationCaptureEnabled() {
		t.Error("Expected location capture to be enabled")
	}

	err1 := New("test1", CategoryInternal)
	if err1.Location == nil {
		t.Error("Expected location when capture is enabled")
	}

	// Test disabling
	SetLocationCapture(false)
	if IsLocationCaptureEnabled() {
		t.Error("Expected location capture to be disabled")
	}

	err2 := New("test2", CategoryInternal)
	if err2.Location != nil {
		t.Error("Expected no location when capture is disabled")
	}
}

// Benchmark location capture overhead
func BenchmarkLocationCapture(b *testing.B) {
	EnableLocationCapture = true
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = captureLocation(0)
	}
}

func BenchmarkLocationCaptureDisabled(b *testing.B) {
	EnableLocationCapture = false
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = captureLocation(0)
	}
}

func BenchmarkNewWithLocation(b *testing.B) {
	EnableLocationCapture = true
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = New("benchmark error", CategoryInternal)
	}
}

func BenchmarkNewWithoutLocation(b *testing.B) {
	EnableLocationCapture = false
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = New("benchmark error", CategoryInternal)
	}
}
