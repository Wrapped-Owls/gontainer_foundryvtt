package dashboard

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func registerHandlers(mux *http.ServeMux, refs []profileRef, sw Switcher, logger *slog.Logger) {
	mux.HandleFunc("GET /profiles", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, logger, http.StatusOK, profilesResponse{
			Active:   sw.Active(),
			Profiles: refs,
		})
	})
	mux.HandleFunc("POST /switch", func(w http.ResponseWriter, r *http.Request) {
		var body switchBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, logger, http.StatusBadRequest,
				errorResponse{Error: "invalid request body"})
			return
		}
		if err := sw.RequestSwitch(body.Profile); err != nil {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, logger, http.StatusOK, statusResponse{
			Active:  sw.Active(),
			Version: sw.Version(),
		})
	})
}

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error("dashboard: failed to encode response", "err", err)
	}
}
