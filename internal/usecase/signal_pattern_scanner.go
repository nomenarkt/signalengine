package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nomenarkt/signalengine/internal/entity"
	"github.com/nomenarkt/signalengine/internal/ports"
)

// ScanSignalPatterns aggregates various scoring algorithms over recent market data.
// It returns unique signals or an error if the input is invalid.
func ScanSignalPatterns(ctx context.Context, logger *slog.Logger, symbol string, candles []ports.Candle, rsi, ema8, ema21 []float64) ([]entity.Signal, error) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.InfoContext(ctx, "scan signal patterns", "symbol", symbol)

	n := len(candles)
	if n < 20 || n != len(rsi) || n != len(ema8) || n != len(ema21) {
		err := fmt.Errorf("invalid input lengths")
		logger.ErrorContext(ctx, "scan patterns", "error", err, "candles", n, "rsi_len", len(rsi), "ema8_len", len(ema8), "ema21_len", len(ema21))
		return nil, err
	}

	rsiSigs := ScoreRSIDivergence(ctx, logger, symbol, candles, rsi)
	emaSigs := ScoreEMAInteractions(ctx, logger, symbol, candles, ema8, ema21)
	candleSigs := ScoreCandlestickPatterns(ctx, logger, symbol, candles)

	merged := make([]entity.Signal, 0, len(rsiSigs)+len(emaSigs)+len(candleSigs))
	seen := map[string]struct{}{}
	add := func(s entity.Signal) {
		key := fmt.Sprintf("%s|%s|%d", s.Symbol, s.Direction, s.TTL)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		merged = append(merged, s)
	}
	for _, s := range rsiSigs {
		add(s)
	}
	for _, s := range emaSigs {
		add(s)
	}
	for _, s := range candleSigs {
		add(s)
	}

	return merged, nil
}
