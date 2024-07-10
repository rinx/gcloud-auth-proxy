package runner

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/rinx/gcloud-auth-proxy/pkg/router"
	"github.com/rinx/gcloud-auth-proxy/pkg/server"
	"github.com/rinx/gcloud-auth-proxy/pkg/service/google/idtoken"
	"github.com/rinx/gcloud-auth-proxy/pkg/service/health"
)

// Proc represents main process
type Proc interface {
	Start(ctx context.Context) (<-chan error, error)
	Stop() error
}

type proc struct {
	*Config

	cancel context.CancelFunc

	server  server.Server
	router  router.Router
	health  health.HealthCheck
	idtoken idtoken.Server
}

func newProc(cfg *Config) (Proc, error) {
	return &proc{
		Config: cfg,
		router: router.New(),
		health: health.New(),
	}, nil
}

func (p *proc) Start(ctx context.Context) (_ <-chan error, err error) {
	ctx, p.cancel = context.WithCancel(ctx)

	ech := make(chan error, 1)

	p.idtoken, err = idtoken.New(
		idtoken.WithDefaultAudience(p.DefaultAudience),
		idtoken.WithTokenSourceCacheDuration(p.TokenSourceCacheDuration),
		idtoken.WithDebugGoproxy(p.DebugGoproxy),
	)
	if err != nil {
		return nil, err
	}

	p.registerRoutes()
	p.health.Register(
		p.idtoken,
	)

	p.server, err = server.New(
		server.WithHost(p.ServerHost),
		server.WithPort(p.ServerPort),
		server.WithHandler(p.router),
	)
	if err != nil {
		return nil, err
	}

	idTokenEch := p.idtoken.Start(ctx)

	srvEch := p.server.Start(ctx)

	go func() {
		defer close(ech)
		defer p.idtoken.Stop()
		defer p.server.Stop()

		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil && err != context.Canceled {
					slog.ErrorContext(ctx, "error occurred", "error", err)
				}
				return
			case err := <-srvEch:
				ech <- err
			case err := <-idTokenEch:
				ech <- err
			}
		}
	}()

	return ech, nil
}

func (p *proc) Stop() error {
	p.cancel()

	return nil
}

func (p *proc) registerRoutes() {
	p.router.Register(map[router.RoutePattern]http.Handler{
		"/healthz":       p.health.Healthz(),
		"/readyz":        p.health.Readyz(),
		"/idtoken":       p.idtoken,
		"/idtoken/proxy": p.idtoken,
	})
}
