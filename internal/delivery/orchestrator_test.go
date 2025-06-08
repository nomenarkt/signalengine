package delivery

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
	"github.com/nomenarkt/signalengine/internal/usecase"
)

type mockFeed struct{ candles []ports.Candle }

func (m *mockFeed) StreamCandles(ctx context.Context, symbols []string) (<-chan ports.Candle, error) {
	ch := make(chan ports.Candle)
	go func() {
		defer close(ch)
		for _, c := range m.candles {
			select {
			case <-ctx.Done():
				return
			case ch <- c:
			}
		}
	}()
	return ch, nil
}

type mockPublisher struct{ msgs []string }

func (m *mockPublisher) PublishMessages(ctx context.Context, msgs []string) error {
	m.msgs = append(m.msgs, msgs...)
	return nil
}

func makeCandles(bullish bool) []ports.Candle {
	base := time.Now()
	var candles []ports.Candle
	for i := 0; i < 16; i++ {
		candles = append(candles, ports.Candle{Symbol: "EURUSD", Time: base.Add(time.Duration(i) * time.Minute), Open: 1, High: 1, Low: 1, Close: 1})
	}
	candles = append(candles, ports.Candle{Symbol: "EURUSD", Time: base.Add(16 * time.Minute), Open: 0.7, High: 0.8, Low: 0.5, Close: 0.6})
	candles = append(candles, ports.Candle{Symbol: "EURUSD", Time: base.Add(17 * time.Minute), Open: 0.65, High: 0.75, Low: 0.55, Close: 0.6})
	candles = append(candles, ports.Candle{Symbol: "EURUSD", Time: base.Add(18 * time.Minute), Open: 0.6, High: 0.65, Low: 0.45, Close: 0.5})
	lastClose := 0.45
	if bullish {
		lastClose = 0.8
	}
	candles = append(candles, ports.Candle{Symbol: "EURUSD", Time: base.Add(19 * time.Minute), Open: 0.48, High: 0.9, Low: 0.4, Close: lastClose})
	return candles
}

func TestOrchestrator_Run(t *testing.T) {
	tests := []struct {
		name    string
		bullish bool
		expect  int
	}{
		{name: "signal", bullish: true, expect: 1},
		{name: "no signal", bullish: false, expect: 0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			candles := makeCandles(tt.bullish)
			closes := make([]float64, len(candles))
			for i := range candles {
				closes[i] = candles[i].Close
			}
			rsi := usecase.CalcRSI(closes, 14)
			expected := usecase.ScoreRSIDivergence(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)), "EURUSD", candles, rsi)

			feed := &mockFeed{candles: candles}
			pub := &mockPublisher{}
			o := NewOrchestrator(feed, pub, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if err := o.Run(ctx, []string{"EURUSD"}); err != nil {
				t.Fatalf("run: %v", err)
			}
			if len(pub.msgs) != len(expected) {
				t.Fatalf("expected %d messages, got %d", len(expected), len(pub.msgs))
			}
		})
	}
}
