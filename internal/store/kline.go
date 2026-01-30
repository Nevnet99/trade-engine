package store

import (
	"context"
	"fmt"
	"time"
)

type Candle struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

type Interval string

const (
	Minute         Interval = "1m"
	FiveMinutes    Interval = "5m"
	FifteenMinutes Interval = "15m"
	Hour           Interval = "1h"
	FourHours      Interval = "4h"
	Day            Interval = "1 day"
)

func ParseInterval(s string) (Interval, error) {
	// We use a switch to whitelist valid inputs
	switch Interval(s) {
	case Minute, FiveMinutes, FifteenMinutes, Hour, FourHours, Day:
		return Interval(s), nil
	default:
		return "", fmt.Errorf("invalid interval: %s", s)
	}
}

func (s *Storage) GetKlines(ctx context.Context, symbol string, interval Interval, limit int) ([]Candle, error) {
	klines := []Candle{}

	var seconds int
	switch interval {
	case Minute:
		seconds = 60
	case FiveMinutes:
		seconds = 300
	case FifteenMinutes:
		seconds = 900
	case Hour:
		seconds = 3600
	case FourHours:
		seconds = 14400
	case Day:
		seconds = 86400
	default:
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	query := fmt.Sprintf(`
		SELECT 
			to_timestamp(floor(extract(epoch from t.timestamp) / %d) * %d) AS bucket,
			(array_agg(t.price ORDER BY t.timestamp ASC))[1] AS open_price,
			MAX(t.price) AS high_price,
			MIN(t.price) AS low_price,
			(array_agg(t.price ORDER BY t.timestamp DESC))[1] AS close_price,
			SUM(t.quantity) AS volume
		FROM trades t
		JOIN orders o ON t.bid_order_id::text = o.id::text
		WHERE o.symbol = $1
		GROUP BY 1
		ORDER BY 1 DESC
		LIMIT $2
	`, seconds, seconds)

	rows, err := s.db.Query(ctx, query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute kline query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var k Candle
		err := rows.Scan(
			&k.Time,
			&k.Open,
			&k.High,
			&k.Low,
			&k.Close,
			&k.Volume,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan kline: %w", err)
		}
		klines = append(klines, k)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return klines, nil
}
