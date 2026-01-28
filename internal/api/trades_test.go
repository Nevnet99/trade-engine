package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestHandleGetRecentTrades(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	s := NewServer(store.NewStorage(tx))
	ctx := context.Background()

	_, err := tx.Exec(ctx, `
		INSERT INTO trading_pairs (symbol, base_asset, quote_asset, is_active) 
		VALUES ('BTC-USD', 'BTC', 'USD', true)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		t.Fatal(err)
	}

	var bidID, askID string
	err = tx.QueryRow(ctx, `INSERT INTO orders (symbol, side, price, quantity, status) VALUES ('BTC-USD', 'BUY', 50000, 1, 'FILLED') RETURNING id`).Scan(&bidID)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.QueryRow(ctx, `INSERT INTO orders (symbol, side, price, quantity, status) VALUES ('BTC-USD', 'SELL', 50000, 1, 'FILLED') RETURNING id`).Scan(&askID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO trades (bid_order_id, ask_order_id, price, quantity, timestamp)
		VALUES ($1, $2, 50000, 1, $3)
	`, bidID, askID, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Returns 200 and Trades for valid symbol", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/trades?symbol=BTC-USD", nil)
		rec := httptest.NewRecorder()

		s.HandleGetRecentTrades(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec.Code)
		}

		var response map[string][]store.Trade
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatal("Failed to decode JSON", err)
		}

		trades, ok := response["recentTrades"]
		if !ok {
			t.Fatal("JSON response missing 'recentTrades' key")
		}

		if len(trades) != 1 {
			t.Errorf("Expected 1 trade, got %d", len(trades))
		}
		if len(trades) > 0 && trades[0].Price != 50000 {
			t.Errorf("Expected price 50000, got %v", trades[0].Price)
		}
	})

	t.Run("Returns 400 if symbol is missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/trades", nil)
		rec := httptest.NewRecorder()

		s.HandleGetRecentTrades(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}
	})
}
