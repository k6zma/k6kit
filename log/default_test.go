package log

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAndSetDefault(t *testing.T) {
	t.Run(testCaseName(testPrefixDefault, "default-and-setdefault", 0), func(t *testing.T) {
		original := Default()

		t.Cleanup(func() {
			SetDefault(original)
		})

		require.NotNil(t, Default())

		SetDefault(nil)

		assert.Equal(t, original, Default())

		var out bytes.Buffer

		l, err := New(Config{Level: LevelInfo, Format: FormatJSON, Writer: &out, TimeFormat: defaultJSONTimeFormat})
		require.NoError(t, err)

		SetDefault(l)
		Info("default-info", String("scope", "default"))

		lines := splitNonEmptyLines(out.String())
		require.NotEmpty(t, lines)
		obj := parseJSONLine(t, lines[0])
		assert.Equal(t, "default-info", obj[keyMsg])
		assert.Equal(t, "default", obj["scope"])
	})
}

func TestDefaultWrappersMatrix(t *testing.T) {
	t.Run(testCaseName(testPrefixDefault, "default-wrappers-matrix", 1), func(t *testing.T) {
		original := Default()

		t.Cleanup(func() {
			SetDefault(original)
		})

		var out bytes.Buffer

		exitCode := -1

		l, err := New(Config{
			Level:      LevelInfo,
			Format:     FormatJSON,
			Writer:     &out,
			TimeFormat: defaultJSONTimeFormat,
			ExitFunc: func(code int) {
				exitCode = code
			},
		})
		require.NoError(t, err)

		SetDefault(l)

		ctx := WithRequestID(context.Background(), "req-default")
		assert.False(t, Enabled(ctx, LevelDebug))

		Debug("hidden")
		Info("info")
		Warnf("warn-%d", 1)
		ErrorCtx(ctx, "error-ctx")
		Fatal("fatal")

		assert.Equal(t, 1, exitCode)

		lines := splitNonEmptyLines(out.String())
		assert.Len(t, lines, 4)
	})
}
