package testutils

import (
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

// MakeCandles returns a slice of candles forming either a bullish or neutral pattern.
func MakeCandles(bullish bool) []ports.Candle {
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
