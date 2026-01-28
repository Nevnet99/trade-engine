package store

import (
	"context"
	"fmt"
	"time"
)

type Trade struct {
	ID        string    `json:"id"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}

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

	buyerQuery := `
	UPDATE orders
  SET quantity = quantity - $1
  WHERE id = $2
	`

	if _, err := tx.Exec(ctx, buyerQuery, qty, buyerOrderID); err != nil {
		return fmt.Errorf("failed to update buyer: %w", err)
	}

	sellerQuery := `
	UPDATE orders
  SET quantity = quantity - $1
  WHERE id = $2
	`

	if _, err := tx.Exec(ctx, sellerQuery, qty, sellerOrderID); err != nil {
		return fmt.Errorf("failed to update seller: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *Storage) GetRecentTrades(ctx context.Context, symbol string) ([]Trade, error) {
	trades := []Trade{}

	query := `
		SELECT t.id, t.bid_order_id, t.ask_order_id, t.price, t.quantity, t.timestamp 
		FROM trades t
		JOIN orders o ON t.bid_order_id = o.id
		WHERE o.symbol = $1 
		ORDER BY t.timestamp DESC
		LIMIT 50
	`

	rows, err := s.db.Query(ctx, query, symbol)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch Recent Trades: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		t := Trade{}

		if err := rows.Scan(
			&t.ID,
			&t.BuyerID,
			&t.SellerID,
			&t.Price,
			&t.Quantity,
			&t.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}

		trades = append(trades, t)
	}

	return trades, nil

}
