package configversion

import (
	"context"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/leg100/surl"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/tfeapi"
)

type (
	Service struct {
		workspace internal.Authorizer

		logger *slog.Logger
		db     *pgdb
		cache  internal.Cache
		tfeapi *tfe
		api    *api
	}

	Options struct {
		Logger *slog.Logger

		WorkspaceAuthorizer internal.Authorizer
		MaxConfigSize       int64

		internal.Cache
		*sql.DB
		*surl.Signer
		*tfeapi.Responder
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		logger: opts.Logger,
	}

	svc.workspace = opts.WorkspaceAuthorizer

	svc.db = &pgdb{opts.DB}
	svc.cache = opts.Cache
	svc.tfeapi = &tfe{
		logger:        opts.Logger,
		tfeClient:     &svc,
		Signer:        opts.Signer,
		Responder:     opts.Responder,
		maxConfigSize: opts.MaxConfigSize,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}

	// Fetch config version when API requests config version be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeConfig, svc.tfeapi.include)
	// Fetch ingress attributes when API requests ingress attributes be included
	// in the response
	opts.Responder.Register(tfeapi.IncludeIngress, svc.tfeapi.includeIngressAttributes)

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
}

func (s *Service) Create(ctx context.Context, workspaceID string, opts CreateOptions) (*ConfigurationVersion, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.CreateConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		s.logger.Error("constructing configuration version", "id", cv.ID, "subject", subject, "err", err)
		return nil, err
	}
	if err := s.db.CreateConfigurationVersion(ctx, cv); err != nil {
		s.logger.Error("creating configuration version", "id", cv.ID, "subject", subject, "err", err)
		return nil, err
	}
	s.logger.Info("created configuration version", "id", cv.ID, "subject", subject)
	return cv, nil
}

func (s *Service) List(ctx context.Context, workspaceID string, opts ListOptions) (*resource.Page[*ConfigurationVersion], error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.ListConfigurationVersionsAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cvl, err := s.db.ListConfigurationVersions(ctx, workspaceID, ListOptions{PageOptions: opts.PageOptions})
	if err != nil {
		s.logger.Error("listing configuration versions", "err", err)
		return nil, err
	}

	s.logger.Debug("listed configuration versions", "subject", subject)
	return cvl, nil
}

func (s *Service) Get(ctx context.Context, cvID string) (*ConfigurationVersion, error) {
	subject, err := s.canAccess(ctx, rbac.GetConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		s.logger.Error("retrieving configuration version", "id", cvID, "subject", subject, "err", err)
		return nil, err
	}
	s.logger.Debug("retrieved configuration version", "id", cvID, "subject", subject)
	return cv, nil
}

func (s *Service) GetLatest(ctx context.Context, workspaceID string) (*ConfigurationVersion, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.GetConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		s.logger.Error("retrieving latest configuration version", "workspace_id", workspaceID, "subject", subject, "err", err)
		return nil, err
	}
	s.logger.Debug("retrieved latest configuration version", "workspace_id", workspaceID, "subject", subject)
	return cv, nil
}

func (s *Service) Delete(ctx context.Context, cvID string) error {
	subject, err := s.canAccess(ctx, rbac.DeleteConfigurationVersionAction, cvID)
	if err != nil {
		return err
	}

	err = s.db.DeleteConfigurationVersion(ctx, cvID)
	if err != nil {
		s.logger.Error("deleting configuration version", "id", cvID, "subject", subject, "err", err)
		return err
	}
	s.logger.Debug("deleted configuration version", "id", cvID, "subject", subject)
	return nil
}

func (s *Service) canAccess(ctx context.Context, action rbac.Action, cvID string) (internal.Subject, error) {
	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		return nil, err
	}
	return s.workspace.CanAccess(ctx, action, cv.WorkspaceID)
}
