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
			rsi := calcRSI(closes, rsiPeriod)
			signals := usecase.ScoreRSIDivergence(c.Symbol, candles, rsi)
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

func calcRSI(closes []float64, period int) []float64 {
	rsi := make([]float64, len(closes))
	if len(closes) <= period {
		return rsi
	}
	var gain, loss float64
	for i := 1; i <= period; i++ {
		diff := closes[i] - closes[i-1]
		if diff > 0 {
			gain += diff
		} else {
			loss -= diff
		}
	}
	avgGain := gain / float64(period)
	avgLoss := loss / float64(period)
	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - 100/(1+rs)
	}
	for i := period + 1; i < len(closes); i++ {
		diff := closes[i] - closes[i-1]
		if diff > 0 {
			avgGain = (avgGain*float64(period-1) + diff) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) - diff) / float64(period)
		}
		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - 100/(1+rs)
		}
	}
	return rsi
}
