package api

import (
	"bytes"
	"encoding/json"
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
