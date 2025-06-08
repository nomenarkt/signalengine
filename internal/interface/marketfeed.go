package ports

import (
	"context"
	"time"
)

// Candle represents aggregated market data for a symbol at a point in time.
type Candle struct {
	Symbol    string
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// MarketFeedPort streams market candles for given symbols.
type MarketFeedPort interface {
	// Subscribe returns a read-only channel of candles. It begins streaming
	// immediately and stops when the provided context is canceled.
	Subscribe(ctx context.Context, symbols []string) (<-chan Candle, error)
}
