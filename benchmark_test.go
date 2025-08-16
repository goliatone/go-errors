package errors

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
)

// BenchmarkLocationCaptureOverhead measures the overhead of location capture
func BenchmarkLocationCaptureOverhead(b *testing.B) {
	b.Run("with_location_capture", func(b *testing.B) {
		EnableLocationCapture = true
		b.ResetTimer()
		for range b.N {
			New("benchmark error", CategoryInternal)
		}
	})

	b.Run("without_location_capture", func(b *testing.B) {
		EnableLocationCapture = false
		b.ResetTimer()
		for range b.N {
			New("benchmark error", CategoryInternal)
		}
		EnableLocationCapture = true // Restore
	})
}

// BenchmarkSeverityOperations measures severity-related performance
func BenchmarkSeverityOperations(b *testing.B) {
	err := New("test error", CategoryInternal).WithSeverity(SeverityWarning)

	b.Run("get_severity", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.GetSeverity()
		}
	})

	b.Run("has_severity", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.HasSeverity(SeverityWarning)
		}
	})

	b.Run("is_above_severity", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.IsAboveSeverity(SeverityInfo)
		}
	})

	b.Run("with_severity", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.WithSeverity(SeverityError)
		}
	})
}

// BenchmarkErrorConstruction measures different error creation patterns
func BenchmarkErrorConstruction(b *testing.B) {
	sourceErr := fmt.Errorf("source error")

	b.Run("new_basic", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			New("basic error")
		}
	})

	b.Run("new_with_category", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			New("error with category", CategoryValidation)
		}
	})

	b.Run("new_critical", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			NewCritical("critical error", CategoryInternal)
		}
	})

	b.Run("new_warning", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			NewWarning("warning", CategoryOperation)
		}
	})

	b.Run("wrap_error", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			Wrap(sourceErr, CategoryExternal, "wrapped error")
		}
	})

	b.Run("new_validation", func(b *testing.B) {
		fieldError := FieldError{Field: "email", Message: "invalid"}
		b.ResetTimer()
		for range b.N {
			NewValidation("validation failed", fieldError)
		}
	})

	b.Run("new_retryable", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			NewRetryable("retryable error", CategoryExternal)
		}
	})
}

// BenchmarkErrorMethods measures the performance of error methods
func BenchmarkErrorMethods(b *testing.B) {
	err := New("test error", CategoryValidation).
		WithSeverity(SeverityWarning).
		WithCode(400).
		WithTextCode("BAD_REQUEST").
		WithMetadata(map[string]any{"key": "value"})

	b.Run("error_string", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			_ = err.Error()
		}
	})

	b.Run("error_string_with_stack", func(b *testing.B) {
		errWithStack := err.WithStackTrace()
		b.ResetTimer()
		for range b.N {
			_ = errWithStack.ErrorWithStack()
		}
	})

	b.Run("clone", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.Clone()
		}
	})

	b.Run("marshal_json", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.MarshalJSON()
		}
	})

	b.Run("to_error_response", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.ToErrorResponse(false, nil)
		}
	})
}

// BenchmarkCollectorOperations measures ErrorCollector performance
func BenchmarkCollectorOperations(b *testing.B) {
	b.Run("collector_creation", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			NewCollector()
		}
	})

	b.Run("collector_add_single", func(b *testing.B) {
		c := NewCollector()
		err := New("benchmark error", CategoryInternal)
		b.ResetTimer()
		for range b.N {
			c.Add(err)
		}
	})

	b.Run("collector_add_different_errors", func(b *testing.B) {
		c := NewCollector()
		b.ResetTimer()
		for i := range b.N {
			err := New(fmt.Sprintf("error %d", i), CategoryInternal)
			c.Add(err)
		}
	})

	// Test with pre-populated collector
	c := NewCollector()
	for i := range 100 {
		c.Add(New(fmt.Sprintf("error %d", i), CategoryInternal))
	}

	b.Run("collector_count", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.Count()
		}
	})

	b.Run("collector_has_errors", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.HasErrors()
		}
	})

	b.Run("collector_errors_copy", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.Errors()
		}
	})

	b.Run("collector_merge", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.Merge()
		}
	})

	b.Run("collector_category_stats", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.CategoryStats()
		}
	})

	b.Run("collector_severity_distribution", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.SeverityDistribution()
		}
	})

	b.Run("collector_filter_by_severity", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.FilterBySeverity(SeverityWarning)
		}
	})

	b.Run("collector_filter_by_category", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.FilterByCategory(CategoryInternal)
		}
	})

	b.Run("collector_to_error_response", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.ToErrorResponse(false)
		}
	})
}

// BenchmarkCollectorConcurrency measures concurrent performance
func BenchmarkCollectorConcurrency(b *testing.B) {
	c := NewCollector(WithMaxErrors(10000))

	b.Run("concurrent_add", func(b *testing.B) {
		err := New("concurrent error", CategoryInternal)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				c.Add(err)
			}
		})
	})

	// Pre-populate for read operations
	for i := range 1000 {
		c.Add(New(fmt.Sprintf("error %d", i), CategoryInternal))
	}

	b.Run("concurrent_read_operations", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				switch i % 4 {
				case 0:
					c.Count()
				case 1:
					c.HasErrors()
				case 2:
					c.CategoryStats()
				case 3:
					c.SeverityDistribution()
				}
				i++
			}
		})
	})
}

