package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestCreateOrderAPI(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "Happy Path",
			body:           map[string]interface{}{"symbol": "BTC-USD", "price": 100, "quantity": 1, "side": "BUY"},
			expectedStatus: 202,
		},
		{
			name:           "Bad Input: Empty Body",
			body:           nil,
			expectedStatus: 400,
		},
		{
			name:           "Parsing Error",
			body:           map[string]interface{}{"price": "100"},
			expectedStatus: 400,
		},
		{
			name:           "Logic Error (Negative Price)",
			body:           map[string]interface{}{"symbol": "BTC", "price": -50, "quantity": 1, "side": "BUY"},
			expectedStatus: 400,
		},
		{
			name:           "Logic Error (Zero Quantity)",
			body:           map[string]interface{}{"symbol": "BTC", "price": 100, "quantity": 0, "side": "BUY"},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutils.SetupTestDB(t)
			storage := store.NewStorage(tx)

			server := NewServer(storage)

			b, _ := json.Marshal(tt.body)

			buffer := bytes.NewBuffer(b)

			request := httptest.NewRequest("POST", "/trade", buffer)
			response := httptest.NewRecorder()

			server.CreateOrder(response, request)

			if response.Code != tt.expectedStatus {
				t.Errorf("Test %s failed: expected status %d, got %d", tt.name, tt.expectedStatus, response.Code)
			}

		})
	}
}

func TestHandleGetOrderBook(t *testing.T) {

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

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (symbol, side, price, quantity, status) VALUES 
		('BTC-USD', 'BUY', 50000, 1, 'PENDING'),
		('BTC-USD', 'BUY', 50000, 2, 'PENDING'), 
		('BTC-USD', 'SELL', 51000, 5, 'PENDING')
	`)

	if err != nil {
		t.Fatal(err)
	}

	t.Run("Returns 200 and OrderBook for valid symbol", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/orderbook?symbol=BTC-USD", nil)
		rec := httptest.NewRecorder()

		s.HandleGetOrderBook(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec.Code)
		}

		var response map[string]store.OrderBook

		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatal("Failed to decode JSON response", err)
		}

		book, exists := response["orderBook"]

		if !exists {
			t.Fatal("Response missing 'orderBook' key")
		}

		if len(book.Bids) != 1 {
			t.Errorf("Expected 1 bid level, got %d", len(book.Bids))
		}

		if book.Bids[0].Quantity != 3 {
			t.Errorf("Expected aggregated bid quantity 3, got %v", book.Bids[0].Quantity)
		}

		if len(book.Asks) != 1 {
			t.Errorf("Expected 1 ask level, got %d", len(book.Asks))
		}
	})

	t.Run("Returns 400 when symbol is missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/orderbook", nil) // No query param
		rec := httptest.NewRecorder()

		s.HandleGetOrderBook(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}
	})
}
