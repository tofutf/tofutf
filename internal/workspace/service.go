package workspace

import (
	"context"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/connections"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/pubsub"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
	"github.com/tofutf/tofutf/internal/team"
	"github.com/tofutf/tofutf/internal/tfeapi"
	"github.com/tofutf/tofutf/internal/user"
	"github.com/tofutf/tofutf/internal/vcsprovider"
)

type (
	Service struct {
		site                internal.Authorizer
		organization        internal.Authorizer
		internal.Authorizer // workspace authorizer

		logger      *slog.Logger
		db          *pgdb
		web         *webHandlers
		tfeapi      *tfe
		api         *api
		broker      *pubsub.Broker[*Workspace]
		connections *connections.Service

		beforeCreateHooks []func(context.Context, *Workspace) error
		afterCreateHooks  []func(context.Context, *Workspace) error
		beforeUpdateHooks []func(context.Context, *Workspace) error
	}

	Options struct {
		*sql.DB
		*sql.Listener
		*tfeapi.Responder
		html.Renderer

		Logger *slog.Logger

		OrganizationService *organization.Service
		VCSProviderService  *vcsprovider.Service
		TeamService         *team.Service
		ConnectionService   *connections.Service
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		logger: opts.Logger,
		Authorizer: &authorizer{
			logger: opts.Logger,
			db:     db,
		},
		db:           db,
		connections:  opts.ConnectionService,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
	}
	svc.web = &webHandlers{
		Renderer:     opts.Renderer,
		teams:        opts.TeamService,
		vcsproviders: opts.VCSProviderService,
		client:       &svc,
	}
	svc.tfeapi = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.broker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"workspaces",
		func(ctx context.Context, id string, action sql.Action) (*Workspace, error) {
			if action == sql.DeleteAction {
				return &Workspace{ID: id}, nil
			}
			return db.get(ctx, id)
		},
	)
	// Fetch workspace when API calls request workspace be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeWorkspace, svc.tfeapi.include)
	opts.Responder.Register(tfeapi.IncludeWorkspaces, svc.tfeapi.includeMany)
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.tfeapi.addHandlers(r)
	s.web.addTagHandlers(r)
	s.tfeapi.addTagHandlers(r)
	s.api.addHandlers(r)
}

func (s *Service) Watch(ctx context.Context) (<-chan pubsub.Event[*Workspace], func()) {
	return s.broker.Subscribe(ctx)
}

func (s *Service) Create(ctx context.Context, opts CreateOptions) (*Workspace, error) {
	ws, err := NewWorkspace(opts)
	if err != nil {
		s.logger.Error("constructing workspace", "err", err)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateWorkspaceAction, ws.Organization)
	if err != nil {
		return nil, err
	}

	err = s.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, hook := range s.beforeCreateHooks {
			if err := hook(ctx, ws); err != nil {
				return err
			}
		}
		if err := s.db.create(ctx, ws); err != nil {
			return err
		}
		// Optionally connect workspace to repo.
		if ws.Connection != nil {
			if err := s.connect(ctx, ws.ID, ws.Connection); err != nil {
				return err
			}
		}
		// Optionally create tags.
		if len(opts.Tags) > 0 {
			added, err := s.addTags(ctx, ws, opts.Tags)
			if err != nil {
				return err
			}
			ws.Tags = added
		}
		for _, hook := range s.afterCreateHooks {
			if err := hook(ctx, ws); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.logger.Error("creating workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject, "err", err)
		return nil, err
	}

	s.logger.Info("created workspace", "id", ws.ID, "name", ws.Name, "organization", ws.Organization, "subject", subject)

	return ws, nil
}

func (s *Service) BeforeCreateWorkspace(hook func(context.Context, *Workspace) error) {
	s.beforeCreateHooks = append(s.beforeCreateHooks, hook)
}

func (s *Service) AfterCreateWorkspace(hook func(context.Context, *Workspace) error) {
	s.afterCreateHooks = append(s.afterCreateHooks, hook)
}

func (s *Service) Get(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.GetWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		s.logger.Error("retrieving workspace", "subject", subject, "workspace", workspaceID, "err", err)
		return nil, err
	}

	s.logger.Debug("retrieved workspace", "subject", subject, "workspace", workspaceID)

	return ws, nil
}

func (s *Service) GetByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	ws, err := s.db.getByName(ctx, organization, workspace)
	if err != nil {
		s.logger.Error("retrieving workspace", "organization", organization, "workspace", workspace, "err", err)
		return nil, err
	}

	subject, err := s.CanAccess(ctx, rbac.GetWorkspaceAction, ws.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("retrieved workspace", "subject", subject, "organization", organization, "workspace", workspace)

	return ws, nil
}

