package log

import "context"

// Logger is the public logging interface exposed by this package
type Logger interface {
	// Enabled reports whether level is enabled for the provided context
	Enabled(ctx context.Context, level Level) bool

	// Debug logs a DEBUG message
	Debug(msg string, fields ...Field)

	// Info logs an INFO message
	Info(msg string, fields ...Field)

	// Warn logs a WARN message
	Warn(msg string, fields ...Field)

	// Error logs an ERROR message
	Error(msg string, fields ...Field)

	// Fatal logs a FATAL message and terminates the process via ExitFunc
	Fatal(msg string, fields ...Field)

	// Debugf logs a formatted DEBUG message
	Debugf(format string, args ...any)

	// Infof logs a formatted INFO message
	Infof(format string, args ...any)

	// Warnf logs a formatted WARN message
	Warnf(format string, args ...any)

	// Errorf logs a formatted ERROR message
	Errorf(format string, args ...any)

	// Fatalf logs a formatted FATAL message and terminates via ExitFunc
	Fatalf(format string, args ...any)

	// DebugCtx logs a context-aware DEBUG message
	DebugCtx(ctx context.Context, msg string, fields ...Field)

	// InfoCtx logs a context-aware INFO message
	InfoCtx(ctx context.Context, msg string, fields ...Field)

	// WarnCtx logs a context-aware WARN message
	WarnCtx(ctx context.Context, msg string, fields ...Field)

	// ErrorCtx logs a context-aware ERROR message
	ErrorCtx(ctx context.Context, msg string, fields ...Field)

	// FatalCtx logs a context-aware FATAL message and terminates via ExitFunc
	FatalCtx(ctx context.Context, msg string, fields ...Field)

	// With returns a child logger enriched with additional fields
	With(fields ...Field) Logger

	// WithErr returns a child logger enriched with a canonical error field
	WithErr(err error) Logger

	// WithGroup returns a child logger that sets the record group section
	WithGroup(name string) Logger
}

// renderer serializes normalized records into byte slices
type renderer interface {
	// Append serializes a normalized record into dst
	Append(dst []byte, rec record) ([]byte, error)
}
