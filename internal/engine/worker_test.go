package engine

import (
	"context"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestProcessMatches_MultiFill(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := store.NewStorage(tx)
	ctx := context.Background()
	engine := New(storage)

	user := store.User{Username: "engine_tester", PasswordHash: "hash"}
	u, err := storage.CreateUser(ctx, &user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	_, err = tx.Exec(ctx, `
    INSERT INTO trading_pairs (symbol, base_asset, quote_asset, is_active) 
    VALUES ('BTC-USD', 'BTC', 'USD', true)
    ON CONFLICT (symbol) DO NOTHING
  `)
	if err != nil {
		t.Fatalf("Failed to seed trading pair: %v", err)
	}

	whaleOrder := store.Order{
		UserID: u.ID,
		Symbol: "BTC-USD", Side: "BUY", Price: 50000, Quantity: 10,
	}

	whaleID, err := storage.CreateOrder(ctx, whaleOrder)
	if err != nil {
		t.Fatalf("Failed to create whale order: %v", err)
	}

	sellerA := store.Order{
		UserID: u.ID,
		Symbol: "BTC-USD", Side: "SELL", Price: 49000, Quantity: 4,
	}

	sellerB := store.Order{
		UserID: u.ID,
		Symbol: "BTC-USD", Side: "SELL", Price: 49500, Quantity: 4,
	}

	if _, err := storage.CreateOrder(ctx, sellerA); err != nil {
		t.Fatalf("Failed to create sellerA: %v", err)
	}
	if _, err := storage.CreateOrder(ctx, sellerB); err != nil {
		t.Fatalf("Failed to create sellerB: %v", err)
	}

	engine.runMatchingCycle(ctx, "BTC-USD")

	var remainingQty int
	query := "SELECT quantity FROM orders WHERE id = $1"

	err = tx.QueryRow(ctx, query, whaleID).Scan(&remainingQty)
	if err != nil {
		t.Fatalf("Failed to fetch whale order: %v", err)
	}

	if remainingQty != 2 {
		t.Errorf("Expected Whale Quantity 2, got %d", remainingQty)
	}
}
