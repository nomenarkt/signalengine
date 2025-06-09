package delivery

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/ports"
	"github.com/nomenarkt/signalengine/internal/testutils"
	"github.com/nomenarkt/signalengine/internal/usecase"
)

func makeSeries(symbol string) []ports.Candle {
	base := time.Now()
	candles := make([]ports.Candle, 35)
	for i := 0; i < 35; i++ {
		ts := base.Add(time.Duration(i) * time.Minute)
		candles[i] = ports.Candle{Symbol: symbol, Time: ts, Open: 1, High: 1, Low: 1, Close: 1}
	}
	pattern := testutils.MakeCandles(true)
	for i, c := range pattern {
		c.Symbol = symbol
		c.Time = base.Add(time.Duration(35+i) * time.Minute)
		candles = append(candles, c)
	}
	start := len(candles)
	for i := 0; i < 5; i++ {
		ts := base.Add(time.Duration(start+i) * time.Minute)
		close := 1.1 + float64(i)*0.05
		candles = append(candles, ports.Candle{Symbol: symbol, Time: ts, Open: close, High: close, Low: close, Close: close})
	}
	return candles
}

func TestBacktestSignals_ExportIntegration(t *testing.T) {
	tmpDir := filepath.Join("testdata", "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	data := map[string][]ports.Candle{
		"EURUSD": makeSeries("EURUSD"),
	}
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rep := usecase.BacktestSignals(ctx, logger, data, 3*time.Minute, 2*time.Minute)
	if len(rep.Results) == 0 {
		t.Fatalf("expected results from backtest")
	}

	jsonPath := filepath.Join(tmpDir, "out.json")
	csvPath := filepath.Join(tmpDir, "out.csv")

	if err := ExportBacktestReport(rep, jsonPath, "json"); err != nil {
		t.Fatalf("export json: %v", err)
	}
	if err := ExportBacktestReport(rep, csvPath, "csv"); err != nil {
		t.Fatalf("export csv: %v", err)
	}

	jdata, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var jrep usecase.BacktestReport
	if err := json.Unmarshal(jdata, &jrep); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if jrep.Total != rep.Total {
		t.Fatalf("expected %d total, got %d", rep.Total, jrep.Total)
	}

	cdata, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	r := csv.NewReader(strings.NewReader(string(cdata)))
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv parse: %v", err)
	}
	if len(recs) != len(rep.Results)+1 {
		t.Fatalf("expected %d rows, got %d", len(rep.Results)+1, len(recs))
	}
}
