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
