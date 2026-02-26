package log

import (
	"fmt"
	"strconv"
)

// textRenderer serializes normalized records into bracketed text lines
type textRenderer struct {
	// color toggles ANSI colorization for level and keys
	color bool

	// timeFormat is the layout used for leading timestamp
	timeFormat string
}

const (
	// colorReset resets terminal color sequence
	colorReset = "\033[0m"

	// colorBlue is ANSI blue color code
	colorBlue = "\033[34m"

	// colorGreen is ANSI green color code
	colorGreen = "\033[32m"

	// colorYellow is ANSI yellow color code
	colorYellow = "\033[33m"

	// colorRed is ANSI red color code
	colorRed = "\033[31m"

	// colorCyan is ANSI cyan color code
	colorCyan = "\033[36m"

	// colorGray is ANSI gray color code
	colorGray = "\033[90m"
)

// colorForLevel returns ANSI color prefix for canonical level label
func colorForLevel(level string) string {
	switch level {
	case levelLabelDebug:
		return colorGray
	case levelLabelInfo:
		return colorGreen
	case levelLabelWarn:
		return colorYellow
	case levelLabelError:
		return colorRed
	case levelLabelFatal, levelLabelPanic:
		return colorBlue
	default:
		return ""
	}
}

// render writes one normalized record in text format
func (r textRenderer) render(dst []byte, rec normalizedRecord) ([]byte, error) {
	dst = append(dst, '[')
	dst = rec.Time.AppendFormat(dst, r.timeFormat)
	dst = append(dst, ']')

	if rec.TraceID != "" || rec.SpanID != "" {
		dst = append(dst, ' ', '[')

		if rec.TraceID != "" {
			dst = append(dst, keyTraceID...)
			dst = append(dst, '=')
			dst = append(dst, rec.TraceID...)
		}

		if rec.SpanID != "" {
			if rec.TraceID != "" {
				dst = append(dst, ' ')
			}

			dst = append(dst, keySpanID...)
			dst = append(dst, '=')
			dst = append(dst, rec.SpanID...)
		}

		dst = append(dst, ']')
	}

	if rec.SourceTrace != "" {
		dst = append(dst, ' ', '[')
		dst = append(dst, rec.SourceTrace...)
		dst = append(dst, ']')
	}

	dst = append(dst, ' ', '[')
	if r.color {
		dst = append(dst, colorForLevel(rec.Level)...)
	}

	dst = append(dst, rec.Level...)
	if r.color {
		dst = append(dst, colorReset...)
	}

	dst = append(dst, ']')

	if rec.Group != "" {
		dst = append(dst, ' ', '[')
		dst = append(dst, rec.Group...)
		dst = append(dst, ']')
	}

	dst = append(dst, ' ')
	dst = append(dst, rec.Msg...)

	for i, a := range rec.Attrs {
		if i == 0 {
			dst = append(dst, ' ')
		} else {
			dst = append(dst, ',')
		}

		dst = append(dst, '{')
		if r.color {
			dst = append(dst, colorCyan...)
		}

		dst = append(dst, a.Key...)
		if r.color {
			dst = append(dst, colorReset...)
		}

		dst = append(dst, '=')
		dst = appendTextValue(dst, a.Value)
		dst = append(dst, '}')
	}

	dst = append(dst, '\n')

	return dst, nil
}

func appendTextValue(dst []byte, v any) []byte {
	switch t := v.(type) {
	case nil:
		return append(dst, "<nil>"...)
	case bool:
		return strconv.AppendBool(dst, t)
	case string:
		return append(dst, t...)
	case int:
		return strconv.AppendInt(dst, int64(t), 10)
	case int8:
		return strconv.AppendInt(dst, int64(t), 10)
	case int16:
		return strconv.AppendInt(dst, int64(t), 10)
	case int32:
		return strconv.AppendInt(dst, int64(t), 10)
	case int64:
		return strconv.AppendInt(dst, t, 10)
	case uint:
		return strconv.AppendUint(dst, uint64(t), 10)
	case uint8:
		return strconv.AppendUint(dst, uint64(t), 10)
	case uint16:
		return strconv.AppendUint(dst, uint64(t), 10)
	case uint32:
		return strconv.AppendUint(dst, uint64(t), 10)
	case uint64:
		return strconv.AppendUint(dst, t, 10)
	case float32:
		return strconv.AppendFloat(dst, float64(t), 'g', -1, 32)
	case float64:
		return strconv.AppendFloat(dst, t, 'g', -1, 64)
	case []error:
		dst = append(dst, '[')

		for i := range t {
			if i > 0 {
				dst = append(dst, ' ')
			}

			if t[i] != nil {
				dst = append(dst, t[i].Error()...)
			}
		}

		return append(dst, ']')
	default:
		return append(dst, fmt.Sprint(v)...)
	}
}
