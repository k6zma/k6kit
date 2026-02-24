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

// normalizeRecord builds canonical record used by renderers
func normalizeRecord(
	ctx context.Context,
	rec slog.Record,
	staticAttrs []slog.Attr,
	loggerAttrs []slog.Attr,
	group string,
	enableSource bool,
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

	if ot, ok := OtelTraceFromContext(ctx); ok {
		out.TraceID = ot.TraceID
		out.SpanID = ot.SpanID
	}

	all := make([]normalizedAttr, 0, len(staticAttrs)+len(loggerAttrs)+rec.NumAttrs()+1)

	all = appendFlattened(all, staticAttrs)
	all = appendFlattened(all, loggerAttrs)

	if rid, ok := RequestID(ctx); ok {
		all = append(all, normalizedAttr{Key: keyRequestID, Value: rid})
	}

	rec.Attrs(func(a slog.Attr) bool {
		all = flattenAttr(all, "", a)

		return true
	})

	out.Attrs = dedupAndSort(all)

	return out
}

// appendFlattened flattens attrs into dst using dot notation for groups
func appendFlattened(dst []normalizedAttr, attrs []slog.Attr) []normalizedAttr {
	for _, a := range attrs {
		dst = flattenAttr(dst, "", a)
	}

	return dst
}

// flattenAttr flattens one attribute recursively
func flattenAttr(dst []normalizedAttr, prefix string, a slog.Attr) []normalizedAttr {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return dst
	}

	if a.Value.Kind() == slog.KindGroup {
		nextPrefix := prefix
		if a.Key != "" {
			nextPrefix = joinKey(prefix, a.Key)
		}

		for _, child := range a.Value.Group() {
			dst = flattenAttr(dst, nextPrefix, child)
		}

		return dst
	}

	if a.Key == "" {
		return dst
	}

	return append(dst, normalizedAttr{Key: joinKey(prefix, a.Key), Value: normalizedValue(a.Value.Any())})
}

// normalizedValue normalizes special payload types
func normalizedValue(v any) any {
	switch t := v.(type) {
	case error:
		if t == nil {
			return ""
		}

		return t.Error()
	case []error:
		out := make([]string, len(t))
		for i := range t {
			if t[i] != nil {
				out[i] = t[i].Error()
			}
		}

		return out
	default:
		return t
	}
}

// dedupAndSort drops empty keys, keeps last value, and sorts by key
func dedupAndSort(attrs []normalizedAttr) []normalizedAttr {
	if len(attrs) == 0 {
		return nil
	}

	m := make(map[string]normalizedAttr, len(attrs))
	keys := make([]string, 0, len(attrs))

	for _, kv := range attrs {
		if strings.TrimSpace(kv.Key) == "" {
			continue
		}

		if _, ok := m[kv.Key]; !ok {
			keys = append(keys, kv.Key)
		}

		m[kv.Key] = kv
	}

	sort.Strings(keys)

	out := make([]normalizedAttr, 0, len(keys))
	for _, key := range keys {
		out = append(out, m[key])
	}

	return out
}

// joinKey joins prefix and key with dot separator
func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}

	return prefix + "." + key
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
