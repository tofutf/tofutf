// Package gpgkeys handles the gpg key management functionality of tofutf.
package gpgkeys

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/gorilla/mux"
	"github.com/leg100/surl"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/organization"
	"github.com/tofutf/tofutf/internal/rbac"
	"github.com/tofutf/tofutf/internal/sql"
	"github.com/tofutf/tofutf/internal/tfeapi"
)

type (
	Options struct {
		Logger *slog.Logger
		*sql.DB
		*internal.HostnameService
		*surl.Signer
		html.Renderer
		*tfeapi.Responder
		OrganizationAuthorizer internal.Authorizer
	}

	Service struct {
		logger       *slog.Logger
		organization internal.Authorizer

		db *pgdb

		api *apiHandlers
		web *webHandlers
		tfe *tfeHandlers

		client        *http.Client
		orgAuthorizer internal.Authorizer
	}
)

// NewService returns a new instance of the service.
func NewService(opts Options) (*Service, error) {
	svc := Service{
		logger:        opts.Logger,
		organization:  &organization.Authorizer{Logger: opts.Logger},
		client:        &http.Client{},
		db:            &pgdb{opts.DB},
		orgAuthorizer: opts.OrganizationAuthorizer,
	}
	svc.api = &apiHandlers{
		svc: &svc,
	}
	svc.web = &webHandlers{
		client:   &svc,
		signer:   opts.Signer,
		renderer: opts.Renderer,
		system:   opts.HostnameService,
	}
	svc.tfe = &tfeHandlers{
		svc:       &svc,
		Responder: opts.Responder,
	}

	return &svc, nil
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
	s.tfe.addHandlers(r)
}

type CreateOptions struct {
	RegistryName string `schema:"registry_name,required"`
	Organization string
	ASCIIArmor   string
	Type         string
}

func (s *Service) Create(ctx context.Context, opts CreateOptions) (*GPGKey, error) {
	if opts.RegistryName != "private" {
		return nil, fmt.Errorf("invalid registry_name, only private registry supports gpg key functionality")
	}

	_, err := s.orgAuthorizer.CanAccess(ctx, rbac.CreateGPGKeyAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	keyReader := bytes.NewReader([]byte(opts.ASCIIArmor))
	entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
	if err != nil {
		return nil, err
	}

	keyID := strings.ToUpper(hex.EncodeToString(entityList[0].PrimaryKey.Fingerprint))

	key := &GPGKey{
		ID:               internal.NewID("gpg"),
		OrganizationName: opts.Organization,
		ASCIIArmor:       opts.ASCIIArmor,
		CreatedAt:        internal.CurrentTimestamp(nil),
		UpdatedAt:        internal.CurrentTimestamp(nil),
		KeyID:            keyID,
	}

	err = s.db.createRegistryGPGKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

type GetOptions struct {
	RegistryName string
	Organization string
	KeyID        string
}

func (s *Service) Get(ctx context.Context, opts GetOptions) (*GPGKey, error) {
	if opts.RegistryName != "private" {
		return nil, fmt.Errorf("invalid registry_name, only private registry supports gpg key functionality")
	}

	_, err := s.orgAuthorizer.CanAccess(ctx, rbac.GetGPGKeyAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	key, err := s.db.getRegistryGPGKey(ctx, pgGetOptions{
		organization: opts.Organization,
		keyID:        opts.KeyID,
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

type DeleteOptions struct {
	RegistryName string
	Organization string
	KeyID        string
}

func (s *Service) Delete(ctx context.Context, opts DeleteOptions) (*GPGKey, error) {
	if opts.RegistryName != "private" {
		return nil, fmt.Errorf("invalid registry_name, only private registry supports gpg key functionality")
	}

	_, err := s.orgAuthorizer.CanAccess(ctx, rbac.GetGPGKeyAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	key, err := s.db.getRegistryGPGKey(ctx, pgGetOptions{
		organization: opts.Organization,
		keyID:        opts.KeyID,
	})
	if err != nil {
		return nil, err
	}

	err = s.db.deleteRegistryGPGKey(ctx, pgDeleteOpts{
		organization: opts.Organization,
		keyID:        opts.KeyID,
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

type ListOptions struct {
	RegistryName string
	Namespaces   []string
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*GPGKey, error) {
	if opts.RegistryName != "private" {
		return nil, fmt.Errorf("invalid registry_name, only private registry supports gpg key functionality")
	}

	for _, namespace := range opts.Namespaces {
		_, err := s.orgAuthorizer.CanAccess(ctx, rbac.GetGPGKeyAction, namespace)
		if err != nil {
			return nil, err
		}
	}

	keys, err := s.db.listRegistryGPGKeys(ctx, opts.Namespaces)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

type UpdateOptions struct {
	RegistryName string `schema:"registry_name,required"`
	Organization string `schema:"namespace,required"`
	KeyID        string `schema:"key_id,required"`

	NewType      string
	NewNamespace string
}

func (s *Service) Update(ctx context.Context, opts UpdateOptions) (*GPGKey, error) {
	if opts.RegistryName != "private" {
		return nil, fmt.Errorf("invalid registry_name, only private registry supports gpg key functionality")
	}

	if opts.NewType != "gpg-keys" {
		return nil, fmt.Errorf("invalid type, only gpg-keys is supported")
	}

	_, err := s.orgAuthorizer.CanAccess(ctx, rbac.GetGPGKeyAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	_, err = s.orgAuthorizer.CanAccess(ctx, rbac.GetGPGKeyAction, opts.NewNamespace)
	if err != nil {
		return nil, err
	}

	err = s.db.updateRegistryGPGKey(ctx, pgUpdateOpts{
		organizationName:    opts.Organization,
		keyID:               opts.KeyID,
		newOrganizationName: opts.NewNamespace,
		updatedAt:           internal.CurrentTimestamp(nil),
	})
	if err != nil {
		return nil, err
	}

	key, err := s.db.getRegistryGPGKey(ctx, pgGetOptions{
		organization: opts.NewNamespace,
		keyID:        opts.KeyID,
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}
