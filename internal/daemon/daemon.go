// Package daemon configures and starts the otfd daemon and its subsystems.
package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/agent"
	"github.com/tofutf/tofutf/internal/api"
	"github.com/tofutf/tofutf/internal/authenticator"
	"github.com/tofutf/tofutf/internal/bitbucketserver"
	"github.com/tofutf/tofutf/internal/configversion"
	"github.com/tofutf/tofutf/internal/connections"
	"github.com/tofutf/tofutf/internal/disco"
	"github.com/tofutf/tofutf/internal/ghapphandler"
	"github.com/tofutf/tofutf/internal/github"
	"github.com/tofutf/tofutf/internal/gitlab"
	"github.com/tofutf/tofutf/internal/http"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/inmem"
	"github.com/tofutf/tofutf/internal/loginserver"
	"github.com/tofutf/tofutf/internal/logs"
	"github.com/tofutf/tofutf/internal/module"
	"github.com/tofutf/tofutf/internal/notifications"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/provider"
	"github.com/tofutf/tofutf/internal/releases"
	"github.com/tofutf/tofutf/internal/repohooks"
	"github.com/tofutf/tofutf/internal/run"
	"github.com/tofutf/tofutf/internal/scheduler"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/state"
	"github.com/tofutf/tofutf/internal/team"
	"github.com/tofutf/tofutf/internal/tfeapi"
	"github.com/tofutf/tofutf/internal/tokens"
	"github.com/tofutf/tofutf/internal/user"
	"github.com/tofutf/tofutf/internal/variable"
	"github.com/tofutf/tofutf/internal/vcs"
	"github.com/tofutf/tofutf/internal/vcsprovider"
	"github.com/tofutf/tofutf/internal/workspace"
	"golang.org/x/sync/errgroup"
)

type (
	Daemon struct {
		Config
		logr.Logger

		*sql.DB

		Organizations *organization.Service
		Runs          *run.Service
		Workspaces    *workspace.Service
		Variables     *variable.Service
		Notifications *notifications.Service
		Logs          *logs.Service
		State         *state.Service
		Configs       *configversion.Service
		Modules       *module.Service
		Providers     *provider.Service
		VCSProviders  *vcsprovider.Service
		Tokens        *tokens.Service
		Teams         *team.Service
		Users         *user.Service
		GithubApp     *github.Service
		RepoHooks     *repohooks.Service
		Agents        *agent.Service
		Connections   *connections.Service
		System        *internal.HostnameService

		handlers []internal.Handlers
		listener *sql.Listener
		agent    agentDaemon
	}

	agentDaemon interface {
		Start(context.Context) error
		Registered() <-chan *agent.Agent
	}
)

