package module

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/semver"
	"github.com/tofutf/tofutf/internal/vcs"
	"github.com/tofutf/tofutf/internal/vcsprovider"
)

type (
	// publisher publishes new versions of terraform modules from VCS tags
	publisher struct {
		logger *slog.Logger

		modules      *Service
		vcsproviders *vcsprovider.Service
	}
)

func (p *publisher) handle(event vcs.Event) {
	logger := p.logger.With(
		"sha", event.CommitSHA,
		"type", event.Type,
		"action", event.Action,
		"branch", event.Branch,
		"tag", event.Tag,
	)

	if err := p.handleWithError(logger, event); err != nil {
		p.logger.Error("handling event", "err", err)
	}
}

// handlerWithError publishes a module version in response to a vcs event.
func (p *publisher) handleWithError(logger *slog.Logger, event vcs.Event) error {
	// no parent context; handler is called asynchronously
	ctx := context.Background()
	// give spawner unlimited powers
	ctx = internal.AddSubjectToContext(ctx, &internal.Superuser{Username: "run-spawner"})

	// only create-tag events trigger the publishing of new module version
	if event.Type != vcs.EventTypeTag {
		return nil
	}
	if event.Action != vcs.ActionCreated {
		return nil
	}
	// only interested in tags that look like semantic versions
	if !semver.IsValid(event.Tag) {
		return nil
	}
	// TODO: we're only retrieving *one* module, but can not *multiple* modules
	// be connected to a repo?
	module, err := p.modules.GetModuleByConnection(ctx, event.VCSProviderID, event.RepoPath)
	if err != nil {
		return err
	}
	if module.Connection == nil {
		return fmt.Errorf("module is not connected to a repo: %s", module.ID)
	}
	client, err := p.vcsproviders.GetVCSClient(ctx, module.Connection.VCSProviderID)
	if err != nil {
		return err
	}
	return p.modules.PublishVersion(ctx, PublishVersionOptions{
		ModuleID: module.ID,
		// strip off v prefix if it has one
		Version: strings.TrimPrefix(event.Tag, "v"),
		Ref:     event.CommitSHA,
		Repo:    Repo(module.Connection.Repo),
		Client:  client,
	})
}
