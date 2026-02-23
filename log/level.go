package log

import (
	"fmt"
	"log/slog"
	"strings"
)

// Level defines supported logger levels
type Level int

// slogLevel returns the slog.Level for the given Level
func (l Level) slogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError + 4
	default:
		return slog.LevelInfo
	}
}

// String returns the string representation of the Level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return levelLabelDebug
	case LevelInfo:
		return levelLabelInfo
	case LevelWarn:
		return levelLabelWarn
	case LevelError:
		return levelLabelError
	case LevelFatal:
		return levelLabelFatal
	default:
		return levelLabelInfo
	}
}

// ParseLevel converts a string level name into Level
func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case levelNameDebug:
		return LevelDebug, nil
	case levelNameInfo:
		return LevelInfo, nil
	case levelNameWarn:
		return LevelWarn, nil
	case levelNameError:
		return LevelError, nil
	case levelNameFatal:
		return LevelFatal, nil
	default:
		return 0, fmt.Errorf("unsupported log level: %q", value)
	}
}
