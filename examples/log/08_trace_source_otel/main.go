package main

import (
	"context"

	"github.com/k6zma/k6kit/log"
)

func main() {
	l, err := log.New(log.Config{
		Level:             log.LevelDebug,
		Format:            log.FormatText,
		Color:             true,
		EnableSourceTrace: true,
	})
	if err != nil {
		panic(err)
	}

	ctx := log.WithOtelTraceContext(context.Background(), "0123456789abcdef0123456789abcdef", "0123456789abcdef")

	l.InfoCtx(ctx, "trace + source demo", log.String("operation", "checkout"))
}
