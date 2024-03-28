package bitbucketserver

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tofutf/tofutf/internal/vcs"
)

func TestEventHandler(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		want   *vcs.EventPayload
		ignore bool
	}{
		// {
		// 	"push",
		// 	"push",
		// 	"./testdata/github_push.json",
		// 	&vcs.EventPayload{
		// 		VCSKind:         vcs.GithubKind,
		// 		Type:            vcs.EventTypePush,
		// 		RepoPath:        "leg100/tfc-workspaces",
		// 		Branch:          "master",
		// 		DefaultBranch:   "master",
		// 		CommitSHA:       "42d6fc7dac35cc7945231195e248af2f6256b522",
		// 		CommitURL:       "https://github.com/leg100/tfc-workspaces/commit/42d6fc7dac35cc7945231195e248af2f6256b522",
		// 		Action:          vcs.ActionCreated,
		// 		Paths:           []string{"main.tf", "networks.tf", "servers.tf"},
		// 		SenderUsername:  "leg100",
		// 		SenderAvatarURL: "https://avatars.githubusercontent.com/u/75728?v=4",
		// 		SenderHTMLURL:   "https://github.com/leg100",
		// 	},
		// 	false,
		// },
		{
			"push",
			"./testdata/bitbucketserver_push.json",
			&vcs.EventPayload{
				VCSKind:         vcs.BitbucketServerKind,
				Type:            vcs.EventTypePush,
				RepoPath:        "tft/terraform-tofutf-test",
				Branch:          "main",
				DefaultBranch:   "main",
				CommitSHA:       "3a32194600dbd0f39bc921d15a785e93994b26da",
				CommitURL:       "https://bitbucket.tofutf.io/projects/tft/repos/terraform-tofutf-test/commits/3a32194600dbd0f39bc921d15a785e93994b26da",
				Action:          vcs.ActionCreated,
				Paths:           nil,
				SenderUsername:  "johnrowl",
				SenderAvatarURL: "https://bitbucket.tofutf.io/users/johnrowl/avatar.png?s=192",
				SenderHTMLURL:   "https://bitbucket.tofutf.io/users/johnrowl",
			},
			false,
		},
		// {
		// 	"pull request opened",
		// 	"pull_request",
		// 	"./testdata/github_pull_opened.json",
		// 	&vcs.EventPayload{
		// 		VCSKind:           vcs.GithubKind,
		// 		Type:              vcs.EventTypePull,
		// 		RepoPath:          "leg100/otf-workspaces",
		// 		Branch:            "pr-2",
		// 		DefaultBranch:     "master",
		// 		CommitSHA:         "c560613b228f5e189520fbab4078284ea8312bcb",
		// 		CommitURL:         "https://github.com/tofutf/tofutf-workspaces/commit/c560613b228f5e189520fbab4078284ea8312bcb",
		// 		PullRequestNumber: 2,
		// 		PullRequestURL:    "https://github.com/tofutf/tofutf-workspaces/pull/2",
		// 		PullRequestTitle:  "pr-2",
		// 		Action:            vcs.ActionCreated,
		// 		SenderUsername:    "leg100",
		// 		SenderAvatarURL:   "https://avatars.githubusercontent.com/u/75728?v=4",
		// 		SenderHTMLURL:     "https://github.com/leg100",
		// 	},
		// 	false,
		// },
		// {
		// 	"pull request updated",
		// 	"pull_request",
		// 	"./testdata/github_pull_update.json",
		// 	&vcs.EventPayload{
		// 		VCSKind:           vcs.GithubKind,
		// 		Type:              vcs.EventTypePull,
		// 		RepoPath:          "leg100/otf-workspaces",
		// 		Branch:            "pr-1",
		// 		DefaultBranch:     "master",
		// 		CommitSHA:         "067e2b4c6394b3dad3c0ec89ffc428ab60ae7e5d",
		// 		CommitURL:         "https://github.com/tofutf/tofutf-workspaces/commit/067e2b4c6394b3dad3c0ec89ffc428ab60ae7e5d",
		// 		PullRequestNumber: 1,
		// 		PullRequestURL:    "https://github.com/tofutf/tofutf-workspaces/pull/1",
		// 		PullRequestTitle:  "pr-1",
		// 		Action:            vcs.ActionUpdated,
		// 		SenderUsername:    "leg100",
		// 		SenderAvatarURL:   "https://avatars.githubusercontent.com/u/75728?v=4",
		// 		SenderHTMLURL:     "https://github.com/leg100",
		// 	},
		// 	false,
		// },
		{
			"tag pushed",
			"./testdata/bitbucketserver_push_tag.json",
			&vcs.EventPayload{
				VCSKind:         vcs.BitbucketServerKind,
				Type:            vcs.EventTypeTag,
				RepoPath:        "tft/terraform-tofutf-test",
				Tag:             "v1.2.3",
				DefaultBranch:   "main",
				CommitSHA:       "3a32194600dbd0f39bc921d15a785e93994b26da",
				CommitURL:       "https://bitbucket.tofutf.io/projects/tft/repos/terraform-tofutf-test/commits/3a32194600dbd0f39bc921d15a785e93994b26da",
				Action:          vcs.ActionCreated,
				SenderUsername:  "johnrowl",
				SenderAvatarURL: "https://bitbucket.tofutf.io/users/johnrowl/avatar.png?s=192",
				SenderHTMLURL:   "https://bitbucket.tofutf.io/users/johnrowl",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.body)
			require.NoError(t, err)
			defer f.Close()

			payload, err := io.ReadAll(f)
			require.Nil(t, err)

			secret := "test-secret"

			hash := hmac.New(sha256.New, []byte(secret))
			_, err = hash.Write(payload)
			require.Nil(t, err)

			signature := hex.EncodeToString(hash.Sum(nil))

			r := httptest.NewRequest("POST", "/", bytes.NewBuffer(payload))
			r.Header.Add("Content-type", "application/json")
			r.Header.Add(SignatureHeader, "sha256="+signature)

			w := httptest.NewRecorder()

			got, err := HandleEvent(r, secret)
			if tt.ignore {
				var ignore vcs.ErrIgnoreEvent
				assert.True(t, errors.As(err, &ignore))
			} else {
				require.NoError(t, err)
				assert.Equal(t, 200, w.Code, w.Body.String())
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
