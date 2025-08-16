package errors

import "os"

// LocationConfig holds configuration for location capture behavior
type LocationConfig struct {
	// EnableLocationCapture controls whether error locations are captured
	EnableLocationCapture bool
}

// Global configuration for location capture
var (
	// EnableLocationCapture controls whether location information is captured for new errors
	// This can be disabled in production for performance optimization
	EnableLocationCapture = true
)

// init initializes configuration from environment variables
func init() {
	// Allow disabling location capture via environment variable
	if os.Getenv("GO_ERRORS_DISABLE_LOCATION") == "true" {
		EnableLocationCapture = false
	}
}

// SetLocationCapture globally enables or disables location capture
func SetLocationCapture(enabled bool) {
	EnableLocationCapture = enabled
}

// IsLocationCaptureEnabled returns whether location capture is currently enabled
func IsLocationCaptureEnabled() bool {
	return EnableLocationCapture
}
