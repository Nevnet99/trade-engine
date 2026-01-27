package store

import (
	"context"
	"fmt"
)

type TradingPair struct {
	Symbol     string `json:"symbol"`
	BaseAsset  string `json:"base_asset"`
	QuoteAsset string `json:"quote_asset"`
	IsActive   bool   `json:"is_active"`
}

func (s *Storage) GetActiveTradingPairs(ctx context.Context) ([]TradingPair, error) {
	query := `
        SELECT symbol, base_asset, quote_asset, is_active 
        FROM trading_pairs 
        WHERE is_active = true
    `

	rows, err := s.db.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch Trading Pairs: %w", err)
	}

	defer rows.Close()

	pairs := []TradingPair{}

	for rows.Next() {
		pair := TradingPair{}

		rowError := rows.Scan(&pair.Symbol, &pair.BaseAsset, &pair.QuoteAsset, &pair.IsActive)

		if rowError != nil {
			return nil, fmt.Errorf("failed to fetch Trading Pair Row: %w", rowError)
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil

}
