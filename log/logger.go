package log

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// writeErrorCount tracks internal handler write failures for observability
var writeErrorCount atomic.Uint64

// logger is the concrete Logger implementation over slog
type logger struct {
	// slog is the underlying slog logger used for record dispatch
	slog *slog.Logger

	// now provides timestamps
	now func() time.Time

	// source enables caller PC capture for source trace extraction
	source bool

	// exitFn is invoked by Fatal-level methods
	exitFn func(code int)
}

var _ Logger = (*logger)(nil)

// New constructs a logger from Config
func New(cfg Config) (Logger, error) {
	var (
		globalAttrs []slog.Attr
		base        *handler
	)

	merged := defaultConfig()
	applyConfig(&merged, cfg)

	if err := validateConfig(merged); err != nil {
		return nil, err
	}

	w := resolveWriter(merged)

	if merged.AppName != "" {
		globalAttrs = append(globalAttrs, slog.String(keyApp, merged.AppName))
	}

	if merged.Env != "" {
		globalAttrs = append(globalAttrs, slog.String(keyEnv, merged.Env))
	}

	if merged.Version != "" {
		globalAttrs = append(globalAttrs, slog.String(keyVersion, merged.Version))
	}

	if merged.TimeFormat == "" {
		merged.TimeFormat = defaultTimeFormat(merged.Format)
	}

	handlerOpts := handlerOptions{
		Level:         merged.Level.slogLevel(),
		IncludeSource: merged.EnableSourceTrace,
		Now:           merged.now,
		StaticAttrs:   globalAttrs,
		OTELTrace:     otelTraceExtractor(merged.EnableOTEL),
	}

	switch merged.Format {
	case FormatJSON:
		base = newJSONHandler(w, handlerOpts, merged.TimeFormat)
	default:
		base = newTextHandler(w, handlerOpts, merged.Color, merged.TimeFormat)
	}

	return &logger{
		slog:   slog.New(base),
		now:    merged.now,
		source: merged.EnableSourceTrace,
		exitFn: merged.ExitFunc,
	}, nil
}

// validateConfig validates userfacing configuration values
func validateConfig(cfg Config) error {
	switch cfg.Level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal:
	default:
		return fmt.Errorf("unsupported log level: %d", cfg.Level)
	}

	switch cfg.Format {
	case FormatText, FormatJSON:
	default:
		return fmt.Errorf("unsupported log format: %q", cfg.Format)
	}

	return nil
}

// applyConfig merges custom config into base defaults
func applyConfig(base *Config, custom Config) {
	base.Level = custom.Level
	if custom.Format != "" {
		base.Format = custom.Format
	}

	base.Color = custom.Color
	base.EnableSourceTrace = custom.EnableSourceTrace

	base.EnableOTEL = custom.EnableOTEL
	if custom.Env != "" {
		base.Env = custom.Env
	}

	if custom.AppName != "" {
		base.AppName = custom.AppName
	}

	if custom.Version != "" {
		base.Version = custom.Version
	}

	if custom.Writer != nil {
		base.Writer = custom.Writer
	}

	if custom.TimeFormat != "" {
		base.TimeFormat = custom.TimeFormat
	}

	if custom.ExitFunc != nil {
		base.ExitFunc = custom.ExitFunc
	}

	if custom.now != nil {
		base.now = custom.now
	}
}

// resolveWriter resolves the output sink from config
func resolveWriter(cfg Config) io.Writer {
	if cfg.Writer != nil {
		return cfg.Writer
	}

	return os.Stdout
}

// otelTraceExtractor builds OTEL context extraction function when enabled
func otelTraceExtractor(enabled bool) otelTraceExtractorFunc {
	if !enabled {
		return nil
	}

	return func(ctx context.Context, traceOut *[32]byte, spanOut *[16]byte) (int, int) {
		sc := trace.SpanContextFromContext(ctx)
		if !sc.IsValid() {
			return 0, 0
		}

		traceID := sc.TraceID()
		spanID := sc.SpanID()

		hex.Encode(traceOut[:], traceID[:])
		hex.Encode(spanOut[:], spanID[:])

		return len(traceOut), len(spanOut)
	}
}

// Enabled implements Logger.Enabled
func (l *logger) Enabled(ctx context.Context, level Level) bool {
	return l.slog.Enabled(normalizeContext(ctx), level.slogLevel())
}

// Debug implements Logger.Debug
func (l *logger) Debug(msg string, fields ...Field) {
	l.log(context.Background(), LevelDebug, msg, fields...)
}

// Info implements Logger.Info
func (l *logger) Info(msg string, fields ...Field) {
	l.log(context.Background(), LevelInfo, msg, fields...)
}

// Warn implements Logger.Warn
func (l *logger) Warn(msg string, fields ...Field) {
	l.log(context.Background(), LevelWarn, msg, fields...)
}

