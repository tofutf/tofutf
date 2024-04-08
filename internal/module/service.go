package module

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/surl"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/connections"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/repohooks"
	"github.com/tofutf/tofutf/internal/semver"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/sql/pggen"
	"github.com/tofutf/tofutf/internal/vcs"
	"github.com/tofutf/tofutf/internal/vcsprovider"
)

type (
	Service struct {
		db           *pgdb
		logger       *slog.Logger
		organization internal.Authorizer

		api          *api
		web          *webHandlers
		vcsproviders *vcsprovider.Service
		connections  *connections.Service
	}

	Options struct {
		Logger *slog.Logger

		*sql.DB
		*internal.HostnameService
		*surl.Signer
		html.Renderer

		RepohookService    *repohooks.Service
		VCSProviderService *vcsprovider.Service
		ConnectionsService *connections.Service
		VCSEventSubscriber vcs.Subscriber
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		logger:       opts.Logger,
		connections:  opts.ConnectionsService,
		organization: &organization.Authorizer{Logger: opts.Logger},
		db:           &pgdb{opts.DB},
		vcsproviders: opts.VCSProviderService,
	}
	svc.api = &api{
		svc:    &svc,
		Signer: opts.Signer,
	}
	svc.web = &webHandlers{
		Renderer:     opts.Renderer,
		client:       &svc,
		vcsproviders: opts.VCSProviderService,
		system:       opts.HostnameService,
	}
	publisher := &publisher{
		logger:       opts.Logger.With("component", "publisher"),
		vcsproviders: opts.VCSProviderService,
		modules:      &svc,
	}
	// Subscribe module publisher to incoming vcs events
	opts.VCSEventSubscriber.Subscribe(publisher.handle)

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// PublishModule publishes a new module from a VCS repository, enumerating through
// its git tags and releasing a module version for each tag.
func (s *Service) PublishModule(ctx context.Context, opts PublishOptions) (*Module, error) {
	vcsprov, err := s.vcsproviders.Get(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateModuleAction, vcsprov.Organization)
	if err != nil {
		return nil, err
	}

	module, err := s.publishModule(ctx, vcsprov.Organization, opts)
	if err != nil {
		s.logger.Error("publishing module", "subject", subject, "repo", opts.Repo, "err", err)
		return nil, err
	}
	s.logger.Info("published module", "subject", subject, "module", module)

	return module, nil
}

func (s *Service) publishModule(ctx context.Context, organization string, opts PublishOptions) (*Module, error) {
	name, provider, err := opts.Repo.Split()
	if err != nil {
		return nil, err
	}

	mod := newModule(CreateOptions{
		Name:         name,
		Provider:     provider,
		Organization: organization,
	})

	// persist module to db and connect to repository
	if err := s.db.createModule(ctx, mod); err != nil {
		return nil, err
	}

	var (
		client vcs.Client
		tags   []string
	)
	setup := func() (err error) {
		mod.Connection, err = s.connections.Connect(ctx, connections.ConnectOptions{
			ConnectionType: connections.ModuleConnection,
			ResourceID:     mod.ID,
			VCSProviderID:  opts.VCSProviderID,
			RepoPath:       string(opts.Repo),
		})
		if err != nil {
			return err
		}

		client, err = s.vcsproviders.GetVCSClient(ctx, opts.VCSProviderID)
		if err != nil {
			return err
		}

		tags, err = client.ListTags(ctx, vcs.ListTagsOptions{
			Repo: string(opts.Repo),
		})

		if err != nil {
			return err
		}
		return nil
	}
	if err := setup(); err != nil {
		return s.updateModuleStatus(ctx, mod, ModuleStatusSetupFailed)
	}
	if len(tags) == 0 {
		return s.updateModuleStatus(ctx, mod, ModuleStatusNoVersionTags)
	}
	for _, tag := range tags {
		// tags/<version> -> <version>
		_, version, found := strings.Cut(tag, "/")
		if !found {
			return nil, fmt.Errorf("malformed git ref: %s", tag)
		}
		// skip tags that are not semantic versions
		if !semver.IsValid(version) {
			continue
		}
		err := s.PublishVersion(ctx, PublishVersionOptions{
			ModuleID: mod.ID,
			// strip off v prefix if it has one
			Version: strings.TrimPrefix(version, "v"),
			Ref:     tag,
			Repo:    opts.Repo,
			Client:  client,
		})
		if err != nil {
			return nil, err
		}
	}
	if _, err := s.updateModuleStatus(ctx, mod, ModuleStatusSetupComplete); err != nil {
		return nil, err
	}
	// return module from db complete with versions
	return s.GetModuleByID(ctx, mod.ID)
}

// PublishVersion publishes a module version, retrieving its contents from a repository and
// uploading it to the module store.
func (s *Service) PublishVersion(ctx context.Context, opts PublishVersionOptions) error {
	modver, err := s.CreateVersion(ctx, CreateModuleVersionOptions{
		ModuleID: opts.ModuleID,
		Version:  opts.Version,
	})
	if err != nil {
		return err
	}

	tarball, _, err := opts.Client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: string(opts.Repo),
		Ref:  &opts.Ref,
	})
	if err != nil {
		return s.db.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     modver.ID,
			Status: ModuleVersionStatusCloneFailed,
			Error:  err.Error(),
		})
	}

	return s.uploadVersion(ctx, modver.ID, tarball)
}

