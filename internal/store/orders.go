package store

import (
	"context"
	"fmt"
)

type Order struct {
	ID       string
	Symbol   string
	Price    float64
	Quantity int
	Side     string
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
