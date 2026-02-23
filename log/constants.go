package log

const (
	// FormatText writes records in bracketed text form
	FormatText Format = "text"

	// FormatJSON writes records as flattened JSON objects
	FormatJSON Format = "json"
)

const (
	// LevelDebug is debug logger level
	LevelDebug Level = iota - 1

	// LevelInfo is info logger level
	LevelInfo

	// LevelWarn is warn logger level
	LevelWarn

	// LevelError is error logger level
	LevelError

	// LevelFatal is fatal logger level
	LevelFatal
)

const (
	// textTimeFormat is the default timestamp layout for text output
	textTimeFormat = "2006-01-02 15:04:05"

	// jsonTimeFormat is the default timestamp layout for JSON output
	jsonTimeFormat = "2006-01-02T15:04:05"
)

const (
	// keyTime is the canonical timestamp field key
	keyTime = "time"

	// keyTraceID is the canonical OpenTelemetry trace id field key
	keyTraceID = "trace_id"

	// keySpanID is the canonical OpenTelemetry span id field key
	keySpanID = "span_id"

	// keySourceTrace is the canonical source trace field key
	keySourceTrace = "source_trace"

	// keyLevel is the canonical log level field key
	keyLevel = "level"

	// keyGroup is the canonical group path field key
	keyGroup = "group"

	// keyMsg is the canonical message field key
	keyMsg = "msg"

	// keyRequestID is the request id metadata field key
	keyRequestID = "request_id"

	// keyApp is the application name metadata field key
	keyApp = "app"

	// keyEnv is the environment metadata field key
	keyEnv = "env"

	// keyVersion is the application version metadata field key
	keyVersion = "version"

	// keyError is the canonical error field key
	keyError = "error"
)

const (
	// levelLabelDebug is the canonical uppercase DEBUG level label
	levelLabelDebug = "DEBUG"

	// levelLabelInfo is the canonical uppercase INFO level label
	levelLabelInfo = "INFO"

	// levelLabelWarn is the canonical uppercase WARN level label
	levelLabelWarn = "WARN"

	// levelLabelError is the canonical uppercase ERROR level label
	levelLabelError = "ERROR"

	// levelLabelFatal is the canonical uppercase FATAL level label
	levelLabelFatal = "FATAL"
)

const (
	// levelNameDebug is the lowercase DEBUG token accepted by ParseLevel
	levelNameDebug = "debug"

	// levelNameInfo is the lowercase INFO token accepted by ParseLevel
	levelNameInfo = "info"

	// levelNameWarn is the lowercase WARN token accepted by ParseLevel
	levelNameWarn = "warn"

	// levelNameError is the lowercase ERROR token accepted by ParseLevel
	levelNameError = "error"

	// levelNameFatal is the lowercase FATAL token accepted by ParseLevel
	levelNameFatal = "fatal"
)

const (
	// Color reset ASCII code
	colorReset = "\033[0m"

	// Blue color ASCII code
	colorBlue = "\033[34m"

	// Green color ASCII code
	colorGreen = "\033[32m"

	// Yellow color ASCII code
	colorYellow = "\033[33m"

	// Red color ASCII code
	colorRed = "\033[31m"

	// Purple color ASCII code
	colorPurple = "\033[35m"

	// Cyan color ASCII code
	colorCyan = "\033[36m"

	// lowerHex is used to render escaped control bytes as \xNN
	lowerHex = "0123456789abcdef"
)

// timeKey is an inlined prefix for the first canonical JSON field
const timeKey = `"` + keyTime + `":"`
