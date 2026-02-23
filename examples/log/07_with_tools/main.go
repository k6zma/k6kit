package main

import (
	"context"
	"errors"

	"github.com/k6zma/k6kit/log"
)

func main() {
	base, err := log.New(log.Config{Level: log.LevelDebug, Format: log.FormatText, Color: true})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx = log.WithLogger(ctx, base)
	ctx = log.WithRequestID(ctx, "req-with-777")
	ctx = log.WithRequestMetadata(ctx, log.String("region", "eu-west"))

	fromCtx := log.FromContext(ctx, base)
	child := fromCtx.With(log.String("component", "worker")).WithGroup("jobs").WithErr(errors.New("retryable"))
	child.InfoCtx(ctx, "with-tools example", log.String("job_id", "job-42"))
}
