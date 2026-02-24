package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestLoggerMethodsAndWithHelpers(t *testing.T) {
	var out bytes.Buffer

	exitCode := -1

	l, err := New(Config{
		Level:             LevelDebug,
		Format:            FormatJSON,
		EnableSourceTrace: true,
		Writer:            &out,
		TimeFormat:        defaultJSONTimeFormat,
		ExitFunc: func(code int) {
			exitCode = code
		},
	})
	require.NoError(t, err)

	ctx := WithRequestID(context.Background(), "req-1")
	ctx = WithOtelTraceContext(ctx, "trace-1", "span-1")

	l.With(String("scope", "child")).WithGroup("api").WithErr(nil).InfoCtx(ctx, "hello", String("service", "x"))
	l.Debugf("debug-%d", 1)
	l.Fatal("fatal")

	assert.Equal(t, 1, exitCode)

	lines := splitNonEmptyLines(out.String())
	require.Len(t, lines, 3)

	info := parseJSONLine(t, lines[0])
	assert.Equal(t, "req-1", info[keyRequestID])
	assert.Equal(t, "trace-1", info[keyTraceID])
	assert.Equal(t, "span-1", info[keySpanID])
	assert.Equal(t, "api", info[keyGroup])
	_, ok := info[keySourceTrace].(string)
	assert.True(t, ok)
}

func TestLoggerPanicVariants(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{Level: LevelDebug, Format: FormatJSON, Writer: &out, TimeFormat: defaultJSONTimeFormat})
	require.NoError(t, err)

	cases := []panicVariantCase{
		{name: "panic", run: func() { l.Panic("panic-msg") }},
		{name: "panicf", run: func() { l.Panicf("panicf-%d", 1) }},
		{name: "panicctx", run: func() { l.PanicCtx(context.Background(), "panicctx") }},
	}

	for i, tc := range cases {
		t.Run(testCaseName(testPrefixLogger, tc.name, i), func(t *testing.T) {
			assert.Panics(t, tc.run)
		})
	}

	assert.Len(t, splitNonEmptyLines(out.String()), 3)
}

func TestEnabledAndOtelFromSpanContext(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{Level: LevelInfo, Format: FormatJSON, Writer: &out, TimeFormat: defaultJSONTimeFormat})
	require.NoError(t, err)

	assert.False(t, l.Enabled(context.Background(), LevelDebug))

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		SpanID:     trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	l.InfoCtx(ctx, "otel-from-span")

	lines := splitNonEmptyLines(out.String())
	require.NotEmpty(t, lines)
	obj := parseJSONLine(t, lines[0])
	assert.NotEmpty(t, obj[keyTraceID])
	assert.NotEmpty(t, obj[keySpanID])
	msg, ok := obj[keyMsg].(string)
	require.True(t, ok)
	assert.Equal(t, "otel-from-span", strings.TrimSpace(msg))
}

func TestLoggerBoundContextOverridesContextValues(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{Level: LevelInfo, Format: FormatJSON, Writer: &out, TimeFormat: defaultJSONTimeFormat})
	require.NoError(t, err)

	ctx := WithRequestID(context.Background(), "req-ctx")
	ctx = WithOtelTraceContext(ctx, "trace-ctx", "span-ctx")

	l.WithRequestID("req-bound").WithOtelTrace("trace-bound", "span-bound").InfoCtx(ctx, "bound-overrides")

	lines := splitNonEmptyLines(out.String())
	require.Len(t, lines, 1)

	obj := parseJSONLine(t, lines[0])
	assert.Equal(t, "req-bound", obj[keyRequestID])
	assert.Equal(t, "trace-bound", obj[keyTraceID])
	assert.Equal(t, "span-bound", obj[keySpanID])
}
