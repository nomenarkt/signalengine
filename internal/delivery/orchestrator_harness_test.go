package delivery

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
	"github.com/nomenarkt/signalengine/internal/testutils"
	"github.com/nomenarkt/signalengine/internal/usecase"
)

func expectedSignals(ctx context.Context, candles []ports.Candle) int {
	const keepBars = 50
	const rsiPeriod = 14

	data := make([]ports.Candle, 0, keepBars)
	count := 0
	for _, c := range candles {
		data = append(data, c)
		if len(data) > keepBars {
			data = data[len(data)-keepBars:]
		}
		if len(data) < 20 {
			continue
		}
		closes := make([]float64, len(data))
		for i := range data {
			closes[i] = data[i].Close
		}
		rsi := usecase.CalcRSI(closes, rsiPeriod)
		signals := usecase.ScoreRSIDivergence(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)), c.Symbol, data, rsi)
		count += len(signals)
	}
	return count
}

func TestOrchestrator_ReconnectPublish(t *testing.T) {
	ctx := context.Background()

	seq := testutils.MakeCandles(true)
	feed := &testutils.MockMarketFeed{Sequences: [][]ports.Candle{seq, seq}, Delay: 10 * time.Millisecond}
	pub := &testutils.MockPublisher{FailFirst: true}
	o := NewOrchestrator(feed, pub, slog.New(slog.NewTextHandler(io.Discard, nil)))

	if err := o.Run(ctx, []string{"EURUSD"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	got := 0
	for _, m := range pub.Messages {
		got += len(m)
	}

	all := append(append([]ports.Candle{}, seq...), seq...)
	want := expectedSignals(ctx, all)

	if got != want {
		t.Fatalf("expected %d messages, got %d", want, got)
	}

	if pub.Calls != len(pub.Messages) {
		t.Fatalf("expected %d publish calls, got %d", len(pub.Messages), pub.Calls)
	}

	if pub.Calls == 0 {
		t.Fatalf("expected publisher to be called")
	}
}
