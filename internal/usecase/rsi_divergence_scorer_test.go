package usecase

import (
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

func TestScoreRSIDivergence(t *testing.T) {
	baseTime := time.Now()

	makeCandles := func(lastBullish bool) ([]ports.Candle, []float64) {
		candles := make([]ports.Candle, 20)
		rsi := make([]float64, 20)
		for i := 0; i < 20; i++ {
			candles[i] = ports.Candle{
				Symbol: "EURUSD",
				Time:   baseTime.Add(time.Duration(i) * time.Minute),
				Open:   1,
				High:   1,
				Low:    1,
				Close:  1,
			}
			rsi[i] = 50 - float64(i)
		}
		// previous swing low at index 16
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
		if lastBullish {
			candles[19].Close = 0.8 // bullish engulfing
		} else {
			candles[19].Close = 0.45 // bearish continuation
		}
		rsi[19] = 40
		return candles, rsi
	}

	t.Run("bullish divergence with reversal", func(t *testing.T) {
		candles, r := makeCandles(true)
		sigs := ScoreRSIDivergence("EURUSD", candles, r)
		if len(sigs) != 1 {
			t.Fatalf("expected 1 signal, got %d", len(sigs))
		}
		s := sigs[0]
		if s.Direction != "UP" {
			t.Errorf("expected UP, got %s", s.Direction)
		}
	})

	t.Run("no signal without reversal candle", func(t *testing.T) {
		candles, r := makeCandles(false)
		sigs := ScoreRSIDivergence("EURUSD", candles, r)
		if len(sigs) != 0 {
			t.Fatalf("expected no signals, got %d", len(sigs))
		}
	})
}
