package engine

import (
	"context"
	"testing"

	// "time" <--- You can remove this import now

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestProcessMatches_MultiFill(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := store.NewStorage(tx)
	ctx := context.Background()
	engine := New(storage)

	_, err := tx.Exec(ctx, `
		INSERT INTO trading_pairs (symbol, base_asset, quote_asset, is_active) 
		VALUES ('BTC-USD', 'BTC', 'USD', true)
		ON CONFLICT (symbol) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("Failed to seed trading pair: %v", err)
	}

	whaleOrder := store.Order{
		Symbol: "BTC-USD", Side: "BUY", Price: 50000, Quantity: 10,
	}

	whaleID, _ := storage.CreateOrder(ctx, whaleOrder)

	sellerA := store.Order{
		Symbol: "BTC-USD", Side: "SELL", Price: 49000, Quantity: 4,
	}

	sellerB := store.Order{
		Symbol: "BTC-USD", Side: "SELL", Price: 49500, Quantity: 4,
	}
	_, _ = storage.CreateOrder(ctx, sellerA)
	_, _ = storage.CreateOrder(ctx, sellerB)

	engine.runMatchingCycle(ctx, "BTC-USD")

	var remainingQty float64

	query := "SELECT quantity FROM orders WHERE id = $1"

	err = tx.QueryRow(ctx, query, whaleID).Scan(&remainingQty)
	if err != nil {
		t.Fatalf("Failed to fetch whale order: %v", err)
	}

	if remainingQty != 2 {
		t.Errorf("Expected Whale Quantity 2, got %f", remainingQty)
	}
}
