package log

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type (
	// ctxLoggerKey stores k6kit logger instance in context
	ctxLoggerKey struct{}

	// ctxRequestIDKey stores request id in context
	ctxRequestIDKey struct{}

	// ctxOtelKey stores explicit OTEL trace/span in context
	ctxOtelKey struct{}
)

// OtelTrace carries OTEL trace and span identifiers
type OtelTrace struct {
	TraceID string
	SpanID  string
}

// WithLogger stores k6kit logger in context
func WithLogger(ctx context.Context, l Logger) context.Context {
	ctx = normalizeContext(ctx)
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, ctxLoggerKey{}, l)
}

// FromContext returns context k6kit logger or fallback
func FromContext(ctx context.Context, fallback Logger) Logger {
	if ctx == nil {
		return fallback
	}

	l, _ := ctx.Value(ctxLoggerKey{}).(Logger)
	if l != nil {
		return l
	}

	return fallback
}

// WithRequestID stores request id in context
func WithRequestID(ctx context.Context, id string) context.Context {
	ctx = normalizeContext(ctx)

	return context.WithValue(ctx, ctxRequestIDKey{}, id)
}

// RequestID returns request id from context when present
func RequestID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	id, ok := ctx.Value(ctxRequestIDKey{}).(string)
	if !ok || id == "" {
		return "", false
	}

	return id, true
}

// WithOtelTraceContext stores explicit OTEL trace/span into context
func WithOtelTraceContext(ctx context.Context, traceID, spanID string) context.Context {
	ctx = normalizeContext(ctx)

	return context.WithValue(ctx, ctxOtelKey{}, OtelTrace{TraceID: traceID, SpanID: spanID})
}

// OtelTraceFromContext extracts OTEL trace/span from context
func OtelTraceFromContext(ctx context.Context) (OtelTrace, bool) {
	if ctx == nil {
		return OtelTrace{}, false
	}

	if v, ok := ctx.Value(ctxOtelKey{}).(OtelTrace); ok && (v.TraceID != "" || v.SpanID != "") {
		return v, true
	}

	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return OtelTrace{}, false
	}

	return OtelTrace{TraceID: sc.TraceID().String(), SpanID: sc.SpanID().String()}, true
}
