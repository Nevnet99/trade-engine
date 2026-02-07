package store

import (
	"context"
	"fmt"
	"time"
)

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Side      string    `json:"side"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
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
    INSERT INTO orders (user_id, symbol, price, quantity, side, status) 
    VALUES ($1, $2, $3, $4, $5, 'PENDING') 
    RETURNING id`

	err := s.db.QueryRow(ctx, query,
		order.UserID,
		order.Symbol,
		order.Price,
		order.Quantity,
		order.Side,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) GetBestBuyOrder(ctx context.Context, symbol string) (*Order, error) {
	var o Order

	query := `
    SELECT id, user_id, symbol, quantity, price, side, status, created_at 
    FROM orders 
    WHERE symbol = $1 AND side = 'BUY' AND quantity > 0 AND status = 'PENDING'
    ORDER BY price DESC, created_at ASC 
    LIMIT 1
    `

	err := s.db.QueryRow(ctx, query, symbol).Scan(
		&o.ID,
		&o.UserID,
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

	query := `
    SELECT id, user_id, symbol, quantity, price, side, status, created_at
    FROM orders
    WHERE symbol = $1 AND side = 'SELL' AND quantity > 0 AND status = 'PENDING'
    ORDER BY price ASC, created_at ASC
    LIMIT 1`

	err := s.db.QueryRow(ctx, query, symbol).Scan(
		&o.ID,
		&o.UserID,
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

type OrderBookEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

type OrderBook struct {
	Bids []OrderBookEntry `json:"bids"`
	Asks []OrderBookEntry `json:"asks"`
}

func (s *Storage) GetOrderBook(ctx context.Context, symbol string) (*OrderBook, error) {

	bidsSlice := []OrderBookEntry{}
	asksSlice := []OrderBookEntry{}
	var o OrderBook

	buyQuery := `
	SELECT price, SUM(quantity) 
		FROM orders 
		WHERE symbol = $1 AND side = 'BUY' AND status = 'PENDING' 
		GROUP BY price 
		ORDER BY price DESC 
		LIMIT 20
	`

	bidsRows, bidsErrors := s.db.Query(ctx, buyQuery, symbol)

	if bidsErrors != nil {
		return nil, bidsErrors
	}

	defer bidsRows.Close()

	for bidsRows.Next() {
		oe := OrderBookEntry{}

		rowError := bidsRows.Scan(&oe.Price, &oe.Quantity)

		if rowError != nil {
			return nil, fmt.Errorf("failed to fetch Trading Pair Row: %w", rowError)
		}

		bidsSlice = append(bidsSlice, oe)
	}

	sellQuery := `
	SELECT price, SUM(quantity) 
		FROM orders 
		WHERE symbol = $1 AND side = 'SELL' AND status = 'PENDING' 
		GROUP BY price 
		ORDER BY price ASC 
		LIMIT 20
	`

	asksRows, asksErrors := s.db.Query(ctx, sellQuery, symbol)

	if asksErrors != nil {
		return nil, asksErrors
	}

	defer asksRows.Close()

	for asksRows.Next() {
		oe := OrderBookEntry{}

		rowError := asksRows.Scan(&oe.Price, &oe.Quantity)

		if rowError != nil {
			return nil, fmt.Errorf("failed to fetch Trading Pair Row: %w", rowError)
		}

		asksSlice = append(asksSlice, oe)
	}

	o.Asks = asksSlice
	o.Bids = bidsSlice

	return &o, nil

}