// Error implements Logger.Error
func (l *logger) Error(msg string, fields ...Field) {
	l.log(context.Background(), LevelError, msg, fields...)
}

// Fatal implements Logger.Fatal
func (l *logger) Fatal(msg string, fields ...Field) {
	l.log(context.Background(), LevelFatal, msg, fields...)
}

// Debugf implements Logger.Debugf
func (l *logger) Debugf(format string, args ...any) {
	l.logf(context.Background(), LevelDebug, format, args...)
}

// Infof implements Logger.Infof
func (l *logger) Infof(format string, args ...any) {
	l.logf(context.Background(), LevelInfo, format, args...)
}

// Warnf implements Logger.Warnf
func (l *logger) Warnf(format string, args ...any) {
	l.logf(context.Background(), LevelWarn, format, args...)
}

// Errorf implements Logger.Errorf
func (l *logger) Errorf(format string, args ...any) {
	l.logf(context.Background(), LevelError, format, args...)
}

// Fatalf implements Logger.Fatalf
func (l *logger) Fatalf(format string, args ...any) {
	l.logf(context.Background(), LevelFatal, format, args...)
}

// DebugCtx implements Logger.DebugCtx
func (l *logger) DebugCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelDebug, msg, fields...)
}

// InfoCtx implements Logger.InfoCtx
func (l *logger) InfoCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelInfo, msg, fields...)
}

// WarnCtx implements Logger.WarnCtx
func (l *logger) WarnCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelWarn, msg, fields...)
}

// ErrorCtx implements Logger.ErrorCtx
func (l *logger) ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelError, msg, fields...)
}

// FatalCtx implements Logger.FatalCtx
func (l *logger) FatalCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelFatal, msg, fields...)
}

// With implements Logger.With
func (l *logger) With(fields ...Field) Logger {
	attrs := make([]any, 0, len(fields))
	for _, f := range fields {
		attrs = append(attrs, f.toAttr())
	}

	return &logger{
		slog:   l.slog.With(attrs...),
		now:    l.now,
		source: l.source,
		exitFn: l.exitFn,
	}
}

// WithErr implements Logger.WithErr
func (l *logger) WithErr(err error) Logger {
	return l.With(Error(keyError, err))
}

// WithGroup implements Logger.WithGroup
func (l *logger) WithGroup(name string) Logger {
	return &logger{
		slog:   l.slog.WithGroup(name),
		now:    l.now,
		source: l.source,
		exitFn: l.exitFn,
	}
}

// logf is the internal formatted logging path with level gating
func (l *logger) logf(ctx context.Context, level Level, format string, args ...any) {
	ctx = normalizeContext(ctx)
	if !l.slog.Enabled(ctx, level.slogLevel()) {
		return
	}

	l.logChecked(ctx, level, fmt.Sprintf(format, args...))
}

// log is the internal structured logging path with level gating
func (l *logger) log(ctx context.Context, level Level, msg string, fields ...Field) {
	ctx = normalizeContext(ctx)
	if !l.slog.Enabled(ctx, level.slogLevel()) {
		return
	}

	l.logChecked(ctx, level, msg, fields...)
}

// logChecked writes a prepared record when level has already been checked
func (l *logger) logChecked(ctx context.Context, level Level, msg string, fields ...Field) {
	pc := uintptr(0)
	if l.source {
		pc = callerPC()
	}

	rec := slog.NewRecord(l.now(), level.slogLevel(), msg, pc)
	for i := range fields {
		rec.AddAttrs(fields[i].toAttr())
	}

	l.emitRecord(ctx, level, rec)
}

// emitRecord dispatches a record to slog handler and applies fatal-exit policy
func (l *logger) emitRecord(ctx context.Context, level Level, rec slog.Record) {
	if err := l.slog.Handler().Handle(ctx, rec); err != nil {
		writeErrorCount.Add(1)

		_ = err
	}

	if level == LevelFatal {
		l.exitFn(1)
	}
}

// callerPC captures the call-site program counter for source tracing
func callerPC() uintptr {
	var pcs [16]uintptr

	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return 0
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !isInternalSourceFrame(frame.Function) {
			return frame.PC
		}

		if !more {
			break
		}
	}

	return pcs[0]
}

// isInternalSourceFrame reports whether frame belongs to logging internals
func isInternalSourceFrame(fn string) bool {
	if fn == "" {
		return true
	}

	if strings.HasPrefix(fn, "runtime.") || strings.HasPrefix(fn, "log/slog.") {
		return true
	}

	if strings.Contains(fn, "github.com/k6zma/k6kit/log.(*logger).") {
		return true
	}

	if strings.HasSuffix(fn, ".callerPC") || strings.HasSuffix(fn, ".isInternalSourceFrame") {
		return true
	}

	return false
}
