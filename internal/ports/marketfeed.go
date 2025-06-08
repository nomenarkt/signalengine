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

// MarketFeedAdapter represents a provider-specific implementation of a
// market feed.
type MarketFeedAdapter interface {
	MarketFeedPort
}

// BackoffStrategy computes the delay before retrying a failed connection.
type BackoffStrategy interface {
	// Next returns the duration to wait before the given retry attempt.
	Next(retry int) time.Duration
}
