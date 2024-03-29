package provider

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/surl"
	"github.com/tofutf/tofutf/internal/http/decode"
	"github.com/tofutf/tofutf/internal/tfeapi"
)

type apiHandlers struct {
	*surl.Signer

	svc *Service
}

func (h *apiHandlers) addHandlers(r *mux.Router) {
	// authenticated module api routes
	//
	// Implements the Module Registry Protocol:
	//
	// https://developer.hashicorp.com/terraform/internals/provider-registry-protocol
	r = r.PathPrefix(tfeapi.ProviderV1Prefix).Subrouter()

	r.HandleFunc("/{namespace}/{type}/versions", h.listAvailableVersions).Methods("GET")
	r.HandleFunc("/{namespace}/{type}/{version}/download/{os}/{arch}", h.findProviderPackage).Methods("GET")
}

// List Available Versions for a Specific Module.
//
// https://developer.hashicorp.com/terraform/internals/provider-registry-protocol#list-available-versions
func (h *apiHandlers) listAvailableVersions(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Type      string `schema:"type,required"`
		Namespace string `schema:"namespace,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	versions, err := h.svc.GetProviderVersions(r.Context(), GetProviderVersionsOptions{
		Type:      params.Type,
		Namespace: params.Namespace,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")

	if err := json.NewEncoder(w).Encode(versions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// findProviderPackage retrieves a specific provider package.
//
// https://developer.hashicorp.com/terraform/internals/provider-registry-protocol#find-a-provider-package
func (h *apiHandlers) findProviderPackage(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Type      string `schema:"type,required"`
		Namespace string `schema:"namespace,required"`
		Version   string `schema:"version"`
		OS        string `schema:"os"`
		Arch      string `schema:"arch"`
	}
	if err := decode.Route(&params, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pkg, err := h.svc.FindProviderPackage(r.Context(), FindProviderPackageOptions{
		Type:      params.Type,
		Namespace: params.Namespace,
		Version:   params.Version,
		OS:        params.OS,
		Arch:      params.Arch,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")

	if err := json.NewEncoder(w).Encode(pkg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
