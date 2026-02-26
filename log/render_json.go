package log

import (
	"math"
	"sort"
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

	dst = append(dst, '"')
	dst = rec.Time.AppendFormat(dst, r.timeFormat)
	dst = append(dst, '"')

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
	case []string:
		return appendJSONStringArray(dst, t), true
	case []int:
		return appendJSONIntArray(dst, t), true
	case map[string]string:
		return appendJSONMapStringString(dst, t), true
	case []error:
		return appendJSONErrorArray(dst, t), true
	default:
		return dst, false
	}
}

func appendJSONStringArray(dst []byte, arr []string) []byte {
	dst = append(dst, '[')

	for i := range arr {
		if i > 0 {
			dst = append(dst, ',')
		}

		dst = strconv.AppendQuote(dst, arr[i])
	}

	dst = append(dst, ']')

	return dst
}

func appendJSONIntArray(dst []byte, arr []int) []byte {
	dst = append(dst, '[')

	for i := range arr {
		if i > 0 {
			dst = append(dst, ',')
		}

		dst = strconv.AppendInt(dst, int64(arr[i]), 10)
	}

	dst = append(dst, ']')

	return dst
}

func appendJSONMapStringString(dst []byte, m map[string]string) []byte {
	dst = append(dst, '{')
	if len(m) == 0 {
		dst = append(dst, '}')

		return dst
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for i := range keys {
		if i > 0 {
			dst = append(dst, ',')
		}

		k := keys[i]
		dst = strconv.AppendQuote(dst, k)
		dst = append(dst, ':')
		dst = strconv.AppendQuote(dst, m[k])
	}

	dst = append(dst, '}')

	return dst
}

func appendJSONErrorArray(dst []byte, arr []error) []byte {
	dst = append(dst, '[')

	for i := range arr {
		if i > 0 {
			dst = append(dst, ',')
		}

		msg := ""
		if arr[i] != nil {
			msg = arr[i].Error()
		}

		dst = strconv.AppendQuote(dst, msg)
	}

	dst = append(dst, ']')

	return dst
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
