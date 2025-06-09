package delivery

import (
	"context"
	"log/slog"

	"github.com/nomenarkt/signalengine/internal/ports"
	"github.com/nomenarkt/signalengine/internal/usecase"
)

// Orchestrator streams market data, scores signals and publishes alerts.
type Orchestrator struct {
	feed      ports.MarketFeedPort
	publisher ports.TelegramPublisher
	logger    *slog.Logger
}

// NewOrchestrator initializes an Orchestrator.
func NewOrchestrator(feed ports.MarketFeedPort, pub ports.TelegramPublisher, logger *slog.Logger) *Orchestrator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Orchestrator{feed: feed, publisher: pub, logger: logger}
}

// Run starts streaming candles for the given symbols and processes signals.
func (o *Orchestrator) Run(ctx context.Context, symbols []string) error {
	ch, err := o.feed.StreamCandles(ctx, symbols)
	if err != nil {
		return err
	}

	const (
		keepBars  = 50
		rsiPeriod = 14
	)

	data := make(map[string][]ports.Candle)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c, ok := <-ch:
			if !ok {
				return nil
			}
			candles := append(data[c.Symbol], c)
			if len(candles) > keepBars {
				candles = candles[len(candles)-keepBars:]
			}
			data[c.Symbol] = candles
			if len(candles) < 20 {
				continue
			}
			closes := make([]float64, len(candles))
			for i := range candles {
				closes[i] = candles[i].Close
			}
			rsi := usecase.CalcRSI(closes, rsiPeriod)
			ema8 := usecase.CalcEMA(closes, 8)
			ema21 := usecase.CalcEMA(closes, 21)

			signals, err := usecase.ScanSignalPatterns(ctx, o.logger, c.Symbol, candles, rsi, ema8, ema21)
			if err != nil {
				o.logger.ErrorContext(ctx, "scan patterns", "error", err)
				continue
			}
			if len(signals) == 0 {
				continue
			}
			msgs := FormatSignals(signals)
			if err := o.publisher.PublishMessages(ctx, msgs); err != nil {
				o.logger.ErrorContext(ctx, "publish telegram", "error", err)
			}
		}
	}
}
