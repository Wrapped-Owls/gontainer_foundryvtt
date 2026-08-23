package dashboard

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

func registerSwitchHandlers(mux *http.ServeMux, sup Supervisor, logger *slog.Logger) {
	mux.HandleFunc("POST /switch", func(w http.ResponseWriter, r *http.Request) {
		var body switchBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: msgInvalidBody})
			return
		}
		if !body.Force {
			if online := onlinePlayers(r.Context(), sup); online > 0 {
				writeOnlineConflict(w, logger, online)
				return
			}
		}
		if err := sup.RequestSwitch(body.Profile); err != nil {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("POST /restart", func(w http.ResponseWriter, r *http.Request) {
		var body restartBody
		// restartBody has one optional field, so an empty body means force:false.
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: msgInvalidBody})
			return
		}
		if !body.Force {
			if online := onlinePlayers(r.Context(), sup); online > 0 {
				writeOnlineConflict(w, logger, online)
				return
			}
		}
		if err := sup.RequestRestart(); err != nil {
			writeJSON(w, logger, http.StatusConflict, errorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		resp := statusResponse{Active: sup.Active(), Version: sup.Version()}
		if st, err := sup.FoundryStatus(r.Context()); err == nil {
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
}
