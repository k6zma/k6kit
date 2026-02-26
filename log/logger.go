package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

// slogLogger is the concrete k6kit logger implementation backed by slog
type slogLogger struct {
	// s is underlying slog logger
	s *slog.Logger

	// exitFn is called by Fatal methods
	exitFn func(int)

	// request stores persistent request id bound via WithRequestID
	request string

	// otel stores persistent trace/span bound via WithOtelTrace
	otel OtelTrace

	// srcTrace toggles source trace capture
	srcTrace bool
}

// New creates a new k6kit logger from config
func New(cfg Config) (Logger, error) {
	merged := cfg.merged()
	if err := merged.validate(); err != nil {
		return nil, err
	}

	h := newHandler(merged)

	return &slogLogger{s: slog.New(h), exitFn: merged.ExitFunc, srcTrace: merged.EnableSourceTrace}, nil
}

// Enabled reports whether level is enabled for context
func (l *slogLogger) Enabled(ctx context.Context, level Level) bool {
	return l.s.Enabled(normalizeContext(ctx), level.slogLevel())
}

// Debug logs a DEBUG message
func (l *slogLogger) Debug(msg string, fields ...Field) {
	l.log(context.Background(), LevelDebug, msg, fields...)
}

// Info logs an INFO message
func (l *slogLogger) Info(msg string, fields ...Field) {
	l.log(context.Background(), LevelInfo, msg, fields...)
}

// Warn logs a WARN message
func (l *slogLogger) Warn(msg string, fields ...Field) {
	l.log(context.Background(), LevelWarn, msg, fields...)
}

// Error logs an ERROR message
func (l *slogLogger) Error(msg string, fields ...Field) {
	l.log(context.Background(), LevelError, msg, fields...)
}

// Fatal logs a FATAL message and terminates via ExitFunc
func (l *slogLogger) Fatal(msg string, fields ...Field) {
	l.log(context.Background(), LevelFatal, msg, fields...)
}

// Panic logs a PANIC message and then panics
func (l *slogLogger) Panic(msg string, fields ...Field) {
	l.log(context.Background(), LevelPanic, msg, fields...)
}

// Debugf logs a formatted DEBUG message
func (l *slogLogger) Debugf(format string, args ...any) {
	l.logf(context.Background(), LevelDebug, format, args...)
}

// Infof logs a formatted INFO message
func (l *slogLogger) Infof(format string, args ...any) {
	l.logf(context.Background(), LevelInfo, format, args...)
}

// Warnf logs a formatted WARN message
func (l *slogLogger) Warnf(format string, args ...any) {
	l.logf(context.Background(), LevelWarn, format, args...)
}

// Errorf logs a formatted ERROR messag.
func (l *slogLogger) Errorf(format string, args ...any) {
	l.logf(context.Background(), LevelError, format, args...)
}

// Fatalf logs a formatted FATAL message and terminates via ExitFunc
func (l *slogLogger) Fatalf(format string, args ...any) {
	l.logf(context.Background(), LevelFatal, format, args...)
}

// Panicf logs a formatted PANIC message and then panics
func (l *slogLogger) Panicf(format string, args ...any) {
	l.logf(context.Background(), LevelPanic, format, args...)
}

// DebugCtx logs a context aware DEBUG message
func (l *slogLogger) DebugCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelDebug, msg, fields...)
}

// InfoCtx logs a context aware INFO message
func (l *slogLogger) InfoCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelInfo, msg, fields...)
}

// WarnCtx logs a context aware WARN message
func (l *slogLogger) WarnCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelWarn, msg, fields...)
}

// ErrorCtx logs a context aware ERROR message
func (l *slogLogger) ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelError, msg, fields...)
}

// FatalCtx logs a context aware FATAL message and terminates via ExitFunc
func (l *slogLogger) FatalCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelFatal, msg, fields...)
}

// PanicCtx logs a context aware PANIC message and then panics
func (l *slogLogger) PanicCtx(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelPanic, msg, fields...)
}

// With returns child k6kit logger with persistent fields.
func (l *slogLogger) With(fields ...Field) Logger {
	args := make([]any, 0, len(fields))
	for _, f := range fields {
		args = append(args, f)
	}

	return &slogLogger{s: l.s.With(args...), exitFn: l.exitFn, request: l.request, otel: l.otel, srcTrace: l.srcTrace}
}

// WithErr returns child k6kit logger enriched with canonical error field
func (l *slogLogger) WithErr(err error) Logger {
	if err == nil {
		return l.With(String(keyError, ""))
	}

	return l.With(String(keyError, err.Error()))
}

// WithGroup returns child k6kit logger with composed group path
func (l *slogLogger) WithGroup(name string) Logger {
	return &slogLogger{s: l.s.WithGroup(name), exitFn: l.exitFn, request: l.request, otel: l.otel, srcTrace: l.srcTrace}
}

// WithRequestID binds request id to child k6kit logger context
func (l *slogLogger) WithRequestID(id string) Logger {
	next := *l
	next.request = id

	return &next
}

// WithOtelTrace binds trace and span ids to child k6kit logger context
func (l *slogLogger) WithOtelTrace(traceID, spanID string) Logger {
	next := *l
	next.otel = OtelTrace{TraceID: traceID, SpanID: spanID}

	return &next
}

// logf logs formatted message through structured logging pipeline
func (l *slogLogger) logf(ctx context.Context, level Level, format string, args ...any) {
	l.log(ctx, level, fmt.Sprintf(format, args...))
}

// log logs one message using normalized context and attrs
func (l *slogLogger) log(ctx context.Context, level Level, msg string, fields ...Field) {
	ctx = normalizeContext(ctx)

	sl := level.slogLevel()
	if !l.s.Enabled(ctx, sl) {
		return
	}

	pc := uintptr(0)

	if l.srcTrace {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}

	rec := slog.NewRecord(time.Now(), sl, msg, pc)

	err := func() error {
		if h, ok := l.s.Handler().(*pipelineHandler); ok {
			return h.handleWithBound(ctx, rec, l.request, l.otel, fields)
		}

		ctx = l.bindContextValues(ctx)

		if len(fields) > 0 {
			attrs := make([]slog.Attr, len(fields))

			copy(attrs, fields)

			rec.AddAttrs(attrs...)
		}

		return l.s.Handler().Handle(ctx, rec)
	}()
	if err != nil {
		_ = err
	}

	if level == LevelFatal {
		l.exitFn(1)
	}

	if level == LevelPanic {
		panic(msg)
	}
}

func (l *slogLogger) bindContextValues(ctx context.Context) context.Context {
	if l.request != "" {
		if current, ok := ctx.Value(ctxRequestIDKey{}).(string); !ok || current != l.request {
			ctx = context.WithValue(ctx, ctxRequestIDKey{}, l.request)
		}
	}

	if l.otel.TraceID != "" || l.otel.SpanID != "" {
		if current, ok := ctx.Value(ctxOtelKey{}).(OtelTrace); !ok || current != l.otel {
			ctx = context.WithValue(ctx, ctxOtelKey{}, l.otel)
		}
	}

	return ctx
}
