package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
)

// Server represents server process
type Server interface {
	Start(ctx context.Context) <-chan error
	Stop() error
}

type server struct {
	cancel  context.CancelFunc
	handler http.Handler
	srv     *http.Server

	host string
	port string
}

// New returns Server instance
func New(opts ...Option) (Server, error) {
	s := &server{}

	for _, opt := range opts {
		if err := opt.Apply(s); err != nil {
			return nil, err
		}
	}

	s.srv = &http.Server{
		Addr:    net.JoinHostPort(s.host, s.port),
		Handler: s.handler,
	}

	return s, nil
}

func (s *server) Start(ctx context.Context) <-chan error {
	ctx, s.cancel = context.WithCancel(ctx)

	ech := make(chan error, 1)

	go func() {
		defer close(ech)

		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil && err != context.Canceled {
					slog.ErrorContext(ctx, "error occurred", "error", err)
				}
			default:
			}

			slog.Info("server started", "addr", s.srv.Addr)

			if err := s.srv.ListenAndServe(); err != nil {
				ech <- err
			}
		}
	}()

	return ech
}

func (s *server) Stop() error {
	s.cancel()

	return s.srv.Close()
}
