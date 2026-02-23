package log

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"strconv"
)

type textRenderer struct {
	// Color toggles ANSI level/key colorization in text mode
	Color bool

	// TimeFormat is the timestamp layout used for text records
	TimeFormat string
}

// newTextHandler builds a base handler configured with the text renderer
func newTextHandler(writer io.Writer, opts handlerOptions, color bool, timeFormat string) *handler {
	return newHandler(writer, textRenderer{Color: color, TimeFormat: timeFormat}, opts)
}

// Append writes a normalized record in bracketed text format
func (r textRenderer) Append(dst []byte, rec record) ([]byte, error) {
	dst = append(dst, '[')
	dst = rec.Time.AppendFormat(dst, r.TimeFormat)
	dst = append(dst, ']')

	if rec.TraceIDLen > 0 || rec.SpanIDLen > 0 {
		dst = append(dst, ' ', '[')
		wrote := false

		if rec.TraceIDLen > 0 {
			dst = append(dst, "trace_id="...)
			dst = append(dst, rec.TraceID[:rec.TraceIDLen]...)
			wrote = true
		}

		if rec.SpanIDLen > 0 {
			if wrote {
				dst = append(dst, ' ')
			}

			dst = append(dst, "span_id="...)
			dst = append(dst, rec.SpanID[:rec.SpanIDLen]...)
		}

		dst = append(dst, ']')
	}

	if rec.SourceTrace != "" {
		dst = append(dst, ' ', '[')
		dst = appendTextSafe(dst, rec.SourceTrace)
		dst = append(dst, ']')
	}

	dst = append(dst, ' ', '[')
	if r.Color {
		dst = append(dst, colorForLevel(rec.Level)...)
	}

	dst = append(dst, rec.Level...)
	if r.Color {
		dst = append(dst, colorReset...)
	}

	dst = append(dst, ']')

	if rec.Group != "" {
		dst = append(dst, ' ', '[')
		dst = appendTextSafe(dst, rec.Group)
		dst = append(dst, ']')
	}

	dst = append(dst, ' ')
	dst = appendTextSafe(dst, rec.Msg)

	if len(rec.Attrs) > 0 {
		dst = append(dst, ' ')

		for i, item := range rec.Attrs {
			if i > 0 {
				dst = append(dst, ',')
			}

			dst = append(dst, '{')
			if r.Color {
				dst = append(dst, colorCyan...)
			}

			dst = appendTextSafe(dst, item.Key)
			if r.Color {
				dst = append(dst, colorReset...)
			}

			dst = append(dst, '=')
			dst = appendTextKVValue(dst, item)
			dst = append(dst, '}')
		}
	}

	dst = append(dst, '\n')

	return dst, nil
}

// appendTextKVValue appends one attribute value in text-safe form
func appendTextKVValue(dst []byte, item kv) []byte {
	if item.UseAny {
		return appendTextSafe(dst, fmt.Sprint(item.Any))
	}

	switch item.Value.Kind() {
	case slog.KindString:
		return appendTextSafe(dst, item.Value.String())
	case slog.KindInt64:
		return strconv.AppendInt(dst, item.Value.Int64(), 10)
	case slog.KindUint64:
		return strconv.AppendUint(dst, item.Value.Uint64(), 10)
	case slog.KindFloat64:
		v := item.Value.Float64()
		if math.IsNaN(v) {
			return append(dst, "NaN"...)
		}

		if math.IsInf(v, 1) {
			return append(dst, "+Inf"...)
		}

		if math.IsInf(v, -1) {
			return append(dst, "-Inf"...)
		}

		return strconv.AppendFloat(dst, v, 'f', -1, 64)
	case slog.KindBool:
		return strconv.AppendBool(dst, item.Value.Bool())
	case slog.KindDuration:
		return appendTextSafe(dst, item.Value.Duration().String())
	case slog.KindTime:
		return appendTextSafe(dst, item.Value.Time().String())
	case slog.KindAny:
		return appendTextSafe(dst, fmt.Sprint(item.Value.Any()))
	default:
		return appendTextSafe(dst, fmt.Sprint(item.Value.Any()))
	}
}

// appendTextSafe escapes control characters for single-line text output
func appendTextSafe(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			dst = append(dst, '\\', '\\')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			if s[i] < 0x20 || s[i] == 0x7f {
				dst = append(dst, '\\', 'x')
				dst = append(dst, lowerHex[s[i]>>4], lowerHex[s[i]&0x0f])

				continue
			}

			dst = append(dst, s[i])
		}
	}

	return dst
}

// colorForLevel returns ANSI color prefix for a log level label
func colorForLevel(level string) string {
	switch level {
	case levelLabelDebug:
		return colorBlue
	case levelLabelInfo:
		return colorGreen
	case levelLabelWarn:
		return colorYellow
	case levelLabelError:
		return colorRed
	case levelLabelFatal:
		return colorPurple
	default:
		return ""
	}
}
