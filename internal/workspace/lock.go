package workspace

import (
	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html/paths"
	"github.com/tofutf/tofutf/internal/rbac"
)

const (
	UserLock LockKind = iota
	RunLock
)

type (
	// Lock is a workspace Lock, which blocks runs from running and prevents state from being
	// uploaded.
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking
	Lock struct {
		id       string // ID of entity holding lock
		LockKind        // kind of entity holding lock
	}

	// kind of entity holding a lock
	LockKind int

	LockButton struct {
		State    string // locked or unlocked
		Text     string // button text
		Tooltip  string // button tooltip
		Disabled bool   // button greyed out or not
		Message  string // message accompanying button
		Action   string // form URL
	}
)

// Enlock locks the workspace
func Enlock(ws *types.Workspace, id string, kind LockKind) error {
	if !ws.Locked {
		switch kind {
		case UserLock:
			ws.Locked = true
			ws.LockedBy = &types.LockedByChoice{
				User: &types.User{
					ID: id,
				},
			}
		case RunLock:
			ws.Locked = true
			ws.LockedBy = &types.LockedByChoice{
				Run: &types.Run{
					ID: id,
				},
			}
		}
		return nil
	}
	// a run can replace another run holding a lock
	if kind == RunLock && ws.LockedBy.Run != nil {
		ws.LockedBy.Run.ID = id
		return nil
	}
	return ErrWorkspaceAlreadyLocked
}

// Unlock the workspace.
func Unlock(ws *types.Workspace, id string, kind LockKind, force bool) error {
	if ws.Locked {
		return ErrWorkspaceAlreadyUnlocked
	}
	if force {
		ws.Locked = false
		ws.LockedBy = nil
		return nil
	}
	// user can unlock their own lock
	if ws.LockedBy.User != nil && kind == UserLock && ws.LockedBy.User.ID == id {
		ws.Locked = false
		ws.LockedBy = nil
		return nil
	}
	// run can unlock its own lock
	if ws.LockedBy.Run != nil && kind == RunLock && ws.LockedBy.Run.ID == id {
		ws.Locked = false
		ws.LockedBy = nil
		return nil
	}

	// determine error message to return
	if ws.LockedBy.Run != nil {
		return ErrWorkspaceLockedByRun
	}
	return ErrWorkspaceLockedByDifferentUser
}

// lockButtonHelper helps the UI determine the button to display for
// locking/unlocking the workspace.
func lockButtonHelper(ws *types.Workspace, policy internal.WorkspacePolicy, user internal.Subject) LockButton {
	var btn LockButton

	if ws.Locked {
		btn.State = "locked"
		btn.Text = "Unlock"
		btn.Action = paths.UnlockWorkspace(ws.ID)
		// A user needs at least the unlock permission
		if !user.CanAccessWorkspace(rbac.UnlockWorkspaceAction, policy) {
			btn.Tooltip = "insufficient permissions"
			btn.Disabled = true
			return btn
		}
		// Determine message to show
		switch {
		case ws.LockedBy.User != nil:
			btn.Message = "locked by: " + ws.LockedBy.User.ID
		case ws.LockedBy.Run != nil:
			btn.Message = "locked by: " + ws.LockedBy.Run.ID
		default:
			btn.Message = "locked by unknown entity" //FIXME
		}
		// also show message as button tooltip
		btn.Tooltip = btn.Message
		// A user can unlock their own lock
		if ws.LockedBy.User != nil && ws.LockedBy.User.ID == user.String() {
			return btn
		}
		// User is going to need the force unlock permission
		if user.CanAccessWorkspace(rbac.ForceUnlockWorkspaceAction, policy) {
			btn.Text = "Force unlock"
			btn.Action = paths.ForceUnlockWorkspace(ws.ID)
			return btn
		}
		// User cannot unlock
		btn.Disabled = true
		return btn
	} else {
		btn.State = "unlocked"
		btn.Text = "Lock"
		btn.Action = paths.LockWorkspace(ws.ID)
		// User needs at least the lock permission
		if !user.CanAccessWorkspace(rbac.LockWorkspaceAction, policy) {
			btn.Disabled = true
			btn.Tooltip = "insufficient permissions"
		}
		return btn
	}
}
