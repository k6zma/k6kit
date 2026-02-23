package main

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/k6zma/k6kit/log"
)

func main() {
	l, err := log.New(log.Config{
		Level:             log.LevelDebug,
		Format:            log.FormatText,
		Color:             true,
		EnableSourceTrace: true,
		EnableOTEL:        true,
	})
	if err != nil {
		panic(err)
	}

	traceID, err := trace.TraceIDFromHex("0123456789abcdef0123456789abcdef")
	if err != nil {
		panic(err)
	}

	spanID, err := trace.SpanIDFromHex("0123456789abcdef")
	if err != nil {
		panic(err)
	}

	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     false,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	l.InfoCtx(ctx, "trace + source demo", log.String("operation", "checkout"))
}
