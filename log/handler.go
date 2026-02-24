package log

import (
	"context"
	"io"
	"log/slog"
	"sync"
)

// renderer serializes normalized records into byte slices
type renderer interface {
	// render serializes normalized record into destination bytes
	render(dst []byte, rec normalizedRecord) ([]byte, error)
}

// pipelineHandler is slog.Handler implementation used by k6kit logger runtime
type pipelineHandler struct {
	writer       io.Writer
	renderer     renderer
	minLevel     slog.Level
	staticAttrs  []slog.Attr
	loggerAttrs  []slog.Attr
	group        string
	enableSource bool
	mu           *sync.Mutex
	bufPool      *sync.Pool
}

const (
	handlerBufferInitialCap = 1024
	handlerBufferMaxCap     = 64 * 1024
)

// newHandler builds handler pipeline from merged config
func newHandler(cfg Config) *pipelineHandler {
	var r renderer = textRenderer{color: cfg.Color, timeFormat: cfg.TimeFormat}

	if cfg.Format == FormatJSON {
		r = jsonRenderer{timeFormat: cfg.TimeFormat}
	}

	static := make([]slog.Attr, 0, 3)
	if cfg.AppName != "" {
		static = append(static, slog.String(keyApp, cfg.AppName))
	}

	if cfg.Environment != "" {
		static = append(static, slog.String(keyEnv, cfg.Environment))
	}

	if cfg.Version != "" {
		static = append(static, slog.String(keyVersion, cfg.Version))
	}

	return &pipelineHandler{
		writer:       cfg.Writer,
		renderer:     r,
		minLevel:     cfg.Level.slogLevel(),
		staticAttrs:  static,
		enableSource: cfg.EnableSourceTrace,
		mu:           &sync.Mutex{},
		bufPool: &sync.Pool{New: func() any {
			b := make([]byte, 0, handlerBufferInitialCap)

			return &b
		}},
	}
}

// Enabled reports whether level passes minimum threshold
func (h *pipelineHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

// Handle normalizes record, renders it and writes
func (h *pipelineHandler) Handle(ctx context.Context, rec slog.Record) error {
	norm := normalizeRecord(ctx, rec, h.staticAttrs, h.loggerAttrs, h.group, h.enableSource)

	bp := h.bufPool.Get().(*[]byte)
	b := (*bp)[:0]

	b, err := h.renderer.render(b, norm)
	if err != nil {
		h.putBuffer(bp, b)

		return err
	}

	h.mu.Lock()
	_, err = h.writer.Write(b)
	h.mu.Unlock()

	h.putBuffer(bp, b)

	return err
}

func (h *pipelineHandler) putBuffer(bp *[]byte, b []byte) {
	if cap(b) > handlerBufferMaxCap {
		return
	}

	*bp = b[:0]
	h.bufPool.Put(bp)
}

// WithAttrs returns handler clone enriched with attrs
func (h *pipelineHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := *h
	next.loggerAttrs = append(append([]slog.Attr{}, h.loggerAttrs...), attrs...)

	return &next
}

// WithGroup returns handler clone with composed group path
func (h *pipelineHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	next := *h
	if next.group == "" {
		next.group = name
	} else {
		next.group += "." + name
	}

	return &next
}
