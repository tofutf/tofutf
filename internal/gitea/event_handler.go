package gitea

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tofutf/tofutf/internal/vcs"
)

// HandleEvent handles the Gitea event.
func HandleEvent(r *http.Request, secret string) (*vcs.EventPayload, error) {
	decoder := json.NewDecoder(r.Body)

	var event WebhookEvent
	err := decoder.Decode(&event)
	if err != nil {
		return nil, fmt.Errorf("failed to decode gitea webhook event json: %w", err)
	}

	return &vcs.EventPayload{
		VCSKind:         vcs.GiteaKind,
		Branch:          event.GetBranch(),
		DefaultBranch:   event.GetDefaultBranch(),
		Type:            vcs.EventTypePush,
		Action:          vcs.ActionCreated,
		SenderUsername:  event.GetSenderUsername(),
		SenderAvatarURL: event.GetSenderAvatarURL(),
		SenderHTMLURL:   event.GetSenderHTMLURL(),
		RepoPath:        event.GetRepoPath(),
		CommitSHA:       event.GetCommitSHA(),
		CommitURL:       event.GetCommitURL(),
	}, nil
}

type WebhookEvent struct {
	Secret     string `json:"secret"`
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	CompareURL string `json:"compare_url"`
	Commits    []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"committer"`
		Timestamp string `json:"timestamp"`
	} `json:"commits"`
	Repository struct {
		ID    int `json:"id"`
		Owner struct {
			ID        int    `json:"id"`
			Login     string `json:"login"`
			FullName  string `json:"full_name"`
			Email     string `json:"email"`
			AvatarURL string `json:"avatar_url"`
			Username  string `json:"username"`
		} `json:"owner"`
		Name            string `json:"name"`
		FullName        string `json:"full_name"`
		Description     string `json:"description"`
		Private         bool   `json:"private"`
		Fork            bool   `json:"fork"`
		HTMLURL         string `json:"html_url"`
		SSHURL          string `json:"ssh_url"`
		CloneURL        string `json:"clone_url"`
		Website         string `json:"website"`
		StarsCount      int    `json:"stars_count"`
		ForksCount      int    `json:"forks_count"`
		WatchersCount   int    `json:"watchers_count"`
		OpenIssuesCount int    `json:"open_issues_count"`
		DefaultBranch   string `json:"default_branch"`
		CreatedAt       string `json:"created_at"`
		UpdatedAt       string `json:"updated_at"`
	} `json:"repository"`
	Pusher struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		FullName  string `json:"full_name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Username  string `json:"username"`
	} `json:"pusher"`
	Sender struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		FullName  string `json:"full_name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Username  string `json:"username"`
	} `json:"sender"`
}

func (e WebhookEvent) GetDefaultBranch() string {
	return e.Repository.DefaultBranch
}

func (e WebhookEvent) GetBranch() string {
	return strings.TrimPrefix(e.Ref, "refs/heads/")
}

func (e WebhookEvent) GetSenderUsername() string {
	return e.Sender.Login
}

func (e WebhookEvent) GetSenderAvatarURL() string {
	return e.Sender.AvatarURL
}

func (e WebhookEvent) GetSenderHTMLURL() string {
	return strings.TrimSuffix(e.Sender.AvatarURL, fmt.Sprintf("avatars/%d", e.Sender.ID)) + e.Sender.Username
}

func (e WebhookEvent) GetRepoPath() string {
	return e.Repository.FullName
}

func (e WebhookEvent) GetCommitSHA() string {
	return e.After
}

func (e WebhookEvent) GetCommitURL() string {
	return e.Commits[0].URL
}
