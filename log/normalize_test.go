package log

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeDedupAndSort(t *testing.T) {
	t.Run(testCaseName(testPrefixNormalize, "dedup-and-sort", 0), func(t *testing.T) {
		rec := slog.NewRecord(time.Unix(1700000000, 0), slog.LevelInfo, "msg", 0)
		rec.AddAttrs(String("b", "1"), String("a", "1"), String("b", "2"), String("", "drop"))

		norm := normalizeRecord(context.Background(), rec, nil, nil, "", false, "", OtelTrace{}, nil)
		require.Len(t, norm.Attrs, 2)
		assert.Equal(t, "a", norm.Attrs[0].Key)
		assert.Equal(t, "b", norm.Attrs[1].Key)
		assert.Equal(t, "2", norm.Attrs[1].Value)
	})
}

func TestNormalizeFlattensGroup(t *testing.T) {
	t.Run(testCaseName(testPrefixNormalize, "flattens-group", 1), func(t *testing.T) {
		rec := slog.NewRecord(time.Unix(1700000000, 0), slog.LevelInfo, "msg", 0)
		rec.AddAttrs(Group("req", String("id", "123"), Group("nested", String("k", "v"))))

		norm := normalizeRecord(context.Background(), rec, nil, nil, "", false, "", OtelTrace{}, nil)
		require.Len(t, norm.Attrs, 2)
		assert.Equal(t, "req.id", norm.Attrs[0].Key)
		assert.Equal(t, "req.nested.k", norm.Attrs[1].Key)
	})
}
