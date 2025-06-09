package testutils

import (
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

// MakeScannerDistinctData returns candles and indicators producing
// unique signals from each scoring algorithm.
func MakeScannerDistinctData() ([]ports.Candle, []float64, []float64, []float64) {
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
	candles[16] = ports.Candle{Symbol: "EURUSD", Time: base.Add(16 * time.Minute), Open: 0.7, High: 0.8, Low: 0.5, Close: 0.6}
	rsi[16] = 30
	candles[17] = ports.Candle{Symbol: "EURUSD", Time: base.Add(17 * time.Minute), Open: 0.65, High: 0.75, Low: 0.55, Close: 0.6}
	rsi[17] = 32
	candles[18] = ports.Candle{Symbol: "EURUSD", Time: base.Add(18 * time.Minute), Open: 0.6, High: 0.65, Low: 0.45, Close: 0.5}
	rsi[18] = 31
	candles[19] = ports.Candle{Symbol: "EURUSD", Time: base.Add(19 * time.Minute), Open: 0.48, High: 0.9, Low: 0.4, Close: 0.8}
	rsi[19] = 40

	// EMA cross down distinct from candlestick bullish pattern
	ema8[18] = 1.0
	ema21[18] = 0.9
	ema8[19] = 0.85
	ema21[19] = 0.9
	return candles, rsi, ema8, ema21
}

// MakeScannerDuplicateData returns candles and indicators that cause
// duplicate signals from different scoring algorithms.
func MakeScannerDuplicateData() ([]ports.Candle, []float64, []float64, []float64) {
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
	candles[19] = ports.Candle{Symbol: "EURUSD", Time: base.Add(19 * time.Minute), Open: 0.98, High: 1.15, Low: 0.95, Close: 1.1}
	rsi[18] = 50
	rsi[19] = 50
	ema8[18] = 1
	ema21[18] = 1
	ema8[19] = 1.05
	ema21[19] = 1
	return candles, rsi, ema8, ema21
}
