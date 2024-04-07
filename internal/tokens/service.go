package tokens

import (
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/tofutf/tofutf/internal"
)

type (
	Service struct {
		*factory
		*registry
		*sessionFactory

		site       internal.Authorizer // authorizes site access
		logger     *slog.Logger
		middleware mux.MiddlewareFunc
	}

	Options struct {
		GoogleIAPConfig

		Logger *slog.Logger
		Secret []byte
	}
)

func NewService(opts Options) (*Service, error) {
	svc := Service{
		logger: opts.Logger,
		site:   &internal.SiteAuthorizer{Logger: opts.Logger},
	}
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.factory = &factory{key: key}
	svc.sessionFactory = &sessionFactory{factory: svc.factory}
	svc.registry = &registry{
		kinds: make(map[Kind]SubjectGetter),
	}
	svc.middleware = newMiddleware(middlewareOptions{
		logger:          opts.Logger,
		GoogleIAPConfig: opts.GoogleIAPConfig,
		key:             key,
		registry:        svc.registry,
	})
	return &svc, nil
}

// Middleware returns middleware for authenticating tokens
func (a *Service) Middleware() mux.MiddlewareFunc { return a.middleware }
