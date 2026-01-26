package store

import (
	"context"
	"fmt"
	"time"
)

type Trade struct {
	ID        int       `json:"id"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}

// WIP

func (s *Storage) CreateTrade(ctx context.Context, price float64, qty int, buyerOrderID, sellerOrderID string) error {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	tradeQuery := `
	INSERT INTO trades (bid_order_id, ask_order_id, price, quantity)
	VALUES ($1, $2, $3, $4)
	`

	if _, err := tx.Exec(ctx, tradeQuery, buyerOrderID, sellerOrderID, price, qty); err != nil {
		return fmt.Errorf("failed to insert trade: %w", err)
	}

	buyerQuery := `...`

	if _, err := tx.Exec(ctx, buyerQuery, qty, buyerOrderID); err != nil {
		return fmt.Errorf("failed to update buyer: %w", err)
	}

	sellerQuery := `...`

	if _, err := tx.Exec(ctx, sellerQuery, qty, sellerOrderID); err != nil {
		return fmt.Errorf("failed to update seller: %w", err)
	}

	return tx.Commit(ctx)
}
