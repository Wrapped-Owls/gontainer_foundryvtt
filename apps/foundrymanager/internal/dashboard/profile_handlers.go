package dashboard

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func registerProfileHandlers(
	mux *http.ServeMux,
	sup Supervisor,
	ps ProfileStore,
	logger *slog.Logger,
) {
	mux.HandleFunc("GET /profiles", func(w http.ResponseWriter, _ *http.Request) {
		profiles := ps.ListProfiles()
		refs := make([]profileRef, len(profiles))
		for i, p := range profiles {
			refs[i] = profileRef{Name: p.Name, Label: p.Label, Version: p.Version, World: p.World}
		}
		writeJSON(w, logger, http.StatusOK, profilesResponse{Active: sup.Active(), Profiles: refs})
	})
	mux.HandleFunc("GET /profiles/{name}", func(w http.ResponseWriter, r *http.Request) {
		p, ok := ps.GetProfile(r.PathValue("name"))
		if !ok {
			writeJSON(w, logger, http.StatusNotFound, errorResponse{Error: "profile not found"})
			return
		}
		writeJSON(w, logger, http.StatusOK, toDetail(p))
	})
	mux.HandleFunc("POST /profiles", func(w http.ResponseWriter, r *http.Request) {
		p, ok := decodeProfile(w, r, logger)
		if !ok {
			return
		}
		if err := ps.CreateProfile(p); err != nil {
			writeProfileError(w, logger, "create profile", err)
			return
		}
		writeJSON(w, logger, http.StatusCreated, toDetail(p))
	})
	mux.HandleFunc("PUT /profiles/{name}", func(w http.ResponseWriter, r *http.Request) {
		p, ok := decodeProfile(w, r, logger)
		if !ok {
			return
		}
		name := r.PathValue("name")
		if err := ps.UpdateProfile(name, p); err != nil {
			writeProfileError(w, logger, "update profile", err)
			return
		}
		updated, _ := ps.GetProfile(name)
		writeJSON(w, logger, http.StatusOK, toDetail(updated))
	})
	mux.HandleFunc("DELETE /profiles/{name}", func(w http.ResponseWriter, r *http.Request) {
		if err := ps.DeleteProfile(r.PathValue("name")); err != nil {
			writeProfileError(w, logger, "delete profile", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func decodeProfile(
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
) (profile.Profile, bool) {
	var p profile.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, logger, http.StatusBadRequest, errorResponse{Error: msgInvalidBody})
		return profile.Profile{}, false
	}
	return p, true
}

// writeProfileError maps a profile domain error to its HTTP status and message.
func writeProfileError(w http.ResponseWriter, logger *slog.Logger, op string, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, profile.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, profile.ErrExists):
		status = http.StatusConflict
	case errors.Is(err, profile.ErrInvalid):
		status = http.StatusBadRequest
	}
	if status == http.StatusInternalServerError {
		logger.Error("dashboard: "+op+" failed", "err", err)
	}
	writeJSON(w, logger, status, errorResponse{Error: err.Error()})
}
