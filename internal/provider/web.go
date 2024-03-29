package provider

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html"
)

type (
	webHandlers struct {
		signer   internal.Signer
		renderer html.Renderer

		client webProvidersClient
		system webHostnameClient
	}

	webProvidersClient interface {
		GetProviderVersions(context.Context, GetProviderVersionsOptions) (*ProviderVersions, error)
	}

	webHostnameClient interface {
		Hostname() string
	}
)

func (h *webHandlers) addHandlers(r *mux.Router) {

}
