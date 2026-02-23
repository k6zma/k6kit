# k6kit/log

## What this logger is

`k6kit/log` is a structured logger built on top of Go `log/slog` with a stable API for application code and predictable output for operators

Design goals:

- Keep call sites small and explicit (`Info`, `Infof`, `InfoCtx`, typed `Field` helpers)
- Emit stable, parseable records in either text or JSON
- Preserve deterministic field ordering and duplicate-key behavior
- Support request metadata, optional source trace, and optional OTEL trace/span extraction
- Stay safe under concurrent use by many goroutines

## Import and quick start

Import:

```go
import "github.com/k6zma/k6kit/log"
```

Quick start:

```go
package main

import "github.com/k6zma/k6kit/log"

func main() {
    l, err := log.New(log.Config{
        Level:  log.LevelInfo,
        Format: log.FormatJSON,
    })
    if err != nil {
        panic(err)
    }
    l.Info("service started", log.String("service", "api"), log.Int("port", 8080))
}
```

## Logger API overview

Interface: `log.Logger`

Levels:

- `Debug(msg, fields...)`
- `Info(msg, fields...)`
- `Warn(msg, fields...)`
- `Error(msg, fields...)`
- `Fatal(msg, fields...)` (writes record, then calls `ExitFunc(1)`)

Formatted variants:

- `Debugf(format, args...)`
- `Infof(format, args...)`
- `Warnf(format, args...)`
- `Errorf(format, args...)`
- `Fatalf(format, args...)`

Context-aware variants:

- `DebugCtx(ctx, msg, fields...)`
- `InfoCtx(ctx, msg, fields...)`
- `WarnCtx(ctx, msg, fields...)`
- `ErrorCtx(ctx, msg, fields...)`
- `FatalCtx(ctx, msg, fields...)`

Other API:

- `Enabled(ctx, level)` checks level gating
- `With(fields...)` returns a child logger with persistent fields
- `WithErr(err)` shorthand for `With(log.Error("error", err))`
- `WithGroup(name)` returns a child logger with a group label

## Configuration reference

Constructor: `log.New(log.Config)`

Validation rules:

- `Level` must be one of `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError`, `LevelFatal`
- `Format` must be `FormatText` or `FormatJSON`
- Invalid values return error from `New`

Defaults (`log.DefaultConfig()`):

| Field | Default | Semantics |
|---|---|---|
| `Level` | `LevelInfo` | Minimum enabled level |
| `Format` | `FormatText` | Output renderer |
| `Color` | `false` | ANSI coloring for text mode only. |
| `EnableSourceTrace` | `false` | Adds `source_trace` (JSON) or `[source]` block (text) |
| `EnableOTEL` | `false` | Extracts `trace_id` and `span_id` from context span context |
| `Env` | `""` | When non-empty, global `env` field on every record |
| `AppName` | `""` | When non-empty, global `app` field on every record |
| `Version` | `""` | When non-empty, global `version` field on every record |
| `Writer` | `os.Stdout` | Destination `io.Writer` when nil in config |
| `TimeFormat` | format-specific | Uses text default `2006-01-02 15:04:05` or JSON default `2006-01-02T15:04:05` if empty |
| `ExitFunc` | `os.Exit` | Called by fatal methods with code `1` after write attempt |

Notes:

- `Color` has no effect in JSON format
- `TimeFormat` only changes top-level record time field formatting
- `Env`, `AppName`, `Version` are injected as static attrs in this order: `app`, `env`, `version`

## Levels behavior and ParseLevel

Supported level constants:

- `LevelDebug`
- `LevelInfo`
- `LevelWarn`
- `LevelError`
- `LevelFatal`

Behavior:

- Logger drops records below configured minimum level
- `Enabled(ctx, level)` matches that level gate
- Fatal methods log at `FATAL` and then call `ExitFunc(1)`

`ParseLevel` behavior:

- Accepts case-insensitive tokens: `debug`, `info`, `warn`, `error`, `fatal`
- Trims surrounding whitespace before parsing
- Rejects aliases such as `warning`
- Returns `unsupported log level: "<input>"` for unknown values

## Output formats and field semantics

### Text format (`FormatText`)

Shape:

```text
[time] [trace_id=... span_id=...] [source_trace] [LEVEL] [group] msg {k=v},{k2=v2}
```

Semantics:

- `time` always first, formatted by text `TimeFormat`
- Optional OTEL block appears only when enabled and IDs exist in context
- Optional source block appears only when `EnableSourceTrace=true`
- `group` is shown in its own bracket section
- Attributes are flattened and rendered as `{key=value}` pairs in deterministic order

Example:

```text
[2026-02-23 12:34:56] [INFO] [api.http] request handled {request_id=req-1},{status=200}
```

### JSON format (`FormatJSON`)

Canonical top-level keys order:

1. `time`
2. `trace_id` (optional)
3. `span_id` (optional)
4. `source_trace` (optional)
5. `level`
6. `group` (optional)
7. `msg`
8. flattened attrs in merge order

Semantics:

- Single flat object; nested groups are flattened with dot notation
- If an attribute key conflicts with canonical keys (`time`, `trace_id`, `span_id`, `source_trace`, `level`, `group`, `msg`), it is emitted as `attr.<key>`
- JSON record ends with `\n`

Example:

```json
{"time":"2026-02-23T12:34:56","level":"INFO","group":"api","msg":"request handled","request_id":"req-1","status":200}
```

## Output destination (`io.Writer`) and ownership

