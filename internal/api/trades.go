package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func (s *Server) HandleGetRecentTrades(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	symbol := params.Get("symbol")

	if symbol == "" || !params.Has("symbol") {
		http.Error(w, "No query parameter set", http.StatusBadRequest)
		return
	}

	recentTrades, err := s.store.GetRecentTrades(r.Context(), symbol)

	if err != nil {
		slog.Error("Failed to get recent trades", "error", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]any{
		"recentTrades": recentTrades,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
