package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

func makeRSIDivData() ([]ports.Candle, []float64, []float64, []float64) {
	base := time.Now()
	candles := make([]ports.Candle, 20)
	rsi := make([]float64, 20)
	ema8 := make([]float64, 20)
	ema21 := make([]float64, 20)
	for i := 0; i < 20; i++ {
		candles[i] = ports.Candle{Symbol: "EURUSD", Time: base.Add(time.Duration(i) * time.Minute), Open: 1, High: 1, Low: 1, Close: 1}
		rsi[i] = 50 - float64(i)
		ema8[i] = 1
		ema21[i] = 1
	}
	candles[16].Open = 0.7
	candles[16].High = 0.8
	candles[16].Low = 0.5
	candles[16].Close = 0.6
	rsi[16] = 30

	candles[17].Open = 0.65
	candles[17].High = 0.75
	candles[17].Low = 0.55
	candles[17].Close = 0.6
	rsi[17] = 32

	candles[18].Open = 0.6
	candles[18].High = 0.65
	candles[18].Low = 0.45
	candles[18].Close = 0.5
	rsi[18] = 31

	candles[19].Open = 0.48
	candles[19].High = 0.9
	candles[19].Low = 0.4
	candles[19].Close = 0.8
	rsi[19] = 40
	return candles, rsi, ema8, ema21
}

func makeDupData() ([]ports.Candle, []float64, []float64, []float64) {
	base := time.Now()
	candles := make([]ports.Candle, 20)
	rsi := make([]float64, 20)
	ema8 := make([]float64, 20)
	ema21 := make([]float64, 20)
	for i := 0; i < 18; i++ {
		candles[i] = ports.Candle{Symbol: "EURUSD", Time: base.Add(time.Duration(i) * time.Minute), Open: 1, High: 1, Low: 1, Close: 1}
		rsi[i] = 50
		ema8[i] = 1
		ema21[i] = 1
	}
	candles[18] = ports.Candle{Symbol: "EURUSD", Time: base.Add(18 * time.Minute), Open: 1.05, High: 1.1, Low: 0.95, Close: 1.0}
	candles[19] = ports.Candle{Symbol: "EURUSD", Time: base.Add(19 * time.Minute), Open: 1.0, High: 1.15, Low: 0.95, Close: 1.1}
	rsi[18] = 50
	rsi[19] = 50
	ema8[18] = 1
	ema21[18] = 1
	ema8[19] = 1.05
	ema21[19] = 1
	return candles, rsi, ema8, ema21
}

func TestScanSignalPatterns(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid input", func(t *testing.T) {
		_, err := ScanSignalPatterns(ctx, nil, "EURUSD", make([]ports.Candle, 10), make([]float64, 10), make([]float64, 10), make([]float64, 9))
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("rsi divergence", func(t *testing.T) {
		candles, rsi, ema8, ema21 := makeRSIDivData()
		sigs, err := ScanSignalPatterns(ctx, nil, "EURUSD", candles, rsi, ema8, ema21)
		if err != nil || len(sigs) != 2 {
			t.Fatalf("unexpected result: %v %v", sigs, err)
		}
	})

	t.Run("deduplicate", func(t *testing.T) {
		candles, rsi, ema8, ema21 := makeDupData()
		sigs, err := ScanSignalPatterns(ctx, nil, "EURUSD", candles, rsi, ema8, ema21)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		if len(sigs) != 1 {
			t.Fatalf("expected 1 signal, got %d", len(sigs))
		}
	})
}
