package scheduler

import (
	"context"

	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal/run"
)

type fakeQueueFactory struct {
	q *fakeQueue
}

func (f *fakeQueueFactory) newQueue(queueOptions) eventHandler {
	f.q = &fakeQueue{}
	return f.q
}

type fakeQueue struct {
	gotWorkspace *types.Workspace
	gotRun       *run.Run
}

func (q *fakeQueue) handleWorkspace(ctx context.Context, ws *types.Workspace) error {
	q.gotWorkspace = ws
	return nil
}

func (q *fakeQueue) handleRun(ctx context.Context, run *run.Run) error {
	q.gotRun = run
	return nil
}
