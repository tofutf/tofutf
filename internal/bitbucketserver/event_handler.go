package bitbucketserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/tofutf/tofutf/internal/vcs"
	"golang.org/x/exp/slog"
)

func HandleEvent(r *http.Request, secret string) (*vcs.EventPayload, error) {
	slog.Debug("handling webhook")

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}

	var event bitbucketHookEvent
	err = json.Unmarshal(payload, &event)
	if err != nil {
		return nil, fmt.Errorf("failed un unmarshal webhook: %w", err)
	}

	{
		signature := strings.TrimPrefix(r.Header.Get("X-Hub-Signature"), "sha256=")

		hash := hmac.New(sha256.New, []byte(secret))
		_, err = hash.Write(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate checksum: %w", err)
		}

		sha := hex.EncodeToString(hash.Sum(nil))

		slog.Debug("checking signatures", slog.String("signature", signature), slog.String("sha", sha))

		if !hmac.Equal([]byte(signature), []byte(sha)) {
			return nil, errors.New("token validation failed")
		}
	}

	spew.Dump(event)

	repoPath := ""

	switch event.EventKey {
	case eventPush:
		slog.Info("handling push event")
		if len(event.Changes) == 0 {
			return nil, fmt.Errorf("invalid event with no changes")
		}

		if len(event.Changes) != 1 {
			return nil, fmt.Errorf("unable to handle multiple changes in a single push")
		}

		changeType := event.Changes[0].Ref.Type
		refID := event.Changes[0].Ref.ID

		if changeType == "TAG" {
			refParts := strings.Split(refID, "/")
			if len(refParts) != 3 {
				return nil, fmt.Errorf("malformed ref: %s", refID)
			}

			return &vcs.EventPayload{
				RepoPath:      repoPath,
				VCSKind:       vcs.BitbucketServer,
				Tag:           refParts[2],
				Action:        vcs.ActionCreated,
				CommitSHA:     event.Changes[0].ToHash,
				DefaultBranch: "main", // TODO(johnrowl) need to change this.
			}, nil
		} else if changeType == "BRANCH" {
			refParts := strings.Split(refID, "/")
			if len(refParts) != 3 {
				return nil, fmt.Errorf("malformed ref: %s", refID)
			}

			return &vcs.EventPayload{
				RepoPath:      repoPath,
				VCSKind:       vcs.BitbucketServer,
				Branch:        refParts[2],
				CommitSHA:     event.Changes[0].ToHash,
				DefaultBranch: "main", // TODO(johnrowl) need to change this.
			}, nil
		}

		// return &cloud.VCSEvent{

		// }, nil

		return nil, fmt.Errorf("failed to handle push event")
	}

	// switch event := rawEvent.(type) {
	// case *gitlab.PushEvent:
	// 	refParts := strings.Split(event.Ref, "/")
	// 	if len(refParts) != 3 {
	// 		return nil, fmt.Errorf("malformed ref: %s", event.Ref)
	// 	}
	// 	return &cloud.VCSEvent{
	// 		Branch:        refParts[2],
	// 		CommitSHA:     event.After,
	// 		DefaultBranch: event.Project.DefaultBranch,
	// 	}, nil
	// case *gitlab.TagEvent:
	// 	refParts := strings.Split(event.Ref, "/")
	// 	if len(refParts) != 3 {
	// 		return nil, fmt.Errorf("malformed ref: %s", event.Ref)
	// 	}
	// 	return &cloud.VCSEvent{
	// 		Tag: refParts[2],
	// 		// Action:     action,
	// 		CommitSHA:     event.After,
	// 		DefaultBranch: event.Project.DefaultBranch,
	// 	}, nil
	// case *gitlab.MergeEvent:
	// }

	return nil, nil
}

type bitbucketHookEvent struct {
	EventKey    string                  `json:"eventKey"`
	Date        string                  `json:"date"`
	Actor       bitbucketv1.Actor       `json:"actor"`
	PullRequest bitbucketv1.PullRequest `json:"pullRequest"`
	Changes     []struct {
		Ref struct {
			ID        string `json:"id"`
			DisplayID string `json:"displayId"`
			Type      string `json:"type"`
		} `json:"ref"`
		RefID    string `json:"refId"`
		FromHash string `json:"fromHash"`
		ToHash   string `json:"toHash"`
		Type     string `json:"type"`
	} `json:"changes"`
}
