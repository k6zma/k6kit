package log

import "context"

// context keys isolate package values stored in context
type (
	loggerCtxKey          struct{}
	requestIDCtxKey       struct{}
	requestMetadataCtxKey struct{}
)

// normalizeContext converts nil context to context.Background
func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}

// WithLogger stores logger in context and ignores nil logger
func WithLogger(ctx context.Context, l Logger) context.Context {
	ctx = normalizeContext(ctx)
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, loggerCtxKey{}, l)
}

// FromContext returns logger from context or fallback when missing
func FromContext(ctx context.Context, fallback Logger) Logger {
	if ctx == nil {
		return fallback
	}

	v, _ := ctx.Value(loggerCtxKey{}).(Logger)
	if v != nil {
		return v
	}

	return fallback
}

// WithRequestID stores request_id in context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	ctx = normalizeContext(ctx)

	return context.WithValue(ctx, requestIDCtxKey{}, requestID)
}

// RequestID returns request_id from context when present and non empty
func RequestID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	v, ok := ctx.Value(requestIDCtxKey{}).(string)
	if !ok || v == "" {
		return "", false
	}

	return v, true
}

// WithRequestMetadata appends request-scoped metadata fields in context
func WithRequestMetadata(ctx context.Context, fields ...Field) context.Context {
	ctx = normalizeContext(ctx)
	if len(fields) == 0 {
		return ctx
	}

	existing := RequestMetadata(ctx)
	out := make([]Field, 0, len(existing)+len(fields))

	out = append(out, existing...)
	out = append(out, fields...)

	return context.WithValue(ctx, requestMetadataCtxKey{}, out)
}

// RequestMetadata returns request-scoped metadata copied from context
func RequestMetadata(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}

	v := requestMetadataUnsafe(ctx)
	if len(v) == 0 {
		return nil
	}

	out := make([]Field, len(v))
	copy(out, v)

	return out
}

// requestMetadataUnsafe returns request metadata from context without copying
func requestMetadataUnsafe(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}

	v, _ := ctx.Value(requestMetadataCtxKey{}).([]Field)

	return v
}
