package log

import (
	"io"
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
)

// sonicJSON is the JSON engine used for fallback Any value marshaling
var sonicJSON = sonic.ConfigStd

// jsonRenderer is a JSON renderer that flattens records as canonical JSON objects
type jsonRenderer struct {
	// TimeFormat is the timestamp layout used for JSON records
	TimeFormat string
}

// newJSONHandler builds a base handler configured with the JSON renderer
func newJSONHandler(writer io.Writer, opts handlerOptions, timeFormat string) *handler {
	return newHandler(writer, jsonRenderer{TimeFormat: timeFormat}, opts)
}

// Append writes a normalized record as a flattened canonical JSON object
func (r jsonRenderer) Append(dst []byte, rec record) ([]byte, error) {
	var err error

	dst = append(dst, '{')
	dst = append(dst, timeKey...)
	dst = rec.Time.AppendFormat(dst, r.TimeFormat)
	dst = append(dst, '"')

	if rec.TraceIDLen > 0 {
		dst = append(dst, ',', '"')
		dst = append(dst, keyTraceID...)
		dst = append(dst, '"', ':')
		dst = append(dst, '"')
		dst = append(dst, rec.TraceID[:rec.TraceIDLen]...)
		dst = append(dst, '"')
	}

	if rec.SpanIDLen > 0 {
		dst = append(dst, ',', '"')
		dst = append(dst, keySpanID...)
		dst = append(dst, '"', ':')
		dst = append(dst, '"')
		dst = append(dst, rec.SpanID[:rec.SpanIDLen]...)
		dst = append(dst, '"')
	}

	if rec.SourceTrace != "" {
		dst = append(dst, ',', '"')
		dst = append(dst, keySourceTrace...)
		dst = append(dst, '"', ':')
		dst = strconv.AppendQuote(dst, rec.SourceTrace)
	}

	dst = append(dst, ',', '"')
	dst = append(dst, keyLevel...)
	dst = append(dst, '"', ':')

	dst = strconv.AppendQuote(dst, rec.Level)
	if rec.Group != "" {
		dst = append(dst, ',', '"')
		dst = append(dst, keyGroup...)
		dst = append(dst, '"', ':')
		dst = strconv.AppendQuote(dst, rec.Group)
	}

	dst = append(dst, ',', '"')
	dst = append(dst, keyMsg...)
	dst = append(dst, '"', ':')
	dst = strconv.AppendQuote(dst, rec.Msg)

	for _, item := range rec.Attrs {
		key := item.Key
		if isCanonicalJSONKey(key) {
			key = "attr." + key
		}

		dst = append(dst, ',')
		dst = strconv.AppendQuote(dst, key)
		dst = append(dst, ':')

		dst, err = appendJSONKVValue(dst, item)
		if err != nil {
			dst = append(dst, "null"...)
		}
	}

	dst = append(dst, '}', '\n')

	return dst, nil
}

// appendJSONValue appends a JSON-encoded value with scalar fast paths
func appendJSONValue(dst []byte, value any) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return append(dst, "null"...), nil
	case string:
		return strconv.AppendQuote(dst, v), nil
	case bool:
		return strconv.AppendBool(dst, v), nil
	case int:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case int8:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case int16:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case int32:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case int64:
		return strconv.AppendInt(dst, v, 10), nil
	case uint:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case uint8:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case uint16:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case uint32:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case uint64:
		return strconv.AppendUint(dst, v, 10), nil
	case float32:
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return append(dst, "null"...), nil
		}

		return strconv.AppendFloat(dst, float64(v), 'f', -1, 32), nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return append(dst, "null"...), nil
		}

		return strconv.AppendFloat(dst, v, 'f', -1, 64), nil
	}

	b, err := sonicJSON.Marshal(value)
	if err != nil {
		return dst, err
	}

	return append(dst, b...), nil
}

// appendJSONKVValue appends one normalized key/value pair value payload
func appendJSONKVValue(dst []byte, item kv) ([]byte, error) {
	if item.UseAny {
		return appendJSONValue(dst, item.Any)
	}

	switch item.Value.Kind() {
	case slog.KindString:
		return strconv.AppendQuote(dst, item.Value.String()), nil
	case slog.KindInt64:
		return strconv.AppendInt(dst, item.Value.Int64(), 10), nil
	case slog.KindUint64:
		return strconv.AppendUint(dst, item.Value.Uint64(), 10), nil
	case slog.KindFloat64:
		v := item.Value.Float64()
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return append(dst, "null"...), nil
		}

		return strconv.AppendFloat(dst, v, 'f', -1, 64), nil
	case slog.KindBool:
		return strconv.AppendBool(dst, item.Value.Bool()), nil
	case slog.KindDuration:
		return strconv.AppendInt(dst, int64(item.Value.Duration()), 10), nil
	case slog.KindTime:
		return strconv.AppendQuote(dst, item.Value.Time().Format(time.RFC3339Nano)), nil
	case slog.KindAny:
		return appendJSONValue(dst, item.Value.Any())
	default:
		return appendJSONValue(dst, item.Value.Any())
	}
}

// isCanonicalJSONKey reports whether key belongs to canonical record fields
func isCanonicalJSONKey(key string) bool {
	switch key {
	case keyTime, keyTraceID, keySpanID, keySourceTrace, keyLevel, keyGroup, keyMsg:
		return true
	default:
		return false
	}
}
