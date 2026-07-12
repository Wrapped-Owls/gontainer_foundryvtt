package dashboard

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func registerHandlers(
	mux *http.ServeMux,
	sw Switcher,
	vm VersionManager,
	ps ProfileStore,
	logger *slog.Logger,
) {
	registerProfileHandlers(mux, ps, logger)
	mux.HandleFunc("GET /profiles", func(w http.ResponseWriter, _ *http.Request) {
		profiles := ps.ListProfiles()
		refs := make([]profileRef, len(profiles))
		for i, p := range profiles {
			refs[i] = profileRef{Name: p.Name, Label: p.Label}
		}
		writeJSON(w, logger, http.StatusOK, profilesResponse{Active: sw.Active(), Profiles: refs})
	})
	mux.HandleFunc("POST /switch", func(w http.ResponseWriter, r *http.Request) {
		var body switchBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, logger, http.StatusBadRequest,
				errorResponse{Error: "invalid request body"})
			return
		}
		if !body.Force {
			if st, err := sw.FoundryStatus(r.Context()); err == nil && st.Active && st.Users > 0 {
				writeJSON(w, logger, http.StatusConflict, errorResponse{Error: fmt.Sprintf(
					"%d user(s) currently online; resend with force to switch anyway", st.Users)})
				return
			}
		}
		if err := sw.RequestSwitch(body.Profile); err != nil {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		resp := statusResponse{Active: sw.Active(), Version: sw.Version()}
		if st, err := sw.FoundryStatus(r.Context()); err == nil {
			resp.Online = true
			resp.WorldActive = st.Active
			resp.World = st.World
			resp.System = st.System
			resp.SystemVersion = st.SystemVersion
			resp.Users = st.Users
			resp.UptimeMS = st.UptimeMS
			if st.Version != "" {
				resp.Version = st.Version
			}
		}
		writeJSON(w, logger, http.StatusOK, resp)
	})
	mux.HandleFunc("GET /versions", func(w http.ResponseWriter, r *http.Request) {
		installed, err := vm.Installed(r.Context())
		if err != nil {
			logger.Error("dashboard: list versions failed", "err", err)
			writeJSON(w, logger, http.StatusInternalServerError,
				errorResponse{Error: "failed to list installed versions"})
			return
		}
		writeJSON(w, logger, http.StatusOK, versionsResponse{Active: sw.Version(), Installed: installed})
	})
	mux.HandleFunc("POST /versions/download", func(w http.ResponseWriter, r *http.Request) {
		var body downloadBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
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

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error("dashboard: failed to encode response", "err", err)
	}
}
