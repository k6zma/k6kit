package log

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
)

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

func benchmarkLogger(b *testing.B, format Format, level Level, out *benchmarkBlackhole) Logger {
	b.Helper()

	l, err := New(Config{
		Level:  level,
		Format: format,
		Writer: out,
	})
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}

	return l
}

func benchmarkFormats(b *testing.B, run func(b *testing.B, format Format)) {
	b.Run("json", func(b *testing.B) {
		run(b, FormatJSON)
	})

	b.Run("text", func(b *testing.B) {
		run(b, FormatText)
	})
}

func benchmarkFields(n int) []Field {
	fields := make([]Field, 0, n)
	for i := 0; i < n; i++ {
		fields = append(fields, String("k"+strconv.Itoa(i), "v"+strconv.Itoa(i)))
	}

	return fields
}

func benchmarkContext() context.Context {
	ctx := WithRequestID(context.Background(), "req-bench")

	return WithOtelTraceContext(ctx, "trace-bench", "span-bench")
}

func BenchmarkNoFields(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("no-fields")
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkOneField(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)
		field := String("service", "api")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("one-field", field)
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkFiveFields(b *testing.B) {
	fields := []Field{
		String("service", "api"),
		String("method", "GET"),
		Int("status", 200),
		Bool("ok", true),
		String("region", "eu-west"),
	}

	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("five-fields", fields...)
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkTenFields(b *testing.B) {
	fields := benchmarkFields(10)

	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("ten-fields", fields...)
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkHundredFields(b *testing.B) {
	fields := benchmarkFields(100)

	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("hundred-fields", fields...)
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkWithContext(b *testing.B) {
	ctx := benchmarkContext()
	fields := []Field{String("service", "api"), Int("status", 200)}

	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.InfoCtx(ctx, "with-context", fields...)
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkParallelNoFields(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("parallel-no-fields")
			}
		})

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkParallelTenFields(b *testing.B) {
	fields := benchmarkFields(10)

	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("parallel-ten-fields", fields...)
			}
		})

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkParallelWithContext(b *testing.B) {
	ctx := benchmarkContext()
	fields := []Field{String("service", "api"), Int("status", 200)}

	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelInfo, out)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.InfoCtx(ctx, "parallel-context", fields...)
			}
		})

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkParallelDisabled(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelWarn, out)

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Debug("parallel-disabled", String("k", "v"))
			}
		})

		if out.WriteCount() != 0 {
			b.Fatalf("disabled path must not write: got=%d", out.WriteCount())
		}
	})
}

func BenchmarkWithRequestID(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		base := benchmarkLogger(b, format, LevelInfo, out)
		l := base.WithRequestID("req-fixed")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("with-request-id", String("service", "api"))
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkWithOtelTrace(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		base := benchmarkLogger(b, format, LevelInfo, out)
		l := base.WithOtelTrace("trace-fixed", "span-fixed")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("with-otel-trace", String("service", "api"))
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkWithGroup(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		base := benchmarkLogger(b, format, LevelInfo, out)
		l := base.WithGroup("api").WithGroup("http")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("with-group", String("service", "api"))
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkWithErr(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		base := benchmarkLogger(b, format, LevelInfo, out)
		l := base.WithErr(errors.New("bench-error"))

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Info("with-err", String("service", "api"))
		}

		if out.WriteCount() != uint64(b.N) {
			b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
		}
	})
}

func BenchmarkTextRendererColored(b *testing.B) {
	out := &benchmarkBlackhole{}

	l, err := New(Config{Level: LevelInfo, Format: FormatText, Color: true, Writer: out})
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}

	fields := []Field{String("service", "api"), Int("status", 200), Bool("ok", true)}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Info("text-colored", fields...)
	}

	if out.WriteCount() != uint64(b.N) {
		b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
	}
}

func BenchmarkTextRendererNoColor(b *testing.B) {
	out := &benchmarkBlackhole{}

	l, err := New(Config{Level: LevelInfo, Format: FormatText, Color: false, Writer: out})
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}

	fields := []Field{String("service", "api"), Int("status", 200), Bool("ok", true)}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Info("text-no-color", fields...)
	}

	if out.WriteCount() != uint64(b.N) {
		b.Fatalf("write count mismatch: got=%d want=%d", out.WriteCount(), b.N)
	}
}

func BenchmarkDebugDisabled(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelWarn, out)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l.Debug("debug-disabled", String("k", "v"))
		}

		if out.WriteCount() != 0 {
			b.Fatalf("disabled path must not write: got=%d", out.WriteCount())
		}
	})
}

func BenchmarkDebugDisabledWithEnabledGuard(b *testing.B) {
	benchmarkFormats(b, func(b *testing.B, format Format) {
		out := &benchmarkBlackhole{}
		l := benchmarkLogger(b, format, LevelWarn, out)
		ctx := context.Background()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			if l.Enabled(ctx, LevelDebug) {
				l.Debug("debug-disabled-guard", String("k", strconv.Itoa(i)))
			}
		}

		if out.WriteCount() != 0 {
			b.Fatalf("guarded disabled path must not write: got=%d", out.WriteCount())
		}
	})
}
