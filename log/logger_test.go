package log

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestJSONFormattingAndBehavior(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{
		Level:   LevelDebug,
		Format:  FormatJSON,
		Writer:  &out,
		AppName: "k6kit",
		Env:     "test",
		Version: "v1",
		now:     fixedTestTime,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := WithRequestID(context.Background(), "req-123")
	ctx = WithRequestMetadata(ctx, String("meta", "from-ctx"))

	l.WithGroup("api").With(String("base", "child")).InfoCtx(
		ctx,
		"hello",
		String("a", "first"),
		String("dup", "one"),
		Group("grp", String("k", "v"), Group("inner", String("x", "y"))),
		Error("err", errors.New("boom")),
		Errors("errs", []error{errors.New("e1"), nil, errors.New("e3")}),
		String("dup", "two"),
	)

	line := readOneLine(t, &out)
	obj := parseJSONLine(t, line)
	assertJSONTimestamp(t, obj)

	if got := obj[keyLevel]; got != levelLabelInfo {
		t.Fatalf("unexpected %q: got=%v want=%q", keyLevel, got, levelLabelInfo)
	}

	if got := obj[keyGroup]; got != "api" {
		t.Fatalf("unexpected %q: got=%v want=%q", keyGroup, got, "api")
	}

	if _, exists := obj["api.base"]; exists {
		t.Fatalf("logger group must not prefix keys, found %q", "api.base")
	}

	if got := obj["base"]; got != "child" {
		t.Fatalf("unexpected base attr: got=%v want=%q", got, "child")
	}

	if got := obj[keyRequestID]; got != "req-123" {
		t.Fatalf("unexpected request_id: got=%v want=%q", got, "req-123")
	}

	if got := obj["dup"]; got != "two" {
		t.Fatalf("duplicate key should keep latest value: got=%v want=%q", got, "two")
	}

	if got := obj["err"]; got != "boom" {
		t.Fatalf("error should render as message-only string: got=%v", got)
	}

	errs, ok := obj["errs"].([]any)
	if !ok {
		t.Fatalf("errs should be []any, got %T", obj["errs"])
	}

	if len(errs) != 3 || errs[0] != "e1" || errs[1] != "" || errs[2] != "e3" {
		t.Fatalf("unexpected errors payload: %#v", errs)
	}

	gotOrder := jsonKeyOrder(t, line)
	wantOrder := []string{
		keyTime,
		keyLevel,
		keyGroup,
		keyMsg,
		keyApp,
		keyEnv,
		keyVersion,
		"base",
		keyRequestID,
		"meta",
		"a",
		"dup",
		"grp.k",
		"grp.inner.x",
		"err",
		"errs",
	}

	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("unexpected JSON key order:\n got: %#v\nwant: %#v", gotOrder, wantOrder)
	}
}

func TestTextFormattingAndErrorSerialization(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{
		Level:  LevelDebug,
		Format: FormatText,
		Writer: &out,
		now:    fixedTestTime,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l.WithGroup("worker").Info(
		"running",
		Error("error", errors.New("boom")),
		Errors("errors", []error{errors.New("e1"), errors.New("e2")}),
		Int("a", 1),
		Int("b", 2),
		Int("a", 3),
	)

	line := readOneLine(t, &out)
	assertTextTimestamp(t, line)

	parts := []string{"[INFO]", "[worker]", " running ", "{error=boom}", "{errors=[e1 e2]}", "{a=3},{b=2}"}
	for i, part := range parts {
		t.Run(testCaseName(testPrefixLogger, fmt.Sprintf("text-contains-%d", i+1), i), func(t *testing.T) {
			if !strings.Contains(line, part) {
				t.Fatalf("text log missing %q in %q", part, line)
			}
		})
	}
}

func TestContextHelpers(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{Format: FormatJSON, Writer: &out, now: fixedTestTime})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var nilCtx context.Context

	ctx := WithLogger(nilCtx, l)

	tests := []contextTestCase{
		{
			name: "from-context-injected",
			run: func(t *testing.T, ctx context.Context) {
				if got := FromContext(ctx, nil); got != l {
					t.Fatal("FromContext should return injected logger")
				}
			},
		},
		{
			name: "from-context-fallback",
			run: func(t *testing.T, _ context.Context) {
				if got := FromContext(nilCtx, l); got != l {
					t.Fatal("FromContext(nil, fallback) should return fallback")
				}
			},
		},
		{
			name: "request-id-and-metadata-copy",
			run: func(t *testing.T, base context.Context) {
				withValues := WithRequestID(base, "r-42")
				withValues = WithRequestMetadata(withValues, String("tenant", "acme"))

				if rid, ok := RequestID(withValues); !ok || rid != "r-42" {
					t.Fatalf("RequestID() = (%q, %v), want (%q, true)", rid, ok, "r-42")
				}

				md := RequestMetadata(withValues)
				if len(md) != 1 {
					t.Fatalf("RequestMetadata length = %d, want 1", len(md))
				}

				md[0] = String("tenant", "mutated")

				md2 := RequestMetadata(withValues)
				if md2[0].toAttr().Value.String() != "acme" {
					t.Fatal("RequestMetadata must return defensive copy")
				}
			},
		},
	}

	for i, tc := range tests {
		t.Run(testCaseName(testPrefixLogger, tc.name, i), func(t *testing.T) {
			tc.run(t, ctx)
		})
	}
}

