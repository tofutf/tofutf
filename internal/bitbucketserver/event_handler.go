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

	"github.com/tofutf/tofutf/internal/vcs"
	"golang.org/x/exp/slog"
)

var HandleEvent = (func(r *http.Request, secret string) (*vcs.EventPayload, error))(ApplyMiddleware(BaseHandleEvent, HandleEventWithLogging))

// SignatureHeader is the header that contains the sha256 signature of the payload content.
const SignatureHeader = "X-Hub-Signature"

// ValidateEvent validates the request.
func ValidateEvent(r *http.Request, secret string, payload []byte) error {
	signature := strings.TrimPrefix(r.Header.Get(SignatureHeader), "sha256=")

	hash := hmac.New(sha256.New, []byte(secret))
	_, err := hash.Write(payload)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	sha := hex.EncodeToString(hash.Sum(nil))

	slog.Debug("checking signatures", slog.String("signature", signature), slog.String("sha", sha))

	if !hmac.Equal([]byte(signature), []byte(sha)) {
		return errors.New("token validation failed")
	}

	return nil
}

// ApplyMiddleware applies middleware to the base EventHandler.
func ApplyMiddleware(base EventHandler, factories ...EventFactory) EventHandler {
	decorated := base

	for _, factory := range factories {
		decorated = factory(decorated)
	}

	return decorated
}

type EventFactory func(next EventHandler) EventHandler

type EventHandler func(r *http.Request, secret string) (*vcs.EventPayload, error)

func HandleEventWithLogging(next EventHandler) EventHandler {
	return func(r *http.Request, secret string) (*vcs.EventPayload, error) {
		event, err := next(r, secret)
		if err != nil {
			return nil, err
		}

		slog.Info("handle bitbucket event", "event", event)

		return event, nil
	}
}

func BaseHandleEvent(r *http.Request, secret string) (*vcs.EventPayload, error) {
	slog.Debug("handling webhook")

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}

	err = ValidateEvent(r, secret, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	var event BitbucketHookEvent
	err = json.Unmarshal(payload, &event)
	if err != nil {
		return nil, fmt.Errorf("failed un unmarshal webhook: %w", err)
	}

	repoPath := strings.ToLower(event.Repository.Project.Key) + "/" + event.Repository.Slug

	switch event.EventKey {
	case eventPush:
		slog.Info("handling push event")
		if len(event.Changes) == 0 {
			return nil, fmt.Errorf("invalid event with no changes")
		}

		if len(event.Changes) != 1 {
			return nil, fmt.Errorf("unable to handle multiple changes in a single push")
		}

		refType, err := event.getRefType()
		if err != nil {
			return nil, err
		}

		actionType, err := event.getActionType()
		if err != nil {
			return nil, err
		}

		refID, err := event.getRefID()
		if err != nil {
			return nil, err
		}

		if refType == "TAG" {
			if actionType == "ADD" {
				tag := strings.TrimPrefix(refID, "refs/")
				return &vcs.EventPayload{
					RepoPath:      repoPath,
					VCSKind:       vcs.BitbucketServer,
					Tag:           tag,
					Action:        vcs.ActionCreated,
					CommitSHA:     event.Changes[0].ToHash,
					DefaultBranch: "main", // TODO(johnrowl) need to change this.
				}, nil
			} else if actionType == "DELETE" {
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
		} else if refType == "BRANCH" {
			if actionType == "UPDATE" {
				commitSHA, err := event.GetCommitSHA()
				if err != nil {
					return nil, err
				}

				commitURL, err := event.GetCommitURL()
				if err != nil {
					return nil, err
				}

				actorURL, err := event.GetActorURL()
				if err != nil {
					return nil, err
				}

				actorAvatarURL, err := event.GetActorAvatarURL()
				if err != nil {
					return nil, err
				}

				actorUsername, err := event.GetActorUsername()
				if err != nil {
					return nil, err
				}

				branch := strings.TrimPrefix(event.Changes[0].Ref.ID, "refs/heads/")

				return &vcs.EventPayload{
					RepoPath:        repoPath,
					VCSKind:         vcs.BitbucketServer,
					Type:            vcs.EventTypePush,
					Action:          vcs.ActionCreated,
					CommitSHA:       commitSHA,
					CommitURL:       commitURL,
					Branch:          branch,
					SenderHTMLURL:   actorURL,
					SenderAvatarURL: actorAvatarURL,
					SenderUsername:  actorUsername,

					// TODO(robbert229): figure out a way to calculate the
					// default branch for bitbucket which doesn't include
					// it in the actual webhook.
					DefaultBranch: "main",
				}, nil
			}
		}

		slog.Error("unhandled push event", "event", event)

		return nil, fmt.Errorf("failed to handle push event")
	}

	return nil, nil
}

func (e BitbucketHookEvent) getRefType() (string, error) {
	return e.Changes[0].Ref.Type, nil

}

func (e BitbucketHookEvent) getActionType() (string, error) {
	return e.Changes[0].Type, nil
}

func (e BitbucketHookEvent) getRefID() (string, error) {
	return e.Changes[0].Ref.ID, nil
}

func (e BitbucketHookEvent) IsBranchPushEvent() (bool, error) {
	refType, err := e.getRefType()
	if err != nil {
		return false, err
	}

	return refType == "BRANCH", nil
}

func (e BitbucketHookEvent) GetCommitURL() (string, error) {
	commitSHA, err := e.GetCommitSHA()
	if err != nil {
		return "", err
	}

	repositoryURL := e.Repository.Links.Self[0].Href
	commitURL := strings.TrimSuffix(repositoryURL, "browse") + "commits/" + commitSHA
	return commitURL, nil
}

func (e BitbucketHookEvent) GetCommitSHA() (string, error) {
	return e.Changes[0].ToHash, nil
}

func (e BitbucketHookEvent) GetActorURL() (string, error) {
	return e.Actor.Links.Self[0].Href, nil
}

func (e BitbucketHookEvent) GetActorAvatarURL() (string, error) {
	actorURL, err := e.GetActorURL()
	if err != nil {
		return "", err
	}

	return actorURL + "/avatar.png?s=192", nil
}

func (e BitbucketHookEvent) GetActorUsername() (string, error) {
	return e.Actor.Slug, nil
}

type BitbucketHookEvent struct {
	EventKey string `json:"eventKey"`
	Date     string `json:"date"`
	Actor    struct {
		Name         string `json:"name"`
		EmailAddress string `json:"emailAddress"`
		ID           int    `json:"id"`
		DisplayName  string `json:"displayName"`
		Active       bool   `json:"active"`
		Slug         string `json:"slug"`
		Type         string `json:"type"`
		Links        struct {
			Self []struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
	} `json:"actor"`
	Repository struct {
		Slug          string `json:"slug"`
		ID            int    `json:"id"`
		Name          string `json:"name"`
		HierarchyID   string `json:"hierarchyId"`
		ScmID         string `json:"scmId"`
		State         string `json:"state"`
		StatusMessage string `json:"statusMessage"`
		Forkable      bool   `json:"forkable"`
		Project       struct {
			Key         string `json:"key"`
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Public      bool   `json:"public"`
			Type        string `json:"type"`
			Links       struct {
				Self []struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"project"`
		Public bool `json:"public"`
		Links  struct {
			Clone []struct {
				Href string `json:"href"`
				Name string `json:"name"`
			} `json:"clone"`
			Self []struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
	} `json:"repository"`
	Changes []struct {
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
