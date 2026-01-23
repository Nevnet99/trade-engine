package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Nevnet99/trade-engine/internal/store"
)

type TradeParams struct {
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Side     string  `json:"side"`
}

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	params := TradeParams{}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	order := store.Order{
		Symbol:   params.Symbol,
		Price:    params.Price,
		Quantity: params.Quantity,
		Side:     params.Side,
	}

	id, err := s.store.CreateOrder(r.Context(), order)

	if err != nil {

		if errors.Is(err, store.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		slog.Error("Failed to create order", "error", err, "symbol", params.Symbol)
		http.Error(w, "Internal System Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"trade_id": id})
}
