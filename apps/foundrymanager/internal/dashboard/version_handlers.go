package dashboard

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func registerVersionHandlers(
	mux *http.ServeMux,
	sup Supervisor,
	vm VersionManager,
	logger *slog.Logger,
) {
	mux.HandleFunc("GET /versions", func(w http.ResponseWriter, r *http.Request) {
		installed, err := vm.Installed(r.Context())
		if err != nil {
			logger.Error("dashboard: list versions failed", "err", err)
			writeJSON(w, logger, http.StatusInternalServerError,
				errorResponse{Error: "failed to list installed versions"})
			return
		}
		writeJSON(
			w,
			logger,
			http.StatusOK,
			versionsResponse{Active: sup.Version(), Installed: installed},
		)
	})
	mux.HandleFunc("POST /versions/download", func(w http.ResponseWriter, r *http.Request) {
		var body downloadBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: msgInvalidBody})
			return
		}
		if body.Version == "" {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: "version is required"})
			return
		}
		if err := vm.Download(r.Context(), body.Version, body.URL); err != nil {
			logger.Error("dashboard: download version failed", "version", body.Version, "err", err)
			writeJSON(w, logger, http.StatusBadGateway, errorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}
