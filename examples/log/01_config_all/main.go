package main

import (
	"bytes"
	"io"
	"os"

	"github.com/k6zma/k6kit/log"
)

func main() {
	var buffer bytes.Buffer

	l, err := log.New(log.Config{
		Level:             log.LevelDebug,
		Format:            log.FormatText,
		Color:             true,
		EnableSourceTrace: true,
		EnableOTEL:        true,
		Env:               "local",
		AppName:           "logger-example",
		Version:           "1.0.0",
		Writer:            io.MultiWriter(os.Stdout, &buffer),
		TimeFormat:        "15:04:05",
	})
	if err != nil {
		panic(err)
	}

	l.Debug("all config knobs set", log.Bool("demo", true))
	l.Infof("captured-by-writer-bytes=%d", buffer.Len())
}
