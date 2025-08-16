package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorLocation represents the file, line, and function where an error was created
type ErrorLocation struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// String returns a formatted string representation of the location
func (loc *ErrorLocation) String() string {
	if loc == nil {
		return ""
	}

	// Extract just the filename from the full path for readability
	filename := loc.File
	if lastSlash := strings.LastIndex(filename, "/"); lastSlash >= 0 {
		filename = filename[lastSlash+1:]
	}

	return fmt.Sprintf("%s:%d", filename, loc.Line)
}

// captureLocation captures the current call stack location
// skip indicates how many stack frames to skip (0 = captureLocation itself)
func captureLocation(skip int) *ErrorLocation {
	// Check if location capture is enabled
	if !EnableLocationCapture {
		return nil
	}

	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return nil
	}

	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn != nil {
		funcName = fn.Name()
	}

	return &ErrorLocation{
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

// Here captures the location where this function is called
// This is a convenience function for explicit location capture
func Here() *ErrorLocation {
	return captureLocation(1) // Skip Here() itself
}
