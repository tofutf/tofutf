package pubsub

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/xslog"
)

type foo struct {
	id string
}

func fooGetter(ctx context.Context, id string, action sql.Action) (*foo, error) {
	return &foo{id: id}, nil
}

func TestBroker_Subscribe(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](slog.New(&xslog.NoopHandler{}), &fakeListener{}, "foos", nil)

	sub, unsub := broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	unsub()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_UnsubscribeViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	broker := NewBroker[*foo](slog.New(&xslog.NoopHandler{}), &fakeListener{}, "foos", nil)

	sub, _ := broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	cancel()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_forward(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](slog.New(&xslog.NoopHandler{}), &fakeListener{}, "foos", fooGetter)

	sub, unsub := broker.Subscribe(ctx)
	defer unsub()

	broker.forward(ctx, "bar", sql.InsertAction)
	want := Event[*foo]{
		Type:    CreatedEvent,
		Payload: &foo{id: "bar"},
	}
	assert.Equal(t, want, <-sub)
}

func TestBroker_UnsubscribeFullSubscriber(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](slog.New(&xslog.NoopHandler{}), &fakeListener{}, "foos", fooGetter)

	broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	// deliberating publish more than subBufferSize events to trigger broker to
	// unsubscribe the sub
	for i := 0; i < subBufferSize+1; i++ {
		broker.forward(ctx, "bar", sql.InsertAction)
	}
	assert.Equal(t, 0, len(broker.subs))
}
