package notifications

import (
	"context"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/pubsub"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/tfeapi"
)

type (
	Service struct {
		logger              *slog.Logger
		workspaceAuthorizer internal.Authorizer // authorize workspaces actions
		db                  *pgdb
		api                 *tfe
		broker              *pubsub.Broker[*Config]
	}

	Options struct {
		*sql.DB
		*sql.Listener
		*tfeapi.Responder
		Logger *slog.Logger

		WorkspaceAuthorizer internal.Authorizer
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		logger:              opts.Logger,
		workspaceAuthorizer: opts.WorkspaceAuthorizer,
		db:                  &pgdb{opts.DB},
	}
	svc.api = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	// Register with broker so that it can relay run events
	svc.broker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"notification_configurations",
		func(ctx context.Context, id string, action sql.Action) (*Config, error) {
			if action == sql.DeleteAction {
				return &Config{ID: id}, nil
			}
			return svc.db.get(ctx, id)
		},
	)
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
}

func (s *Service) Watch(ctx context.Context) (<-chan pubsub.Event[*Config], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Create(ctx context.Context, workspaceID string, opts CreateConfigOptions) (*Config, error) {
	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.CreateNotificationConfigurationAction, workspaceID)
	if err != nil {
		return nil, err
	}

	nc, err := NewConfig(workspaceID, opts)
	if err != nil {
		s.logger.Error("constructing notification config", "subject", subject, "err", err)
		return nil, err
	}

	if err := s.db.create(ctx, nc); err != nil {
		s.logger.Error("creating notification config", "config", nc, "subject", subject, "err", err)
		return nil, err
	}

	s.logger.Info("creating notification config", "config", nc, "subject", subject)
	return nc, nil
}

func (s *Service) Update(ctx context.Context, id string, opts UpdateConfigOptions) (*Config, error) {
	var subject internal.Subject
	updated, err := s.db.update(ctx, id, func(nc *Config) (err error) {
		subject, err = s.workspaceAuthorizer.CanAccess(ctx, rbac.UpdateNotificationConfigurationAction, nc.WorkspaceID)
		if err != nil {
			return err
		}
		return nc.update(opts)
	})
	if err != nil {
		s.logger.Error("updating notification config", "id", id, "subject", subject, "err", err)
		return nil, err
	}
	s.logger.Info("updated notification config", "updated", updated, "subject", subject)
	return updated, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Config, error) {
	nc, err := s.db.get(ctx, id)
	if err != nil {
		s.logger.Error("retrieving notification config", "id", id, "err", err)
		return nil, err
	}

	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.GetNotificationConfigurationAction, nc.WorkspaceID)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("retrieved notification config", "config", nc, "subject", subject)
	return nc, nil
}

func (s *Service) List(ctx context.Context, workspaceID string) ([]*Config, error) {
	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.ListNotificationConfigurationsAction, workspaceID)
	if err != nil {
		return nil, err
	}
	configs, err := s.db.list(ctx, workspaceID)
	if err != nil {
		s.logger.Error("listing notification configs", "id", workspaceID, "err", err)
		return nil, err
	}
	s.logger.Debug("listed notification configs", "total", len(configs), "subject", subject)
	return configs, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	nc, err := s.db.get(ctx, id)
	if err != nil {
		s.logger.Error("retrieving notification config", "id", id, "err", err)
		return err
	}

	subject, err := s.workspaceAuthorizer.CanAccess(ctx, rbac.DeleteNotificationConfigurationAction, nc.WorkspaceID)
	if err != nil {
		return err
	}

	if err := s.db.delete(ctx, id); err != nil {
		s.logger.Error("deleting notification config", "id", id, "err", err)
		return err
	}

	s.logger.Info("deleted notification config", "config", nc, "subject", subject)
	return nil
}
