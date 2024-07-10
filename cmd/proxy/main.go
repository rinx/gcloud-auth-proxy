package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/rinx/gcloud-auth-proxy/pkg/runner"
	"github.com/spf13/cobra"
)

var (
	Version = "unknown"

	serverHost               string
	serverPort               string
	defaultAudience          string
	tokenSourceCacheDuration string
	debugGoproxy             bool
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cmd := &cobra.Command{
		Use:     "gcloud-auth-proxy",
		Short:   "",
		Version: Version,
		RunE:    run,
	}

	cmd.Flags().StringVar(&defaultAudience, "audience", "", "default audience (required)")
	if err := cmd.MarkFlagRequired("audience"); err != nil {
		slog.Error("error on mark flag required", "error", err)
	}

	cmd.Flags().StringVar(&serverHost, "host", "0.0.0.0", "server host")
	cmd.Flags().StringVar(&serverPort, "port", "8100", "server port")
	cmd.Flags().StringVar(&tokenSourceCacheDuration, "token-source-cache-duration", "30m", "token source cache duration")
	cmd.Flags().BoolVar(&debugGoproxy, "debug-goproxy", false, "verbose logging about elazarl/goproxy")

	if err := cmd.Execute(); err != nil {
		slog.Error("error occurred", "error", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg := &runner.Config{
		ServerHost:               serverHost,
		ServerPort:               serverPort,
		DefaultAudience:          defaultAudience,
		TokenSourceCacheDuration: tokenSourceCacheDuration,
		DebugGoproxy:             debugGoproxy,
	}

	r, err := runner.New(cfg)
	if err != nil {
		return err
	}

	return r.Start(context.Background())
}