- If `Config.Writer` is nil, logger writes to `os.Stdout`
- If `Config.Writer` is set, logger writes to that writer directly
- Logger does not close or manage lifecycle of external writers
- Logger has no `Close` API; writer lifecycle is always owned by the caller
- Use `io.MultiWriter(...)` when fan-out is needed

## Fields reference

`Field` is a thin wrapper around typed `slog.Attr` construction

Scalar helpers:

- `Rune`, `Byte`
- `Int`, `Int8`, `Int16`, `Int32`, `Int64`
- `Uint8`, `Uint16`, `Uint32`, `Uint64`
- `Float32`, `Float64`
- `Bool`, `String`
- `Duration`, `Time`
- `Any`

Slice helpers:

- `Bytes`, `Strings`, `Runes`, `Bools`
- `Ints`, `Int8s`, `Int16s`, `Int32s`, `Int64s`
- `Uint8s`, `Uint16s`, `Uint32s`, `Uint64s`
- `Float32s`, `Float64s`
- `Errors` (`[]error`)

Error/group helpers:

- `Error(key, err)`
- `Errors(key, []error)`
- `Group(name, fields...)`

Group flattening rule:

- Nested `Group` fields are flattened with `.` key paths (`group.inner.key`)

## Child loggers (`With`, `WithErr`, `WithGroup`)

`With(fields...)`:

- Returns a child logger inheriting parent config and adding persistent fields
- Parent logger remains unchanged
- Child fields are merged before per-call fields

`WithErr(err)`:

- Equivalent to `With(log.Error("error", err))`
- Uses canonical key `error` by default

`WithGroup(name)`:

- Returns a child logger with record group metadata
- Group is a separate record field (`group`) and text section (`[group]`), not a key prefix for attrs
- Nested `WithGroup` calls compose with dot notation (`parent.child`)

## Context integration

Helpers:

- `WithLogger(ctx, logger)` stores logger in context (nil logger is ignored)
- `FromContext(ctx, fallback)` returns context logger, else fallback
- `WithRequestID(ctx, id)` stores request id
- `RequestID(ctx)` returns `(id, true)` only for non-empty ids
- `WithRequestMetadata(ctx, fields...)` appends metadata fields
- `RequestMetadata(ctx)` returns a defensive copy

Logging behavior:

- Context-derived fields (`request_id`, request metadata, OTEL fields) are only considered for `*Ctx` methods
- Non-ctx methods (`Info`, `Warn`, etc.) log with `context.Background()`

Example:

```go
ctx := context.Background()
ctx = log.WithRequestID(ctx, "req-123")
ctx = log.WithRequestMetadata(ctx, log.String("tenant", "acme"))

l.InfoCtx(ctx, "checkout", log.Int("items", 2))
```

## Tracing (`EnableSourceTrace`, `EnableOTEL`)

`EnableSourceTrace`:

- Captures caller program counter and resolves `file:line function`
- Uses short file name (base name), not full path
- Omitted when disabled or if source cannot be resolved

`EnableOTEL`:

- Extracts `TraceID` and `SpanID` from `trace.SpanContextFromContext(ctx)`
- Emits lowercase hex strings in `trace_id` and `span_id`
- Omitted when disabled or when span context is invalid

Example:

```go
l, _ := log.New(log.Config{Format: log.FormatJSON, EnableSourceTrace: true, EnableOTEL: true})
ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)
l.InfoCtx(ctx, "operation")
```

## Error serialization behavior

- `error` values are serialized as message-only strings (`err.Error()`)
- `nil` error serializes as empty string
- `[]error` values serialize as `[]string`
- This applies to `Error`, `Errors`, and `Any` if the underlying value type is `error` or `[]error`

Examples:

- `log.Error("error", errors.New("boom"))` -> `"error":"boom"` (JSON) / `{error=boom}` (text)
- `log.Errors("errors", []error{errors.New("e1"), nil})` -> `"errors":["e1",""]` (JSON)

## Ordering and dedup rules

Attribute merge order is deterministic:

1. Static config attrs (`app`, `env`, `version` when set)
2. Child logger attrs from `With(...)`
3. Context request id (`request_id`)
4. Context request metadata (`WithRequestMetadata` append order)
5. Per-call attrs

Dedup semantics:

- Empty keys are dropped
- Duplicate keys are deduplicated by key
- Last value wins, but key keeps first insertion position

Example:

- Input attrs: `dup=one`, `x=1`, `dup=two`
- Output order: `dup`, `x`
- Output values: `dup=two`, `x=1`

## Concurrency, thread safety, fatal behavior

Concurrency/thread safety:

- Logger and child loggers are safe for concurrent use
- Handler uses a shared mutex to serialize writes so each record is emitted atomically
- Internal buffers are pooled; behavior is transparent to callers

Fatal behavior and `ExitFunc`:

- `Fatal`, `Fatalf`, `FatalCtx` write the record and then invoke `ExitFunc(1)`
- Default `ExitFunc` is `os.Exit`
- In tests or controlled programs, set `ExitFunc` to a custom function to avoid process termination
- Exit hook is called even if writer returns an error after handle attempt

## Examples and validation commands

Example programs:

- `examples/log/00_handlers/main.go`
- `examples/log/01_config_all/main.go`
- `examples/log/02_fields_all/main.go`
- `examples/log/03_context_all/main.go`
- `examples/log/04_child_loggers/main.go`
- `examples/log/05_levels/main.go`
- `examples/log/06_writers/main.go`
- `examples/log/07_with_tools/main.go`
- `examples/log/08_trace_source_otel/main.go`

Run examples:

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
