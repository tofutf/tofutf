package gpgkeys

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal/http/decode"
	"github.com/tofutf/tofutf/internal/resource"
	"github.com/tofutf/tofutf/internal/tfeapi"
	"github.com/tofutf/tofutf/internal/tfeapi/types"
)

type (
	tfeHandlers struct {
		svc *Service
		*tfeapi.Responder
	}
)

func (h *tfeHandlers) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV1).Subrouter()

	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/private-registry/gpg-keys#gpg-keys-api
	r.HandleFunc("/registry/{registry_name}/v2/gpg-keys", h.list).Methods(http.MethodGet)
	r.HandleFunc("/registry/{registry_name}/v2/gpg-keys", h.create).Methods(http.MethodPost)
	r.HandleFunc("/registry/{registry_name}/v2/gpg-keys/{namespace}/{key_id}", h.get).Methods(http.MethodGet)
	r.HandleFunc("/registry/{registry_name}/v2/gpg-keys/{namespace}/{key_id}", h.update).Methods(http.MethodPatch)
	r.HandleFunc("/registry/{registry_name}/v2/gpg-keys/{namespace}/{key_id}", h.delete).Methods(http.MethodDelete)
}

type listRouteParams struct {
	RegistryName string   `schema:"registry_name,required"`
	Namespaces   []string `schema:"filter[namespace]"`
	types.ListOptions
}

func (h *tfeHandlers) list(w http.ResponseWriter, r *http.Request) {
	var params listRouteParams
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	keys, err := h.svc.List(r.Context(), ListOptions{
		RegistryName: params.RegistryName,
		Namespaces:   params.Namespaces,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// client expects a page, whereas List returns full result set, so
	// convert to page first
	page := resource.NewPage(keys, resource.PageOptions(params.ListOptions), nil)

	// convert items

	items := make([]*types.GPGKey, len(page.Items))
	for i, from := range page.Items {
		items[i] = h.toGPGKey(from)
	}

	h.RespondWithPage(w, r, items, page.Pagination)
}

type getRouteParams struct {
	RegistryName string `schema:"registry_name,required"`
	Namespace    string `schema:"namespace,required"`
	KeyID        string `schema:"key_id,required"`
}

func (h *tfeHandlers) get(w http.ResponseWriter, r *http.Request) {
	var params getRouteParams
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	key, err := h.svc.Get(r.Context(), GetOptions{
		RegistryName: params.RegistryName,
		Organization: params.Namespace,
		KeyID:        params.KeyID,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	h.Respond(w, r, h.toGPGKey(key), http.StatusOK)
}

type deleteRouteParams struct {
	RegistryName string `schema:"registry_name,required"`
	Namespace    string `schema:"namespace,required"`
	KeyID        string `schema:"key_id,required"`
}

func (h *tfeHandlers) delete(w http.ResponseWriter, r *http.Request) {
	var params deleteRouteParams
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	key, err := h.svc.Delete(r.Context(), DeleteOptions{
		RegistryName: params.RegistryName,
		Organization: params.Namespace,
		KeyID:        params.KeyID,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	h.Respond(w, r, h.toGPGKey(key), http.StatusCreated)
}

type updateRouteParams struct {
	RegistryName string `schema:"registry_name,required"`
	Namespace    string `schema:"namespace,required"`
	KeyID        string `schema:"key_id,required"`
}

func (h *tfeHandlers) update(w http.ResponseWriter, r *http.Request) {
	var params updateRouteParams
	err := decode.Route(&params, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var payload types.GPGKeyUpdateOptions
	if err := decode.All(&payload, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	key, err := h.svc.Update(r.Context(), UpdateOptions{
		RegistryName: params.RegistryName,
		Organization: params.Namespace,
		KeyID:        params.KeyID,
		NewType:      payload.Type,
		NewNamespace: payload.Namespace,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	h.Respond(w, r, h.toGPGKey(key), http.StatusCreated)
}

type createRouteParams struct {
	RegistryName string `schema:"registry_name,required"`
}

func (h *tfeHandlers) create(w http.ResponseWriter, r *http.Request) {
	var params createRouteParams
	err := decode.Route(&params, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var opts types.GPGKeyCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	key, err := h.svc.Create(r.Context(), CreateOptions{
		RegistryName: params.RegistryName,
		Organization: opts.Namespace,
		ASCIIArmor:   opts.ASCIIArmor,
		Type:         opts.Type,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	h.Respond(w, r, h.toGPGKey(key), http.StatusCreated)
}

func (h *tfeHandlers) toGPGKey(key *GPGKey) *types.GPGKey {
	return &types.GPGKey{
		ID:             key.ID,
		ASCIIArmor:     key.ASCIIArmor,
		Namespace:      key.OrganizationName,
		KeyID:          key.KeyID,
		TrustSignature: "",
		Source:         "",
		SourceURL:      nil,
		CreatedAt:      key.CreatedAt.Format(types.ISO8601),
		UpdatedAt:      key.UpdatedAt.Format(types.ISO8601),
	}
}
