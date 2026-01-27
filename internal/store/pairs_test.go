package store

import (
	"context"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestGetActiveTradingPairs(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := NewStorage(tx)
	ctx := context.Background()

	// DELETES THE SEEDED DATA FROM THE MIGRATION
	if _, err := tx.Exec(ctx, "DELETE FROM trading_pairs"); err != nil {
		t.Fatalf("Failed to clean trading_pairs table: %v", err)
	}

	defer tx.Conn().Close(ctx)

	pairs := []TradingPair{
		{
			Symbol:     "BTC-LUNA",
			IsActive:   true,
			BaseAsset:  "BTC",
			QuoteAsset: "LUNA",
		},
		{
			Symbol:     "LUNA-USD",
			IsActive:   false,
			BaseAsset:  "LUNA",
			QuoteAsset: "USD",
		},
	}

	query := `	
	INSERT INTO trading_pairs (symbol, base_asset, quote_asset, is_active)
	VALUES ($1, $2, $3, $4)
	`

	for _, pair := range pairs {
		_, err := tx.Exec(ctx, query, pair.Symbol, pair.BaseAsset, pair.QuoteAsset, pair.IsActive)

		if err != nil {
			t.Fatalf("Failed to seed DB: %v", err)
		}

	}

	pairs, err := storage.GetActiveTradingPairs(ctx)

	if err != nil {
		t.Fatalf("Failed to get active trading pairs: %v", err)
	}

	if len(pairs) != 1 {
		t.Fatalf("Returned too many pairs got: %v", len(pairs))
	}

	for _, pair := range pairs {
		if pair.Symbol != "BTC-LUNA" {
			t.Fatalf("Failed to get correct trading pairs: %v", pair.Symbol)
		}
	}

}
