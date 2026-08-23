package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const msgInvalidBody = "invalid request body"

func registerHandlers(
	mux *http.ServeMux,
	sup Supervisor,
	vm VersionManager,
	ps ProfileStore,
	logger *slog.Logger,
) {
	registerProfileHandlers(mux, sup, ps, logger)
	registerSwitchHandlers(mux, sup, logger)
	registerVersionHandlers(mux, sup, vm, logger)
}

// onlinePlayers treats an unreachable server as zero, so a restart is never blocked
// by the outage it would fix.
func onlinePlayers(ctx context.Context, sup Supervisor) int {
	st, err := sup.FoundryStatus(ctx)
	if err != nil || !st.Active {
		return 0
	}
	return st.Users
}

func writeOnlineConflict(w http.ResponseWriter, logger *slog.Logger, online int) {
	writeJSON(w, logger, http.StatusConflict, errorResponse{Error: fmt.Sprintf(
		"%d user(s) currently online; resend with force to proceed anyway", online,
	)})
}

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error("dashboard: failed to encode response", "err", err)
	}
}
