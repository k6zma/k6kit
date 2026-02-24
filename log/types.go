package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Format defines output renderer
type Format string

const (
	// FormatText renders records as bracketed human-readable lines
	FormatText Format = "text"

	// FormatJSON renders records as flat JSON objects
	FormatJSON Format = "json"
)

// Level defines k6kit logger severity
type Level int

const (
	// LevelDebug enables debug and above
	LevelDebug Level = iota - 1

	// LevelInfo enables info and above
	LevelInfo

	// LevelWarn enables warn and above
	LevelWarn

	// LevelError enables error and above
	LevelError

	// LevelFatal enables fatal and above
	LevelFatal

	// LevelPanic enables panic and above
	LevelPanic
)

// slogLevel maps log level to slog level scale
func (l Level) slogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError + 4
	case LevelPanic:
		return slog.LevelError + 8
	default:
		return slog.LevelInfo
	}
}

// Config configures k6kit logger behavior and output
type Config struct {
	// Level sets minimum enabled severity
	Level Level

	// Format selects the output renderer
	Format Format

	// Color enables ANSI colors in text mode
	Color bool

	// EnableSourceTrace adds source file/line/function data
	EnableSourceTrace bool

	// Environment is emitted as a static metadata field
	Environment string

	// AppName is emitted as a static metadata field
	AppName string

	// Version is emitted as a static metadata field
	Version string

	// Writer is the destination sink
	Writer io.Writer

	// TimeFormat overrides default timestamp layout
	TimeFormat string

	// ExitFunc is called by Fatal methods after write attempt
	ExitFunc func(code int)
}

// DefaultConfig returns baseline k6kit logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      LevelInfo,
		Format:     FormatText,
		Writer:     os.Stdout,
		TimeFormat: "",
		ExitFunc:   os.Exit,
	}
}

// merged applies config defaults and explicit overrides
func (c Config) merged() Config {
	b := DefaultConfig()

	b.Level = c.Level
	if c.Format != "" {
		b.Format = c.Format
	}

	b.Color = c.Color
	b.EnableSourceTrace = c.EnableSourceTrace
	b.Environment = c.Environment
	b.AppName = c.AppName
	b.Version = c.Version

	if c.Writer != nil {
		b.Writer = c.Writer
	}

	if c.TimeFormat != "" {
		b.TimeFormat = c.TimeFormat
	}

	if c.ExitFunc != nil {
		b.ExitFunc = c.ExitFunc
	}

	if b.TimeFormat == "" {
		if b.Format == FormatJSON {
			b.TimeFormat = defaultJSONTimeFormat
		} else {
			b.TimeFormat = defaultTextTimeFormat
		}
	}

	return b
}

// validate checks configuration for supported values
func (c Config) validate() error {
	switch c.Level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelPanic:
	default:
		return fmt.Errorf("unsupported log level: %d", c.Level)
	}

	switch c.Format {
	case FormatText, FormatJSON:
	default:
		return fmt.Errorf("unsupported log format: %q", c.Format)
	}

	if c.Writer == nil {
		return fmt.Errorf("writer cannot be nil")
	}

	if strings.TrimSpace(c.TimeFormat) == "" {
		return fmt.Errorf("time format cannot be empty")
	}

	return nil
}

// Logger is the public logging interface for k6kit logger
type Logger interface {
	// Debug logs a DEBUG message
	Debug(msg string, fields ...Field)

	// Info logs an INFO message
	Info(msg string, fields ...Field)

	// Warn logs a WARN message
	Warn(msg string, fields ...Field)

	// Error logs an ERROR message
	Error(msg string, fields ...Field)

	// Fatal logs a FATAL message and terminates via ExitFunc
	Fatal(msg string, fields ...Field)

	// Panic logs a PANIC message and then panics
	Panic(msg string, fields ...Field)

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

	// Panicf logs a formatted PANIC message and then panics
	Panicf(format string, args ...any)

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

	// PanicCtx logs a context-aware PANIC message and then panics
	PanicCtx(ctx context.Context, msg string, fields ...Field)

	// With returns a child k6kit logger with persistent fields
	With(fields ...Field) Logger

	// WithErr returns a child k6kit logger with canonical error field
	WithErr(err error) Logger

	// WithGroup returns a child k6kit logger with composed group path
	WithGroup(name string) Logger

	// WithRequestID returns a child k6kit logger with fixed request id
	WithRequestID(id string) Logger

	// WithOtelTrace returns a child k6kit logger with fixed trace and span ids
	WithOtelTrace(traceID, spanID string) Logger

	// Enabled reports whether level is enabled for context
	Enabled(ctx context.Context, level Level) bool
}
