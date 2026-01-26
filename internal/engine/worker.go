package engine

import (
	"context"
	"log/slog"

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
	var symbol = "BTC-USD"

	buyOrder, buyOrderError := m.store.GetBestBuyOrder(ctx, symbol)

	if buyOrderError != nil {
		slog.Info("Buy Order Error")
		return
	}

	if buyOrder == nil {
		return
	}

	sellOrder, sellOrderError := m.store.GetBestSellOrder(ctx, symbol)

	if sellOrderError != nil {
		slog.Info("Sell Order Error")
		return
	}

	if sellOrder == nil {
		return
	}

	if buyOrder.Price >= sellOrder.Price {
		var buyQuantity = buyOrder.Quantity - buyOrder.FilledQuantity
		var sellQuantity = sellOrder.Quantity - sellOrder.FilledQuantity
		var tradeQuantity = min(buyQuantity, sellQuantity)
		var tradePrice float64

		if buyOrder.CreatedAt.Before(sellOrder.CreatedAt) {
			tradePrice = buyOrder.Price
		} else {
			tradePrice = sellOrder.Price
		}

		slog.Info("Match Found", "qty", tradeQuantity, "price", tradePrice)
	}

}
