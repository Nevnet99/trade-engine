package engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/Nevnet99/trade-engine/internal/store"
)

type MatchingEngine struct {
	store *store.Storage
}

func New(s *store.Storage) *MatchingEngine {
	return &MatchingEngine{
		store: s,
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (m *MatchingEngine) ProcessMatches(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	slog.Info("Matching Engine Worker Started")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Matching Engine shutting down...")
			return
		case <-ticker.C:
			tradingPairs, err := m.store.GetActiveTradingPairs(ctx)

			if err != nil {
				slog.Error("failed to get active trading pairs", "error", err)
				continue
			}

			for _, pair := range tradingPairs {
				m.runMatchingCycle(ctx, pair.Symbol)
			}

		}
	}
}

func (m *MatchingEngine) runMatchingCycle(ctx context.Context, symbol string) {

	for {
		buyOrder, err := m.store.GetBestBuyOrder(ctx, symbol)
		if err != nil {
			slog.Error("Failed to fetch best buy order", "error", err)
			return
		}

		if buyOrder == nil {
			return
		}

		sellOrder, err := m.store.GetBestSellOrder(ctx, symbol)
		if err != nil {
			slog.Error("Failed to fetch best sell order", "error", err)
			return
		}

		if sellOrder == nil {
			return
		}

		if buyOrder.Price < sellOrder.Price {
			return
		}

		tradeQuantity := min(buyOrder.Quantity, sellOrder.Quantity)
		if tradeQuantity <= 0 {
			slog.Info("Order filled or empty, skipping match")
			return
		}

		tradePrice := buyOrder.Price
		if sellOrder.CreatedAt.Before(buyOrder.CreatedAt) {
			tradePrice = sellOrder.Price
		}

		slog.Info("Match Found", "qty", tradeQuantity, "price", tradePrice)

		err = m.store.CreateTrade(ctx, tradePrice, tradeQuantity, buyOrder.ID, sellOrder.ID)
		if err != nil {
			slog.Error("Failed to execute trade", "error", err)
			return
		}
	}
}
