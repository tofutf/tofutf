package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/http/html"
	"github.com/tofutf/tofutf/internal/json"
)

const (
	// shutdownTimeout is the time given for outstanding requests to finish
	// before shutdown.
	shutdownTimeout     = 1 * time.Second
	headersKey      key = "headers"
)

var (
	healthzPayload = json.MustMarshal(struct {
		Version string
		Commit  string
		Built   string
	}{
		Version: internal.Version,
		Commit:  internal.Commit,
		Built:   internal.Built,
	})
)

type (
	// ServerConfig is the http server config
	ServerConfig struct {
		SSL                  bool
		CertFile, KeyFile    string
		EnableRequestLogging bool
		DevMode              bool

		Handlers []internal.Handlers
		// middleware to intercept requests, executed in the order given.
		Middleware []mux.MiddlewareFunc
	}

	// Server is the http server for OTF
	Server struct {
		ServerConfig

		logger *slog.Logger
		server *http.Server
	}

	// unexported type for use with embedding values in contexts
	key string
)

// NewServer constructs the http server for OTF
func NewServer(logger *slog.Logger, cfg ServerConfig) (*Server, error) {
	if cfg.SSL {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("must provide both --cert-file and --key-file")
		}
	}

	r := mux.NewRouter()

	// Catch panics and return 500s
	r.Use(gorillaHandlers.RecoveryHandler(gorillaHandlers.PrintRecoveryStack(true)))

	r.Handle("/", otelhttp.WithRouteTag("/", http.RedirectHandler("/app/organizations", http.StatusFound)))

	// Serve static files
	if err := html.AddStaticHandler(logger, r, cfg.DevMode); err != nil {
		return nil, err
	}

	// Prometheus metrics
	r.Handle("/metrics", otelhttp.WithRouteTag("/metrics", promhttp.Handler()))

	r.Handle("/version", otelhttp.WithRouteTag("/version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.Write(healthzPayload) //nolint:errcheck
	})))

	r.Handle("/healthz", otelhttp.WithRouteTag("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.Write([]byte(`{"status":"OK"}`)) //nolint:errcheck
	})))

	// Subrouter for service routes
	svcRouter := r.NewRoute().Subrouter()

	// this middleware adds http headers from the request to the context
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), headersKey, r.Header)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Subject service routes to provided middleware, verifying tokens,
	// sessions.
	svcRouter.Use(cfg.Middleware...)

	// Add handlers for each service
	for _, h := range cfg.Handlers {
		h.AddHandlers(svcRouter)
	}

	// Optionally log every request
	if cfg.EnableRequestLogging {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				m := httpsnoop.CaptureMetrics(next, w, r)
				logger.Info("request",
					"duration", fmt.Sprintf("%dms", m.Duration.Milliseconds()),
					"status", m.Code,
					"method", r.Method,
					"path", fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery))
			})
		})
	}

	return &Server{
		logger:       logger,
		ServerConfig: cfg,
		server:       &http.Server{Handler: otelhttp.NewHandler(r, "/")},
	}, nil
}

// Start starts serving http traffic on the given listener and waits until the server exits due to
// error or the context is cancelled.
func (s *Server) Start(ctx context.Context, ln net.Listener) (err error) {
	errch := make(chan error)

	go func() {
		if s.SSL {
			errch <- s.server.ServeTLS(ln, s.CertFile, s.KeyFile)
		} else {
			errch <- s.server.Serve(ln)
		}
	}()

	s.logger.Info("started server", "address", ln.Addr().String(), "ssl", s.SSL)

	// Block until server stops listening or context is cancelled.
	select {
	case err := <-errch:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		s.logger.Info("gracefully shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			return s.server.Close()
		}

		return nil
	}
}

func HeadersFromContext(ctx context.Context) (http.Header, error) {
	headers, ok := ctx.Value(headersKey).(http.Header)
	if !ok {
		return nil, errors.New("no http headers found in context")
	}
	return headers, nil
}
