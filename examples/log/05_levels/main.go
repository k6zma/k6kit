package main

import (
	"context"
	"fmt"

	"github.com/k6zma/k6kit/log"
)

func main() {
	l, err := log.New(log.Config{
		Level:  log.LevelInfo,
		Format: log.FormatText,
		Color:  true,
		ExitFunc: func(int) {
		},
	})
	if err != nil {
		panic(err)
	}

	selected := log.LevelWarn

	fmt.Println("selected level: WARN", selected)
	fmt.Println("debug enabled:", l.Enabled(context.Background(), log.LevelDebug))
	fmt.Println("info enabled:", l.Enabled(context.Background(), log.LevelInfo))

	l.Debug("this is filtered")
	l.Info("info message")
	l.Warnf("warnf code=%d", 7)
	l.ErrorCtx(context.Background(), "error with context", log.String("module", "levels"))

	l.Fatal("fatal without process exit", log.String("reason", "demo"))
}
