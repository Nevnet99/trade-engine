package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestHandleGetPairs(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := store.NewStorage(tx)

	server := NewServer(storage)

	_, err := tx.Exec(context.Background(), "DELETE FROM trading_pairs")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec(context.Background(), `
		INSERT INTO trading_pairs (symbol, base_asset, quote_asset, is_active)
		VALUES ('SOL-USD', 'SOL', 'USD', true)
	`)

	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest("GET", "/pairs", nil)
	response := httptest.NewRecorder()

	server.HandleGetPairs(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var pairs []store.TradingPair

	if err := json.NewDecoder(response.Body).Decode(&pairs); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(pairs) != 1 {
		t.Fatalf("Expected 1 pair, got %d", len(pairs))
	}

	if pairs[0].Symbol != "SOL-USD" {
		t.Errorf("Expected symbol SOL-USD, got %s", pairs[0].Symbol)
	}
}
