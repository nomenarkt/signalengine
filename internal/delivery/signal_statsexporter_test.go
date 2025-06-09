package delivery

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/usecase"
)

func sampleReport() usecase.BacktestReport {
	return usecase.BacktestReport{
		Results: []usecase.BacktestResult{
			{
				Symbol:     "EURUSD",
				Direction:  "UP",
				EntryTime:  time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
				ExpiryTime: time.Date(2024, 1, 2, 3, 9, 5, 0, time.UTC),
				Outcome:    "WIN",
				Reason:     "test",
			},
		},
		Total:  1,
		Wins:   1,
		Losses: 0,
	}
}

func TestExportBacktestReport(t *testing.T) {
	tmpDir := filepath.Join("testdata", "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	rep := sampleReport()

	tests := []struct {
		name   string
		file   string
		format string
	}{
		{name: "json", file: filepath.Join(tmpDir, "report.json"), format: "json"},
		{name: "csv", file: filepath.Join(tmpDir, "report.csv"), format: "csv"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := ExportBacktestReport(rep, tt.file, tt.format); err != nil {
				t.Fatalf("export: %v", err)
			}

			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}

			switch tt.format {
			case "csv":
				r := csv.NewReader(strings.NewReader(string(data)))
				recs, err := r.ReadAll()
				if err != nil {
					t.Fatalf("read csv: %v", err)
				}
				if len(recs) != len(rep.Results)+1 {
					t.Fatalf("expected %d records, got %d", len(rep.Results)+1, len(recs))
				}
			case "json":
				var out usecase.BacktestReport
				if err := json.Unmarshal(data, &out); err != nil {
					t.Fatalf("json parse: %v", err)
				}
				if out.Total != rep.Total {
					t.Fatalf("expected total %d, got %d", rep.Total, out.Total)
				}
			}
		})
	}
}

func TestExportBacktestReport_Empty(t *testing.T) {
	tmpDir := filepath.Join("testdata", "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	err := ExportBacktestReport(usecase.BacktestReport{}, filepath.Join(tmpDir, "r.json"), "json")
	if err == nil {
		t.Fatalf("expected error for empty report")
	}
}
