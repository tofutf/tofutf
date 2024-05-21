package workspace

import (
	"testing"

	types "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/rbac"
)

func TestWorkspace_Lock(t *testing.T) {
	t.Run("lock an unlocked workspace", func(t *testing.T) {
		ws := &types.Workspace{}
		err := Enlock(ws, "janitor", UserLock)
		require.NoError(t, err)
		assert.True(t, ws.Locked)
	})
	t.Run("replace run lock with another run lock", func(t *testing.T) {
		ws := &types.Workspace{
			Locked: true,
			LockedBy: &types.LockedByChoice{
				Run: &types.Run{
					ID: "run-123",
				},
			},
		}
		err := Enlock(ws, "run-456", RunLock)
		require.NoError(t, err)
		assert.True(t, ws.Locked)
	})
	t.Run("user cannot lock a locked workspace", func(t *testing.T) {
		ws := &types.Workspace{
			Locked: true,
			LockedBy: &types.LockedByChoice{
				Run: &types.Run{
					ID: "run-123",
				},
			},
		}
		err := Enlock(ws, "janitor", UserLock)
		require.Equal(t, ErrWorkspaceAlreadyLocked, err)
	})
}

func TestWorkspace_Unlock(t *testing.T) {
	t.Run("cannot unlock workspace already unlocked", func(t *testing.T) {
		err := Unlock(&types.Workspace{}, "janitor", UserLock, false)
		require.Equal(t, ErrWorkspaceAlreadyUnlocked, err)
	})
	t.Run("user can unlock their own lock", func(t *testing.T) {
		ws := &types.Workspace{
			Locked: true,
			LockedBy: &types.LockedByChoice{
				User: &types.User{
					ID: "janitor",
				},
			},
		}
		err := Unlock(ws, "janitor", UserLock, false)
		require.NoError(t, err)
		assert.False(t, ws.Locked)
	})
	t.Run("user cannot unlock another user's lock", func(t *testing.T) {
		ws := &types.Workspace{
			Locked: true,
			LockedBy: &types.LockedByChoice{
				User: &types.User{
					ID: "janitor",
				},
			},
		}
		err := Unlock(ws, "burglar", UserLock, false)
		require.Equal(t, ErrWorkspaceLockedByDifferentUser, err)
	})
	t.Run("user can unlock a lock by force", func(t *testing.T) {
		ws := &types.Workspace{
			Locked: true,
			LockedBy: &types.LockedByChoice{
				User: &types.User{
					ID: "janitor",
				},
			},
		}
		err := Unlock(ws, "headmaster", UserLock, true)
		require.NoError(t, err)
		assert.False(t, ws.Locked)
	})
	t.Run("run can unlock its own lock", func(t *testing.T) {
		ws := &types.Workspace{
			Locked: true,
			LockedBy: &types.LockedByChoice{
				Run: &types.Run{
					ID: "run-123",
				},
			},
		}
		err := Unlock(ws, "run-123", RunLock, false)
		require.NoError(t, err)
		assert.False(t, ws.Locked)
	})
}

func TestWorkspace_LockButtonHelper(t *testing.T) {
	tests := []struct {
		name    string
		ws      *types.Workspace
		subject *fakeSubject
		want    LockButton
	}{
		{
			"unlocked state",
			&types.Workspace{ID: "ws-123"},
			&fakeSubject{canLock: true},
			LockButton{
				State:  "unlocked",
				Text:   "Lock",
				Action: "/app/workspaces/ws-123/lock",
			},
		},
		{
			"insufficient permissions to lock",
			&types.Workspace{ID: "ws-123"},
			&fakeSubject{},
			LockButton{
				State:    "unlocked",
				Text:     "Lock",
				Tooltip:  "insufficient permissions",
				Action:   "/app/workspaces/ws-123/lock",
				Disabled: true,
			},
		},
		{
			"insufficient permissions to unlock",
			&types.Workspace{
				Locked: true,
				LockedBy: &types.LockedByChoice{
					User: &types.User{
						ID: "janitor",
					},
				},
			},
			&fakeSubject{},
			LockButton{
				State:    "locked",
				Text:     "Unlock",
				Tooltip:  "insufficient permissions",
				Action:   "/app/workspaces//unlock",
				Disabled: true,
			},
		},
		{
			"user can unlock their own lock",
			&types.Workspace{
				Locked: true,
				LockedBy: &types.LockedByChoice{
					User: &types.User{
						ID: "janitor",
					},
				},
			},
			&fakeSubject{id: "janitor", canUnlock: true},
			LockButton{
				State:   "locked",
				Text:    "Unlock",
				Message: "locked by: janitor",
				Tooltip: "locked by: janitor",
				Action:  "/app/workspaces//unlock",
			},
		},
		{
			"can unlock lock held by a different user",
			&types.Workspace{
				Locked: true,
				LockedBy: &types.LockedByChoice{
					User: &types.User{
						ID: "janitor",
					},
				},
			},
			&fakeSubject{id: "burglar", canUnlock: true},
			LockButton{
				State:    "locked",
				Text:     "Unlock",
				Action:   "/app/workspaces//unlock",
				Message:  "locked by: janitor",
				Tooltip:  "locked by: janitor",
				Disabled: true,
			},
		},
		{
			"user can force unlock",
			&types.Workspace{
				Locked: true,
				LockedBy: &types.LockedByChoice{
					User: &types.User{
						ID: "janitor",
					},
				},
			},
			&fakeSubject{id: "headmaster", canUnlock: true, canForceUnlock: true},
			LockButton{
				State:   "locked",
				Text:    "Force unlock",
				Action:  "/app/workspaces//force-unlock",
				Message: "locked by: janitor",
				Tooltip: "locked by: janitor",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lockButtonHelper(tt.ws, internal.WorkspacePolicy{}, tt.subject)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeSubject struct {
	id                                 string
	canUnlock, canForceUnlock, canLock bool

	internal.Subject
}

func (f *fakeSubject) String() string { return f.id }

func (f *fakeSubject) CanAccessWorkspace(action rbac.Action, _ internal.WorkspacePolicy) bool {
	switch action {
	case rbac.UnlockWorkspaceAction:
		return f.canUnlock
	case rbac.ForceUnlockWorkspaceAction:
		return f.canForceUnlock
	case rbac.LockWorkspaceAction:
		return f.canLock
	default:
		return false

	}
}
