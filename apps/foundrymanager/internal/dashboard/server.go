package dashboard

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Params struct {
	Logger     *slog.Logger
	Addr       string
	Supervisor Supervisor
	Versions   VersionManager
	Profiles   ProfileStore
	Logs       LogReader
}

func Start(ctx context.Context, params Params) <-chan error {
	const readHeaderTimeout = 3 * time.Second
	mux := http.NewServeMux()
	registerHandlers(mux, params.Supervisor, params.Versions, params.Profiles, params.Logger)
	registerLogHandlers(mux, params.Logs, params.Logger)

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
			errCh <- err
		}
		close(errCh)
	}()

	params.Logger.Info("dashboard server listening", "params.Addr", params.Addr)
	return errCh
}