// New builds a new daemon and establishes a connection to the database and
// migrates it to the latest schema. Close() should be called to close this
// connection.
func New(ctx context.Context, logger logr.Logger, cfg Config) (*Daemon, error) {
	if cfg.DevMode {
		logger.Info("enabled developer mode")
	}
	if err := cfg.Valid(); err != nil {
		return nil, err
	}

	hostnameService := internal.NewHostnameService(cfg.Host)
	hostnameService.SetWebhookHostname(cfg.WebhookHost)

	renderer, err := html.NewRenderer(cfg.DevMode)
	if err != nil {
		return nil, fmt.Errorf("setting up web page renderer: %w", err)
	}
	cache, err := inmem.NewCache(*cfg.CacheConfig)
	if err != nil {
		return nil, err
	}
	logger.Info("started cache", "max_size", cfg.CacheConfig.Size, "ttl", cfg.CacheConfig.TTL)

	var db *sql.DB
	const maxRetries = 10
	retryInterval := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.New(ctx, sql.Options{
			Logger:     logger,
			ConnString: cfg.Database,
		})
		if err == nil {
			break
		}

		logger.Info("waiting for database", "err", err)

		time.Sleep(retryInterval)
	}

	if err != nil {
		return nil, err
	}

	// listener listens to database events
	listener := sql.NewListener(logger, db)

	// responder responds to TFE API requests
	responder := tfeapi.NewResponder()

	// Setup url signer
	signer := internal.NewSigner(cfg.Secret)

	tokensService, err := tokens.NewService(tokens.Options{
		Logger:          logger,
		GoogleIAPConfig: cfg.GoogleIAPConfig,
		Secret:          cfg.Secret,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up authentication middleware: %w", err)
	}

	orgService := organization.NewService(organization.Options{
		Logger:                       logger,
		DB:                           db,
		Listener:                     listener,
		Renderer:                     renderer,
		Responder:                    responder,
		RestrictOrganizationCreation: cfg.RestrictOrganizationCreation,
		TokensService:                tokensService,
	})

	teamService := team.NewService(team.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Responder:           responder,
		OrganizationService: orgService,
		TokensService:       tokensService,
	})
	userService := user.NewService(user.Options{
		Logger:        logger,
		DB:            db,
		Renderer:      renderer,
		Responder:     responder,
		TokensService: tokensService,
		SiteToken:     cfg.SiteToken,
		TeamService:   teamService,
	})
	// promote nominated users to site admin
	if err := userService.SetSiteAdmins(ctx, cfg.SiteAdmins...); err != nil {
		return nil, err
	}

	githubAppService := github.NewService(github.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		HostnameService:     hostnameService,
		GithubHostname:      cfg.GithubHostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
	})

	vcsEventBroker := &vcs.Broker{}

	vcsProviderService := vcsprovider.NewService(vcsprovider.Options{
		Logger:                  logger,
		DB:                      db,
		Renderer:                renderer,
		Responder:               responder,
		HostnameService:         hostnameService,
		GithubAppService:        githubAppService,
		GithubHostname:          cfg.GithubHostname,
		GitlabHostname:          cfg.GitlabHostname,
		BitbucketServerHostname: cfg.BitbucketServerHostname,
		SkipTLSVerification:     cfg.SkipTLSVerification,
		Subscriber:              vcsEventBroker,
	})
	repoService := repohooks.NewService(ctx, repohooks.Options{
		Logger:              logger,
		DB:                  db,
		HostnameService:     hostnameService,
		OrganizationService: orgService,
		VCSProviderService:  vcsProviderService,
		GithubAppService:    githubAppService,
		VCSEventBroker:      vcsEventBroker,
	})
	repoService.RegisterCloudHandler(vcs.GithubKind, github.HandleEvent)
	repoService.RegisterCloudHandler(vcs.GitlabKind, gitlab.HandleEvent)
	repoService.RegisterCloudHandler(vcs.BitbucketServer, bitbucketserver.HandleEvent)

	connectionService := connections.NewService(ctx, connections.Options{
		Logger:             logger,
		DB:                 db,
		VCSProviderService: vcsProviderService,
		RepoHooksService:   repoService,
	})
	releasesService := releases.NewService(releases.Options{
		Logger: logger,
		DB:     db,
	})
	if cfg.DisableLatestChecker == nil || !*cfg.DisableLatestChecker {
		releasesService.StartLatestChecker(ctx)
	}
	workspaceService := workspace.NewService(workspace.Options{
		Logger:              logger,
		DB:                  db,
		Listener:            listener,
		Renderer:            renderer,
		Responder:           responder,
		ConnectionService:   connectionService,
		TeamService:         teamService,
		OrganizationService: orgService,
		VCSProviderService:  vcsProviderService,
	})
	configService := configversion.NewService(configversion.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Responder:           responder,
		Cache:               cache,
		Signer:              signer,
		MaxConfigSize:       cfg.MaxConfigSize,
	})

	runService := run.NewService(run.Options{
		Logger:               logger,
		DB:                   db,
		Listener:             listener,
		Renderer:             renderer,
		Responder:            responder,
		WorkspaceAuthorizer:  workspaceService,
		OrganizationService:  orgService,
		WorkspaceService:     workspaceService,
		ConfigVersionService: configService,
		VCSProviderService:   vcsProviderService,
		Cache:                cache,
		VCSEventSubscriber:   vcsEventBroker,
		Signer:               signer,
		ReleasesService:      releasesService,
		TokensService:        tokensService,
	})
	logsService := logs.NewService(logs.Options{
		Logger:        logger,
		DB:            db,
		RunAuthorizer: runService,
		Cache:         cache,
		Listener:      listener,
		Verifier:      signer,
	})
	moduleService := module.NewService(module.Options{
		Logger:             logger,
		DB:                 db,
		Renderer:           renderer,
		HostnameService:    hostnameService,
		VCSProviderService: vcsProviderService,
		Signer:             signer,
		ConnectionsService: connectionService,
		RepohookService:    repoService,
		VCSEventSubscriber: vcsEventBroker,
	})
	providerService := provider.NewService(provider.Options{
		Logger:             logger,
		DB:                 db,
		HostnameService:    hostnameService,
		Signer:             signer,
		Renderer:           renderer,
		ProxyURL:           cfg.ProviderProxy.URL,
		ProxyIsArtifactory: cfg.ProviderProxy.IsArtifactory,
	})
	stateService := state.NewService(state.Options{
		Logger:           logger,
		DB:               db,
		WorkspaceService: workspaceService,
		Cache:            cache,
		Renderer:         renderer,
		Responder:        responder,
		Signer:           signer,
	})
	variableService := variable.NewService(variable.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Responder:           responder,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
		RunClient:           runService,
	})

	agentService := agent.NewService(agent.ServiceOptions{
		Logger:           logger,
		DB:               db,
		Renderer:         renderer,
		Responder:        responder,
		RunService:       runService,
		WorkspaceService: workspaceService,
		TokensService:    tokensService,
		Listener:         listener,
	})

	agentDaemon, err := agent.NewServerDaemon(
		logger.WithValues("component", "agent"),
		*cfg.AgentConfig,
		agent.ServerDaemonOptions{
			WorkspaceService:            workspaceService,
			VariableService:             variableService,
			StateService:                stateService,
			ConfigurationVersionService: configService,
			RunService:                  runService,
			LogsService:                 logsService,
			AgentService:                agentService,
			HostnameService:             hostnameService,
		},
	)
	if err != nil {
		return nil, err
	}

	authenticatorService, err := authenticator.NewAuthenticatorService(ctx, authenticator.Options{
		Logger:          logger,
		Renderer:        renderer,
		HostnameService: hostnameService,
		TokensService:   tokensService,
		OpaqueHandlerConfigs: []authenticator.OpaqueHandlerConfig{
			{
				ClientConstructor: github.NewOAuthClient,
				OAuthConfig: authenticator.OAuthConfig{
					Hostname:     cfg.GithubHostname,
					Name:         string(vcs.GithubKind),
					Endpoint:     github.OAuthEndpoint,
					Scopes:       github.OAuthScopes,
					ClientID:     cfg.GithubClientID,
					ClientSecret: cfg.GithubClientSecret,
				},
			},
			{
				ClientConstructor: gitlab.NewOAuthClient,
				OAuthConfig: authenticator.OAuthConfig{
					Hostname:     cfg.GitlabHostname,
					Name:         string(vcs.GitlabKind),
					Endpoint:     gitlab.OAuthEndpoint,
					Scopes:       gitlab.OAuthScopes,
					ClientID:     cfg.GitlabClientID,
					ClientSecret: cfg.GitlabClientSecret,
				},
			},
		},
		IDTokenHandlerConfig: cfg.OIDC,
		SkipTLSVerification:  cfg.SkipTLSVerification,
	})
	if err != nil {
		return nil, err
	}

	notificationService := notifications.NewService(notifications.Options{
		Logger:              logger,
		DB:                  db,
		Listener:            listener,
		Responder:           responder,
		WorkspaceAuthorizer: workspaceService,
	})

	handlers := []internal.Handlers{
		teamService,
		userService,
		workspaceService,
		stateService,
		orgService,
		variableService,
		vcsProviderService,
		moduleService,
		providerService,
		runService,
		logsService,
		repoService,
		authenticatorService,
		loginserver.NewServer(loginserver.Options{
			Secret:      cfg.Secret,
			Renderer:    renderer,
			UserService: userService,
		}),
		configService,
		notificationService,
		githubAppService,
		agentService,
		disco.Service{},
		&ghapphandler.Handler{
			Logger:       logger,
			Publisher:    vcsEventBroker,
			GithubApps:   githubAppService,
			VCSProviders: vcsProviderService,
		},
		&api.Handlers{},
		&tfeapi.Handlers{},
	}

	return &Daemon{
		Config:        cfg,
		Logger:        logger,
		handlers:      handlers,
		Organizations: orgService,
		System:        hostnameService,
		Runs:          runService,
		Workspaces:    workspaceService,
		Variables:     variableService,
		Notifications: notificationService,
		Logs:          logsService,
		State:         stateService,
		Configs:       configService,
		Modules:       moduleService,
		Providers:     providerService,
		VCSProviders:  vcsProviderService,
		Tokens:        tokensService,
		Teams:         teamService,
		Users:         userService,
		RepoHooks:     repoService,
		GithubApp:     githubAppService,
		Connections:   connectionService,
		Agents:        agentService,
		DB:            db,
		agent:         agentDaemon,
		listener:      listener,
	}, nil
}

