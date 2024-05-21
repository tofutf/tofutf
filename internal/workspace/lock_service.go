package workspace

import (
	"context"
	"fmt"

	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/user"
)

// Lock locks the workspace. A workspace can only be locked on behalf of a run or a
// user. If the former then runID must be populated. Otherwise a user is
// extracted from the context.
func (s *Service) Lock(ctx context.Context, workspaceID string, runID *string) (*types.Workspace, error) {
	var (
		id   string
		kind LockKind
	)
	if runID != nil {
		id = *runID
		kind = RunLock
	} else {
		subject, err := s.CanAccess(ctx, rbac.LockWorkspaceAction, workspaceID)
		if err != nil {
			return nil, err
		}
		user, ok := subject.(*user.User)
		if !ok {
			return nil, fmt.Errorf("only a run or a user can lock a workspace")
		}
		id = user.Username
		kind = UserLock
	}

	ws, err := s.db.toggleLock(ctx, workspaceID, func(ws *types.Workspace) error {
		return Enlock(ws, id, kind)
	})
	if err != nil {
		s.logger.Error("locking workspace", "subject", id, "workspace", workspaceID, "err", err)
		return nil, err
	}

	s.logger.Info("locked workspace", "subject", id, "workspace", workspaceID)

	return ws, nil
}

// Unlock unlocks the workspace. A workspace can only be unlocked on behalf of a run or
// a user. If the former then runID must be non-nil; otherwise a user is
// extracted from the context.
func (s *Service) Unlock(ctx context.Context, workspaceID string, runID *string, force bool) (*types.Workspace, error) {
	var (
		id   string
		kind LockKind
	)
	if runID != nil {
		id = *runID
		kind = RunLock
	} else {
		var action rbac.Action
		if force {
			action = rbac.ForceUnlockWorkspaceAction
		} else {
			action = rbac.UnlockWorkspaceAction
		}
		subject, err := s.CanAccess(ctx, action, workspaceID)
		if err != nil {
			return nil, err
		}
		user, ok := subject.(*user.User)
		if !ok {
			return nil, fmt.Errorf("only a run or a user can unlock a workspace")
		}
		id = user.Username
		kind = UserLock
	}

	ws, err := s.db.toggleLock(ctx, workspaceID, func(ws *types.Workspace) error {
		return Unlock(ws, id, kind, force)
	})
	if err != nil {
		s.logger.Error("unlocking workspace", "subject", id, "workspace", workspaceID, "forced", force, "err", err)
		return nil, err
	}
	s.logger.Info("unlocked workspace", "subject", id, "workspace", workspaceID, "forced", force)

	return ws, nil
}
