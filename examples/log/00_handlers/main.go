package main

import (
	"fmt"

	"github.com/k6zma/k6kit/log"
)

func main() {
	examples := []struct {
		name string
		cfg  log.Config
	}{
		{
			name: "text_color",
			cfg: log.Config{
				Level:  log.LevelDebug,
				Format: log.FormatText,
				Color:  true,
			},
		},
		{
			name: "text_no_color",
			cfg: log.Config{
				Level:  log.LevelDebug,
				Format: log.FormatText,
				Color:  false,
			},
		},
		{
			name: "json",
			cfg: log.Config{
				Level:  log.LevelDebug,
				Format: log.FormatJSON,
			},
		},
	}

	for _, item := range examples {
		l, err := log.New(item.cfg)
		if err != nil {
			panic(err)
		}

		l.Info("handler demo", log.String("handler", item.name), log.Int("attempt", 1))

		fmt.Println("---")
	}
}
