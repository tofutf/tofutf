package run

import (
	"context"
	"log/slog"
	"testing"

	types "github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal/configversion"
	"github.com/tofutf/tofutf/internal/vcs"
	"github.com/tofutf/tofutf/internal/workspace"
	"github.com/tofutf/tofutf/internal/xslog"
)

func TestSpawner(t *testing.T) {
	tests := []struct {
		name string
		ws   *types.Workspace
		// incoming event
		event vcs.Event
		// file paths to return from stubbed client.ListPullRequestFiles
		pullFiles []string
		// want spawned run
		spawn bool
	}{
		{
			name: "spawn run for push to default branch",
			ws:   &types.Workspace{VCSRepo: &types.VCSRepo{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:          vcs.EventTypePush,
					Action:        vcs.ActionCreated,
					Branch:        "main",
					DefaultBranch: "main",
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push to non-default branch",
			ws:   &types.Workspace{VCSRepo: &types.VCSRepo{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:          vcs.EventTypePush,
					Action:        vcs.ActionCreated,
					Branch:        "dev",
					DefaultBranch: "main",
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for push event for a workspace with user-specified branch",
			ws:   &types.Workspace{VCSRepo: &types.VCSRepo{Branch: "dev"}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Branch: "dev",
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push event for a workspace with non-matching, user-specified branch",
			ws:   &types.Workspace{VCSRepo: &types.VCSRepo{Branch: "dev"}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Branch: "staging",
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for opened pull request",
			ws:   &types.Workspace{VCSRepo: &types.VCSRepo{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type: vcs.EventTypePull, Action: vcs.ActionCreated,
				},
			},
			spawn: true,
		},
		{
			name: "spawn run for update to pull request",
			ws:   &types.Workspace{VCSRepo: &types.VCSRepo{}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionUpdated,
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push event for workspace with tags regex",
			ws:   &types.Workspace{VCSRepo: &types.VCSRepo{TagsRegex: "0.1.2"}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{Type: vcs.EventTypePush, Action: vcs.ActionCreated},
			},
			spawn: false,
		},
		{
			name: "spawn run for tag event for workspace with matching tags regex",
			ws: &types.Workspace{VCSRepo: &types.VCSRepo{
				TagsRegex: `^\d+\.\d+\.\d+$`,
			}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypeTag,
					Action: vcs.ActionCreated,
					Tag:    "0.1.2",
				},
			},
			spawn: true,
		},
		{
			name: "skip run for tag event for workspace with non-matching tags regex",
			ws: &types.Workspace{VCSRepo: &types.VCSRepo{
				TagsRegex: `^\d+\.\d+\.\d+$`,
			}},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypeTag,
					Action: vcs.ActionCreated,
					Tag:    "v0.1.2",
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for push event for workspace with matching file trigger pattern",
			ws: &types.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				VCSRepo:         &types.VCSRepo{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Paths:  []string{"/foo/bar.tf"},
				},
			},
			spawn: true,
		},
		{
			name: "skip run for push event for workspace with non-matching file trigger pattern",
			ws: &types.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				VCSRepo:         &types.VCSRepo{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePush,
					Action: vcs.ActionCreated,
					Paths:  []string{"README.md", ".gitignore"},
				},
			},
			spawn: false,
		},
		{
			name: "spawn run for pull event for workspace with matching file trigger pattern",
			ws: &types.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				VCSRepo:         &types.VCSRepo{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionUpdated,
				},
			},
			pullFiles: []string{"/foo/bar.tf"},
			spawn:     true,
		},
		{
			name: "skip run for pull event for workspace with non-matching file trigger pattern",
			ws: &types.Workspace{
				TriggerPatterns: []string{"/foo/*.tf"},
				VCSRepo:         &types.VCSRepo{},
			},
			event: vcs.Event{
				EventPayload: vcs.EventPayload{
					Type:   vcs.EventTypePull,
					Action: vcs.ActionUpdated,
				},
			},
			pullFiles: []string{"README.md", ".gitignore"},
			spawn:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runClient := &fakeSpawnerRunClient{}
			spawner := Spawner{
				configs: &configversion.FakeService{},
				workspaces: &workspace.FakeService{
					Workspaces: []*types.Workspace{tt.ws},
				},
				runs: runClient,
				vcs: &fakeSpawnerVCSProviderClient{
					pullFiles: tt.pullFiles,
				},
			}
			err := spawner.handleWithError(slog.New(&xslog.NoopHandler{}), tt.event)
			require.NoError(t, err)

			assert.Equal(t, tt.spawn, runClient.spawned)
		})
	}
}

type fakeSpawnerRunClient struct {
	// whether a run was spawned
	spawned bool
}

func (f *fakeSpawnerRunClient) Create(context.Context, string, CreateOptions) (*Run, error) {
	f.spawned = true
	return nil, nil
}

type fakeSpawnerVCSProviderClient struct {
	// list of file paths to return from stubbed ListPullRequestFiles()
	pullFiles []string
}

func (f *fakeSpawnerVCSProviderClient) GetVCSClient(context.Context, string) (vcs.Client, error) {
	return &fakeSpawnerCloudClient{pullFiles: f.pullFiles}, nil
}

type fakeSpawnerCloudClient struct {
	vcs.Client
	pullFiles []string
}

func (f *fakeSpawnerCloudClient) GetRepoTarball(context.Context, vcs.GetRepoTarballOptions) ([]byte, string, error) {
	return nil, "", nil
}

func (f *fakeSpawnerCloudClient) ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error) {
	return f.pullFiles, nil
}
