package main

import (
	"errors"

	"github.com/k6zma/k6kit/log"
)

func main() {
	root, err := log.New(log.Config{Level: log.LevelDebug, Format: log.FormatText, Color: true})
	if err != nil {
		panic(err)
	}

	child := root.With(log.String("service", "billing"))
	grouped := child.WithGroup("http")
	withErr := grouped.WithErr(errors.New("payment provider timeout"))

	root.Info("root logger")
	child.Info("child with field")
	grouped.Info("grouped child", log.Int("status", 502))
	withErr.Error("child with canonical error")
}
