package main

import (
	"context"
	"fmt"

	"github.com/k6zma/k6kit/log"
)

func main() {
	base, err := log.New(log.Config{Level: log.LevelDebug, Format: log.FormatText, Color: true})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx = log.WithLogger(ctx, base)
	ctx = log.WithRequestID(ctx, "req-123")
	ctx = log.WithOtelTraceContext(ctx, "0123456789abcdef0123456789abcdef", "0123456789abcdef")

	if rid, ok := log.RequestID(ctx); ok {
		fmt.Println("request_id:", rid)
	}
	if otel, ok := log.OtelTraceFromContext(ctx); ok {
		fmt.Printf("otel_trace: %s/%s\n", otel.TraceID, otel.SpanID)
	}

	fromCtx := log.FromContext(ctx, base)
	fromCtx.InfoCtx(ctx, "context-aware log",
		log.String("event", "checkout"),
		log.String("method", "GET"),
		log.String("path", "/v1/orders"),
		log.String("tenant", "acme"),
	)
}
