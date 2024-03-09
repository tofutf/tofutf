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

		refType := event.Changes[0].Ref.Type
		actionType := event.Changes[0].Type
		refID := event.Changes[0].Ref.ID

		if refType == "TAG" && actionType == "ADD" {
			tag := strings.TrimPrefix(refID, "refs/")
			return &vcs.EventPayload{
				RepoPath:      repoPath,
				VCSKind:       vcs.BitbucketServer,
				Tag:           tag,
				Action:        vcs.ActionCreated,
				CommitSHA:     event.Changes[0].ToHash,
				DefaultBranch: "main", // TODO(johnrowl) need to change this.
			}, nil
		} else if refType == "TAG" && actionType == "DELETE" {
			tag := strings.TrimPrefix(refID, "refs/")
			return &vcs.EventPayload{
				RepoPath:      repoPath,
				VCSKind:       vcs.BitbucketServer,
				Tag:           tag,
				Action:        vcs.ActionCreated,
				CommitSHA:     event.Changes[0].ToHash,
				DefaultBranch: "main", // TODO(johnrowl) need to change this.
			}, nil
		}

		slog.Error("unhandled push event", "event", event)

		return nil, fmt.Errorf("failed to handle push event")
	}

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
