package log

import (
	"io"
	"os"
	"time"
)

// Format controls log output serialization
type Format string

// Config contains logger behavior and output settings
type Config struct {
	// Level sets the minimum enabled log level
	Level Level

	// Format selects the output serializer (`text` or `json`)
	Format Format

	// Color enables ANSI colors for text output
	Color bool

	// EnableSourceTrace adds call site information (file:line function)
	EnableSourceTrace bool

	// EnableOTEL enables OpenTelemetry trace/span extraction from context
	EnableOTEL bool

	// Env is included as global metadata in every record when non-empty
	Env string

	// AppName is included as global metadata in every record when non-empty
	AppName string

	// Version is included as global metadata in every record when non-empty
	Version string

	// Writer is the output sink, if it's nil, stdout is used
	Writer io.Writer

	// TimeFormat overrides the default format specific timestamp layout
	TimeFormat string

	// ExitFunc is called by Fatal level methods after writing the record
	ExitFunc func(code int)

	// now is an internal clock hook used by tests
	now func() time.Time
}

// defaultConfig - builds internal baseline configuration used by New/DefaultConfig
func defaultConfig() Config {
	return Config{
		Level:             LevelInfo,
		Format:            FormatText,
		Color:             false,
		EnableSourceTrace: false,
		EnableOTEL:        false,
		TimeFormat:        "",
		ExitFunc:          os.Exit,
		now:               time.Now,
	}
}

// DefaultConfig returns defaults for logger configuration
func DefaultConfig() Config {
	return defaultConfig()
}
