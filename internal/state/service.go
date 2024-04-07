package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/leg100/surl"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
	"github.com/tofutf/tofutf/internal/tfeapi"
	"github.com/tofutf/tofutf/internal/workspace"
)

var ErrCurrentVersionDeletionAttempt = errors.New("deleting the current state version is not allowed")

// cacheKey generates a key for caching state files
func cacheKey(svID string) string { return fmt.Sprintf("%s.json", svID) }

type (
	// Service provides access to state and state versions
	Service struct {
		logger    *slog.Logger
		db        *pgdb
		cache     internal.Cache // cache state file
		workspace internal.Authorizer
		web       *webHandlers
		tfeapi    *tfe
		api       *api

		*factory // for creating state versions
	}

	Options struct {
		Logger *slog.Logger
		html.Renderer
		internal.Cache
		*sql.DB
		*tfeapi.Responder
		*surl.Signer

		WorkspaceService *workspace.Service
	}

	// StateVersionListOptions represents the options for listing state versions.
	StateVersionListOptions struct {
		resource.PageOptions
		Organization string `schema:"filter[organization][name],required"`
		Workspace    string `schema:"filter[workspace][name],required"`
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		logger:    opts.Logger,
		cache:     opts.Cache,
		db:        db,
		workspace: opts.WorkspaceService,
		factory:   &factory{db},
	}

	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		Service:  &svc,
	}

	svc.tfeapi = &tfe{
		Responder:  opts.Responder,
		Signer:     opts.Signer,
		state:      &svc,
		workspaces: opts.WorkspaceService,
	}

	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
		tfeapi:    svc.tfeapi,
	}

	// include state version outputs in api responses when requested.
	opts.Responder.Register(tfeapi.IncludeOutputs, svc.tfeapi.includeOutputs)
	opts.Responder.Register(tfeapi.IncludeOutputs, svc.tfeapi.includeWorkspaceCurrentOutputs)
	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.tfeapi.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *Service) Create(ctx context.Context, opts CreateStateVersionOptions) (*Version, error) {
	if opts.WorkspaceID == nil {
		return nil, errors.New("workspace ID is required")
	}
	subject, err := a.workspace.CanAccess(ctx, rbac.CreateStateVersionAction, *opts.WorkspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := a.new(ctx, opts)
	if err != nil {
		a.logger.Error("creating state version", "subject", subject, "err", err)
		return nil, err
	}

	if err := a.cache.Set(cacheKey(sv.ID), sv.State); err != nil {
		a.logger.Error("caching state file", "err", err)
	}

	a.logger.Info("created state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *Service) DownloadCurrent(ctx context.Context, workspaceID string) ([]byte, error) {
	v, err := a.GetCurrent(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return a.Download(ctx, v.ID)
}

func (a *Service) List(ctx context.Context, workspaceID string, opts resource.PageOptions) (*resource.Page[*Version], error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.ListStateVersionsAction, workspaceID)
	if err != nil {
		return nil, err
	}

	svl, err := a.db.listVersions(ctx, workspaceID, opts)
	if err != nil {
		a.logger.Error("listing state versions", "workspace", workspaceID, "subject", subject, "err", err)
		return nil, err
	}
	a.logger.Debug("listed state versions", "workspace", workspaceID, "subject", subject)
	return svl, nil
}

func (a *Service) GetCurrent(ctx context.Context, workspaceID string) (*Version, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.GetStateVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.getCurrentVersion(ctx, workspaceID)
	if errors.Is(err, internal.ErrResourceNotFound) {
		// not found error occurs legitimately with a new workspace without any
		// state, so we log these errors at low level instead
		a.logger.Debug("retrieving current state version: workspace has no state yet", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	} else if err != nil {
		a.logger.Error("retrieving current state version", "workspace_id", workspaceID, "subject", subject, "err", err)
		return nil, err
	}

	a.logger.Debug("retrieved current state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *Service) Get(ctx context.Context, versionID string) (*Version, error) {
	subject, err := a.CanAccess(ctx, rbac.GetStateVersionAction, versionID)
	if err != nil {
		return nil, err
	}

	sv, err := a.db.getVersion(ctx, versionID)
	if err != nil {
		a.logger.Error("retrieving state version", "id", versionID, "subject", subject, "err", err)
		return nil, err
	}

	a.logger.Debug("retrieved state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *Service) Delete(ctx context.Context, versionID string) error {
	subject, err := a.CanAccess(ctx, rbac.DeleteStateVersionAction, versionID)
	if err != nil {
		return err
	}

	if err := a.db.deleteVersion(ctx, versionID); err != nil {
		a.logger.Error("deleting state version", "id", versionID, "subject", subject, "err", err)
		return err
	}

	a.logger.Info("deleted state version", "id", versionID, "subject", subject)
	return nil
}

func (a *Service) Rollback(ctx context.Context, versionID string) (*Version, error) {
	subject, err := a.CanAccess(ctx, rbac.RollbackStateVersionAction, versionID)
	if err != nil {
		return nil, err
	}

	sv, err := a.rollback(ctx, versionID)
	if err != nil {
		a.logger.Error("rolling back state version", "id", versionID, "subject", subject, "err", err)
		return nil, err
	}

	a.logger.Info("rolled back state version", "state_version", sv, "subject", subject)
	return sv, nil
}

func (a *Service) Upload(ctx context.Context, svID string, state []byte) error {
	var sv *Version
	err := a.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		var err error
		sv, err = a.db.getVersionForUpdate(ctx, svID)
		if err != nil {
			return err
		}
		sv, err = a.uploadStateAndOutputs(ctx, sv, state)
		if err != nil {
			return err
		}
		if err := a.cache.Set(cacheKey(svID), state); err != nil {
			a.logger.Error("caching state file", "err", err)
		}
		return nil
	})
	if err != nil {
		a.logger.Error("uploading state", "id", svID, "err", err)
		return err
	}

	a.logger.Debug("uploading state", "state_version", sv)
	return nil
}

func (a *Service) Download(ctx context.Context, svID string) ([]byte, error) {
	subject, err := a.CanAccess(ctx, rbac.DownloadStateAction, svID)
	if err != nil {
		return nil, err
	}

	if state, err := a.cache.Get(cacheKey(svID)); err == nil {
		a.logger.Debug("downloaded state", "id", svID, "subject", subject)
		return state, nil
	}

	state, err := a.db.getState(ctx, svID)
	if err != nil {
		a.logger.Error("downloading state", "id", svID, "subject", subject, "err", err)
		return nil, err
	}

	if err := a.cache.Set(cacheKey(svID), state); err != nil {
		a.logger.Error("caching state file", "err", err)
	}
	a.logger.Debug("downloaded state", "id", svID, "subject", subject)
	return state, nil
}

func (a *Service) GetOutput(ctx context.Context, outputID string) (*Output, error) {
	out, err := a.db.getOutput(ctx, outputID)
	if err != nil {
		a.logger.Error("retrieving state version output", "id", outputID, "err", err)
		return nil, err
	}

	subject, err := a.CanAccess(ctx, rbac.GetStateVersionOutputAction, out.StateVersionID)
	if err != nil {
		return nil, err
	}

	a.logger.Debug("retrieved state version output", "id", outputID, "subject", subject)
	return out, nil
}

func (a *Service) CanAccess(ctx context.Context, action rbac.Action, svID string) (internal.Subject, error) {
	sv, err := a.db.getVersion(ctx, svID)
	if err != nil {
		return nil, err
	}

	return a.workspace.CanAccess(ctx, action, sv.WorkspaceID)
}
