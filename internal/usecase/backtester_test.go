package usecase

import (
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
	"github.com/nomenarkt/signalengine/internal/testutils"
)

// makeSeries returns 60 candles with a bullish divergence around index 54.
// If win is true, the trade closes higher; otherwise lower.
func makeSeries(sym string, win bool) []ports.Candle {
	base := time.Now()
	candles := make([]ports.Candle, 35)
	for i := 0; i < 35; i++ {
		ts := base.Add(time.Duration(i) * time.Minute)
		candles[i] = ports.Candle{Symbol: sym, Time: ts, Open: 1, High: 1, Low: 1, Close: 1}
	}
	pattern := testutils.MakeCandles(true)
	for i, c := range pattern {
		c.Symbol = sym
		c.Time = base.Add(time.Duration(35+i) * time.Minute)
		candles = append(candles, c)
	}
	start := len(candles)
	for i := 0; i < 5; i++ {
		ts := base.Add(time.Duration(start+i) * time.Minute)
		close := 1.1 + float64(i)*0.05
		if !win {
			close = 0.6 - float64(i)*0.05
		}
		candles = append(candles, ports.Candle{Symbol: sym, Time: ts, Open: close, High: close, Low: close, Close: close})
	}
	return candles
}

func TestBacktestSignals_Generate(t *testing.T) {
	data := map[string][]ports.Candle{
		"EURUSD": makeSeries("EURUSD", true),
	}
	rep := BacktestSignals(data, 3*time.Minute, 2*time.Minute)
	if rep.Total == 0 {
		t.Fatalf("expected results")
	}
	if rep.Results[0].Outcome == "NEUTRAL" {
		t.Errorf("expected resolved outcome, got %s", rep.Results[0].Outcome)
	}
}

func TestBacktestSignals_ReportCounts(t *testing.T) {
	data := map[string][]ports.Candle{
		"WIN":  makeSeries("WIN", true),
		"LOSS": makeSeries("LOSS", false),
	}
	rep := BacktestSignals(data, 3*time.Minute, 2*time.Minute)
	if rep.Total != len(rep.Results) {
		t.Fatalf("total mismatch")
	}
	if rep.Wins == 0 || rep.Losses == 0 {
		t.Fatalf("expected wins and losses")
	}
	acc := float64(rep.Wins) / float64(rep.Wins+rep.Losses)
	if rep.Accuracy != acc {
		t.Errorf("accuracy mismatch")
	}
}
