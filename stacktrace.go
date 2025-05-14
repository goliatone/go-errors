package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// StackFrame is a single frame in the stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// StackTrace is a collection of stack frames
type StackTrace []StackFrame

func (s StackTrace) String() string {
	var parts []string
	for _, frame := range s {
		parts = append(parts, fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))
	}
	return strings.Join(parts, "\n")
}

func CaptureStackTrace(skip int) StackTrace {
	var frames StackTrace
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip+2, pcs)

	for i := range n {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}
	return frames
}
