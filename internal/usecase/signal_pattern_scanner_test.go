package usecase

import (
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

func makeRSIDivCandles() ([]ports.Candle, []float64, []float64, []float64) {
	base := time.Now()
	candles := make([]ports.Candle, 20)
	rsi := make([]float64, 20)
	ema8 := make([]float64, 20)
	ema21 := make([]float64, 20)
	for i := 0; i < 20; i++ {
		candles[i] = ports.Candle{Symbol: "EURUSD", Time: base.Add(time.Duration(i) * time.Minute), Open: 1, High: 1, Low: 1, Close: 1}
		rsi[i] = 50 - float64(i)
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

func makeEMABounceCandles() ([]ports.Candle, []float64, []float64, []float64) {
	base := time.Now()
	candles := []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1, High: 1.1, Low: 0.8, Close: 0.85},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 0.86, High: 1.05, Low: 0.82, Close: 0.95},
	}
	rsi := []float64{0, 0}
	ema8 := []float64{0.9, 0.9}
	ema21 := []float64{0.85, 0.86}
	return candles, rsi, ema8, ema21
}

func makeFVRCandles() ([]ports.Candle, []float64, []float64, []float64) {
	base := time.Now()
	candles := []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1.1, High: 1.2, Low: 1.05, Close: 1.15},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 1.17, High: 1.2, Low: 1.0, Close: 1.18},
	}
	rsi := []float64{0, 0}
	ema8 := []float64{1.05, 1.15}
	ema21 := []float64{1.0, 1.1}
	return candles, rsi, ema8, ema21
}

func makeNoSignalCandles() ([]ports.Candle, []float64, []float64, []float64) {
	base := time.Now()
	candles := []ports.Candle{
		{Symbol: "EURUSD", Time: base, Open: 1, High: 1, Low: 1, Close: 1},
		{Symbol: "EURUSD", Time: base.Add(time.Minute), Open: 1, High: 1, Low: 1, Close: 1},
	}
	rsi := []float64{0, 0}
	ema8 := []float64{1, 1}
	ema21 := []float64{1, 1}
	return candles, rsi, ema8, ema21
}

func TestScanSignalPatterns(t *testing.T) {
	tests := []struct {
		name   string
		maker  func() ([]ports.Candle, []float64, []float64, []float64)
		expect int
		dir    string
	}{
		{name: "rsi divergence", maker: makeRSIDivCandles, expect: 1, dir: "UP"},
		{name: "ema bounce", maker: makeEMABounceCandles, expect: 1, dir: "UP"},
		{name: "fair value rejection", maker: makeFVRCandles, expect: 1, dir: "UP"},
		{name: "none", maker: makeNoSignalCandles, expect: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candles, rsi, ema8, ema21 := tt.maker()
			sigs := ScanSignalPatterns("EURUSD", candles, rsi, ema8, ema21)
			if len(sigs) != tt.expect {
				t.Fatalf("expected %d signals, got %d", tt.expect, len(sigs))
			}
			if tt.expect > 0 && sigs[0].Direction != tt.dir {
				t.Errorf("expected %s, got %s", tt.dir, sigs[0].Direction)
			}
		})
	}
}
