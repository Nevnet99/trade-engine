package engine

import (
	"context"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestProcessMatches_LogsMatch(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := store.NewStorage(tx)
	ctx := context.Background()

	orders := []store.Order{
		{Symbol: "BTC-USD", Price: 60000, Quantity: 1, Side: "BUY"},
		{Symbol: "BTC-USD", Price: 50000, Quantity: 1, Side: "SELL"},
	}

	for _, o := range orders {
		_, err := storage.CreateOrder(ctx, o)
		if err != nil {
			t.Fatalf("Failed to seed order: %v", err)
		}
	}

	eng := New(storage)

	eng.ProcessMatches(ctx)

}
