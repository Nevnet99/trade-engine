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

	user := User{Username: "test_trader", PasswordHash: "hashed_password"}

	u, err := storage.CreateUser(ctx, &user)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	newOrder := Order{
		Symbol:   "TEST-BTC",
		Price:    50000.00,
		Quantity: 1,
		Side:     "BUY",
		UserID:   u.ID,
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

	user := User{Username: "test_trader", PasswordHash: "hashed_password"}

	u, err := storage.CreateUser(ctx, &user)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	orders := []Order{
		{Symbol: "BTC-USD", Price: 50000, Quantity: 1, Side: "BUY", UserID: u.ID},
		{Symbol: "BTC-USD", Price: 52000, Quantity: 1, Side: "BUY", UserID: u.ID},
		{Symbol: "BTC-USD", Price: 51000, Quantity: 1, Side: "BUY", UserID: u.ID},
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

	user := User{Username: "test_trader", PasswordHash: "hashed_password"}

	u, err := storage.CreateUser(ctx, &user)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	orders := []Order{
		{Symbol: "BTC-USD", Price: 60000, Quantity: 1, Side: "SELL", UserID: u.ID}, // Expensive
		{Symbol: "BTC-USD", Price: 49000, Quantity: 1, Side: "SELL", UserID: u.ID}, // Winner!
		{Symbol: "BTC-USD", Price: 50000, Quantity: 1, Side: "SELL", UserID: u.ID}, // Mid
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

func TestGetOrderBook(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := NewStorage(tx)
	ctx := context.Background()

	_, err := tx.Exec(ctx, `
		INSERT INTO trading_pairs (symbol, base_asset, quote_asset, is_active) 
		VALUES ('BTC-USD', 'BTC', 'USD', true)
		ON CONFLICT DO NOTHING
	`)

	if err != nil {
		t.Fatal(err)
	}

	user := User{Username: "test_trader", PasswordHash: "hashed_password"}

	u, userErr := storage.CreateUser(ctx, &user)

	if userErr != nil {
		t.Fatalf("Failed to create test user: %v", userErr)
	}

	orders := []Order{
		{Symbol: "BTC-USD", Side: "BUY", Price: 50000, Quantity: 1, Status: "PENDING", UserID: u.ID},
		{Symbol: "BTC-USD", Side: "BUY", Price: 50000, Quantity: 2, Status: "PENDING", UserID: u.ID},
		{Symbol: "BTC-USD", Side: "BUY", Price: 49000, Quantity: 5, Status: "PENDING", UserID: u.ID},
		{Symbol: "BTC-USD", Side: "BUY", Price: 55000, Quantity: 10, Status: "FILLED", UserID: u.ID},
		{Symbol: "BTC-USD", Side: "SELL", Price: 51000, Quantity: 10, Status: "PENDING", UserID: u.ID},
	}

	for _, o := range orders {
		_, err := tx.Exec(ctx, `
			INSERT INTO orders (symbol, side, price, quantity, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			o.Symbol, o.Side, o.Price, o.Quantity, o.Status)
		if err != nil {
			t.Fatalf("Failed to seed order: %v", err)
		}
	}

	book, err := storage.GetOrderBook(ctx, "BTC-USD")
	if err != nil {
		t.Fatalf("Failed to get orderbook: %v", err)
	}

	if len(book.Bids) != 2 {
		t.Errorf("Expected 2 bid levels, got %d", len(book.Bids))
	}

	if book.Bids[0].Price != 50000 || book.Bids[0].Quantity != 3 {
		t.Errorf("Top bid incorrect. Expected 50k/3, got %v/%v", book.Bids[0].Price, book.Bids[0].Quantity)
	}

	if book.Bids[1].Price != 49000 {
		t.Errorf("Second bid incorrect. Expected 49k, got %v", book.Bids[1].Price)
	}

	if len(book.Asks) != 1 {
		t.Errorf("Expected 1 ask level, got %d", len(book.Asks))
	}
	if book.Asks[0].Price != 51000 {
		t.Errorf("Top ask incorrect. Expected 51k, got %v", book.Asks[0].Price)
	}
}