func TestOTELAndSourceToggles(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled, Remote: true})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	var withFeatures bytes.Buffer

	l1, err := New(Config{
		Format:            FormatJSON,
		EnableOTEL:        true,
		EnableSourceTrace: true,
		Writer:            &withFeatures,
		now:               fixedTestTime,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l1.InfoCtx(ctx, "features-on")

	obj1 := parseJSONLine(t, readOneLine(t, &withFeatures))
	withCases := []runOnlyCase{
		{
			name: "otel-keys-present",
			run: func(t *testing.T) {
				if obj1[keyTraceID] == nil || obj1[keySpanID] == nil {
					t.Fatalf("expected trace keys when OTEL enabled: %#v", obj1)
				}
			},
		},
		{
			name: "source-trace-present",
			run: func(t *testing.T) {
				src, ok := obj1[keySourceTrace].(string)
				if !ok || src == "" || !strings.Contains(src, "logger_test.go:") {
					t.Fatalf("expected source trace to include test callsite, got %v", obj1[keySourceTrace])
				}
			},
		},
	}

	for i, tc := range withCases {
		t.Run(testCaseName(testPrefixLogger, tc.name, i), tc.run)
	}

	var withoutFeatures bytes.Buffer

	l2, err := New(Config{
		Format:            FormatJSON,
		EnableOTEL:        false,
		EnableSourceTrace: false,
		Writer:            &withoutFeatures,
		now:               fixedTestTime,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l2.InfoCtx(ctx, "features-off")

	obj2 := parseJSONLine(t, readOneLine(t, &withoutFeatures))

	withoutCases := []namedKeyCase{
		{name: "trace-id-omitted", key: keyTraceID},
		{name: "span-id-omitted", key: keySpanID},
		{name: "source-trace-omitted", key: keySourceTrace},
	}

	for i, tc := range withoutCases {
		t.Run(testCaseName(testPrefixLogger, tc.name, i), func(t *testing.T) {
			if _, exists := obj2[tc.key]; exists {
				t.Fatalf("%s must be omitted when feature disabled: %#v", tc.key, obj2)
			}
		})
	}
}

func TestSourceTraceUsesUserCallSite(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{
		Format:            FormatJSON,
		EnableSourceTrace: true,
		Writer:            &out,
		now:               fixedTestTime,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tmp := bytes.NewBufferString("abc")

	l.Info("info-source")
	l.Infof("len=%d", tmp.Len())

	lines := splitNonEmptyLines(out.String())
	if len(lines) != 2 {
		t.Fatalf("line count mismatch: got=%d want=2", len(lines))
	}

	for i, line := range lines {
		obj := parseJSONLine(t, line)

		src, ok := obj[keySourceTrace].(string)
		if !ok || src == "" {
			t.Fatalf("missing source trace for line %d: %#v", i, obj)
		}

		if strings.Contains(src, "bytes.(*Buffer).Len") {
			t.Fatalf("source trace leaked argument helper frame: %q", src)
		}

		if !strings.Contains(src, "logger_test.go:") {
			t.Fatalf("source trace must point to caller in test file: %q", src)
		}
	}
}

func TestFatalUsesExitHook(t *testing.T) {
	var out bytes.Buffer

	called := false
	code := 0

	l, err := New(Config{
		Format: FormatJSON,
		Writer: &out,
		now:    fixedTestTime,
		ExitFunc: func(c int) {
			called = true
			code = c
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l.Fatal("fatal-msg", String("reason", "test"))

	if !called || code != 1 {
		t.Fatalf("fatal exit hook mismatch: called=%v code=%d", called, code)
	}

	obj := parseJSONLine(t, readOneLine(t, &out))
	if obj[keyLevel] != levelLabelFatal {
		t.Fatalf("fatal log level mismatch: got=%v want=%q", obj[keyLevel], levelLabelFatal)
	}

	if obj[keyMsg] != "fatal-msg" {
		t.Fatalf("fatal message mismatch: got=%v", obj[keyMsg])
	}
}

func TestDisabledLevelBehavior(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Writer: &out,
		now:    fixedTestTime,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	enabledCases := []enabledLevelCase{
		{name: "debug-disabled", lvl: LevelDebug, want: false},
		{name: "info-enabled", lvl: LevelInfo, want: true},
	}

	for i, tc := range enabledCases {
		t.Run(testCaseName(testPrefixLogger, tc.name, i), func(t *testing.T) {
			if got := l.Enabled(context.Background(), tc.lvl); got != tc.want {
				t.Fatalf("Enabled(%v) = %v, want %v", tc.lvl, got, tc.want)
			}
		})
	}

	l.Debug("hidden", String("k", "v"))

	if out.Len() != 0 {
		t.Fatalf("disabled log level should produce no output, got: %q", out.String())
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []configValidationCase{
		{name: "unsupported-level", cfg: Config{Level: Level(123), Format: FormatJSON, Writer: &bytes.Buffer{}}},
		{name: "unsupported-format", cfg: Config{Level: LevelInfo, Format: Format("xml"), Writer: &bytes.Buffer{}}},
	}

	for i, tc := range tests {
		t.Run(testCaseName(testPrefixLogger, tc.name, i), func(t *testing.T) {
			if _, err := New(tc.cfg); err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}
