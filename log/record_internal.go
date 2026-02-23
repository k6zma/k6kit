package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// otelTraceExtractorFunc extracts trace/span IDs from context into fixed buffers
type otelTraceExtractorFunc func(context.Context, *[32]byte, *[16]byte) (traceIDLen, spanIDLen int)

// kv is a flattened normalized key/value pair used by renderers
type kv struct {
	// Key is a flattened key path in dot notation
	Key string

	// Value is typed slog value for fast scalar rendering
	Value slog.Value

	// Any stores pre-normalized fallback data for KindAny values
	Any any

	// UseAny selects Any over Value during rendering
	UseAny bool
}

// record is the canonical normalized log record consumed by renderers
type record struct {
	// Time is final timestamp resolved for this record
	Time time.Time

	// TraceID stores OTEL trace id in hex bytes when present
	TraceID [32]byte

	// TraceIDLen indicates valid bytes in TraceID
	TraceIDLen int

	// SpanID stores OTEL span id in hex bytes when present
	SpanID [16]byte

	// SpanIDLen indicates valid bytes in SpanID
	SpanIDLen int

	// SourceTrace stores caller source trace when enabled
	SourceTrace string

	// Level is normalized uppercase level label
	Level string

	// Group is merged group path in dot notation
	Group string

	// Msg is log message payload
	Msg string

	// Attrs are flattened, deduplicated, and ordered attributes
	Attrs []kv
}

// normalizeOptions configures canonical record normalization behavior
type normalizeOptions struct {
	// Group is currently active merged group path
	Group string

	// StaticAttrs are config-level attributes added to every record
	StaticAttrs []slog.Attr

	// GlobalAttrs are attributes accumulated via child loggers
	GlobalAttrs []slog.Attr

	// OTELTrace is optional trace/span extractor
	OTELTrace otelTraceExtractorFunc

	// IncludeSource toggles source trace extraction from PC
	IncludeSource bool

	// Now is a clock hook used by tests and zero-time fallback
	Now func() time.Time
}

// contextFields is a compact context extraction result used in normalize
type contextFields struct {
	// RequestID is request id value if available
	RequestID string

	// HasRequestID indicates whether RequestID is set
	HasRequestID bool

	// Metadata are request-scoped extra fields
	Metadata []Field

	// TraceID is extracted OTEL trace id bytes
	TraceID [32]byte

	// TraceIDLen is valid length in TraceID
	TraceIDLen int

	// SpanID is extracted OTEL span id bytes
	SpanID [16]byte

	// SpanIDLen is valid length in SpanID
	SpanIDLen int
}

// defaultTimeFormat returns default timestamp layout for a given output format
func defaultTimeFormat(format Format) string {
	if format == FormatJSON {
		return jsonTimeFormat
	}

	return textTimeFormat
}

// normalize transforms slog.Record into canonical record for all renderers
func normalize(ctx context.Context, rec slog.Record, flat []kv, opts normalizeOptions) record {
	ts := rec.Time
	if rec.Time.IsZero() {
		now := time.Now
		if opts.Now != nil {
			now = opts.Now
		}

		ts = now()
	}

	out := record{
		Time:  ts,
		Level: normalizeLevel(rec.Level),
		Group: opts.Group,
		Msg:   rec.Message,
	}

	ctxFields := extractContextFields(ctx, opts.OTELTrace)
	out.TraceID = ctxFields.TraceID
	out.TraceIDLen = ctxFields.TraceIDLen
	out.SpanID = ctxFields.SpanID
	out.SpanIDLen = ctxFields.SpanIDLen

	if opts.IncludeSource {
		out.SourceTrace = sourceTraceFromPC(rec.PC)
	}

	estimated := len(opts.StaticAttrs) + len(opts.GlobalAttrs) + rec.NumAttrs()
	requestID := ctxFields.RequestID
	hasRequestID := ctxFields.HasRequestID
	requestMetadata := ctxFields.Metadata

	estimated += len(requestMetadata)
	if hasRequestID {
		estimated++
	}

	if cap(flat) < estimated {
		flat = make([]kv, 0, estimated)
	} else {
		flat = flat[:0]
	}

	for _, attr := range opts.StaticAttrs {
		flattenAttr("", attr, &flat)
	}

	for _, attr := range opts.GlobalAttrs {
		flattenAttr("", attr, &flat)
	}

	if hasRequestID {
		flat = append(flat, kv{Key: keyRequestID, Value: slog.StringValue(requestID)})
	}

	for _, field := range requestMetadata {
		flattenAttr("", field.toAttr(), &flat)
	}

	rec.Attrs(func(attr slog.Attr) bool {
		flattenAttr("", attr, &flat)

		return true
	})

	out.Attrs = mergeOrderedKVs(flat)

	return out
}

