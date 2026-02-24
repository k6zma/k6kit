package log

import (
	"context"
	"sync"
	"sync/atomic"
)

var (
	// defaultOnce ensures package level default k6kit logger is initialized once
	defaultOnce sync.Once

	// defaultLogger stores current package level default k6kit logger
	defaultLogger atomic.Value
)

// initDefault initializes package level default k6kit logger
func initDefault() {
	l, err := New(DefaultConfig())
	if err != nil {
		panic(err)
	}

	defaultLogger.Store(l)
}

// Default returns package level default k6kit logger
func Default() Logger {
	defaultOnce.Do(initDefault)

	if l, ok := defaultLogger.Load().(Logger); ok && l != nil {
		return l
	}

	initDefault()

	return defaultLogger.Load().(Logger)
}

// SetDefault replaces package level default k6kit logger
func SetDefault(l Logger) {
	if l == nil {
		return
	}

	defaultOnce.Do(initDefault)
	defaultLogger.Store(l)
}

// Enabled reports whether level is enabled by default k6kit logger
func Enabled(ctx context.Context, level Level) bool {
	return Default().Enabled(ctx, level)
}

// Debug logs a DEBUG message with the default k6kit logger
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

// Info logs an INFO message with the default k6kit logger
func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

// Warn logs a WARN message with the default k6kit logger
func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

// Error logs an ERROR message with the default k6kit logger
func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

// Fatal logs a FATAL message with the default k6kit logger
func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}

// Panic logs a PANIC message with the default k6kit logger
func Panic(msg string, fields ...Field) {
	Default().Panic(msg, fields...)
}

// Debugf logs a formatted DEBUG message with the default k6kit logger
func Debugf(format string, args ...any) {
	Default().Debugf(format, args...)
}

// Infof logs a formatted INFO message with the default k6kit logger
func Infof(format string, args ...any) {
	Default().Infof(format, args...)
}

// Warnf logs a formatted WARN message with the default k6kit logger
func Warnf(format string, args ...any) {
	Default().Warnf(format, args...)
}

// Errorf logs a formatted ERROR message with the default k6kit logger
func Errorf(format string, args ...any) {
	Default().Errorf(format, args...)
}

// Fatalf logs a formatted FATAL message with the default k6kit logger
func Fatalf(format string, args ...any) {
	Default().Fatalf(format, args...)
}

// Panicf logs a formatted PANIC message with the default k6kit logger
func Panicf(format string, args ...any) {
	Default().Panicf(format, args...)
}

// DebugCtx logs a context aware DEBUG message with the default k6kit logger
func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	Default().DebugCtx(ctx, msg, fields...)
}

// InfoCtx logs a context aware INFO message with the default k6kit logger
func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	Default().InfoCtx(ctx, msg, fields...)
}

// WarnCtx logs a context aware WARN message with the default k6kit logger
func WarnCtx(ctx context.Context, msg string, fields ...Field) {
	Default().WarnCtx(ctx, msg, fields...)
}

// ErrorCtx logs a context aware ERROR message with the default k6kit logger
func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	Default().ErrorCtx(ctx, msg, fields...)
}

// FatalCtx logs a context aware FATAL message with the default k6kit logger
func FatalCtx(ctx context.Context, msg string, fields ...Field) {
	Default().FatalCtx(ctx, msg, fields...)
}

// PanicCtx logs a context aware PANIC message with the default k6kit logger
func PanicCtx(ctx context.Context, msg string, fields ...Field) {
	Default().PanicCtx(ctx, msg, fields...)
}
