package errors

import (
	"encoding/json"
	"fmt"
)

// Severity represents the severity level of an error
type Severity int

const (
	SeverityDebug Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
	SeverityCritical
	SeverityFatal
)

// String representations for severity levels
var severityStrings = map[Severity]string{
	SeverityDebug:    "DEBUG",
	SeverityInfo:     "INFO",
	SeverityWarning:  "WARNING",
	SeverityError:    "ERROR",
	SeverityCritical: "CRITICAL",
	SeverityFatal:    "FATAL",
}

// String returns the string representation of the severity level
func (s Severity) String() string {
	if str, ok := severityStrings[s]; ok {
		return str
	}
	return "UNKNOWN"
}

// MarshalJSON implements JSON marshaling for Severity
func (s Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements JSON unmarshaling for Severity
func (s *Severity) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	for sev, name := range severityStrings {
		if name == str {
			*s = sev
			return nil
		}
	}

	return fmt.Errorf("unknown severity: %s", str)
}
