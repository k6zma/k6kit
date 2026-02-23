package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	testPrefixLogger      = "logger"
	testPrefixConcurrency = "concurrency"
	testPrefixFieldLevel  = "field-level"
	testPrefixWrappers    = "wrappers"
	testPrefixBenchmark   = "benchmark"
)

type namedLevelCase struct {
	name      string
	msg       string
	wantLevel string
}

type namedKeyCase struct {
	name string
	key  string
}

type contextTestCase struct {
	name string
	run  func(t *testing.T, ctx context.Context)
}

type runOnlyCase struct {
	name string
	run  func(t *testing.T)
}

type enabledLevelCase struct {
	name string
	lvl  Level
	want bool
}

type configValidationCase struct {
	name string
	cfg  Config
}

type boolCheckCase struct {
	name string
	ok   bool
}

type fieldConstructorCase struct {
	name  string
	field Field
	key   string
}

type levelStringCase struct {
	name  string
	level Level
	want  string
}

type parseLevelCase struct {
	input string
	want  Level
}

type parseLevelErrorCase struct {
	name        string
	input       string
	wantExact   string
	wantContain string
}

func testCaseName(prefix, caseName string, i int) string {
	return fmt.Sprintf("[%s]-%s-№%d", prefix, caseName, i+1)
}

func fixedTestTime() time.Time {
	return time.Date(2026, 2, 23, 12, 34, 56, 0, time.UTC)
}

func readOneLine(t *testing.T, buf *bytes.Buffer) string {
	t.Helper()

	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		t.Fatal("expected one log line, got empty output")
	}

	if strings.Contains(raw, "\n") {
		t.Fatalf("expected one line, got multiple lines: %q", raw)
	}

	return raw
}

func splitNonEmptyLines(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), "\n")
	if len(parts) == 1 && parts[0] == "" {
		return nil
	}

	return parts
}

func parseJSONLine(t *testing.T, line string) map[string]any {
	t.Helper()

	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("failed to unmarshal JSON line: %v; line=%q", err, line)
	}

	return obj
}

func parseJSONLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	lines := splitNonEmptyLines(buf.String())
	if len(lines) == 0 {
		t.Fatal("expected log output, got empty buffer")
	}

	parsed := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		parsed = append(parsed, parseJSONLine(t, trimmed))
	}

	return parsed
}

func findByMsg(lines []map[string]any, msg string) (map[string]any, bool) {
	for _, line := range lines {
		if line[keyMsg] == msg {
			return line, true
		}
	}

	return nil, false
}

func jsonKeyOrder(t *testing.T, line string) []string {
	t.Helper()

	dec := json.NewDecoder(strings.NewReader(line))

	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("failed to read opening token: %v", err)
	}

	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		t.Fatalf("expected JSON object, got token=%v", tok)
	}

	keys := make([]string, 0, 16)

	for dec.More() {
		ktok, err := dec.Token()
		if err != nil {
			t.Fatalf("failed to read key token: %v", err)
		}

		key, ok := ktok.(string)
		if !ok {
			t.Fatalf("expected string key, got token=%v", ktok)
		}

		keys = append(keys, key)

		var skip any
		if err := dec.Decode(&skip); err != nil {
			t.Fatalf("failed to decode value for key %q: %v", key, err)
		}
	}

	if _, err := dec.Token(); err != nil {
		t.Fatalf("failed to read closing token: %v", err)
	}

	return keys
}

func assertTextTimestamp(t *testing.T, line string) {
	t.Helper()

	if !strings.HasPrefix(line, "[") {
		t.Fatalf("text log line does not start with '[': %q", line)
	}

	end := strings.IndexByte(line, ']')
	if end <= 1 {
		t.Fatalf("text log line missing timestamp closing bracket: %q", line)
	}

	ts := line[1:end]
	if _, err := time.Parse(textTimeFormat, ts); err != nil {
		t.Fatalf("timestamp %q does not match text layout %q: %v", ts, textTimeFormat, err)
	}
}

func assertJSONTimestamp(t *testing.T, obj map[string]any) {
	t.Helper()

	raw, ok := obj[keyTime].(string)
	if !ok {
		t.Fatalf("missing or non-string %q field: %T", keyTime, obj[keyTime])
	}

	if _, err := time.Parse(jsonTimeFormat, raw); err != nil {
		t.Fatalf("timestamp %q does not match JSON layout %q: %v", raw, jsonTimeFormat, err)
	}
}
