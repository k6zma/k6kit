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
	ctx = log.WithRequestMetadata(ctx,
		log.String("method", "GET"),
		log.String("path", "/v1/orders"),
	)

	ctx = log.WithRequestMetadata(ctx, log.String("tenant", "acme"))
	if rid, ok := log.RequestID(ctx); ok {
		fmt.Println("request_id:", rid)
	}
	fmt.Println("request_metadata_count:", len(log.RequestMetadata(ctx)))

	fromCtx := log.FromContext(ctx, base)
	fromCtx.InfoCtx(ctx, "context-aware log", log.String("event", "checkout"))
}
