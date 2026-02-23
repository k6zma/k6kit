package log

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestLoggerWrapperMethodsAndLogf(t *testing.T) {
	var out bytes.Buffer

	exitCodes := make([]int, 0, 2)

	l, err := New(Config{
		Level:  LevelDebug,
		Format: FormatJSON,
		Writer: &out,
		now:    fixedTestTime,
		ExitFunc: func(code int) {
			exitCodes = append(exitCodes, code)
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := WithRequestID(context.Background(), "req-wrap")

	l.Warn("warn-msg", String("kind", "warn"))
	l.Error("error-msg", String("kind", "error"))
	l.Debugf("debugf-%d", 1)
	l.Infof("infof-%s", "ok")
	l.Warnf("warnf-%s", "ok")
	l.Errorf("errorf-%s", "ok")
	l.DebugCtx(ctx, "debug-ctx", String("scope", "ctx"))
	l.WarnCtx(ctx, "warn-ctx")
	l.ErrorCtx(ctx, "error-ctx")
	l.WithErr(errors.New("wrapped-boom")).Info("witherr")
	l.Fatalf("fatalf-%d", 9)
	l.FatalCtx(ctx, "fatal-ctx")

	lines := parseJSONLines(t, &out)
	if len(lines) != 12 {
		t.Fatalf("unexpected log line count: got=%d want=12", len(lines))
	}

	if len(exitCodes) != 2 || exitCodes[0] != 1 || exitCodes[1] != 1 {
		t.Fatalf("unexpected exit codes: %#v", exitCodes)
	}

	levelCases := []namedLevelCase{
		{name: "warn", msg: "warn-msg", wantLevel: levelLabelWarn},
		{name: "error", msg: "error-msg", wantLevel: levelLabelError},
		{name: "debugf", msg: "debugf-1", wantLevel: levelLabelDebug},
		{name: "infof", msg: "infof-ok", wantLevel: levelLabelInfo},
		{name: "warnf", msg: "warnf-ok", wantLevel: levelLabelWarn},
		{name: "errorf", msg: "errorf-ok", wantLevel: levelLabelError},
		{name: "debug-ctx", msg: "debug-ctx", wantLevel: levelLabelDebug},
		{name: "warn-ctx", msg: "warn-ctx", wantLevel: levelLabelWarn},
		{name: "error-ctx", msg: "error-ctx", wantLevel: levelLabelError},
		{name: "witherr", msg: "witherr", wantLevel: levelLabelInfo},
		{name: "fatalf", msg: "fatalf-9", wantLevel: levelLabelFatal},
		{name: "fatal-ctx", msg: "fatal-ctx", wantLevel: levelLabelFatal},
	}

	for i, tc := range levelCases {
		t.Run(testCaseName(testPrefixWrappers, fmt.Sprintf("level-%s", tc.name), i), func(t *testing.T) {
			line, ok := findByMsg(lines, tc.msg)
			if !ok {
				t.Fatalf("missing log line for message %q", tc.msg)
			}

			if got := line[keyLevel]; got != tc.wantLevel {
				t.Fatalf("unexpected level for %q: got=%v want=%q", tc.msg, got, tc.wantLevel)
			}
		})
	}

	warnLine, _ := findByMsg(lines, "warn-msg")
	if _, exists := warnLine[keyRequestID]; exists {
		t.Fatalf("non-context wrappers must not include request metadata: %#v", warnLine)
	}

	ctxLine, _ := findByMsg(lines, "warn-ctx")
	if got := ctxLine[keyRequestID]; got != "req-wrap" {
		t.Fatalf("context wrappers should include request_id: got=%v", got)
	}

	withErrLine, _ := findByMsg(lines, "witherr")
	if got := withErrLine[keyError]; got != "wrapped-boom" {
		t.Fatalf("WithErr should attach canonical error field: got=%v", got)
	}

	var gatedOut bytes.Buffer

	gated, err := New(Config{Level: LevelInfo, Format: FormatJSON, Writer: &gatedOut, now: fixedTestTime})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	gated.Debugf("hidden-%d", 42)

	if strings.TrimSpace(gatedOut.String()) != "" {
		t.Fatalf("disabled Debugf should produce no output, got %q", gatedOut.String())
	}
}
