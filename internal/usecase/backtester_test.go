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
	if rep.Total != 2 {
		t.Fatalf("want 2 results, got %d", rep.Total)
	}
	if rep.Wins != 1 || rep.Losses != 1 || rep.Neutrals != 0 {
		t.Fatalf("unexpected counts: %+v", rep)
	}
	if rep.Accuracy != 0.5 {
		t.Errorf("want accuracy 0.5, got %v", rep.Accuracy)
	}
}
