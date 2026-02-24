package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestContextLoggerAndRequestID(t *testing.T) {
	t.Run(testCaseName(testPrefixContext, "logger-and-request-id", 0), func(t *testing.T) {
		fallback, err := New(Config{Level: LevelInfo, Format: FormatJSON, Writer: &discardWriter{}})
		require.NoError(t, err)

		ctx := WithLogger(context.TODO(), fallback)
		require.NotNil(t, FromContext(ctx, nil))

		ctx = WithRequestID(ctx, "req-42")
		id, ok := RequestID(ctx)
		assert.True(t, ok)
		assert.Equal(t, "req-42", id)

		id, ok = RequestID(context.Background())
		assert.False(t, ok)
		assert.Equal(t, "", id)
	})
}

func TestOtelTraceFromContextPrecedence(t *testing.T) {
	t.Run(testCaseName(testPrefixContext, "otel-trace-precedence", 1), func(t *testing.T) {
		ctx := context.Background()
		ctx = WithOtelTraceContext(ctx, "trace-explicit", "span-explicit")

		ot, ok := OtelTraceFromContext(ctx)
		require.True(t, ok)
		assert.Equal(t, "trace-explicit", ot.TraceID)
		assert.Equal(t, "span-explicit", ot.SpanID)

		sc := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
			SpanID:     trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18},
			TraceFlags: trace.FlagsSampled,
		})

		ctx2 := trace.ContextWithSpanContext(context.Background(), sc)

		ot, ok = OtelTraceFromContext(ctx2)
		require.True(t, ok)
		assert.NotEmpty(t, ot.TraceID)
		assert.NotEmpty(t, ot.SpanID)
	})
}
