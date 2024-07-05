package idtoken

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/rinx/gcloud-auth-proxy/pkg/service/health"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type tokenSource struct {
	tokenSource oauth2.TokenSource
	expire      time.Time
}

type Server interface {
	health.Health
	http.Handler
	Start(ctx context.Context) <-chan error
	Stop() error
}

type server struct {
	mux    *http.ServeMux
	proxy  *httputil.ReverseProxy
	cancel context.CancelFunc

	// ctx for TokenSource
	ctx context.Context

	tss   map[string]tokenSource
	tssMu sync.Mutex

	tsCacheDuration time.Duration
	defaultAudience string

	started bool
}

func New(opts ...Option) (Server, error) {
	s := &server{
		mux:   &http.ServeMux{},
		proxy: &httputil.ReverseProxy{},
		tss:   map[string]tokenSource{},
	}

	for _, opt := range opts {
		if err := opt.Apply(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *server) IsHealthy() bool {
	return true
}

func (s *server) IsReady() bool {
	return s.started
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *server) idTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	aud := r.FormValue("audience")
	if aud == "" {
		aud = s.defaultAudience
	}

	ts, err := s.newTokenSource(aud)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.ErrorContext(r.Context(), "error on creating token-source", "error", err)
		return
	}

	tok, err := ts.Token()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.ErrorContext(r.Context(), "error on fetch token", "error", err)
		return
	}

	res := &Response{
		AccessToken:  tok.AccessToken,
		ExpiresIn:    int(time.Until(tok.Expiry).Seconds()),
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), "error on json encode", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
}

func (s *server) idTokenProxy(w http.ResponseWriter, r *http.Request) {
	ts, err := s.newTokenSource(s.defaultAudience)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tok, err := ts.Token()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tok.SetAuthHeader(r)

	s.proxy.ServeHTTP(w, r)
}

func (s *server) newTokenSource(aud string) (oauth2.TokenSource, error) {
	s.tssMu.Lock()
	defer s.tssMu.Unlock()

	if ts, ok := s.tss[aud]; ok {
		return ts.tokenSource, nil
	}

	nts, err := idtoken.NewTokenSource(s.ctx, aud)
	if err != nil {
		return nil, err
	}

	s.tss[aud] = tokenSource{
		tokenSource: nts,
		expire:      time.Now().Add(s.tsCacheDuration),
	}

	slog.Info("add tokensource cache", "audience", aud)

	return nts, nil
}

func (s *server) Start(ctx context.Context) <-chan error {
	ctx, s.cancel = context.WithCancel(ctx)
	ech := make(chan error, 1)

	ticker := time.NewTicker(time.Minute)

	go func() {
		defer close(ech)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil && err != context.Canceled {
					slog.ErrorContext(ctx, "error occurred", "error", err)
				}
				return
			case <-ticker.C:
				if err := s.refreshCache(); err != nil {
					ech <- err
				}
			}
		}
	}()

	s.ctx = ctx

	s.mux.HandleFunc("/idtoken", s.idTokenHandler)
	s.mux.HandleFunc("/idtoken/proxy", s.idTokenProxy)

	s.started = true

	slog.Info("idtoken service started", "default audience", s.defaultAudience)

	return ech
}

func (s *server) Stop() error {
	s.cancel()

	return nil
}

func (s *server) refreshCache() error {
	s.tssMu.Lock()
	defer s.tssMu.Unlock()

	for aud, ts := range s.tss {
		if ts.expire.After(time.Now()) {
			delete(s.tss, aud)

			slog.Info("delete tokensource cache", "audience", aud)
		}
	}

	return nil
}
