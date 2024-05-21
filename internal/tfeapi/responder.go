package tfeapi

import (
	"bytes"
	"io"
	"net/http"

	"github.com/hashicorp/jsonapi"
)

const mediaType = "application/vnd.api+json"

// Responder handles responding to API requests.
type Responder struct {
	*includer
}

func NewResponder() *Responder {
	return &Responder{
		includer: &includer{
			registrations: make(map[IncludeName][]IncludeFunc),
		},
	}
}

func (res *Responder) Respond(w http.ResponseWriter, r *http.Request, payload any, status int) {
	var b bytes.Buffer
	bw := io.Writer(&b)
	err := jsonapi.MarshalPayload(bw, payload)
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(status)
	w.Write(b.Bytes()) //nolint:errcheck
}
