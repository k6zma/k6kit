package log

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestFieldConstructorsAndGroupFlattening(t *testing.T) {
	constructors := []fieldConstructorCase{
		{name: "Rune", field: Rune("rune", 'r'), key: "rune"},
		{name: "Byte", field: Byte("byte", 7), key: "byte"},
		{name: "Int", field: Int("int", 1), key: "int"},
		{name: "Int8", field: Int8("int8", 8), key: "int8"},
		{name: "Int16", field: Int16("int16", 16), key: "int16"},
		{name: "Int32", field: Int32("int32", 32), key: "int32"},
		{name: "Int64", field: Int64("int64", 64), key: "int64"},
		{name: "Uint8", field: Uint8("uint8", 8), key: "uint8"},
		{name: "Uint16", field: Uint16("uint16", 16), key: "uint16"},
		{name: "Uint32", field: Uint32("uint32", 32), key: "uint32"},
		{name: "Uint64", field: Uint64("uint64", 64), key: "uint64"},
		{name: "Float32", field: Float32("float32", 3.2), key: "float32"},
		{name: "Float64", field: Float64("float64", 6.4), key: "float64"},
		{name: "Bool", field: Bool("bool", true), key: "bool"},
		{name: "String", field: String("string", "value"), key: "string"},
		{name: "Duration", field: Duration("duration", 2*time.Second), key: "duration"},
		{name: "Time", field: Time("time_value", fixedTestTime()), key: "time_value"},
		{name: "Any", field: Any("any", map[string]any{"k": "v"}), key: "any"},
		{name: "Error", field: Error("error", errors.New("boom")), key: "error"},
		{name: "Bytes", field: Bytes("bytes", []byte{1, 2}), key: "bytes"},
		{name: "Strings", field: Strings("strings", []string{"a", "b"}), key: "strings"},
		{name: "Runes", field: Runes("runes", []rune{'x', 'y'}), key: "runes"},
		{name: "Bools", field: Bools("bools", []bool{true, false}), key: "bools"},
		{name: "Ints", field: Ints("ints", []int{1, 2}), key: "ints"},
		{name: "Int8s", field: Int8s("int8s", []int8{1, 2}), key: "int8s"},
		{name: "Int16s", field: Int16s("int16s", []int16{1, 2}), key: "int16s"},
		{name: "Int32s", field: Int32s("int32s", []int32{1, 2}), key: "int32s"},
		{name: "Int64s", field: Int64s("int64s", []int64{1, 2}), key: "int64s"},
		{name: "Uint8s", field: Uint8s("uint8s", []uint8{1, 2}), key: "uint8s"},
		{name: "Uint16s", field: Uint16s("uint16s", []uint16{1, 2}), key: "uint16s"},
		{name: "Uint32s", field: Uint32s("uint32s", []uint32{1, 2}), key: "uint32s"},
		{name: "Uint64s", field: Uint64s("uint64s", []uint64{1, 2}), key: "uint64s"},
		{name: "Float32s", field: Float32s("float32s", []float32{1.1, 2.2}), key: "float32s"},
		{name: "Float64s", field: Float64s("float64s", []float64{1.1, 2.2}), key: "float64s"},
		{name: "Errors", field: Errors("errors", []error{errors.New("e1"), nil}), key: "errors"},
		{name: "Group", field: Group("group", String("inner", "v")), key: "group"},
	}

	for i, tc := range constructors {
		tc := tc
		t.Run(testCaseName(testPrefixFieldLevel, tc.name, i), func(t *testing.T) {
			attr := tc.field.toAttr()
			if attr.Key != tc.key {
				t.Fatalf("unexpected attr key for %s: got=%q want=%q", tc.name, attr.Key, tc.key)
			}
		})
	}

	var out bytes.Buffer

	l, err := New(Config{Level: LevelDebug, Format: FormatJSON, Writer: &out, now: fixedTestTime})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l.Info(
		"all-fields",
		Rune("rune", 'r'),
		Byte("byte", 7),
		Int("int", 1),
		Int8("int8", 8),
		Int16("int16", 16),
		Int32("int32", 32),
		Int64("int64", 64),
		Uint8("uint8", 8),
		Uint16("uint16", 16),
		Uint32("uint32", 32),
		Uint64("uint64", 64),
		Float32("float32", 3.2),
		Float64("float64", 6.4),
		Bool("bool", true),
		String("string", "value"),
		Duration("duration", 2*time.Second),
		Time("time_value", fixedTestTime()),
		Any("any", map[string]any{"k": "v"}),
		Error("error", errors.New("boom")),
		Bytes("bytes", []byte{1, 2}),
		Strings("strings", []string{"a", "b"}),
		Runes("runes", []rune{'x', 'y'}),
		Bools("bools", []bool{true, false}),
		Ints("ints", []int{1, 2}),
		Int8s("int8s", []int8{1, 2}),
		Int16s("int16s", []int16{1, 2}),
		Int32s("int32s", []int32{1, 2}),
		Int64s("int64s", []int64{1, 2}),
		Uint8s("uint8s", []uint8{1, 2}),
		Uint16s("uint16s", []uint16{1, 2}),
		Uint32s("uint32s", []uint32{1, 2}),
		Uint64s("uint64s", []uint64{1, 2}),
		Float32s("float32s", []float32{1.1, 2.2}),
		Float64s("float64s", []float64{1.1, 2.2}),
		Errors("errors", []error{errors.New("e1"), nil}),
		Group("group", String("inner", "v"), Group("deep", Int("n", 1))),
	)

	obj := parseJSONLine(t, readOneLine(t, &out))

	keys := []string{
		"rune", "byte", "int", "int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "string", "duration", "time_value",
		"any", "error", "bytes", "strings", "runes", "bools",
		"ints", "int8s", "int16s", "int32s", "int64s",
		"uint8s", "uint16s", "uint32s", "uint64s",
		"float32s", "float64s", "errors",
		"group.inner", "group.deep.n",
	}

	for i, key := range keys {
		t.Run(testCaseName(testPrefixFieldLevel, fmt.Sprintf("json-key-%s", key), i), func(t *testing.T) {
			if _, ok := obj[key]; !ok {
				t.Fatalf("expected key %q in JSON output", key)
			}
		})
	}

	if got := obj["error"]; got != "boom" {
		t.Fatalf("error should render as message-only string: got=%v", got)
	}

	errSlice, ok := obj["errors"].([]any)
	if !ok || len(errSlice) != 2 || errSlice[0] != "e1" || errSlice[1] != "" {
		t.Fatalf("unexpected errors value: %#v", obj["errors"])
	}
}

