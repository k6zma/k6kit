package log

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRendererCanonicalAndAttrConflict(t *testing.T) {
	t.Run(testCaseName(testPrefixRender, "json-canonical-and-conflict", 0), func(t *testing.T) {
		r := jsonRenderer{timeFormat: defaultJSONTimeFormat}
		rec := normalizedRecord{
			Time:  time.Unix(1700000000, 0),
			Level: levelLabelInfo,
			Msg:   "hello",
			Attrs: []normalizedAttr{{Key: keyLevel, Value: "custom"}, {Key: "service", Value: "api"}},
		}

		out, err := r.render(nil, rec)
		require.NoError(t, err)

		line := string(out)
		assert.True(t, strings.HasSuffix(line, "\n"))

		obj := parseJSONLine(t, line)
		_, ok := obj["attr."+keyLevel]
		assert.True(t, ok)

		assert.Equal(t, levelLabelInfo, obj[keyLevel])
	})
}

func TestTextRendererColoredAndPlain(t *testing.T) {
	t.Run(testCaseName(testPrefixRender, "text-colored-and-plain", 1), func(t *testing.T) {
		rec := normalizedRecord{
			Time:  time.Unix(1700000000, 0),
			Level: levelLabelWarn,
			Msg:   "text",
			Group: "api.http",
			Attrs: []normalizedAttr{{Key: "service", Value: "api"}},
		}

		plain, err := (textRenderer{color: false, timeFormat: defaultTextTimeFormat}).render(nil, rec)
		require.NoError(t, err)

		assert.False(t, strings.Contains(string(plain), "\u001b["))

		color, err := (textRenderer{color: true, timeFormat: defaultTextTimeFormat}).render(nil, rec)
		require.NoError(t, err)

		assert.True(t, strings.Contains(string(color), "\u001b["))
	})
}

func TestJSONRendererNonFiniteFloatsFallbackToNull(t *testing.T) {
	t.Run(testCaseName(testPrefixRender, "json-nonfinite-floats-null", 2), func(t *testing.T) {
		r := jsonRenderer{timeFormat: defaultJSONTimeFormat}
		rec := normalizedRecord{
			Time:  time.Unix(1700000000, 0),
			Level: levelLabelInfo,
			Msg:   "non-finite",
			Attrs: []normalizedAttr{
				{Key: "nan", Value: math.NaN()},
				{Key: "pos_inf", Value: math.Inf(1)},
				{Key: "neg_inf", Value: float32(math.Inf(-1))},
			},
		}

		out, err := r.render(nil, rec)
		require.NoError(t, err)

		obj := parseJSONLine(t, string(out))
		assert.Nil(t, obj["nan"])
		assert.Nil(t, obj["pos_inf"])
		assert.Nil(t, obj["neg_inf"])
	})
}