func (s *Service) CreateModule(ctx context.Context, opts CreateOptions) (*Module, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.CreateModuleAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	module := newModule(opts)

	if err := s.db.createModule(ctx, module); err != nil {
		s.logger.Error("creating module", "subject", subject, "module", module, "err", err)
		return nil, err
	}

	s.logger.Info("created module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) ListModules(ctx context.Context, opts ListModulesOptions) ([]*Module, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.ListModulesAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	modules, err := s.db.listModules(ctx, opts)
	if err != nil {
		s.logger.Error("listing modules", "organization", opts.Organization, "subject", subject, "err", err)
		return nil, err
	}
	s.logger.Debug("listed modules", "organization", opts.Organization, "subject", subject)
	return modules, nil
}

func (s *Service) GetModule(ctx context.Context, opts GetModuleOptions) (*Module, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.GetModuleAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	module, err := s.db.getModule(ctx, opts)
	if err != nil {
		s.logger.Error("retrieving module", "module", opts, "err", err)
		return nil, err
	}

	s.logger.Debug("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) GetModuleByID(ctx context.Context, id string) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, id)
	if err != nil {
		s.logger.Error("retrieving module", "id", id, "err", err)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.GetModuleAction, module.Organization)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("retrieved module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) GetModuleByConnection(ctx context.Context, vcsProviderID, repoPath string) (*Module, error) {
	return s.db.getModuleByConnection(ctx, vcsProviderID, repoPath)
}

func (s *Service) DeleteModule(ctx context.Context, id string) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, id)
	if err != nil {
		s.logger.Error("retrieving module", "id", id, "err", err)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.DeleteModuleAction, module.Organization)
	if err != nil {
		return nil, err
	}

	err = s.db.Tx(ctx, func(ctx context.Context, _ pggen.Querier) error {
		// disconnect module prior to deletion
		if module.Connection != nil {
			err := s.connections.Disconnect(ctx, connections.DisconnectOptions{
				ConnectionType: connections.ModuleConnection,
				ResourceID:     module.ID,
			})
			if err != nil {
				return err
			}
		}
		return s.db.delete(ctx, id)
	})
	if err != nil {
		s.logger.Error("deleting module", "subject", subject, "module", module, "err", err)
		return nil, err
	}
	s.logger.Debug("deleted module", "subject", subject, "module", module)
	return module, nil
}

func (s *Service) RefreshModule(ctx context.Context, id string) (*Module, error) {
	logger := s.logger.With("service", "module", "op", "RefreshModule")

	module, err := s.db.getModuleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = s.organization.CanAccess(ctx, rbac.CreateModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	client, err := s.vcsproviders.GetVCSClient(ctx, module.Connection.VCSProviderID)
	if err != nil {
		return nil, err
	}

	tags, err := client.ListTags(ctx, vcs.ListTagsOptions{
		Repo: module.Connection.Repo,
	})
	if err != nil {
		return nil, err
	}

	exists := map[string]bool{}
	for _, version := range module.Versions {
		exists[version.Version] = true
	}

	logger.Info("listing tags", "exists", exists, "tags", tags)

	for _, tag := range tags {
		// tags/<version> -> <version>
		_, version, found := strings.Cut(tag, "/")
		if !found {
			logger.Info("skipping malformed git ref", "tag", tag)
			return nil, fmt.Errorf("malformed git ref: %s", tag)
		}
		// skip tags that are not semantic versions
		if !semver.IsValid(version) {
			logger.Info("skipping invalid version", "version", version)
			continue
		}

		finalVersion := strings.TrimPrefix(version, "v")

		// if it already exists then continue
		if _, ok := exists[finalVersion]; ok {
			logger.Info("skipping version that already exists", "version", finalVersion)
			continue
		}

		logger.Info("publishing newly detected version", "version", finalVersion)

		err := s.PublishVersion(ctx, PublishVersionOptions{
			ModuleID: module.ID,
			// strip off v prefix if it has one
			Version: finalVersion,
			Ref:     tag,
			Repo:    Repo(module.Connection.Repo),
			Client:  client,
		})
		if err != nil {
			return nil, err
		}
	}

	module, err = s.db.getModuleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (s *Service) CreateVersion(ctx context.Context, opts CreateModuleVersionOptions) (*ModuleVersion, error) {
	module, err := s.db.getModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.CreateModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	modver := newModuleVersion(opts)

	if err := s.db.createModuleVersion(ctx, modver); err != nil {
		s.logger.Error("creating module version", "organization", module.Organization, "subject", subject, "module_version", modver, "err", err)
		return nil, err
	}
	s.logger.Info("created module version", "organization", module.Organization, "subject", subject, "module_version", modver)
	return modver, nil
}

func (s *Service) GetModuleInfo(ctx context.Context, versionID string) (*TerraformModule, error) {
	tarball, err := s.db.getTarball(ctx, versionID)
	if err != nil {
		return nil, err
	}
	return unmarshalTerraformModule(tarball)
}

func (s *Service) updateModuleStatus(ctx context.Context, mod *Module, status ModuleStatus) (*Module, error) {
	mod.Status = status

	if err := s.db.updateModuleStatus(ctx, mod.ID, status); err != nil {
		s.logger.Error("updating module status", "module", mod.ID, "status", status, "err", err)
		return nil, err
	}
	s.logger.Info("updated module status", "module", mod.ID, "status", status)
	return mod, nil
}

func (s *Service) uploadVersion(ctx context.Context, versionID string, tarball []byte) error {
	module, err := s.db.getModuleByVersionID(ctx, versionID)
	if err != nil {
		return err
	}

	// validate tarball
	if _, err := unmarshalTerraformModule(tarball); err != nil {
		s.logger.Error("uploading module version", "module_version", versionID, "err", err)
		return s.db.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusRegIngressFailed,
			Error:  err.Error(),
		})
	}

	// save tarball, set status, and make it the latest version
	err = s.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		if err := s.db.saveTarball(ctx, versionID, tarball); err != nil {
			return err
		}
		err := s.db.updateModuleVersionStatus(ctx, UpdateModuleVersionStatusOptions{
			ID:     versionID,
			Status: ModuleVersionStatusOK,
		})
		if err != nil {
			return err
		}
		// re-retrieve module so that includes the above version with updated
		// status
		_, err = s.db.getModuleByID(ctx, module.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.logger.Error("uploading module version", "module_version_id", versionID, "err", err)
		return err
	}

	s.logger.Info("uploaded module version", "module_version", versionID)
	return nil
}

// downloadVersion should be accessed via signed URL
func (s *Service) downloadVersion(ctx context.Context, versionID string) ([]byte, error) {
	tarball, err := s.db.getTarball(ctx, versionID)
	if err != nil {
		s.logger.Error("downloading module", "module_version_id", versionID, "err", err)
		return nil, err
	}
	s.logger.Debug("downloaded module", "module_version_id", versionID)
	return tarball, nil
}

//lint:ignore U1000 to be used later
func (s *Service) deleteVersion(ctx context.Context, versionID string) (*Module, error) {
	module, err := s.db.getModuleByID(ctx, versionID)
	if err != nil {
		s.logger.Error("retrieving module", "id", versionID, "err", err)
		return nil, err
	}

	subject, err := s.organization.CanAccess(ctx, rbac.DeleteModuleVersionAction, module.Organization)
	if err != nil {
		return nil, err
	}

	if err = s.db.deleteModuleVersion(ctx, versionID); err != nil {
		s.logger.Error("deleting module", "subject", subject, "module", module, "err", err)
		return nil, err
	}
	s.logger.Debug("deleted module", "subject", subject, "module", module)

	// return module w/o deleted version
	return s.db.getModuleByID(ctx, module.ID)
}
