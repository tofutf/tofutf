package gpgkeys

import "github.com/gorilla/mux"

type (
	apiHandlers struct {
		svc *Service
	}
)

func (h *apiHandlers) addHandlers(r *mux.Router) {

}
