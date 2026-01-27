package store

import (
	"context"
	"fmt"
	"time"
)

type Order struct {
	ID        string
	Symbol    string
	Price     float64
	Quantity  int
	Side      string
	Status    string
	CreatedAt time.Time
	// FilledQuantity removed to match DB schema simplicity
}

type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

func (s *Storage) validateOrder(order Order) error {
	if order.Price <= 0 {
		return fmt.Errorf("price must be positive: %w", ErrValidation)
	}
	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive: %w", ErrValidation)
	}
	side := OrderSide(order.Side)
	if side != Buy && side != Sell {
		return fmt.Errorf("side must be BUY or SELL: %w", ErrValidation)
	}
	return nil
}

func (s *Storage) CreateOrder(ctx context.Context, order Order) (string, error) {
	var id string

	if err := s.validateOrder(order); err != nil {
		return "", err
	}

	query := `
    INSERT INTO orders (symbol, price, quantity, side, status) 
    VALUES ($1, $2, $3, $4, 'PENDING') 
    RETURNING id`

	err := s.db.QueryRow(ctx, query,
		order.Symbol, order.Price, order.Quantity, order.Side).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) GetBestBuyOrder(ctx context.Context, symbol string) (*Order, error) {
	var o Order

	// FIX: Select exactly 7 columns to match the 7 Scan variables below
	query := `
    SELECT id, symbol, quantity, price, side, status, created_at 
    FROM orders 
    WHERE symbol = $1 AND side = 'BUY' AND quantity > 0
    ORDER BY price DESC, created_at ASC 
    LIMIT 1
    `

	err := s.db.QueryRow(ctx, query, symbol).Scan(
		&o.ID,
		&o.Symbol,
		&o.Quantity,
		&o.Price,
		&o.Side,
		&o.Status,
		&o.CreatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &o, nil
}

func (s *Storage) GetBestSellOrder(ctx context.Context, symbol string) (*Order, error) {
	var o Order

	// FIX: Removed 'filled_quantity' (likely doesn't exist in DB).
	// FIX: Added 'quantity > 0' check to prevent infinite loops.
	// FIX: Selected exactly 7 columns to match Scan.
	query := `
    SELECT id, symbol, quantity, price, side, status, created_at
    FROM orders
    WHERE symbol = $1 AND side = 'SELL' AND quantity > 0
    ORDER BY price ASC, created_at ASC
    LIMIT 1`

	err := s.db.QueryRow(ctx, query, symbol).Scan(
		&o.ID,
		&o.Symbol,
		&o.Quantity,
		&o.Price,
		&o.Side,
		&o.Status,
		&o.CreatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}
