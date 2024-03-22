package gitea

import (
	"bytes"
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
		{
			"push",
			"./testdata/gitea_push.json",
			&vcs.EventPayload{
				VCSKind:         vcs.GiteaKind,
				Type:            vcs.EventTypePush,
				RepoPath:        "gitea/webhooks",
				Branch:          "develop",
				DefaultBranch:   "master",
				CommitSHA:       "bffeb74224043ba2feb48d137756c8a9331c449a",
				CommitURL:       "http://localhost:3000/gitea/webhooks/commit/bffeb74224043ba2feb48d137756c8a9331c449a",
				Action:          vcs.ActionCreated,
				Paths:           nil,
				SenderUsername:  "gitea",
				SenderAvatarURL: "https://localhost:3000/avatars/1",
				SenderHTMLURL:   "https://localhost:3000/gitea",
			},
			false,
		},
		// {
		// 	"tag pushed",
		// 	"./testdata/bitbucketserver_push_tag.json",
		// 	&vcs.EventPayload{
		// 		VCSKind:         vcs.BitbucketServer,
		// 		Type:            vcs.EventTypeTag,
		// 		RepoPath:        "tft/terraform-tofutf-test",
		// 		Tag:             "v1.2.3",
		// 		DefaultBranch:   "main",
		// 		CommitSHA:       "3a32194600dbd0f39bc921d15a785e93994b26da",
		// 		CommitURL:       "https://bitbucket.tofutf.io/projects/tft/repos/terraform-tofutf-test/commits/3a32194600dbd0f39bc921d15a785e93994b26da",
		// 		Action:          vcs.ActionCreated,
		// 		SenderUsername:  "johnrowl",
		// 		SenderAvatarURL: "https://bitbucket.tofutf.io/users/johnrowl/avatar.png?s=192",
		// 		SenderHTMLURL:   "https://bitbucket.tofutf.io/users/johnrowl",
		// 	},
		// 	false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.body)
			require.NoError(t, err)
			defer f.Close()

			payload, err := io.ReadAll(f)
			require.Nil(t, err)

			secret := "test-secret"

			r := httptest.NewRequest("POST", "/", bytes.NewBuffer(payload))
			r.Header.Add("Content-type", "application/json")

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
