package log

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"time"
)

// renderBufPool is a pool of reusable []byte buffers for rendering
var renderBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)

		return &b
	},
}

// kvPool is a reusable pool of []kv slices for flattening attributes
var kvPool = sync.Pool{
	New: func() any {
		kvs := make([]kv, 0, 16)

		return &kvs
	},
}

type handlerOptions struct {
	// Level is the minimum enabled slog level for this handler
	Level slog.Level

	// IncludeSource toggles source trace extraction
	IncludeSource bool

	// Now is a clock hook used for deterministic tests
	Now func() time.Time

	// StaticAttrs are global attrs injected into every record
	StaticAttrs []slog.Attr

	// OTELTrace extracts trace/span IDs from context when enabled
	OTELTrace otelTraceExtractorFunc
}

type handler struct {
	// writer is the final sink for rendered bytes
	writer io.Writer

	// renderer is format specific (text/json) serialization
	renderer renderer

	// opts contains immutable runtime behavior toggles
	opts handlerOptions

	// attrs are accumulated via WithAttrs and applied during normalize
	attrs []slog.Attr

	// group is the merged dot-path accumulated via WithGroup
	group string

	// mu serializes writes so each log record is written atomically
	mu *sync.Mutex
}

// handler is a log record serializer, accumulating attributes and maintaining
func newHandler(writer io.Writer, r renderer, opts handlerOptions) *handler {
	return &handler{writer: writer, renderer: r, opts: opts, mu: &sync.Mutex{}}
}

// Enabled reports whether the handler will serialize records at or above the specified level
func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level
}

// Handle serializes a log record to the writer using the provided renderer
func (h *handler) Handle(ctx context.Context, rec slog.Record) error {
	pKvs := kvPool.Get().(*[]kv)

	flat := (*pKvs)[:0]

	norm := normalize(ctx, rec, flat, normalizeOptions{
		Group:         h.group,
		StaticAttrs:   h.opts.StaticAttrs,
		GlobalAttrs:   h.attrs,
		OTELTrace:     h.opts.OTELTrace,
		IncludeSource: h.opts.IncludeSource,
		Now:           h.opts.Now,
	})

	pb := renderBufPool.Get().(*[]byte)
	b := (*pb)[:0]

	b, err := h.renderer.Append(b, norm)
	if err != nil {
		if cap(norm.Attrs) > 4096 {
			small := make([]kv, 0, 16)

			*pKvs = small
		} else {
			*pKvs = norm.Attrs[:0]
		}

		kvPool.Put(pKvs)

		*pb = b[:0]

		renderBufPool.Put(pb)

		return err
	}

	h.mu.Lock()
	_, err = h.writer.Write(b)
	h.mu.Unlock()

	if cap(norm.Attrs) > 4096 {
		small := make([]kv, 0, 16)

		*pKvs = small
	} else {
		*pKvs = norm.Attrs[:0]
	}

	kvPool.Put(pKvs)

	if cap(b) > 64*1024 {
		small := make([]byte, 0, 512)

		*pb = small
	} else {
		*pb = b[:0]
	}

	renderBufPool.Put(pb)

	return err
}

// WithAttrs appends additional attributes to the handler, creating a new instance.
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := *h
	next.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)

	return &next
}

// WithGroup creates a new handler instance with an additional group name
func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	next := *h
	if h.group == "" {
		next.group = name
	} else {
		next.group = h.group + "." + name
	}

	return &next
}
