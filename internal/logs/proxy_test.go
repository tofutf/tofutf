package logs

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
)

// TestProxy_Get tests get() with and without a cached entry
func TestProxy_Get(t *testing.T) {
	ctx := context.Background()

	opts := internal.GetChunkOptions{
		RunID:  "run-123",
		Phase:  internal.PlanPhase,
		Offset: 3,
		Limit:  4,
	}

	t.Run("cache hit", func(t *testing.T) {
		cache := newFakeCache("run-123.plan.log", "hello world")
		proxy := &proxy{cache: cache}

		got, err := proxy.get(ctx, opts)
		require.NoError(t, err)

		want := internal.Chunk{RunID: "run-123", Phase: internal.PlanPhase, Offset: 3, Data: []byte("lo w")}
		assert.Equal(t, want, got)
	})

	t.Run("cache miss", func(t *testing.T) {
		db := &fakeDB{data: []byte("hello world")}
		cache := newFakeCache()
		proxy := &proxy{cache: cache, db: db}

		got, err := proxy.get(ctx, opts)
		require.NoError(t, err)

		want := internal.Chunk{RunID: "run-123", Phase: internal.PlanPhase, Offset: 3, Data: []byte("lo w")}
		assert.Equal(t, want, got)

		// cache should be populated now
		assert.Equal(t, "hello world", string(cache.cache["run-123.plan.log"]))
	})
}

func FuzzProxy_Get(f *testing.F) {
	ctx := context.Background()

	f.Add("run-1234", "hello world", 3, 4)

	f.Fuzz(func(t *testing.T, runID, payload string, offset, limit int) {
		if offset < 0 {
			t.Skip()
		}

		if limit <= 0 {
			t.Skip()
		}

		if offset+limit > len(payload) {
			t.Skip()
		}
		opts := internal.GetChunkOptions{
			RunID:  runID,
			Phase:  internal.PlanPhase,
			Offset: offset,
			Limit:  limit,
		}

		cache := newFakeCache(fmt.Sprintf("%s.plan.log", runID), payload)
		proxy := &proxy{cache: cache}

		got, err := proxy.get(ctx, opts)
		require.NoError(t, err)

		want := internal.Chunk{
			RunID:  runID,
			Phase:  internal.PlanPhase,
			Offset: offset,
			Data:   []byte(payload[offset : offset+limit]),
		}

		assert.Equal(t, want, got)
	})
}
