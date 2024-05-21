package workspace

import (
	"context"

	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/team"
	"github.com/tofutf/tofutf/internal/vcs"
	"github.com/tofutf/tofutf/internal/vcsprovider"
)

type FakeService struct {
	Workspaces []*types.Workspace
	Policy     internal.WorkspacePolicy
}

func (f *FakeService) ListConnectedWorkspaces(ctx context.Context, vcsProviderID, repoPath string) ([]*types.Workspace, error) {
	return f.Workspaces, nil
}

func (f *FakeService) Create(context.Context, CreateOptions) (*types.Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Update(_ context.Context, _ string, opts types.WorkspaceUpdateOptions) (*types.Workspace, error) {
	ref := f.Workspaces[0]
	Update(ref, opts) //nolint:errcheck
	return ref, nil
}

func (f *FakeService) List(ctx context.Context, opts ListOptions) (*resource.Page[*types.Workspace], error) {
	return resource.NewPage(f.Workspaces, opts.PageOptions, nil), nil
}

func (f *FakeService) Get(context.Context, string) (*types.Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) GetByName(context.Context, string, string) (*types.Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Delete(context.Context, string) (*types.Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Lock(context.Context, string, *string) (*types.Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Unlock(context.Context, string, *string, bool) (*types.Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) ListTags(context.Context, string, ListTagsOptions) (*resource.Page[*Tag], error) {
	return nil, nil
}

func (f *FakeService) GetPolicy(context.Context, string) (internal.WorkspacePolicy, error) {
	return f.Policy, nil
}

func (f *FakeService) AddTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	return nil
}

func (f *FakeService) RemoveTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	return nil
}

func (f *FakeService) SetPermission(ctx context.Context, workspaceID, teamID string, role rbac.Role) error {
	return nil
}

func (f *FakeService) UnsetPermission(ctx context.Context, workspaceID, teamID string) error {
	return nil
}

type fakeVCSProviderService struct {
	providers []*vcsprovider.VCSProvider
	repos     []string
}

func (f *fakeVCSProviderService) Get(ctx context.Context, providerID string) (*vcsprovider.VCSProvider, error) {
	return f.providers[0], nil
}

func (f *fakeVCSProviderService) List(context.Context, string) ([]*vcsprovider.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeVCSProviderService) GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error) {
	return &fakeVCSClient{repos: f.repos}, nil
}

type fakeVCSClient struct {
	repos []string

	vcs.Client
}

func (f *fakeVCSClient) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]string, error) {
	return f.repos, nil
}

type fakeTeamService struct {
	teams []*team.Team
}

func (f *fakeTeamService) List(context.Context, string) ([]*team.Team, error) {
	return f.teams, nil
}
