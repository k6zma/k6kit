package log

import (
	"math"
	"strconv"

	"github.com/bytedance/sonic"
)

// jsonRenderer serializes normalized records into JSON
type jsonRenderer struct {
	// timeFormat is the layout used for toplevel time field
	timeFormat string
}

// render writes one normalized record in JSON format
func (r jsonRenderer) render(dst []byte, rec normalizedRecord) ([]byte, error) {
	dst = append(dst, '{')
	dst = append(dst, `"`...)
	dst = append(dst, keyTime...)
	dst = append(dst, `":`...)

	dst = strconv.AppendQuote(dst, rec.Time.Format(r.timeFormat))

	if rec.TraceID != "" {
		dst = append(dst, `,"`...)
		dst = append(dst, keyTraceID...)
		dst = append(dst, `":`...)

		dst = strconv.AppendQuote(dst, rec.TraceID)
	}

	if rec.SpanID != "" {
		dst = append(dst, `,"`...)
		dst = append(dst, keySpanID...)
		dst = append(dst, `":`...)

		dst = strconv.AppendQuote(dst, rec.SpanID)
	}

	if rec.SourceTrace != "" {
		dst = append(dst, `,"`...)
		dst = append(dst, keySourceTrace...)
		dst = append(dst, `":`...)

		dst = strconv.AppendQuote(dst, rec.SourceTrace)
	}

	dst = append(dst, `,"`...)
	dst = append(dst, keyLevel...)
	dst = append(dst, `":`...)

	dst = strconv.AppendQuote(dst, rec.Level)

	if rec.Group != "" {
		dst = append(dst, `,"`...)
		dst = append(dst, keyGroup...)
		dst = append(dst, `":`...)

		dst = strconv.AppendQuote(dst, rec.Group)
	}

	dst = append(dst, `,"`...)
	dst = append(dst, keyMsg...)
	dst = append(dst, `":`...)

	dst = strconv.AppendQuote(dst, rec.Msg)

	for _, a := range rec.Attrs {
		key := a.Key
		if isCanonicalKey(key) {
			key = "attr." + key
		}

		dst = append(dst, ',')

		dst = strconv.AppendQuote(dst, key)

		dst = append(dst, ':')

		if fast, ok := appendJSONPrimitive(dst, a.Value); ok {
			dst = fast

			continue
		}

		b, err := sonic.ConfigFastest.Marshal(a.Value)
		if err != nil {
			dst = append(dst, "null"...)

			continue
		}

		dst = append(dst, b...)
	}

	dst = append(dst, '}', '\n')

	return dst, nil
}

// appendJSONPrimitive appends v to dst if it is a primitive type
func appendJSONPrimitive(dst []byte, v any) ([]byte, bool) {
	switch t := v.(type) {
	case nil:
		return append(dst, "null"...), true
	case bool:
		if t {
			return append(dst, "true"...), true
		}

		return append(dst, "false"...), true
	case string:
		return strconv.AppendQuote(dst, t), true
	case int:
		return strconv.AppendInt(dst, int64(t), 10), true
	case int8:
		return strconv.AppendInt(dst, int64(t), 10), true
	case int16:
		return strconv.AppendInt(dst, int64(t), 10), true
	case int32:
		return strconv.AppendInt(dst, int64(t), 10), true
	case int64:
		return strconv.AppendInt(dst, t, 10), true
	case uint:
		return strconv.AppendUint(dst, uint64(t), 10), true
	case uint8:
		return strconv.AppendUint(dst, uint64(t), 10), true
	case uint16:
		return strconv.AppendUint(dst, uint64(t), 10), true
	case uint32:
		return strconv.AppendUint(dst, uint64(t), 10), true
	case uint64:
		return strconv.AppendUint(dst, t, 10), true
	case float32:
		if math.IsNaN(float64(t)) || math.IsInf(float64(t), 0) {
			return append(dst, "null"...), true
		}

		return strconv.AppendFloat(dst, float64(t), 'g', -1, 32), true
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return append(dst, "null"...), true
		}

		return strconv.AppendFloat(dst, t, 'g', -1, 64), true
	default:
		return dst, false
	}
}

// isCanonicalKey reports whether key belongs to reserved top level keys
func isCanonicalKey(key string) bool {
	switch key {
	case keyTime, keyTraceID, keySpanID, keySourceTrace, keyLevel, keyGroup, keyMsg:
		return true
	default:
		return false
	}
}