func TestLevelStringAndParseLevel(t *testing.T) {
	stringCases := []levelStringCase{
		{name: "debug", level: LevelDebug, want: levelLabelDebug},
		{name: "info", level: LevelInfo, want: levelLabelInfo},
		{name: "warn", level: LevelWarn, want: levelLabelWarn},
		{name: "error", level: LevelError, want: levelLabelError},
		{name: "fatal", level: LevelFatal, want: levelLabelFatal},
		{name: "unknown-fallback", level: Level(999), want: levelLabelInfo},
	}

	for i, tc := range stringCases {
		t.Run(testCaseName(testPrefixFieldLevel, "level-string-"+tc.name, i), func(t *testing.T) {
			if got := tc.level.String(); got != tc.want {
				t.Fatalf("Level(%d).String() = %q, want %q", tc.level, got, tc.want)
			}
		})
	}

	parseCases := []parseLevelCase{
		{input: "debug", want: LevelDebug},
		{input: " INFO ", want: LevelInfo},
		{input: "Warn", want: LevelWarn},
		{input: "eRrOr", want: LevelError},
		{input: "FATAL", want: LevelFatal},
	}

	for i, tc := range parseCases {
		t.Run(testCaseName(testPrefixFieldLevel, "parse-level-ok", i), func(t *testing.T) {
			got, err := ParseLevel(tc.input)
			if err != nil {
				t.Fatalf("ParseLevel(%q) unexpected error: %v", tc.input, err)
			}

			if got != tc.want {
				t.Fatalf("ParseLevel(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}

	errCases := []parseLevelErrorCase{
		{name: "warning-alias-rejected", input: "warning", wantExact: `unsupported log level: "warning"`},
		{name: "unknown-rejected", input: " unknown ", wantContain: "unsupported log level"},
	}

	for i, tc := range errCases {
		t.Run(testCaseName(testPrefixFieldLevel, "parse-level-err-"+tc.name, i), func(t *testing.T) {
			_, err := ParseLevel(tc.input)
			if err == nil {
				t.Fatalf("ParseLevel(%s) expected error", tc.input)
			}

			if tc.wantExact != "" && err.Error() != tc.wantExact {
				t.Fatalf("unexpected ParseLevel error: got=%q want=%q", err.Error(), tc.wantExact)
			}

			if tc.wantContain != "" && !strings.Contains(err.Error(), tc.wantContain) {
				t.Fatalf("unexpected ParseLevel error: %v", err)
			}
		})
	}
}
