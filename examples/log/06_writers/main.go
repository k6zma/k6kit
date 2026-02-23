package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/k6zma/k6kit/log"
)

func main() {
	var buffer bytes.Buffer

	file, err := os.CreateTemp("", "k6kit-logger-*.log")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()

	l, err := log.New(log.Config{
		Level:  log.LevelDebug,
		Format: log.FormatText,
		Color:  true,
		Writer: io.MultiWriter(os.Stdout, &buffer, file),
	})
	if err != nil {
		panic(err)
	}

	l.Info("writer fan-out", log.String("file", file.Name()))

	fileBytes, err := os.ReadFile(file.Name())
	if err != nil {
		panic(err)
	}

	fmt.Printf("buffer_bytes=%d file_bytes=%d\n", buffer.Len(), len(fileBytes))
}
