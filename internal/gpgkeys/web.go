package gpgkeys

import (
	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html"
)

type (
	webHandlers struct {
		client   webClient
		system   webHostnameClient
		signer   internal.Signer
		renderer html.Renderer
	}

	webClient interface {
		//
	}

	webHostnameClient interface {
		Hostname() string
	}
)

func (h *webHandlers) addHandlers(r *mux.Router) {

}
