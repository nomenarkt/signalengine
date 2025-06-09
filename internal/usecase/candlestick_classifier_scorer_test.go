package usecase

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

func makeBullishEngulf() []ports.Candle {
	base := time.Now()
	return []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1.0, High: 1.05, Low: 0.95, Close: 0.96},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 0.94, High: 1.1, Low: 0.9, Close: 1.05},
	}
}

func makeBearishPin() []ports.Candle {
	base := time.Now()
	return []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1, High: 1, Low: 1, Close: 1},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 1.0, High: 1.2, Low: 0.9, Close: 0.95},
	}
}

func makeNeutral() []ports.Candle {
	base := time.Now()
	return []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1, High: 1, Low: 1, Close: 1},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 1, High: 1, Low: 1, Close: 1},
	}
}

func TestScoreCandlestickPatterns(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("bullish engulfing", func(t *testing.T) {
		candles := makeBullishEngulf()
		sigs := ScoreCandlestickPatterns(ctx, logger, "EURUSD", candles)
		if len(sigs) != 1 || sigs[0].Direction != "UP" {
			t.Fatalf("expected UP signal, got %+v", sigs)
		}
	})

	t.Run("bearish pinbar", func(t *testing.T) {
		candles := makeBearishPin()
		sigs := ScoreCandlestickPatterns(ctx, logger, "EURUSD", candles)
		if len(sigs) != 1 || sigs[0].Direction != "DOWN" {
			t.Fatalf("expected DOWN signal, got %+v", sigs)
		}
	})

	t.Run("none logs", func(t *testing.T) {
		candles := makeNeutral()
		buf := &bytes.Buffer{}
		h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
		log := slog.New(h)
		ScoreCandlestickPatterns(ctx, log, "EURUSD", candles)
		if !bytes.Contains(buf.Bytes(), []byte("no candlestick patterns found")) {
			t.Errorf("expected debug log, got %s", buf.String())
		}
	})

	t.Run("insufficient", func(t *testing.T) {
		candles := makeBullishEngulf()[:1]
		buf := &bytes.Buffer{}
		log := slog.New(slog.NewTextHandler(buf, nil))
		ScoreCandlestickPatterns(ctx, log, "EURUSD", candles)
		if !bytes.Contains(buf.Bytes(), []byte("insufficient data")) {
			t.Errorf("expected warn log, got %s", buf.String())
		}
	})
}
