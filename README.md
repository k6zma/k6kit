# k6kit
<div align="center">
    <img src="https://img.shields.io/badge/status-active-success.svg"/>
    <img src="https://img.shields.io/github/issues/k6zma/k6kit.svg"/>
    <img src="https://img.shields.io/github/issues-pr/k6zma/k6kit.svg"/>
</div>

<p></p>

<div align="center">
    <img src="https://img.shields.io/badge/Go-black?style=for-the-badge&logo=go&logoColor=#00ADD8"/>
    <img src="https://img.shields.io/badge/OpenTelemetry-black?style=for-the-badge&logo=opentelemetry&logoColor=#00ADD8"/>
</div>

---

## About

`k6kit` is a Go toolkit repository focused on reusable backend components

At the moment, the production module in this repository is `k6kit/log`, a structured logger built on top of `slog`

---

## Installation

Install in your module:

```bash
go get github.com/k6zma/k6kit
```

---

## Project layout

The project tree is structured very simply. Each individual utility has its own root folder, which contains the implementation

```md
├── .github - description of CI/CD pipelines for GitHub Actions
│   └── workflows
│       └── ci.yml - main CI pipeline
├── .golangci.yml - golangci linter config
├── README.md - basic description of the project
├── docs - detailed documentation for each utility
│   └── log
│       └── LOG.md - detailed documentation for the logger
├── examples - examples of using utilities
│   └── log - runnable logger examples by capability
├── go.mod
├── go.sum
└── log - folder with the custom logger implementation
```

---

## Implemented utilities

Currently, the following utilities have been implemented:
- `k6kit/log` - A custom logger built on top of slog with a rich API for use

---

## Detailed information about the utility

### Logger

`k6kit/log` is a context first structured logger built on top of `slog`
It is designed for production use and keeps a clean API focused on practical logging workflows

Key capabilities:
- Text and JSON output formats
- Optional colored text output
- Context helpers (`WithLogger`, `FromContext`, `WithRequestID`, `WithRequestMetadata`)
- Child logger tools (`With`, `WithErr`, `WithGroup`)
- Optional source trace and OpenTelemetry trace/span extraction
- Deterministic output rules for fields and key collisions

Core API surface:
- Level methods: `Debug`, `Info`, `Warn`, `Error`, `Fatal`
- Formatted methods: `Debugf`, `Infof`, `Warnf`, `Errorf`, `Fatalf`
- Context methods: `DebugCtx`, `InfoCtx`, `WarnCtx`, `ErrorCtx`, `FatalCtx`

Config options (`log.Config`):
- `Level` (`Debug`, `Info`, `Warn`, `Error`, `Fatal`)
- `Format` (`text`, `json`)
- `Color` (for text mode)
- `EnableSourceTrace`
- `EnableOTEL`
- `Env`, `AppName`, `Version`
- `Writer` (`io.Writer`, stdout by default)
- `TimeFormat`
- `ExitFunc` (used by `Fatal*`)

Quick start:

```go
package main

import "github.com/k6zma/k6kit/log"

func main() {
	l, err := log.New(
		log.Config{
			Level:   log.LevelInfo,
			Format:  log.FormatJSON,
			AppName: "k6kit",
			Version: "0.1.0",
		},
	)
	if err != nil {
		panic(err)
	}

	l.Info("logger ready", log.String("component", "example"))
}
```

Context example:

```go
ctx := context.Background()
ctx = log.WithRequestID(ctx, "req-123")
ctx = log.WithRequestMetadata(ctx, log.String("route", "GET /health"))

l.InfoCtx(ctx, "request accepted")
```

Also you can run logger examples:

```bash
go run ./examples/log/00_handlers
go run ./examples/log/01_config_all
go run ./examples/log/02_fields_all
go run ./examples/log/03_context_all
go run ./examples/log/04_child_loggers
go run ./examples/log/05_levels
go run ./examples/log/06_writers
go run ./examples/log/07_with_tools
go run ./examples/log/08_trace_source_otel
```
