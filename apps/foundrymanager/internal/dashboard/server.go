package dashboard

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

type Params struct {
	Logger   *slog.Logger
	Addr     string
	Profiles []profile.Profile
	Switcher Switcher
	Versions VersionManager
}

func Start(ctx context.Context, params Params) <-chan error {
	const readHeaderTimeout = 3 * time.Second
	refs := make([]profileRef, len(params.Profiles))
	for i, p := range params.Profiles {
		refs[i] = profileRef{Name: p.Name, Label: p.Label}
	}

	mux := http.NewServeMux()
	registerHandlers(mux, refs, params.Switcher, params.Versions, params.Logger)

	srv := &http.Server{Addr: params.Addr, Handler: mux, ReadHeaderTimeout: readHeaderTimeout}
	errCh := make(chan error, 1)

	context.AfterFunc(ctx, func() {
		const shutdownTimeout = 5 * time.Second
		shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			params.Logger.Error("dashboard server stopped", "err", err)
			errCh <- err
		}
		close(errCh)
	}()

	params.Logger.Info("dashboard server listening", "params.Addr", params.Addr)
	return errCh
}
