package log

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentLoggingJSONLinesAreStable(t *testing.T) {
	t.Run(testCaseName(testPrefixConcurrency, "json-lines-are-stable", 0), func(t *testing.T) {
		const workers = 16

		const perWorker = 200

		var out bytes.Buffer

		l, err := New(Config{Level: LevelInfo, Format: FormatJSON, Writer: &out, TimeFormat: defaultJSONTimeFormat})
		require.NoError(t, err)

		var wg sync.WaitGroup

		for w := 0; w < workers; w++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				child := l.With(Int("worker", w)).WithGroup("jobs")

				for i := 0; i < perWorker; i++ {
					child.Info("concurrent", Int("n", i))
				}
			}()
		}

		wg.Wait()

		lines := splitNonEmptyLines(out.String())
		want := workers * perWorker

		require.Len(t, lines, want)

		for i, line := range lines {
			obj := parseJSONLine(t, line)
			assert.Equal(t, "concurrent", obj[keyMsg], "line %d message mismatch", i)
			assert.Equal(t, "jobs", obj[keyGroup], "line %d group mismatch", i)
		}
	})
}
