package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
)

// BacktestResult represents the outcome of a single simulated trade.
type BacktestResult struct {
	Symbol     string
	Direction  string
	EntryTime  time.Time
	ExpiryTime time.Time
	Outcome    string
	Reason     string
}

// BacktestReport aggregates results from a backtest run.
type BacktestReport struct {
	Results  []BacktestResult
	Accuracy float64
	Total    int
	Wins     int
	Losses   int
	Neutrals int
}

// BacktestSignals replays historical candles and evaluates signal outcomes.
func BacktestSignals(data map[string][]ports.Candle, delayBeforeEntry, expiry time.Duration) BacktestReport {
	const (
		windowSize = 50
		rsiPeriod  = 14
	)

	var rep BacktestReport

	for symbol, candles := range data {
		if len(candles) < windowSize || !sorted(candles) {
			continue
		}

		for i := windowSize - 1; i < len(candles); i++ {
			window := candles[i-windowSize+1 : i+1]
			closes := make([]float64, len(window))
			for j, c := range window {
				closes[j] = c.Close
			}

			rsi := CalcRSI(closes, rsiPeriod)
			ema8 := CalcEMA(closes, 8)
			ema21 := CalcEMA(closes, 21)

			signals, err := ScanSignalPatterns(context.Background(), slog.Default(), symbol, window, rsi, ema8, ema21)
			if err != nil {
				continue
			}
			if len(signals) == 0 {
				continue
			}

			entryIdx := i + int(delayBeforeEntry/time.Minute)
			exitIdx := entryIdx + int(expiry/time.Minute)
			if exitIdx >= len(candles) {
				continue
			}

			entryTime := candles[entryIdx].Time
			expiryTime := candles[exitIdx].Time
			entryClose := candles[entryIdx].Close
			exitClose := candles[exitIdx].Close

			for _, s := range signals {
				res := BacktestResult{
					Symbol:     symbol,
					Direction:  s.Direction,
					EntryTime:  entryTime,
					ExpiryTime: expiryTime,
				}

				switch s.Direction {
				case "UP":
					switch {
					case exitClose > entryClose:
						res.Outcome = "WIN"
						res.Reason = "closed above entry"
					case exitClose < entryClose:
						res.Outcome = "LOSS"
						res.Reason = "closed below entry"
					default:
						res.Outcome = "NEUTRAL"
						res.Reason = "no change"
					}
				case "DOWN":
					switch {
					case exitClose < entryClose:
						res.Outcome = "WIN"
						res.Reason = "closed below entry"
					case exitClose > entryClose:
						res.Outcome = "LOSS"
						res.Reason = "closed above entry"
					default:
						res.Outcome = "NEUTRAL"
						res.Reason = "no change"
					}
				}

				rep.Results = append(rep.Results, res)
				rep.Total++
				switch res.Outcome {
				case "WIN":
					rep.Wins++
				case "LOSS":
					rep.Losses++
				case "NEUTRAL":
					rep.Neutrals++
				}
			}
		}
	}

	if rep.Wins+rep.Losses > 0 {
		rep.Accuracy = float64(rep.Wins) / float64(rep.Wins+rep.Losses)
	}

	return rep
}

func sorted(c []ports.Candle) bool {
	for i := 1; i < len(c); i++ {
		if c[i].Time.Before(c[i-1].Time) {
			return false
		}
	}
	return true
}
