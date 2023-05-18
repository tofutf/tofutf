package integration

import (
	"testing"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Events demonstrates events are triggered and successfully
// received by a subscriber.
func TestIntegration_Events(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	sub, err := daemon.Subscribe(ctx, "")
	require.NoError(t, err)

	org := daemon.createOrganization(t, ctx)
	ws := daemon.createWorkspace(t, ctx, org)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	run := daemon.createRun(t, ctx, ws, cv)

	assert.Equal(t, pubsub.NewCreatedEvent(org), <-sub)
	assert.Equal(t, pubsub.NewCreatedEvent(ws), <-sub)
	assert.Equal(t, pubsub.NewCreatedEvent(run), <-sub)
}
