package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/health"
)

func startHealthServer(
	ctx context.Context,
	logger *slog.Logger,
	addr string,
	foundryPort int,
) <-chan error {
	const (
		probeTimeout      = 3 * time.Second
		readHeaderTimeout = 3 * time.Second
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		probe := health.Default()
		if foundryPort > 0 {
			probe.URL = fmt.Sprintf("http://localhost:%d/api/status", foundryPort)
		}
		cctx, cancel := context.WithTimeout(r.Context(), probeTimeout)
		defer cancel()
		if err := health.Check(cctx, probe); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: readHeaderTimeout}
	errCh := make(chan error, 1)
	context.AfterFunc(ctx, func() {
		_ = srv.Close()
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server stopped", "err", err)
			errCh <- err
		}
	}()
	logger.Info("health server listening", "addr", addr)
	return errCh
}
