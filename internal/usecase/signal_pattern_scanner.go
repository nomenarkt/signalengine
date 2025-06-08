package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/nomenarkt/signalengine/internal/entity"
	"github.com/nomenarkt/signalengine/internal/ports"
)

// ScanSignalPatterns runs multiple pattern matchers over the latest market data.
// It aggregates signals from RSI divergence, EMA bounce and fair-value rejection patterns.
func ScanSignalPatterns(symbol string, candles []ports.Candle, rsi, ema8, ema21 []float64) []entity.Signal {
	signals := matchRSIDivergence(symbol, candles, rsi)
	signals = append(signals, matchEMABounce(symbol, candles, ema8, ema21)...)
	signals = append(signals, matchFairValueRejection(symbol, candles, ema8, ema21)...)
	return signals
}

func matchRSIDivergence(symbol string, candles []ports.Candle, rsi []float64) []entity.Signal {
	return ScoreRSIDivergence(context.Background(), slog.Default(), symbol, candles, rsi)
}

func matchEMABounce(symbol string, candles []ports.Candle, ema8, ema21 []float64) []entity.Signal {
	n := len(candles)
	if n < 2 || n != len(ema8) || n != len(ema21) {
		return nil
	}

	prev := candles[n-2]
	cur := candles[n-1]
	prevEMA8 := ema8[n-2]
	curEMA8 := ema8[n-1]
	curEMA21 := ema21[n-1]

	var signals []entity.Signal

	if curEMA8 > curEMA21 && prev.Close < prevEMA8 && cur.Low <= curEMA8 && cur.Close > curEMA8 {
		signals = append(signals, entity.Signal{
			Symbol:     symbol,
			Direction:  "UP",
			Confidence: 0.7,
			TTL:        time.Minute,
		})
	}

	if curEMA8 < curEMA21 && prev.Close > prevEMA8 && cur.High >= curEMA8 && cur.Close < curEMA8 {
		signals = append(signals, entity.Signal{
			Symbol:     symbol,
			Direction:  "DOWN",
			Confidence: 0.7,
			TTL:        time.Minute,
		})
	}

	return signals
}

func matchFairValueRejection(symbol string, candles []ports.Candle, ema8, ema21 []float64) []entity.Signal {
	n := len(candles)
	if n < 2 || n != len(ema8) || n != len(ema21) {
		return nil
	}

	prev := candles[n-2]
	cur := candles[n-1]
	prevEMA8 := ema8[n-2]
	curEMA8 := ema8[n-1]
	curEMA21 := ema21[n-1]

	var signals []entity.Signal

	if curEMA8 > curEMA21 && prev.Close > prevEMA8 && cur.Low <= curEMA21 && cur.Close > curEMA8 {
		signals = append(signals, entity.Signal{
			Symbol:     symbol,
			Direction:  "UP",
			Confidence: 0.75,
			TTL:        2 * time.Minute,
		})
	}

	if curEMA8 < curEMA21 && prev.Close < prevEMA8 && cur.High >= curEMA21 && cur.Close < curEMA8 {
		signals = append(signals, entity.Signal{
			Symbol:     symbol,
			Direction:  "DOWN",
			Confidence: 0.75,
			TTL:        2 * time.Minute,
		})
	}

	return signals
}
