package ports

import (
	"context"
	"time"
)

// Candle represents a 1-minute OHLCV bar for a symbol.
type Candle struct {
	Symbol string
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// MarketFeedPort streams candles for the given symbols.
type MarketFeedPort interface {
	// StreamCandles returns a channel emitting 1-minute candles for each symbol.
	// The channel is closed when the context is canceled or an error occurs.
	StreamCandles(ctx context.Context, symbols []string) (<-chan Candle, error)
}
