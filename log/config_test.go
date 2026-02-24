package log

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, LevelInfo, cfg.Level)
	assert.Equal(t, FormatText, cfg.Format)
	assert.Equal(t, os.Stdout, cfg.Writer)
	assert.NotNil(t, cfg.ExitFunc)
}

func TestConfigMergedAndValidate(t *testing.T) {
	tests := []configValidationCase{
		{name: "valid-minimal", cfg: Config{Level: LevelInfo, Format: FormatJSON, Writer: &bytes.Buffer{}}},
		{name: "invalid-level", cfg: Config{Level: Level(99), Format: FormatJSON, Writer: &bytes.Buffer{}}, wantErr: true},
		{name: "invalid-format", cfg: Config{Level: LevelInfo, Format: Format("bad"), Writer: &bytes.Buffer{}}, wantErr: true},
	}

	for i, tc := range tests {
		t.Run(testCaseName(testPrefixConfig, tc.name, i), func(t *testing.T) {
			merged := tc.cfg.merged()

			err := merged.validate()

			if tc.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}
