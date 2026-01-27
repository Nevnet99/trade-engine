package store

import (
	"context"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestCreateTrade(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := NewStorage(tx)
	ctx := context.Background()

	defer tx.Conn().Close(ctx)

	tests := []struct {
		name          string
		setupOrders   []Order
		tradeQty      int
		tradePrice    float64
		useInvalidIDs bool
		expectError   bool
		expectedQty   int
	}{
		{
			name: "Partial Fill (Standard)",
			setupOrders: []Order{
				{Symbol: "BTC", Price: 100, Quantity: 10, Side: "BUY"},
				{Symbol: "BTC", Price: 100, Quantity: 10, Side: "SELL"},
			},
			tradeQty:    4,
			expectError: false,
			expectedQty: 6, // 10 - 4
		},
		{
			name: "Full Fill (Liquidity Consumed)",
			setupOrders: []Order{
				{Symbol: "ETH", Price: 2000, Quantity: 5, Side: "BUY"},
				{Symbol: "ETH", Price: 2000, Quantity: 5, Side: "SELL"},
			},
			tradeQty:    5,
			expectError: false,
			expectedQty: 0, // 5 - 5
		},
		{
			name: "Invalid Order IDs (Foreign Key Check)",
			setupOrders: []Order{
				{Symbol: "SOL", Price: 50, Quantity: 10, Side: "BUY"},
				{Symbol: "SOL", Price: 50, Quantity: 10, Side: "SELL"},
			},
			tradeQty:      2,
			useInvalidIDs: true,
			expectError:   true,
			expectedQty:   10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			var buyID, sellID string
			for _, o := range tc.setupOrders {
				id, err := storage.CreateOrder(ctx, o)
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				if o.Side == "BUY" {
					buyID = id
				} else {
					sellID = id
				}
			}

			if tc.useInvalidIDs {
				buyID = "00000000-0000-0000-0000-000000000000"
			}

			err := storage.CreateTrade(ctx, 100.0, tc.tradeQty, buyID, sellID)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Did not expect error but got: %v", err)
			}

			var currentQty int
			var checkQuantityQuery string = "SELECT quantity FROM orders WHERE id = $1"

			err = storage.db.QueryRow(ctx, checkQuantityQuery, buyID).Scan(&currentQty)
			if err != nil {
				t.Fatalf("Failed to fetch buyer qty: %v", err)
			}
			if currentQty != tc.expectedQty {
				t.Errorf("Buyer Qty: want %d, got %d", tc.expectedQty, currentQty)
			}

			err = storage.db.QueryRow(ctx, checkQuantityQuery, sellID).Scan(&currentQty)
			if err != nil {
				t.Fatalf("Failed to fetch seller qty: %v", err)
			}
			if currentQty != tc.expectedQty {
				t.Errorf("Seller Qty: want %d, got %d", tc.expectedQty, currentQty)
			}
		})
	}
}
