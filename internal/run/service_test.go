package run

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/pubsub"
	"github.com/tofutf/tofutf/internal/xslog"
)

func TestService_Watch(t *testing.T) {
	// input event channel
	in := make(chan pubsub.Event[*Run], 1)

	svc := &Service{
		site:   internal.NewAllowAllAuthorizer(),
		logger: slog.New(&xslog.NoopHandler{}),
		broker: &fakeSubService{ch: in},
	}

	// inject input event
	want := pubsub.Event[*Run]{Payload: &Run{}}
	in <- want

	got, err := svc.watchWithOptions(context.Background(), WatchOptions{})
	require.NoError(t, err)

	assert.Equal(t, want, <-got)
}