// Start the otfd daemon and block until ctx is cancelled or an error is
// returned. The started channel is closed once the daemon has started.
func (d *Daemon) Start(ctx context.Context, started chan struct{}) error {
	// Cancel context the first time a func started with g.Go() fails
	g, ctx := errgroup.WithContext(ctx)

	// close all db connections upon exit
	defer d.DB.Close()

	// Construct web server and start listening on port
	server, err := http.NewServer(d.Logger, http.ServerConfig{
		SSL:                  d.SSL,
		CertFile:             d.CertFile,
		KeyFile:              d.KeyFile,
		EnableRequestLogging: d.EnableRequestLogging,
		DevMode:              d.DevMode,
		Middleware:           []mux.MiddlewareFunc{d.Tokens.Middleware()},
		Handlers:             d.handlers,
	})
	if err != nil {
		return fmt.Errorf("setting up http server: %w", err)
	}
	ln, err := net.Listen("tcp", d.Address)
	if err != nil {
		return err
	}
	defer ln.Close()

	// Unless user has set a hostname, set the hostname to the listening address
	// of the http server.
	if d.Host == "" {
		listenAddress := ln.Addr().(*net.TCPAddr)
		d.System.SetHostname(internal.NormalizeAddress(listenAddress))
	}

	d.V(0).Info("set system hostname", "hostname", d.System.Hostname())
	d.V(0).Info("set webhook hostname", "webhook_hostname", d.System.WebhookHostname())

	subsystems := []*Subsystem{
		{
			Name:   "listener",
			Logger: d.Logger,
			System: d.listener,
		},
		{
			Name:   "proxy",
			Logger: d.Logger,
			System: d.Logs,
		},
		{
			Name:      "reporter",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(run.ReporterLockID),
			System: &run.Reporter{
				Logger:          d.Logger.WithValues("component", "reporter"),
				VCS:             d.VCSProviders,
				HostnameService: d.System,
				Workspaces:      d.Workspaces,
				Runs:            d.Runs,
				Configs:         d.Configs,
			},
		},
		{
			Name:      "notifier",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(notifications.LockID),
			System: notifications.NewNotifier(notifications.NotifierOptions{
				Logger:             d.Logger,
				HostnameService:    d.System,
				WorkspaceClient:    d.Workspaces,
				RunClient:          d.Runs,
				NotificationClient: d.Notifications,
				DB:                 d.DB,
			}),
		},
		{
			Name:      "job-allocator",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(agent.AllocatorLockID),
			System:    d.Agents.NewAllocator(d.Logger),
		},
		{
			Name:      "agent-manager",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(agent.ManagerLockID),
			System:    d.Agents.NewManager(),
		},
		{
			Name:   "agent-daemon",
			Logger: d.Logger,
			DB:     d.DB,
			System: d.agent,
		},
	}
	if !d.DisableScheduler {
		subsystems = append(subsystems, &Subsystem{
			Name:      "scheduler",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(scheduler.LockID),
			System: scheduler.NewScheduler(scheduler.Options{
				Logger:          d.Logger,
				WorkspaceClient: d.Workspaces,
				RunClient:       d.Runs,
			}),
		})
	}
	for _, ss := range subsystems {
		if err := ss.Start(ctx, g); err != nil {
			return err
		}
	}

	// Wait for database events listener start listening; otherwise some tests may fail
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 10):
		return fmt.Errorf("timed out waiting for database events listener to start")
	case <-d.listener.Started():
	}
	// Wait for agent to register; otherwise some tests may fail
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-d.agent.Registered():
	}

	// Run HTTP/JSON-API server and web app
	g.Go(func() error {
		if err := server.Start(ctx, ln); err != nil {
			return fmt.Errorf("http server terminated: %w", err)
		}
		return nil
	})

	// Inform the caller the daemon has started
	close(started)

	// Block until error or Ctrl-C received.
	return g.Wait()
}