func (s *Service) List(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	if opts.Organization == nil {
		// subject needs perms on site to list workspaces across site
		_, err := s.site.CanAccess(ctx, rbac.ListWorkspacesAction, "")
		if err != nil {
			return nil, err
		}
	} else {
		// check if subject has perms to list workspaces in organization
		_, err := s.organization.CanAccess(ctx, rbac.ListWorkspacesAction, *opts.Organization)
		if err == internal.ErrAccessNotPermitted {
			// user does not have org-wide perms; fallback to listing workspaces
			// for which they have workspace-level perms.
			subject, err := internal.SubjectFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if user, ok := subject.(*user.User); ok {
				return s.db.listByUsername(ctx, user.Username, *opts.Organization, opts.PageOptions)
			}
		} else if err != nil {
			return nil, err
		}
	}

	return s.db.list(ctx, opts)
}

func (s *Service) ListConnectedWorkspaces(ctx context.Context, vcsProviderID, repoPath string) ([]*Workspace, error) {
	return s.db.listByConnection(ctx, vcsProviderID, repoPath)
}

func (s *Service) BeforeUpdateWorkspace(hook func(context.Context, *Workspace) error) {
	s.beforeUpdateHooks = append(s.beforeUpdateHooks, hook)
}

func (s *Service) Update(ctx context.Context, workspaceID string, opts UpdateOptions) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.UpdateWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	// update the workspace and optionally connect/disconnect to/from vcs repo.
	var updated *Workspace
	err = s.db.Tx(ctx, func(ctx context.Context, _ pggen.Querier) error {
		var connect *bool
		updated, err = s.db.update(ctx, workspaceID, func(ws *Workspace) (err error) {
			for _, hook := range s.beforeUpdateHooks {
				if err := hook(ctx, ws); err != nil {
					return err
				}
			}
			connect, err = ws.Update(opts)
			return err
		})
		if err != nil {
			return err
		}
		if connect != nil {
			if *connect {
				if err := s.connect(ctx, workspaceID, updated.Connection); err != nil {
					return err
				}
			} else {
				if err := s.disconnect(ctx, workspaceID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		s.logger.Error("updating workspace", "workspace", workspaceID, "subject", subject, "err", err)
		return nil, err
	}

	s.logger.Info("updated workspace", "workspace", workspaceID, "subject", subject)

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, workspaceID string) (*Workspace, error) {
	subject, err := s.CanAccess(ctx, rbac.DeleteWorkspaceAction, workspaceID)
	if err != nil {
		return nil, err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// disconnect repo before deleting
	if ws.Connection != nil {
		if err := s.disconnect(ctx, ws.ID); err != nil {
			return nil, err
		}
	}

	if err := s.db.delete(ctx, ws.ID); err != nil {
		s.logger.Error("deleting workspace", "id", ws.ID, "name", ws.Name, "subject", subject, "err", err)
		return nil, err
	}

	s.logger.Info("deleted workspace", "id", ws.ID, "name", ws.Name, "subject", subject)

	return ws, nil
}

// connect connects the workspace to a repo.
func (s *Service) connect(ctx context.Context, workspaceID string, connection *Connection) error {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = s.connections.Connect(ctx, connections.ConnectOptions{
		ConnectionType: connections.WorkspaceConnection,
		ResourceID:     workspaceID,
		VCSProviderID:  connection.VCSProviderID,
		RepoPath:       connection.Repo,
	})
	if err != nil {
		s.logger.Error("connecting workspace", "workspace", workspaceID, "subject", subject, "repo", connection.Repo, "err", err)
		return err
	}
	s.logger.Info("connected workspace repo", "workspace", workspaceID, "subject", subject, "repo", connection.Repo)

	return nil
}

func (s *Service) disconnect(ctx context.Context, workspaceID string) error {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.connections.Disconnect(ctx, connections.DisconnectOptions{
		ConnectionType: connections.WorkspaceConnection,
		ResourceID:     workspaceID,
	})
	if err != nil {
		s.logger.Error("disconnecting workspace", "workspace", workspaceID, "subject", subject, "err", err)
		return err
	}

	s.logger.Info("disconnected workspace", "workspace", workspaceID, "subject", subject)

	return nil
}

// SetCurrentRun sets the current run for the workspace
func (s *Service) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error) {
	return s.db.setCurrentRun(ctx, workspaceID, runID)
}