// BenchmarkLoggingIntegration measures logging performance
func BenchmarkLoggingIntegration(b *testing.B) {
	err := New("log test error", CategoryValidation).
		WithSeverity(SeverityWarning).
		WithMetadata(map[string]any{"key": "value"})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Set to ERROR to reduce output during benchmarks
	}))

	b.Run("to_slog_attributes", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			ToSlogAttributes(err)
		}
	})

	b.Run("log_by_severity", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			LogBySeverity(logger, err)
		}
	})

	// Test collector logging
	c := NewCollector()
	for i := range 10 {
		c.Add(New(fmt.Sprintf("error %d", i), CategoryInternal))
	}

	b.Run("collector_to_slog_attributes", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.ToSlogAttributes()
		}
	})

	b.Run("collector_log_errors", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c.LogErrors(logger)
		}
	})
}

// BenchmarkValidationErrors measures validation error performance
func BenchmarkValidationErrors(b *testing.B) {
	fieldErrors := []FieldError{
		{Field: "email", Message: "invalid format"},
		{Field: "name", Message: "required"},
		{Field: "age", Message: "must be positive"},
	}

	b.Run("new_validation", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			NewValidation("validation failed", fieldErrors...)
		}
	})

	b.Run("validation_from_map", func(b *testing.B) {
		fieldMap := map[string]string{
			"email": "invalid format",
			"name":  "required",
			"age":   "must be positive",
		}
		b.ResetTimer()
		for range b.N {
			NewValidationFromMap("validation failed", fieldMap)
		}
	})

	err := NewValidation("validation test", fieldErrors...)

	b.Run("all_validation_errors", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.AllValidationErrors()
		}
	})

	b.Run("validation_map", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			err.ValidationMap()
		}
	})
}

// BenchmarkRetryableErrors measures retryable error performance
func BenchmarkRetryableErrors(b *testing.B) {
	sourceErr := fmt.Errorf("source error")

	b.Run("new_retryable", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			NewRetryable("retryable error", CategoryExternal)
		}
	})

	b.Run("wrap_retryable", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			WrapRetryable(sourceErr, CategoryExternal, "wrapped retryable")
		}
	})

	retryableErr := NewRetryable("test error", CategoryExternal)

	b.Run("is_retryable", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			retryableErr.IsRetryable()
		}
	})

	b.Run("retry_delay", func(b *testing.B) {
		b.ResetTimer()
		for i := range b.N {
			retryableErr.RetryDealy(i % 10) // Test different attempt numbers
		}
	})
}

// BenchmarkMemoryUsage measures memory allocations
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("error_creation_allocs", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			New("memory test error", CategoryInternal)
		}
	})

	b.Run("collector_add_allocs", func(b *testing.B) {
		c := NewCollector()
		err := New("allocation test", CategoryInternal)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			c.Add(err)
		}
	})

	b.Run("error_clone_allocs", func(b *testing.B) {
		err := New("clone test", CategoryInternal).
			WithMetadata(map[string]any{"key": "value"})
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			err.Clone()
		}
	})

	b.Run("json_marshal_allocs", func(b *testing.B) {
		err := New("json test", CategoryValidation).
			WithSeverity(SeverityWarning).
			WithMetadata(map[string]any{"data": "test"})
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			err.MarshalJSON()
		}
	})
}

// BenchmarkComparisonBaseline provides baseline comparisons
func BenchmarkComparisonBaseline(b *testing.B) {
	b.Run("stdlib_error_creation", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			_ = fmt.Errorf("standard library error")
		}
	})

	b.Run("stdlib_error_wrapping", func(b *testing.B) {
		sourceErr := fmt.Errorf("source")
		b.ResetTimer()
		for range b.N {
			_ = fmt.Errorf("wrapped: %w", sourceErr)
		}
	})

	b.Run("our_error_creation", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			New("our library error")
		}
	})

	b.Run("our_error_wrapping", func(b *testing.B) {
		sourceErr := fmt.Errorf("source")
		b.ResetTimer()
		for range b.N {
			Wrap(sourceErr, CategoryInternal, "wrapped error")
		}
	})
}

// BenchmarkScaleTest measures performance with large numbers of errors
func BenchmarkScaleTest(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("collector_%d_errors", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				c := NewCollector()
				for i := range size {
					c.Add(New(fmt.Sprintf("error %d", i), CategoryInternal))
				}
				c.Merge() // Test merge performance with different sizes
			}
		})
	}

	// Test filtering performance with different scales
	for _, size := range sizes {
		b.Run(fmt.Sprintf("filter_%d_errors", size), func(b *testing.B) {
			c := NewCollector()
			for i := range size {
				severity := []Severity{SeverityDebug, SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}[i%5]
				c.Add(New(fmt.Sprintf("error %d", i), CategoryInternal).WithSeverity(severity))
			}

			b.ResetTimer()
			for range b.N {
				c.FilterBySeverity(SeverityWarning)
			}
		})
	}
}

// BenchmarkFeatureComparison compares performance with and without features
func BenchmarkFeatureComparison(b *testing.B) {
	b.Run("minimal_error", func(b *testing.B) {
		EnableLocationCapture = false
		b.ResetTimer()
		for range b.N {
			New("minimal error")
		}
		EnableLocationCapture = true
	})

	b.Run("full_featured_error", func(b *testing.B) {
		EnableLocationCapture = true
		b.ResetTimer()
		for range b.N {
			New("full error", CategoryValidation).
				WithSeverity(SeverityWarning).
				WithCode(400).
				WithTextCode("BAD_REQUEST").
				WithMetadata(map[string]any{"key": "value"}).
				WithStackTrace()
		}
	})

	// Compare collector performance with different configurations
	b.Run("basic_collector", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c := NewCollector()
			for i := range 10 {
				c.Add(New(fmt.Sprintf("error %d", i)))
			}
		}
	})

	b.Run("configured_collector", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			c := NewCollector(
				WithMaxErrors(100),
				WithStrictMode(true),
			)
			for i := range 10 {
				c.Add(New(fmt.Sprintf("error %d", i), CategoryValidation).
					WithSeverity(SeverityWarning))
			}
		}
	})
}