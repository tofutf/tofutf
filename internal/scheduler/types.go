package scheduler

import (
	"context"

	"github.com/tofutf/tofutf/internal/run"
	"github.com/tofutf/tofutf/internal/workspace"
)

// interfaces purely for faking purposes
type queueFactory interface {
	newQueue(opts queueOptions) eventHandler
}

type eventHandler interface {
	handleRun(context.Context, *run.Run) error
	handleWorkspace(context.Context, *workspace.Workspace) error
}
