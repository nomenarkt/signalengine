package testutils

import (
	"context"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

// MockMarketFeed streams predefined candle sequences with optional delays between sequences.
type MockMarketFeed struct {
	Sequences [][]ports.Candle
	Delay     time.Duration
}

// StreamCandles returns a channel that emits the configured candle sequences.
// The channel closes when all sequences are sent or the context is canceled.
func (m *MockMarketFeed) StreamCandles(ctx context.Context, symbols []string) (<-chan ports.Candle, error) {
	ch := make(chan ports.Candle)
	go func() {
		defer close(ch)
		for i, seq := range m.Sequences {
			for _, c := range seq {
				select {
				case <-ctx.Done():
					return
				case ch <- c:
				}
			}
			if i < len(m.Sequences)-1 && m.Delay > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(m.Delay):
				}
			}
		}
	}()
	return ch, nil
}
