package log

import "context"

const (
	// defaultTextTimeFormat is the default timestamp layout for text output
	defaultTextTimeFormat = "2006-01-02 15:04:05"

	// defaultJSONTimeFormat is the default timestamp layout for JSON output
	defaultJSONTimeFormat = "2006-01-02T15:04:05"
)

const (
	// keyTime is the canonical top level timestamp key
	keyTime = "time"

	// keyTraceID is the canonical OTEL trace id key
	keyTraceID = "trace_id"

	// keySpanID is the canonical OTEL span id key
	keySpanID = "span_id"

	// keySourceTrace is the canonical source trace key
	keySourceTrace = "source_trace"

	// keyLevel is the canonical level key
	keyLevel = "level"

	// keyGroup is the canonical group key
	keyGroup = "group"

	// keyMsg is the canonical message key
	keyMsg = "msg"

	// keyRequestID is the canonical request id key
	keyRequestID = "request_id"

	// keyApp is the canonical app metadata key
	keyApp = "app"

	// keyEnv is the canonical environment metadata key
	keyEnv = "env"

	// keyVersion is the canonical version metadata ke.
	keyVersion = "version"

	// keyError is the canonical error key
	keyError = "error"
)

const (
	// levelLabelDebug is canonical DEBUG level label
	levelLabelDebug = "DEBUG"

	// levelLabelInfo is canonical INFO level label
	levelLabelInfo = "INFO"

	// levelLabelWarn is canonical WARN level label
	levelLabelWarn = "WARN"

	// levelLabelError is canonical ERROR level label
	levelLabelError = "ERROR"

	// levelLabelFatal is canonical FATAL level label
	levelLabelFatal = "FATAL"

	// levelLabelPanic is canonical PANIC level label
	levelLabelPanic = "PANIC"
)

// normalizeContext converts nil context to context.Background
func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
