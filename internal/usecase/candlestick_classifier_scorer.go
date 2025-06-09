package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/nomenarkt/signalengine/internal/entity"
	"github.com/nomenarkt/signalengine/internal/ports"
)

// ScoreCandlestickPatterns evaluates the latest candles for simple candlestick patterns.
// It returns signals when bullish or bearish patterns are detected.
func ScoreCandlestickPatterns(ctx context.Context, logger *slog.Logger, symbol string, candles []ports.Candle) []entity.Signal {
	if logger == nil {
		logger = slog.Default()
	}
	logger.InfoContext(ctx, "score candlestick patterns", "symbol", symbol)

	n := len(candles)
	if n < 2 {
		logger.WarnContext(ctx, "insufficient data", "candles", n)
		return nil
	}

	last := n - 1
	var signals []entity.Signal

	if isBullishEngulfing(candles, last) || isBullishPinBar(candles[last]) {
		signals = append(signals, entity.Signal{Symbol: symbol, Direction: "UP", Confidence: 0.5, TTL: time.Minute})
	}
	if isBearishEngulfing(candles, last) || isBearishPinBar(candles[last]) {
		signals = append(signals, entity.Signal{Symbol: symbol, Direction: "DOWN", Confidence: 0.5, TTL: time.Minute})
	}

	if len(signals) == 0 {
		logger.DebugContext(ctx, "no candlestick patterns found")
	}

	return signals
}
