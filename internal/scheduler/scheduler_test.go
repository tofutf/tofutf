package scheduler

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal/pubsub"
	"github.com/tofutf/tofutf/internal/run"
	"github.com/tofutf/tofutf/internal/workspace"
	"github.com/tofutf/tofutf/internal/xslog"
)

// TestScheduler checks the scheduler is creating workspace queues and
// forwarding events to the queue handlers.
func TestScheduler(t *testing.T) {
	ctx := context.Background()

	t.Run("create workspace queue", func(t *testing.T) {
		qf := &fakeQueueFactory{}
		scheduler := scheduler{
			logger:       slog.New(&xslog.NoopHandler{}),
			queues:       make(map[string]eventHandler),
			queueFactory: qf,
		}
		want := &workspace.Workspace{ID: "ws-123"}
		err := scheduler.handleWorkspaceEvent(ctx, pubsub.Event[*workspace.Workspace]{
			Payload: want,
		})
		require.NoError(t, err)

		assert.Equal(t, 1, len(scheduler.queues))
		assert.Equal(t, want, qf.q.gotWorkspace)
	})

	t.Run("delete workspace queue", func(t *testing.T) {
		scheduler := scheduler{
			logger: slog.New(&xslog.NoopHandler{}),
			queues: map[string]eventHandler{
				"ws-123": &fakeQueue{},
			},
		}
		err := scheduler.handleWorkspaceEvent(ctx, pubsub.Event[*workspace.Workspace]{
			Payload: &workspace.Workspace{ID: "ws-123"},
			Type:    pubsub.DeletedEvent,
		})
		require.NoError(t, err)

		assert.Equal(t, 0, len(scheduler.queues))
	})

	t.Run("relay run to queue", func(t *testing.T) {
		q := &fakeQueue{}
		want := &run.Run{WorkspaceID: "ws-123"}
		scheduler := scheduler{
			logger: slog.New(&xslog.NoopHandler{}),
			queues: map[string]eventHandler{
				"ws-123": q,
			},
		}
		err := scheduler.handleRunEvent(ctx, pubsub.Event[*run.Run]{
			Payload: want,
		})
		require.NoError(t, err)
		assert.Equal(t, want, q.gotRun)
	})
}
