package dashboard

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func Start(
	ctx context.Context,
	logger *slog.Logger,
	addr string,
	profiles []profile.Profile,
	sw Switcher,
) <-chan error {
	const readHeaderTimeout = 3 * time.Second
	refs := make([]profileRef, len(profiles))
	for i, p := range profiles {
		refs[i] = profileRef{Name: p.Name, Label: p.Label}
	}

	mux := http.NewServeMux()
	registerHandlers(mux, refs, sw, logger)

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: readHeaderTimeout}
	errCh := make(chan error, 1)

	context.AfterFunc(ctx, func() {
		const shutdownTimeout = 5 * time.Second
		shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	})
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
