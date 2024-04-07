package vcsprovider

import (
	"context"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/github"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
	"github.com/tofutf/tofutf/internal/tfeapi"
	"github.com/tofutf/tofutf/internal/vcs"
)

type (
	Service struct {
		logger            *slog.Logger
		site              internal.Authorizer
		organization      internal.Authorizer
		db                *pgdb
		web               *webHandlers
		api               *tfe
		beforeDeleteHooks []func(context.Context, *VCSProvider) error
		githubapps        *github.Service

		*internal.HostnameService
		*factory
	}

	Options struct {
		*internal.HostnameService
		*sql.DB
		*tfeapi.Responder
		html.Renderer
		Logger *slog.Logger
		vcs.Subscriber

		GithubAppService        *github.Service
		GithubHostname          string
		GitlabHostname          string
		BitbucketServerHostname string
		SkipTLSVerification     bool
	}
)

func NewService(opts Options) *Service {
	factory := factory{
		githubapps:              opts.GithubAppService,
		githubHostname:          opts.GithubHostname,
		bitbucketServerHostname: opts.BitbucketServerHostname,
		gitlabHostname:          opts.GitlabHostname,
		skipTLSVerification:     opts.SkipTLSVerification,
	}
	svc := Service{
		logger:          opts.Logger,
		HostnameService: opts.HostnameService,
		githubapps:      opts.GithubAppService,
		site:            &internal.SiteAuthorizer{Logger: opts.Logger},
		organization:    &organization.Authorizer{Logger: opts.Logger},
		factory:         &factory,
		db: &pgdb{
			DB:      opts.DB,
			factory: &factory,
		},
	}
	svc.web = &webHandlers{
		Renderer:        opts.Renderer,
		HostnameService: opts.HostnameService,
		GithubHostname:  opts.GithubHostname,
		GitlabHostname:  opts.GitlabHostname,
		client:          &svc,
		githubApps:      opts.GithubAppService,
	}
	svc.api = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	// delete vcs providers when a github app is uninstalled
	opts.Subscribe(func(event vcs.Event) {
		// ignore events other than uninstallation events
		if event.Type != vcs.EventTypeInstallation || event.Action != vcs.ActionDeleted {
			return
		}
		// create user with unlimited permissions
		user := &internal.Superuser{Username: "vcs-provider-service"}
		ctx := internal.AddSubjectToContext(context.Background(), user)
		// list all vcsproviders using the app install
		providers, err := svc.ListVCSProvidersByGithubAppInstall(ctx, *event.GithubAppInstallID)
		if err != nil {
			return
		}
		// and delete them
		for _, prov := range providers {
			if _, err = svc.Delete(ctx, prov.ID); err != nil {
				return
			}
		}
	})
	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *Service) Create(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateVCSProviderAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	provider, err := a.newProvider(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := a.db.create(ctx, provider); err != nil {
		a.logger.Error("creating vcs provider", "provider", provider, "subject", subject, "err", err)
		return nil, err
	}

	a.logger.Info("created vcs provider", "provider", provider, "subject", subject)
	return provider, nil
}

func (a *Service) Update(ctx context.Context, id string, opts UpdateOptions) (*VCSProvider, error) {
	var (
		subject internal.Subject
		before  VCSProvider
		after   *VCSProvider
	)
	err := a.db.update(ctx, id, func(provider *VCSProvider) (err error) {
		subject, err = a.organization.CanAccess(ctx, rbac.UpdateVariableSetAction, provider.Organization)
		if err != nil {
			return err
		}
		// keep copy for logging the differences before and after update
		before = *provider
		after = provider
		if err := after.Update(opts); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		a.logger.Error("updating vcs provider", "vcs_provider_id", id, "err", err)
		return nil, err
	}

	a.logger.Info("updated vcs provider", "before", &before, "after", after, "subject", subject)
	return after, nil
}

func (a *Service) List(ctx context.Context, organization string) ([]*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListVCSProvidersAction, organization)
	if err != nil {
		return nil, err
	}

	providers, err := a.db.listByOrganization(ctx, organization)
	if err != nil {
		a.logger.Error("listing vcs providers", "organization", organization, "subject", subject, "err", err)
		return nil, err
	}

	a.logger.Debug("listed vcs providers", "organization", organization, "subject", subject)
	return providers, nil
}

func (a *Service) ListAllVCSProviders(ctx context.Context) ([]*VCSProvider, error) {
	subject, err := a.site.CanAccess(ctx, rbac.ListVCSProvidersAction, "")
	if err != nil {
		return nil, err
	}

	providers, err := a.db.list(ctx)
	if err != nil {
		a.logger.Error("listing vcs providers", "subject", subject, "err", err)
		return nil, err
	}
	a.logger.Debug("listed vcs providers", "subject", subject)
	return providers, nil
}

// ListVCSProvidersByGithubAppInstall is unauthenticated: only for internal use.
func (a *Service) ListVCSProvidersByGithubAppInstall(ctx context.Context, installID int64) ([]*VCSProvider, error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	providers, err := a.db.listByGithubAppInstall(ctx, installID)
	if err != nil {
		a.logger.Error("listing github app installation vcs providers", "subject", subject, "install", installID, "err", err)
		return nil, err
	}
	a.logger.Debug("listed github app installation vcs providers", "count", len(providers), "subject", subject, "install", installID)
	return providers, nil
}

func (a *Service) Get(ctx context.Context, id string) (*VCSProvider, error) {
	// Parameters only include VCS Provider ID, so we can only determine
	// authorization _after_ retrieving the provider
	provider, err := a.db.get(ctx, id)
	if err != nil {
		a.logger.Error("retrieving vcs provider", "id", id, "err", err)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetVCSProviderAction, provider.Organization)
	if err != nil {
		return nil, err
	}
	a.logger.Debug("retrieved vcs provider", "provider", provider, "subject", subject)

	return provider, nil
}

func (a *Service) GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error) {
	provider, err := a.Get(ctx, providerID)
	if err != nil {
		return nil, err
	}

	return provider.NewClient()
}

func (a *Service) Delete(ctx context.Context, id string) (*VCSProvider, error) {
	var (
		provider *VCSProvider
		subject  internal.Subject
	)
	err := a.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) (err error) {
		// retrieve vcs provider first in order to get organization for authorization
		provider, err = a.db.get(ctx, id)
		if err != nil {
			a.logger.Error("retrieving vcs provider", "id", id, "err", err)
			return err
		}

		subject, err = a.organization.CanAccess(ctx, rbac.DeleteVCSProviderAction, provider.Organization)
		if err != nil {
			return err
		}

		for _, hook := range a.beforeDeleteHooks {
			if err := hook(ctx, provider); err != nil {
				return err
			}
		}
		return a.db.delete(ctx, id)
	})
	if err != nil {
		a.logger.Error("deleting vcs provider", "provider", provider, "subject", subject, "err", err)
		return nil, err
	}

	a.logger.Info("deleted vcs provider", "provider", provider, "subject", subject)
	return provider, nil
}

func (a *Service) BeforeDeleteVCSProvider(hook func(context.Context, *VCSProvider) error) {
	a.beforeDeleteHooks = append(a.beforeDeleteHooks, hook)
}
