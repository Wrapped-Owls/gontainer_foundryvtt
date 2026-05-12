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
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		probe := health.Default()
		if foundryPort > 0 {
			probe.URL = fmt.Sprintf("http://localhost:%d/api/status", foundryPort)
		}
		cctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := health.Check(cctx, probe); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server stopped", "err", err)
			errCh <- err
		}
	}()
	logger.Info("health server listening", "addr", addr)
	return errCh
}
