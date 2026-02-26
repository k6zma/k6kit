package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sort"
	"strings"
	"time"
)

// normalizedAttr is a flattened key/value pair after normalization
type normalizedAttr struct {
	// Key is the canonical flattened attribute key
	Key string

	// Value is the normalized attribute payload
	Value any
}

// normalizedRecord is canonical renderer input
type normalizedRecord struct {
	Time        time.Time
	TraceID     string
	SpanID      string
	SourceTrace string
	Level       string
	Group       string
	Msg         string
	Attrs       []normalizedAttr
}

// normalizedAttrsByKey adapts normalized attributes
type normalizedAttrsByKey []normalizedAttr

// Len returns number of normalized attributes
func (a normalizedAttrsByKey) Len() int {
	return len(a)
}

// Less reports lexical order by attribute key
func (a normalizedAttrsByKey) Less(i, j int) bool {
	return a[i].Key < a[j].Key
}

// Swap exchanges two attribute positions
func (a normalizedAttrsByKey) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// normalizeRecord builds canonical record used by renderers
func normalizeRecord(
	ctx context.Context,
	rec slog.Record,
	staticAttrs []slog.Attr,
	loggerAttrs []slog.Attr,
	group string,
	enableSource bool,
	requestID string,
	otel OtelTrace,
	callAttrs []slog.Attr,
) normalizedRecord {
	out := normalizedRecord{
		Time:  rec.Time,
		Level: levelLabel(rec.Level),
		Group: group,
		Msg:   rec.Message,
	}

	if out.Time.IsZero() {
		out.Time = time.Now()
	}

	if enableSource {
		out.SourceTrace = sourceFromPC(rec.PC)
	}

	if otel.TraceID == "" && otel.SpanID == "" {
		if ot, ok := OtelTraceFromContext(ctx); ok {
			otel = ot
		}
	}

	if otel.TraceID != "" || otel.SpanID != "" {
		out.TraceID = otel.TraceID
		out.SpanID = otel.SpanID
	}

	estimated := len(staticAttrs) + len(loggerAttrs) + len(callAttrs)
	if len(callAttrs) == 0 {
		estimated += rec.NumAttrs()
	}

	if requestID != "" {
		estimated++
	}

	var all []normalizedAttr

	if estimated > 0 {
		all = make([]normalizedAttr, 0, estimated)
	}

	all = appendFlattened(all, staticAttrs)
	all = appendFlattened(all, loggerAttrs)

	rid := requestID
	if rid == "" {
		if fromCtx, ok := RequestID(ctx); ok {
			rid = fromCtx
		}
	}

	if rid != "" {
		all = append(all, normalizedAttr{Key: keyRequestID, Value: rid})
	}

	if len(callAttrs) > 0 {
		all = appendFlattened(all, callAttrs)
	} else {
		rec.Attrs(func(a slog.Attr) bool {
			all = flattenAttr(all, nil, a)

			return true
		})
	}

	if attrsAlreadyCanonical(all) {
		out.Attrs = all

		return out
	}

	out.Attrs = dedupAndSort(all)

	return out
}

// appendFlattened flattens attrs into dst using dot notation for groups
func appendFlattened(dst []normalizedAttr, attrs []slog.Attr) []normalizedAttr {
	for _, a := range attrs {
		dst = flattenAttr(dst, nil, a)
	}

	return dst
}

// flattenAttr flattens one attribute recursively
func flattenAttr(dst []normalizedAttr, path []string, a slog.Attr) []normalizedAttr {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return dst
	}

	if a.Value.Kind() == slog.KindGroup {
		nextPath := path
		if a.Key != "" {
			nextPath = append(nextPath, a.Key)
		}

		for _, child := range a.Value.Group() {
			dst = flattenAttr(dst, nextPath, child)
		}

		return dst
	}

	if a.Key == "" {
		return dst
	}

	return append(dst, normalizedAttr{Key: joinPathKey(path, a.Key), Value: normalizedValue(a.Value)})
}

// normalizedValue normalizes special payload types
func normalizedValue(v slog.Value) any {
	switch v.Kind() {
	case slog.KindBool:
		return v.Bool()
	case slog.KindDuration:
		return v.Duration()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindString:
		return v.String()
	case slog.KindTime:
		return v.Time()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindAny:
		return normalizedAny(v.Any())
	default:
		return normalizedAny(v.Any())
	}
}

func normalizedAny(v any) any {
	switch t := v.(type) {
	case error:
		if t == nil {
			return ""
		}

		return t.Error()
	case []error:
		return t
	default:
		return t
	}
}

func joinPathKey(path []string, key string) string {
	if len(path) == 0 {
		return key
	}

	total := len(key) + len(path)
	for i := range path {
		total += len(path[i])
	}

	b := make([]byte, 0, total)

	for i := range path {
		if i > 0 {
			b = append(b, '.')
		}

		b = append(b, path[i]...)
	}

	b = append(b, '.')
	b = append(b, key...)

	return string(b)
}

// dedupAndSort drops empty keys, keeps last value, and sorts by key
func dedupAndSort(attrs []normalizedAttr) []normalizedAttr {
	if len(attrs) == 0 {
		return nil
	}

	filtered := attrs[:0]
	for _, kv := range attrs {
		if strings.TrimSpace(kv.Key) == "" {
			continue
		}

		filtered = append(filtered, kv)
	}

	if len(filtered) == 0 {
		return nil
	}

	sort.Stable(normalizedAttrsByKey(filtered))

	write := 0

	for i := 0; i < len(filtered); {
		j := i + 1
		for j < len(filtered) && filtered[j].Key == filtered[i].Key {
			j++
		}

		filtered[write] = filtered[j-1]
		write++
		i = j
	}

	return filtered[:write]
}

func attrsAlreadyCanonical(attrs []normalizedAttr) bool {
	if len(attrs) == 0 {
		return true
	}

	prev := ""

	for i := range attrs {
		key := attrs[i].Key
		if strings.TrimSpace(key) == "" {
			return false
		}

		if i > 0 && key <= prev {
			return false
		}

		prev = key
	}

	return true
}

// sourceFromPC resolves source trace from program counter
func sourceFromPC(pc uintptr) string {
	if pc == 0 {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	file, line := fn.FileLine(pc)
	if file == "" {
		return ""
	}

	if idx := strings.LastIndex(file, "/"); idx >= 0 && idx+1 < len(file) {
		file = file[idx+1:]
	}

	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}

// levelLabel converts slog level to canonical uppercase label
func levelLabel(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return levelLabelDebug
	case level < slog.LevelWarn:
		return levelLabelInfo
	case level < slog.LevelError:
		return levelLabelWarn
	case level < slog.LevelError+4:
		return levelLabelError
	case level < slog.LevelError+8:
		return levelLabelFatal
	default:
		return levelLabelPanic
	}
}
