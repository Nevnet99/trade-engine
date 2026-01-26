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

func TestGetBestBuyOrder(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := NewStorage(tx)
	ctx := context.Background()

	orders := []Order{
		{Symbol: "BTC-USD", Price: 50000, Quantity: 1, Side: "BUY"},
		{Symbol: "BTC-USD", Price: 52000, Quantity: 1, Side: "BUY"},
		{Symbol: "BTC-USD", Price: 51000, Quantity: 1, Side: "BUY"},
	}

	for _, o := range orders {
		_, err := storage.CreateOrder(ctx, o)
		if err != nil {
			t.Fatalf("Failed to seed order: %v", err)
		}
	}

	bestOrder, err := storage.GetBestBuyOrder(ctx, "BTC-USD")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestOrder == nil {
		t.Fatal("Expected an order, got nil")
	}

	if bestOrder.Price != 52000 {
		t.Errorf("Expected price 52000, got %f", bestOrder.Price)
	}
}

func TestGetBestSellOrder(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := NewStorage(tx)
	ctx := context.Background()

	orders := []Order{
		{Symbol: "BTC-USD", Price: 60000, Quantity: 1, Side: "SELL"}, // Expensive
		{Symbol: "BTC-USD", Price: 49000, Quantity: 1, Side: "SELL"}, // Winner!
		{Symbol: "BTC-USD", Price: 50000, Quantity: 1, Side: "SELL"}, // Mid
	}

	for _, o := range orders {
		_, err := storage.CreateOrder(ctx, o)
		if err != nil {
			t.Fatalf("Failed to seed order: %v", err)
		}
	}

	bestOrder, err := storage.GetBestSellOrder(ctx, "BTC-USD")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestOrder == nil {
		t.Fatal("Expected an order, got nil")
	}

	if bestOrder.Price != 49000 {
		t.Errorf("Expected price 49000, got %f", bestOrder.Price)
	}
}
