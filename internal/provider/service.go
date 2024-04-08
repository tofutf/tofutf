// Package provider provides terraform provider registry functionality to tofutf.
package provider

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/leg100/surl"
	"github.com/pkg/errors"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/sql"
)

type (
	Options struct {
		Logger *slog.Logger

		*sql.DB
		*internal.HostnameService
		*surl.Signer
		html.Renderer

		ProxyURL           string
		ProxyIsArtifactory bool
	}

	Service struct {
		logger       *slog.Logger
		organization internal.Authorizer

		api *apiHandlers
		web *webHandlers

		proxyURL           string
		proxyIsArtifactory bool

		client *http.Client
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		logger:             opts.Logger,
		organization:       &organization.Authorizer{Logger: opts.Logger},
		proxyURL:           opts.ProxyURL,
		proxyIsArtifactory: opts.ProxyIsArtifactory,
		client:             &http.Client{},
	}
	svc.api = &apiHandlers{
		svc:    &svc,
		Signer: opts.Signer,
	}
	svc.web = &webHandlers{
		client:   &svc,
		signer:   opts.Signer,
		renderer: opts.Renderer,
		system:   opts.HostnameService,
	}

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

type GetProviderVersionsOptions struct {
	Namespace string
	Type      string
}

type ProviderVersions struct {
	Versions []ProviderVersionsVersions `json:"versions"`
}

type ProviderVersionsVersions struct {
	Version   string             `json:"version"`
	Protocols []string           `json:"protocols"`
	Platforms []ProviderPlatform `json:"platforms"`
}

type ProviderPlatform struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`
}

// GetProviderVersions retrieves the list of versions that exist for a given provider.
func (s *Service) GetProviderVersions(ctx context.Context, options GetProviderVersionsOptions) (*ProviderVersions, error) {
	s.logger.Info("proxying provider versions request", "namespace", options.Namespace, "type", options.Type)

	versionsURL, err := url.JoinPath(s.proxyURL, options.Namespace, options.Type, "versions")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	request, err := http.NewRequest(http.MethodGet, versionsURL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if response.StatusCode != http.StatusOK {
		err := errors.Errorf("unexpected error from upstream provider registery: %d", response.StatusCode)

		s.logger.Error("failed to proxy provider versions request", "namespace", options.Namespace, "type", options.Type, "url", versionsURL, "err", err)
		return nil, errors.WithStack(err)
	}

	var versions ProviderVersions
	decoder := json.NewDecoder(response.Body)

	err = decoder.Decode(&versions)
	if err != nil {
		return nil, err
	}

	return &versions, nil
}

type FindProviderPackageOptions struct {
	Namespace string
	Type      string
	Version   string
	OS        string
	Arch      string
}

type ProviderVersionManifest struct {
	Protocols           []string `json:"protocols"`
	Os                  string   `json:"os"`
	Arch                string   `json:"arch"`
	Filename            string   `json:"filename"`
	DownloadURL         string   `json:"download_url"`
	ShasumsURL          string   `json:"shasums_url"`
	ShasumsSignatureURL string   `json:"shasums_signature_url"`
	Shasum              string   `json:"shasum"`
	SigningKeys         struct {
		GpgPublicKeys []struct {
			KeyID          string `json:"key_id"`
			ASCIIArmor     string `json:"ascii_armor"`
			TrustSignature string `json:"trust_signature"`
			Source         string `json:"source"`
			SourceURL      string `json:"source_url"`
		} `json:"gpg_public_keys"`
	} `json:"signing_keys"`
}

func (s Service) FindProviderPackage(ctx context.Context, options FindProviderPackageOptions) (*ProviderVersionManifest, error) {
	s.logger.Info("proxying provider download request", "namespace", options.Namespace, "type", options.Type, "version", options.Version, "os", options.OS, "arch", options.Arch)

	var manifestURL string
	var err error
	if s.proxyIsArtifactory {
		// artifactory is kinda garbage and doesn't actually build correct URLs for terraform at this point in time.
		manifestURL, err = url.JoinPath(s.proxyURL, options.Namespace, options.Type, options.Version, options.OS, options.Arch)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		manifestURL, err = url.JoinPath(s.proxyURL, options.Namespace, options.Type, options.Version, "download", options.OS, options.Arch)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	request, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if response.StatusCode != http.StatusOK {
		err := errors.Errorf("unexpected error from upstream provider registery: %d", response.StatusCode)

		s.logger.Error("failed to proxy provider download request", "namespace", options.Namespace, "type", options.Type, "version", options.Version, "os", options.OS, "arch", options.Arch, "url", manifestURL, "err", err)
		return nil, errors.WithStack(err)
	}

	var manifest ProviderVersionManifest
	decoder := json.NewDecoder(response.Body)

	err = decoder.Decode(&manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}
