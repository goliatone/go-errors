package errors

import (
	"log/slog"
)

func ToSlogAttributes(err error) []slog.Attr {
	var richErr *Error
	if As(err, &richErr) {
		var attrs []slog.Attr
		if richErr.Code != 0 {
			attrs = append(attrs, slog.Int("error_code", richErr.Code))
		}

		if richErr.TextCode != "" {
			attrs = append(attrs, slog.String("text_code", richErr.TextCode))
		}

		if richErr.Category != "" {
			attrs = append(attrs, slog.String("category", richErr.Category.String()))
		}

		// Add severity information
		attrs = append(attrs, slog.String("severity", richErr.Severity.String()))

		if richErr.RequestID != "" {
			attrs = append(attrs, slog.String("request_id", richErr.RequestID))
		}

		if len(richErr.AllValidationErrors()) > 0 {
			attrs = append(attrs, slog.Any("validation_errors", richErr.AllValidationErrors()))
		}

		if len(richErr.Metadata) > 0 {
			attrs = append(attrs, slog.Any("metadata", richErr.Metadata))
		}
		return attrs
	}
	return nil
}

// LogBySeverity logs an error using the appropriate slog level based on its severity
func LogBySeverity(logger *slog.Logger, err *Error) {
	if logger == nil || err == nil {
		return
	}

	attrs := ToSlogAttributes(err)
	// Convert []slog.Attr to []any for logging methods
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}

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
		logger.Error(err.ErrorWithStack(), anyAttrs...) // Include stack trace for critical errors
	case SeverityFatal:
		logger.Error(err.ErrorWithStack(), anyAttrs...) // Include stack trace for fatal errors
	default:
		logger.Error(err.Error(), anyAttrs...)
	}
}
