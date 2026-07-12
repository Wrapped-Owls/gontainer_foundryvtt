package dashboard

import (
	"log/slog"
	"net/http"
	"strconv"
)

func registerLogHandlers(mux *http.ServeMux, lr LogReader, logger *slog.Logger) {
	mux.HandleFunc("GET /logs", func(w http.ResponseWriter, r *http.Request) {
		tail, _ := strconv.Atoi(r.URL.Query().Get("tail"))
		writeJSON(w, logger, http.StatusOK, logsResponse{Lines: lr.Logs(tail)})
	})
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		cursor, _ := strconv.Atoi(r.URL.Query().Get("since"))
		events, next := lr.Events(cursor)
		writeJSON(w, logger, http.StatusOK, eventsResponse{Events: events, Next: next})
	})
}
