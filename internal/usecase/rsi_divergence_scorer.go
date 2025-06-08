package usecase

import (
	"math"
	"time"

	"github.com/nomenarkt/signalengine/internal/entity"
	"github.com/nomenarkt/signalengine/internal/ports"
)

// ScoreRSIDivergence looks for RSI divergence reversal setups over the provided candles and rsi values.
// It returns binary trade signals with a typical TTL of 2 minutes.
func ScoreRSIDivergence(symbol string, candles []ports.Candle, rsi []float64) []entity.Signal {
	n := len(candles)
	if n < 20 || n != len(rsi) {
		return nil
	}

	// Use last 20 bars for analysis
	start := n - 20
	c := candles[start:]
	r := rsi[start:]

	// Find previous swing high/low excluding last 3 bars
	lookback := len(c) - 3
	if lookback <= 0 {
		return nil
	}

	prevHighIdx, prevLowIdx := 0, 0
	for i := 1; i < lookback; i++ {
		if c[i].High > c[prevHighIdx].High {
			prevHighIdx = i
		}
		if c[i].Low < c[prevLowIdx].Low {
			prevLowIdx = i
		}
	}

	latest := len(c) - 1
	signals := []entity.Signal{}

	// Bearish divergence: price higher high but RSI lower high
	if c[latest].High > c[prevHighIdx].High && r[latest] < r[prevHighIdx] {
		if revDir(c[len(c)-3:]) == "DOWN" {
			signals = append(signals, entity.Signal{
				Symbol:     symbol,
				Direction:  "DOWN",
				Confidence: 0.8,
				TTL:        2 * time.Minute,
			})
		}
	}

	// Bullish divergence: price lower low but RSI higher low
	if c[latest].Low < c[prevLowIdx].Low && r[latest] > r[prevLowIdx] {
		if revDir(c[len(c)-3:]) == "UP" {
			signals = append(signals, entity.Signal{
				Symbol:     symbol,
				Direction:  "UP",
				Confidence: 0.8,
				TTL:        2 * time.Minute,
			})
		}
	}

	return signals
}

// revDir checks the last up to 3 candles for a reversal pattern and returns "UP", "DOWN", or "".
func revDir(c []ports.Candle) string {
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if isBullishEngulfing(c, i) || isBullishPinBar(c[i]) {
			return "UP"
		}
		if isBearishEngulfing(c, i) || isBearishPinBar(c[i]) {
			return "DOWN"
		}
	}
	return ""
}

func body(c ports.Candle) float64      { return math.Abs(c.Close - c.Open) }
func rangeSize(c ports.Candle) float64 { return c.High - c.Low }

func isBullishPinBar(c ports.Candle) bool {
	r := rangeSize(c)
	if r == 0 {
		return false
	}
	lowerWick := math.Min(c.Open, c.Close) - c.Low
	return body(c) <= r/3 && lowerWick >= r*2/3
}

func isBearishPinBar(c ports.Candle) bool {
	r := rangeSize(c)
	if r == 0 {
		return false
	}
	upperWick := c.High - math.Max(c.Open, c.Close)
	return body(c) <= r/3 && upperWick >= r*2/3
}

func isBullishEngulfing(c []ports.Candle, i int) bool {
	if i == 0 {
		return false
	}
	p := c[i-1]
	cur := c[i]
	if cur.Close <= cur.Open || p.Close >= p.Open {
		return false
	}
	return cur.Open < p.Close && cur.Close > p.Open
}

func isBearishEngulfing(c []ports.Candle, i int) bool {
	if i == 0 {
		return false
	}
	p := c[i-1]
	cur := c[i]
	if cur.Close >= cur.Open || p.Close <= p.Open {
		return false
	}
	return cur.Open > p.Close && cur.Close < p.Open
}
