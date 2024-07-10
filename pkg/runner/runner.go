package runner

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// Config represents configuration options
type Config struct {
	ServerHost               string
	ServerPort               string
	DefaultAudience          string
	TokenSourceCacheDuration string
	DebugGoproxy             bool
}

// Runner represents main routine interface
type Runner interface {
	Start(ctx context.Context) error
}

type runner struct {
	*Config
}

// New returns Runner instance
func New(config *Config) (Runner, error) {
	return &runner{config}, nil
}

func (r *runner) Start(ctx context.Context) error {
	sigCh := make(chan os.Signal, 1)
	defer close(sigCh)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p, err := newProc(r.Config)
	if err != nil {
		return err
	}

	ech, err := p.Start(ctx)
	if err != nil {
		return err
	}
	defer p.Stop()

	for {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
			return nil
		case err := <-ech:
			if err != context.Canceled {
				slog.ErrorContext(ctx, "error occurred", "error", err)
			}
		}
	}
}
