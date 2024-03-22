package gitea

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/davecgh/go-spew/spew"
	"github.com/tofutf/tofutf/internal/vcs"
)

type TokenClient struct {
	client *gitea.Client
}

var _ vcs.Client = &TokenClient{}

func NewTokenClient(opts vcs.NewTokenClientOptions) (*TokenClient, error) {
	client, err := gitea.NewClient(opts.URL.String(), gitea.SetToken(opts.Token))
	if err != nil {
		return nil, err
	}

	return &TokenClient{client: client}, nil
}

func (g *TokenClient) GetCurrentUser(ctx context.Context) (string, error) {
	user, _, err := g.client.GetMyUserInfo()
	if err != nil {
		return "", err
	}

	return user.UserName, nil
}

func (g *TokenClient) GetRepository(ctx context.Context, identifier string) (vcs.Repository, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return vcs.Repository{}, fmt.Errorf("malformed identifier: %s", identifier)
	}

	repo, _, err := g.client.GetRepo(owner, name)
	if err != nil {
		return vcs.Repository{}, fmt.Errorf("failed to retrieve repository: %w", err)
	}

	return vcs.Repository{
		Path:          identifier,
		DefaultBranch: repo.DefaultBranch,
	}, nil
}

func (g *TokenClient) ListRepositories(ctx context.Context, lopts vcs.ListRepositoriesOptions) ([]string, error) {
	repos, _, err := g.client.ListMyRepos(gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			PageSize: lopts.PageSize,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	var vcsRepos []string
	for _, repo := range repos {
		vcsRepos = append(vcsRepos, repo.FullName)
	}

	return vcsRepos, nil
}

func (g *TokenClient) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error) {
	slog.Info("calling ListTags")

	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	response, _, err := g.client.ListRepoTags(owner, name, gitea.ListRepoTagsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve repository tags: %w", err)
	}

	var tags []string
	for _, ref := range response {
		tags = append(tags, ref.Name)
	}

	return tags, nil
}

func (g *TokenClient) GetRepoTarball(ctx context.Context, opts vcs.GetRepoTarballOptions) ([]byte, string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	archive, _, err := g.client.GetArchive(owner, name, *opts.Ref, gitea.TarGZArchive)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve repository: %w", err)
	}

	return archive, *opts.Ref, nil
}

func (g *TokenClient) CreateWebhook(ctx context.Context, opts vcs.CreateWebhookOptions) (string, error) {
	slog.Info("calling CreateWebhook", slog.String("repo", opts.Repo))

	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	hook, _, err := g.client.CreateRepoHook(owner, name, gitea.CreateHookOption{
		Type:   gitea.HookTypeGitea,
		Active: true,
		Config: map[string]string{
			"url":          opts.Endpoint,
			"content_type": "json",
		},
		BranchFilter:        "*",
		AuthorizationHeader: "Bearer " + opts.Secret,
		Events:              encodeEvents(opts.Events),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create webhook: %w", err)
	}

	return fmt.Sprintf("%d", hook.ID), nil
}

func (g *TokenClient) UpdateWebhook(ctx context.Context, idString string, opts vcs.UpdateWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		return fmt.Errorf("failed to parse webhook id: %s", idString)
	}

	_, err = g.client.EditRepoHook(owner, name, int64(id), gitea.EditHookOption{
		Config: map[string]string{
			"url":          opts.Endpoint,
			"content_type": "json",
		},
		Active:              gitea.OptionalBool(true),
		AuthorizationHeader: "Bearer " + opts.Secret,
		Events:              encodeEvents(opts.Events),
		BranchFilter:        "*",
	})
	if err != nil {
		return fmt.Errorf("failed to edit repository hook: %w", err)
	}

	return nil
}

func (g *TokenClient) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return vcs.Webhook{}, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return vcs.Webhook{}, fmt.Errorf("failed to parse webhook id: %s", opts.ID)
	}

	hook, _, err := g.client.GetRepoHook(owner, name, int64(id))
	if err != nil {
		return vcs.Webhook{}, fmt.Errorf("failed to retrieve githook: %w", err)
	}

	return vcs.Webhook{
		ID:       fmt.Sprintf("%d", hook.ID),
		Repo:     opts.Repo,
		Events:   decodeEvents(hook.Events),
		Endpoint: hook.URL,
	}, nil
}

func decodeEvents(events []string) []vcs.EventType {
	var result []vcs.EventType
	for _, event := range events {
		switch event {
		case "push":
			result = append(result, vcs.EventTypePush)
		case "pull_request":
			result = append(result, vcs.EventTypePull)
		}
	}

	return result
}

func encodeEvents(events []vcs.EventType) []string {
	var result []string
	for _, event := range events {
		switch event {
		case vcs.EventTypePush:
			result = append(result, "push")
		case vcs.EventTypePull:
			result = append(result, "pull_request")
		}
	}

	return result
}

func (g *TokenClient) DeleteWebhook(ctx context.Context, opts vcs.DeleteWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return fmt.Errorf("failed to parse webhook id: %s", opts.ID)
	}

	_, err = g.client.DeleteRepoHook(owner, name, int64(id))
	if err != nil {
		return fmt.Errorf("failed to delete repo hook: %s", err)
	}

	return nil
}

func (g *TokenClient) SetStatus(ctx context.Context, opts vcs.SetStatusOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	// TODO(johnrowl) implement SetStatus to ensure tofutf can push statuses to gitea.

	_ = owner
	_ = name

	return nil
}

func (g *TokenClient) ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error) {
	slog.Info("calling ListPullRequestFiles")

	owner, name, found := strings.Cut(repo, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", repo)
	}

	diff, _, err := g.client.GetPullRequestDiff(owner, name, int64(pull), gitea.PullRequestDiffOptions{Binary: false})
	if err != nil {
		return nil, err
	}

	_ = diff
	spew.Dump(diff)

	// TODO(johnrowl) implement ListPullRequestFiles to get workspace file triggers functioning for gitea.
	return nil, nil
}

func (g *TokenClient) GetCommit(ctx context.Context, repo, ref string) (vcs.Commit, error) {
	slog.Info("calling GetCommit")

	owner, name, found := strings.Cut(repo, "/")
	if !found {
		return vcs.Commit{}, fmt.Errorf("malformed identifier: %s", repo)
	}

	commit, _, err := g.client.GetSingleCommit(owner, name, ref)
	if err != nil {
		return vcs.Commit{}, fmt.Errorf("failed to retrieve: %s", err)
	}

	return vcs.Commit{
		SHA: commit.SHA,
		URL: commit.URL,
		Author: vcs.CommitAuthor{
			AvatarURL: commit.Author.AvatarURL,
			Username:  commit.Author.UserName,
		},
	}, nil
}
