// Package dashboard provides the profile-switching HTTP dashboard.
package dashboard

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

const (
	readHeaderTimeout = 3 * time.Second
	shutdownTimeout   = 5 * time.Second
)

// Start launches the dashboard HTTP server and returns a channel that receives
// the first non-ErrServerClosed error (or is closed on clean shutdown).
func Start(
	ctx context.Context,
	logger *slog.Logger,
	addr string,
	profiles []profile.Profile,
	sw Switcher,
) <-chan error {
	refs := make([]profileRef, len(profiles))
	for i, p := range profiles {
		refs[i] = profileRef{Name: p.Name, Label: p.Label}
	}

	mux := http.NewServeMux()
	registerHandlers(mux, refs, sw, logger)

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: readHeaderTimeout}
	errCh := make(chan error, 1)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("dashboard server stopped", "err", err)
			errCh <- err
		}
		close(errCh)
	}()

	logger.Info("dashboard server listening", "addr", addr)
	return errCh
}
