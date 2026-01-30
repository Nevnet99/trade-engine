package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Nevnet99/trade-engine/internal/store"
)

func (s *Server) HandleGetKlines(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	symbol := params.Get("symbol")
	interval := params.Get("interval")

	if symbol == "" {
		http.Error(w, "No symbol query parameter set", http.StatusBadRequest)
		return
	}
	if interval == "" {
		http.Error(w, "No interval query parameter set", http.StatusBadRequest)
		return
	}

	protectedInterval, err := store.ParseInterval(interval)
	if err != nil {
		http.Error(w, "interval is not of type Interval", http.StatusBadRequest)
		return
	}

	limit := 100

	if params.Has("limit") {
		parsedLimit, err := strconv.Atoi(params.Get("limit"))
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	if limit > 1000 {
		limit = 1000
	}

	klines, err := s.store.GetKlines(r.Context(), symbol, protectedInterval, limit)

	if err != nil {
		slog.Error("failed to fetch klines", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(klines); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