// extractContextFields pulls request metadata and OTEL fields from context
func extractContextFields(ctx context.Context, otelExtractor otelTraceExtractorFunc) contextFields {
	if ctx == nil {
		return contextFields{}
	}

	fields := contextFields{}
	if metadata, ok := ctx.Value(requestMetadataCtxKey{}).([]Field); ok {
		fields.Metadata = metadata
	}

	if requestID, ok := ctx.Value(requestIDCtxKey{}).(string); ok && requestID != "" {
		fields.RequestID = requestID
		fields.HasRequestID = true
	}

	if otelExtractor != nil {
		fields.TraceIDLen, fields.SpanIDLen = otelExtractor(ctx, &fields.TraceID, &fields.SpanID)
	}

	return fields
}

// mergeOrderedKVs filters and deduplicates attrs preserving insertion order
// with stable positions and last-write-wins values.
func mergeOrderedKVs(values []kv) []kv {
	if len(values) == 0 {
		return nil
	}

	filtered := values[:0]
	for i := range values {
		if values[i].Key != "" {
			filtered = append(filtered, values[i])
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	if len(filtered) == 1 {
		return filtered
	}

	out := filtered[:0]
	indices := make(map[string]int, len(filtered))

	for _, item := range filtered {
		if idx, ok := indices[item.Key]; ok {
			out[idx] = item

			continue
		}

		indices[item.Key] = len(out)
		out = append(out, item)
	}

	return out
}

// flattenAttr flattens nested slog attrs/groups into dot-notated key/value pairs
func flattenAttr(prefix string, attr slog.Attr, out *[]kv) {
	a := attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	currentPrefix := prefix

	if a.Kind() == slog.KindGroup {
		if attr.Key != "" {
			currentPrefix = joinKey(prefix, attr.Key)
		}

		for _, nested := range a.Group() {
			flattenAttr(currentPrefix, nested, out)
		}

		return
	}

	if attr.Key == "" {
		return
	}

	key := joinKey(prefix, attr.Key)
	k := kv{Key: key}

	if a.Kind() == slog.KindAny {
		if v, ok := normalizeAnyValue(a.Any()); ok {
			k.Any = v
			k.UseAny = true
		} else {
			k.Value = a
		}
	} else {
		k.Value = a
	}

	*out = append(*out, k)
}

// joinKey joins prefix and key using dot notation
func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}

	return prefix + "." + key
}

// normalizeAnyValue normalizes special Any payloads
func normalizeAnyValue(v any) (any, bool) {
	switch t := v.(type) {
	case error:
		return errorMessage(t), true
	case []error:
		out := make([]string, len(t))
		for i, err := range t {
			out[i] = errorMessage(err)
		}

		return out, true
	default:
		return nil, false
	}
}

// errorMessage converts an error value to a message-only payload.
func errorMessage(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

// sourceTraceFromPC resolves file:line and function name from PC
func sourceTraceFromPC(pc uintptr) string {
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

	short := file
	if idx := strings.LastIndex(short, "/"); idx >= 0 && idx < len(short)-1 {
		short = short[idx+1:]
	}

	return fmt.Sprintf("%s:%d %s", short, line, fn.Name())
}

// normalizeLevel maps slog levels to canonical uppercase labels
func normalizeLevel(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return levelLabelDebug
	case level < slog.LevelWarn:
		return levelLabelInfo
	case level < slog.LevelError:
		return levelLabelWarn
	case level < slog.LevelError+4:
		return levelLabelError
	default:
		return levelLabelFatal
	}
}
