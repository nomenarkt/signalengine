package usecase

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/nomenarkt/signalengine/internal/entity"
	"github.com/nomenarkt/signalengine/internal/ports"
)

// ScoreRSIDivergence looks for RSI divergence reversal setups over the provided candles and rsi values.
// It returns binary trade signals with a typical TTL of 2 minutes.
func ScoreRSIDivergence(ctx context.Context, logger *slog.Logger, symbol string, candles []ports.Candle, rsi []float64) []entity.Signal {
	if logger == nil {
		logger = slog.Default()
	}
	logger.InfoContext(ctx, "score RSI divergence", "symbol", symbol)

	n := len(candles)
	if n < 20 || n != len(rsi) {
		logger.WarnContext(ctx, "insufficient data", "candles", n, "rsi_len", len(rsi))
		return nil
	}

	// Use last 20 bars for analysis
	start := n - 20
	c := candles[start:]
	r := rsi[start:]

	// Find previous swing high/low excluding last 3 bars
	lookback := len(c) - 3
	if lookback <= 0 {
		logger.WarnContext(ctx, "insufficient candles for swing lookup")
		return nil
	}

	prevHighIdx, prevLowIdx := 0, 0
	var highFound, lowFound bool
	for i := 1; i < lookback; i++ {
		if c[i].High > c[prevHighIdx].High {
			prevHighIdx = i
			highFound = true
		}
		if c[i].Low < c[prevLowIdx].Low {
			prevLowIdx = i
			lowFound = true
		}
	}
	if !highFound && !lowFound {
		logger.WarnContext(ctx, "no swing points found")
		return nil
	}

	latest := len(c) - 1
	signals := []entity.Signal{}

	// Bearish divergence: price higher high but RSI lower high
	if c[latest].High > c[prevHighIdx].High && r[latest] < r[prevHighIdx] {
		dir := revDir(ctx, logger, c[len(c)-3:])
		if dir == "DOWN" {
			signals = append(signals, entity.Signal{
				Symbol:     symbol,
				Direction:  "DOWN",
				Confidence: 0.8,
				TTL:        2 * time.Minute,
			})
		} else {
			logger.InfoContext(ctx, "divergence without reversal", "expected", "DOWN", "got", dir)
		}
	}

	// Bullish divergence: price lower low but RSI higher low
	if c[latest].Low < c[prevLowIdx].Low && r[latest] > r[prevLowIdx] {
		dir := revDir(ctx, logger, c[len(c)-3:])
		if dir == "UP" {
			signals = append(signals, entity.Signal{
				Symbol:     symbol,
				Direction:  "UP",
				Confidence: 0.8,
				TTL:        2 * time.Minute,
			})
		} else {
			logger.InfoContext(ctx, "divergence without reversal", "expected", "UP", "got", dir)
		}
	}

	return signals
}

// revDir checks the last up to 3 candles for a reversal pattern and returns "UP", "DOWN", or "".
func revDir(ctx context.Context, logger *slog.Logger, c []ports.Candle) string {
	if logger == nil {
		logger = slog.Default()
	}
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if isBullishEngulfing(c, i) || isBullishPinBar(c[i]) {
			return "UP"
		}
		if isBearishEngulfing(c, i) || isBearishPinBar(c[i]) {
			return "DOWN"
		}
	}
	logger.DebugContext(ctx, "no reversal pattern found")
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
