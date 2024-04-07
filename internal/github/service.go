package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/vcs"
)

type (
	// Service is the service for github app management
	Service struct {
		logger *slog.Logger

		GithubHostname string

		site         internal.Authorizer
		organization internal.Authorizer
		db           *pgdb
		web          *webHandlers
	}

	Options struct {
		*sql.DB
		html.Renderer
		vcs.Publisher
		*internal.HostnameService

		Logger              *slog.Logger
		GithubHostname      string
		SkipTLSVerification bool
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		logger:         opts.Logger,
		GithubHostname: opts.GithubHostname,
		site:           &internal.SiteAuthorizer{Logger: opts.Logger},
		organization:   &organization.Authorizer{Logger: opts.Logger},
		db:             &pgdb{opts.DB},
	}
	svc.web = &webHandlers{
		Renderer:        opts.Renderer,
		HostnameService: opts.HostnameService,
		GithubHostname:  opts.GithubHostname,
		GithubSkipTLS:   opts.SkipTLSVerification,
		svc:             &svc,
	}
	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
}

func (a *Service) CreateApp(ctx context.Context, opts CreateAppOptions) (*App, error) {
	subject, err := a.site.CanAccess(ctx, rbac.CreateGithubAppAction, "")
	if err != nil {
		return nil, err
	}

	app := newApp(opts)

	if err := a.db.create(ctx, app); err != nil {
		a.logger.Error("creating github app", "app", app, "subject", subject, "err", err)
		return nil, err
	}

	a.logger.Info("created github app", "app", app, "subject", subject)
	return app, nil
}

func (a *Service) GetApp(ctx context.Context) (*App, error) {
	subject, err := a.site.CanAccess(ctx, rbac.GetGithubAppAction, "")
	if err != nil {
		return nil, err
	}

	app, err := a.db.get(ctx)
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	a.logger.Debug("retrieved github app", "app", app, "subject", subject)

	return app, nil
}

func (a *Service) DeleteApp(ctx context.Context) error {
	subject, err := a.site.CanAccess(ctx, rbac.DeleteGithubAppAction, "")
	if err != nil {
		return err
	}

	err = a.db.delete(ctx)
	if err != nil {
		a.logger.Error("deleting github app", "subject", subject, "err", err)
		return err
	}

	a.logger.Info("deleted github app", "subject", subject)
	return nil
}

func (a *Service) ListInstallations(ctx context.Context) ([]*Installation, error) {
	app, err := a.db.get(ctx)
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	client, err := a.newClient(app)
	if err != nil {
		return nil, err
	}

	from, err := client.ListInstallations(ctx)
	if err != nil {
		return nil, err
	}

	to := make([]*Installation, len(from))
	for i, f := range from {
		to[i] = &Installation{Installation: f}
	}

	return to, nil
}

func (a *Service) GetInstallCredentials(ctx context.Context, installID int64) (*InstallCredentials, error) {
	app, err := a.db.get(ctx)
	if err != nil {
		return nil, err
	}
	client, err := a.newClient(app)
	if err != nil {
		return nil, err
	}
	install, err := client.GetInstallation(ctx, installID)
	if err != nil {
		return nil, err
	}
	creds := InstallCredentials{
		ID: installID,
		AppCredentials: AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	}
	switch install.GetTargetType() {
	case "Organization":
		creds.Organization = install.GetAccount().Login
	case "User":
		creds.User = install.GetAccount().Login
	default:
		return nil, fmt.Errorf("unexpected target type: %s", install.GetTargetType())
	}
	return &creds, nil
}

func (a *Service) DeleteInstallation(ctx context.Context, installID int64) error {
	app, err := a.db.get(ctx)
	if err != nil {
		return err
	}
	client, err := a.newClient(app)
	if err != nil {
		return err
	}
	if err := client.DeleteInstallation(ctx, installID); err != nil {
		return err
	}
	return nil
}

func (a *Service) newClient(app *App) (*Client, error) {
	return NewClient(ClientOptions{
		Hostname:            a.GithubHostname,
		SkipTLSVerification: true,
		AppCredentials: &AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	})
}
