package log

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func newBenchmarkLogger(b *testing.B, cfg Config) Logger {
	b.Helper()

	if cfg.Writer == nil {
		cfg.Writer = io.Discard
	}

	if cfg.now == nil {
		cfg.now = func() time.Time {
			return time.Unix(1700000000, 0)
		}
	}

	l, err := New(cfg)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}

	return l
}

type benchmarkBlackhole struct {
	writes atomic.Uint64
}

func (b *benchmarkBlackhole) Write(p []byte) (int, error) {
	b.writes.Add(1)

	return len(p), nil
}

func (b *benchmarkBlackhole) WriteCount() uint64 {
	return b.writes.Load()
}

func BenchmarkInfoNoFields(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Info("hello")
	}
}

func BenchmarkInfoWithFields(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Info(
			"hello",
			String("service", "api"),
			String("method", "GET"),
			Int("status", 200),
			Bool("ok", true),
		)
	}
}

func BenchmarkInfoWithContext(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON})

	ctx := WithRequestID(context.Background(), "req-1")
	ctx = WithRequestMetadata(ctx, String("tenant", "acme"), String("region", "us-east-1"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InfoCtx(ctx, "hello", Int("attempt", i&3))
	}
}

func BenchmarkInfoWithOTEL(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, EnableOTEL: true})

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		SpanID:     trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.InfoCtx(ctx, "hello", Int("n", i&7))
	}
}

func BenchmarkDebugDisabled(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Debug("debug-hidden", Int("n", i), String("k", "v"))
	}
}

func BenchmarkInfoWithDuplicateKeys(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Info(
			"dup",
			String("k", "v1"),
			Int("n", 1),
			String("k", "v2"),
			String("x", "x1"),
			String("x", "x2"),
			String("k", "v3"),
		)
	}
}

func BenchmarkSlogDefaultJSONBaseline(b *testing.B) {
	sl := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sl.LogAttrs(
			context.Background(),
			slog.LevelInfo,
			"hello",
			slog.String("service", "api"),
			slog.Int("status", 200),
		)
	}
}

func BenchmarkEventParallel(b *testing.B) {
	out := &benchmarkBlackhole{}
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: out})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("benchmark-event")
		}
	})

	b.StopTimer()

	if out.WriteCount() != uint64(b.N) {
		b.Fatalf("mismatch in write count: expected=%d actual=%d", b.N, out.WriteCount())
	}
}

func BenchmarkDisabledParallel(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: io.Discard})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Debug("benchmark-disabled")
		}
	})
}

func BenchmarkEventFmtParallel(b *testing.B) {
	out := &benchmarkBlackhole{}
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: out})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infof("benchmark-fmt %s %d", "event", 42)
		}
	})

	b.StopTimer()

	if out.WriteCount() != uint64(b.N) {
		b.Fatalf("mismatch in write count: expected=%d actual=%d", b.N, out.WriteCount())
	}
}

func BenchmarkDisabledFmtParallel(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: io.Discard})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Debugf("benchmark-fmt %s %d", "event", 42)
		}
	})
}

func BenchmarkEventCtxParallel(b *testing.B) {
	out := &benchmarkBlackhole{}
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: out})
	ctx := WithRequestID(context.Background(), "req-parallel")
	ctx = WithRequestMetadata(ctx, String("tenant", "acme"), String("region", "us-east-1"))

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.InfoCtx(ctx, "benchmark-ctx", String("service", "api"), Int("status", 200), Bool("ok", true))
		}
	})

	b.StopTimer()

	if out.WriteCount() != uint64(b.N) {
		b.Fatalf("mismatch in write count: expected=%d actual=%d", b.N, out.WriteCount())
	}
}

func BenchmarkDisabledCtxParallel(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: io.Discard})
	ctx := WithRequestID(context.Background(), "req-disabled")
	ctx = WithRequestMetadata(ctx, String("tenant", "acme"), String("region", "us-east-1"))

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.DebugCtx(ctx, "benchmark-disabled-ctx", String("service", "api"), Int("status", 200), Bool("ok", true))
		}
	})
}

func BenchmarkEventCtxWeakParallel(b *testing.B) {
	out := &benchmarkBlackhole{}
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: out})
	child := l.With(
		Any("service", "api"),
		Any("status", 200),
		Any("ok", true),
		Any("meta", map[string]any{"region": "us-east-1"}),
	)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			child.Info("benchmark-ctx-weak")
		}
	})

	b.StopTimer()

	if out.WriteCount() != uint64(b.N) {
		b.Fatalf("mismatch in write count: expected=%d actual=%d", b.N, out.WriteCount())
	}
}

func BenchmarkDisabledCtxWeakParallel(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: io.Discard})
	child := l.With(
		Any("service", "api"),
		Any("status", 200),
		Any("ok", true),
		Any("meta", map[string]any{"region": "us-east-1"}),
	)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			child.Debug("benchmark-disabled-ctx-weak")
		}
	})
}

func BenchmarkEventAccumulatedCtxParallel(b *testing.B) {
	out := &benchmarkBlackhole{}
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: out})
	child := l.With(String("service", "api"), String("region", "us-east-1"), Int("status", 200), Bool("ok", true))

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			child.Info("benchmark-accumulated")
		}
	})

	b.StopTimer()

	if out.WriteCount() != uint64(b.N) {
		b.Fatalf("mismatch in write count: expected=%d actual=%d", b.N, out.WriteCount())
	}
}

func BenchmarkDisabledAccumulatedCtxParallel(b *testing.B) {
	l := newBenchmarkLogger(b, Config{Level: LevelInfo, Format: FormatJSON, Writer: io.Discard})
	child := l.With(String("service", "api"), String("region", "us-east-1"), Int("status", 200), Bool("ok", true))

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			child.Debug("benchmark-disabled-accumulated")
		}
	})
}
