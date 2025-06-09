package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/nomenarkt/signalengine/internal/entity"
	"github.com/nomenarkt/signalengine/internal/ports"
)

// ScoreEMAInteractions looks for EMA crossovers between ema8 and ema21.
// It returns binary trade signals when the fast EMA crosses the slow EMA.
func ScoreEMAInteractions(ctx context.Context, logger *slog.Logger, symbol string, candles []ports.Candle, ema8, ema21 []float64) []entity.Signal {
	if logger == nil {
		logger = slog.Default()
	}
	logger.InfoContext(ctx, "score EMA interactions", "symbol", symbol)

	n := len(candles)
	if n < 2 || n != len(ema8) || n != len(ema21) {
		logger.WarnContext(ctx, "insufficient data", "candles", n, "ema8_len", len(ema8), "ema21_len", len(ema21))
		return nil
	}

	prev := n - 2
	last := n - 1
	var signals []entity.Signal

	if ema8[prev] <= ema21[prev] && ema8[last] > ema21[last] && candles[last].Close > ema8[last] {
		signals = append(signals, entity.Signal{
			Symbol:     symbol,
			Direction:  "UP",
			Confidence: 0.6,
			TTL:        time.Minute,
		})
	}

	if ema8[prev] >= ema21[prev] && ema8[last] < ema21[last] && candles[last].Close < ema8[last] {
		signals = append(signals, entity.Signal{
			Symbol:     symbol,
			Direction:  "DOWN",
			Confidence: 0.6,
			TTL:        time.Minute,
		})
	}

	if len(signals) == 0 {
		logger.DebugContext(ctx, "no ema interaction signals")
	}
	return signals
}
