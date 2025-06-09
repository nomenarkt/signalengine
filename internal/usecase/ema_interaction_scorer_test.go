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

func makeBullishCross() ([]ports.Candle, []float64, []float64) {
	base := time.Now()
	candles := []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1, High: 1, Low: 1, Close: 1},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 1.05, High: 1.1, Low: 1.0, Close: 1.1},
	}
	ema8 := []float64{1, 1.05}
	ema21 := []float64{1, 1}
	return candles, ema8, ema21
}

func makeBearishCross() ([]ports.Candle, []float64, []float64) {
	base := time.Now()
	candles := []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1, High: 1, Low: 1, Close: 1},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 0.95, High: 1.0, Low: 0.9, Close: 0.9},
	}
	ema8 := []float64{1, 0.95}
	ema21 := []float64{1, 1}
	return candles, ema8, ema21
}

func makeNoCross() ([]ports.Candle, []float64, []float64) {
	base := time.Now()
	candles := []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1, High: 1, Low: 1, Close: 1},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 1, High: 1, Low: 1, Close: 1},
	}
	ema8 := []float64{1, 1}
	ema21 := []float64{1, 1}
	return candles, ema8, ema21
}

func TestScoreEMAInteractions(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("bullish cross", func(t *testing.T) {
		candles, ema8, ema21 := makeBullishCross()
		sigs := ScoreEMAInteractions(ctx, logger, "EURUSD", candles, ema8, ema21)
		if len(sigs) != 1 || sigs[0].Direction != "UP" {
			t.Fatalf("expected UP signal, got %+v", sigs)
		}
	})

	t.Run("bearish cross", func(t *testing.T) {
		candles, ema8, ema21 := makeBearishCross()
		sigs := ScoreEMAInteractions(ctx, logger, "EURUSD", candles, ema8, ema21)
		if len(sigs) != 1 || sigs[0].Direction != "DOWN" {
			t.Fatalf("expected DOWN signal, got %+v", sigs)
		}
	})

	t.Run("no cross logs", func(t *testing.T) {
		candles, ema8, ema21 := makeNoCross()
		buf := &bytes.Buffer{}
		h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
		log := slog.New(h)
		ScoreEMAInteractions(ctx, log, "EURUSD", candles, ema8, ema21)
		if !bytes.Contains(buf.Bytes(), []byte("no ema interaction signals")) {
			t.Errorf("expected debug log, got %s", buf.String())
		}
	})

	t.Run("insufficient data", func(t *testing.T) {
		candles, ema8, ema21 := makeBullishCross()
		buf := &bytes.Buffer{}
		log := slog.New(slog.NewTextHandler(buf, nil))
		ScoreEMAInteractions(ctx, log, "EURUSD", candles[:1], ema8, ema21)
		if !bytes.Contains(buf.Bytes(), []byte("insufficient data")) {
			t.Errorf("expected warn log, got %s", buf.String())
		}
	})
}
