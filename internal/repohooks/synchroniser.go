package repohooks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/vcs"
)

type (
	// synchroniser synchronises a hook with the vcs provider
	synchroniser struct {
		logger *slog.Logger
		syncdb
	}

	syncdb interface {
		updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error
	}
)

// sync should be called from within a tx to avoid inconsistent results.
func (s *synchroniser) sync(ctx context.Context, client vcs.Client, hook *hook) error {
	createAndSync := func() error {
		cloudID, err := client.CreateWebhook(ctx, vcs.CreateWebhookOptions{
			Repo:     hook.repoPath,
			Secret:   hook.secret,
			Events:   defaultEvents,
			Endpoint: hook.endpoint,
		})
		if err != nil {
			return err
		}
		s.logger.Info("created webhook", "webhook", hook)
		if err := s.updateHookCloudID(ctx, hook.id, cloudID); err != nil {
			return err
		}
		return nil
	}
	if hook.cloudID == nil {
		return createAndSync()
	}
	cloudHook, err := client.GetWebhook(ctx, vcs.GetWebhookOptions{
		Repo: hook.repoPath,
		ID:   *hook.cloudID,
	})
	if errors.Is(err, internal.ErrResourceNotFound) {
		return createAndSync()
	} else if err != nil {
		return fmt.Errorf("retrieving hook from cloud: %w", err)
	}
	// hook is present on the vcs repo, but we update it anyway just to ensure
	// its configuration is consistent with what we have in the DB
	err = client.UpdateWebhook(ctx, cloudHook.ID, vcs.UpdateWebhookOptions{
		Repo:     hook.repoPath,
		Secret:   hook.secret,
		Events:   defaultEvents,
		Endpoint: hook.endpoint,
	})
	if err != nil {
		return err
	}
	s.logger.Info("updated webhook", "webhook", hook)
	return nil
}
