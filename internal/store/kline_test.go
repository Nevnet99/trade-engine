package store

import (
	"context"
	"testing"
	"time"

	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestGetKlines(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	s := NewStorage(tx)
	ctx := context.Background()

	_, err := tx.Exec(ctx, `INSERT INTO trading_pairs (symbol, base_asset, quote_asset, is_active) VALUES ('BTC-USD', 'BTC', 'USD', true) ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatal(err)
	}

	var bidID, askID string

	err = tx.QueryRow(ctx, `INSERT INTO orders (symbol, side, price, quantity, status) VALUES ('BTC-USD', 'BUY', 1, 1, 'FILLED') RETURNING id`).Scan(&bidID)
	if err != nil {
		t.Fatal("Failed to create dummy Buy Order:", err)
	}

	err = tx.QueryRow(ctx, `INSERT INTO orders (symbol, side, price, quantity, status) VALUES ('BTC-USD', 'SELL', 1, 1, 'FILLED') RETURNING id`).Scan(&askID)
	if err != nil {
		t.Fatal("Failed to create dummy Sell Order:", err)
	}

	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	tradesToInsert := []struct {
		Price  float64
		Offset time.Duration
	}{
		{Price: 100, Offset: 5 * time.Second},
		{Price: 150, Offset: 30 * time.Second},
		{Price: 50, Offset: 55 * time.Second},
		{Price: 200, Offset: 65 * time.Second},
	}

	for _, tr := range tradesToInsert {
		_, err := tx.Exec(ctx, `
			INSERT INTO trades (bid_order_id, ask_order_id, price, quantity, timestamp)
			VALUES ($1, $2, $3, 1, $4)
		`, bidID, askID, tr.Price, baseTime.Add(tr.Offset))
		if err != nil {
			t.Fatal("Failed to seed trade:", err)
		}
	}

	var count int

	err = tx.QueryRow(ctx, "SELECT count(*) FROM trades").Scan(&count)

	if err != nil {
		t.Fatal("Failed to count trades:", err)
	}

	t.Logf("DEBUG: Raw Trades in DB: %d", count)

	var orderCount int

	err = tx.QueryRow(ctx, "SELECT count(*) FROM orders WHERE symbol='BTC-USD'").Scan(&orderCount)

	klines, err := s.GetKlines(ctx, "BTC-USD", Minute, 10)

	if err != nil {
		t.Fatalf("GetKlines failed: %v", err)
	}

	if len(klines) != 2 {
		t.Fatalf("Expected 2 candles, got %d", len(klines))
	}

	latest := klines[0]

	if latest.Close != 200 {
		t.Errorf("Latest candle (10:01) Close wrong. Got %v, want 200", latest.Close)
	}

	target := klines[1]

	if target.Open != 100 {
		t.Errorf("Open: got %v, want 100 (First trade)", target.Open)
	}
	if target.High != 150 {
		t.Errorf("High: got %v, want 150 (Max price)", target.High)
	}
	if target.Low != 50 {
		t.Errorf("Low: got %v, want 50 (Min price)", target.Low)
	}
	if target.Close != 50 {
		t.Errorf("Close: got %v, want 50 (Last trade)", target.Close)
	}
	if target.Volume != 3 {
		t.Errorf("Volume: got %v, want 3 (Sum of qty)", target.Volume)
	}
}
