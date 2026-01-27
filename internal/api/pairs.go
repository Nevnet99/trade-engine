package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) HandleGetPairs(w http.ResponseWriter, r *http.Request) {
	pairs, err := s.store.GetActiveTradingPairs(r.Context())

	if err != nil {
		http.Error(w, "failed to fetch pairs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pairs); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
