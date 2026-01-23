package store

import (
	"context"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestCreateOrder(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := NewStorage(tx)
	ctx := context.Background()

	newOrder := Order{
		Symbol:   "TEST-BTC",
		Price:    50000.00,
		Quantity: 1,
		Side:     "BUY",
	}

	id, err := storage.CreateOrder(ctx, newOrder)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id == "" {
		t.Errorf("Expected a generated ID, got empty string")
	}

	if len(id) != 36 {
		t.Errorf("Expected UUID length 36, got %d (ID: %s)", len(id), id)
	}
}
