package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testPrefixConfig      = "log-config-unit"
	testPrefixContext     = "log-context-unit"
	testPrefixFields      = "log-fields-unit"
	testPrefixNormalize   = "log-normalize-unit"
	testPrefixRender      = "log-render-unit"
	testPrefixLogger      = "log-logger-unit"
	testPrefixDefault     = "log-default-unit"
	testPrefixConcurrency = "log-concurrency-unit"
)

func testCaseName(prefix, name string, idx int) string {
	return fmt.Sprintf("[%s]-%s-№%d", prefix, name, idx+1)
}

func parseJSONLine(t *testing.T, raw string) map[string]any {
	t.Helper()

	var out map[string]any
	require.NoErrorf(t, json.Unmarshal([]byte(raw), &out), "json unmarshal error raw=%q", raw)

	return out
}

func splitNonEmptyLines(raw string) []string {
	buf := bytes.NewBufferString(raw)
	out := make([]string, 0, 8)

	for {
		line, err := buf.ReadString('\n')
		if len(line) > 0 {
			line = line[:len(line)-1]
			if line != "" {
				out = append(out, line)
			}
		}

		if err != nil {
			break
		}
	}

	return out
}

type discardWriter struct{}

type configValidationCase struct {
	name    string
	cfg     Config
	wantErr bool
}

type fieldConstructorCase struct {
	name  string
	field Field
	key   string
}

type panicVariantCase struct {
	name string
	run  func()
}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
