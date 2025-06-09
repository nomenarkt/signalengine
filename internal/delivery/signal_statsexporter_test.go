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
	tmp := t.TempDir()
	rep := sampleReport()

	tests := []struct {
		name      string
		path      string
		format    string
		expectErr bool
	}{
		{name: "json by ext", path: filepath.Join(tmp, "r.json"), format: ""},
		{name: "csv by ext", path: filepath.Join(tmp, "r.csv"), format: ""},
		{name: "override format", path: filepath.Join(tmp, "r.json"), format: "csv"},
		{name: "unwritable path", path: filepath.Join(tmp, "missing", "r.csv"), format: "", expectErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ExportBacktestReport(rep, tt.path, tt.format)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			data, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			switch format := tt.format; {
			case format == "csv" || (format == "" && filepath.Ext(tt.path) == ".csv"):
				r := csv.NewReader(strings.NewReader(string(data)))
				recs, err := r.ReadAll()
				if err != nil {
					t.Fatalf("read csv: %v", err)
				}
				if len(recs) != len(rep.Results)+1 {
					t.Fatalf("expected %d records, got %d", len(rep.Results)+1, len(recs))
				}
			default:
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
	err := ExportBacktestReport(usecase.BacktestReport{}, filepath.Join(t.TempDir(), "r.json"), "")
	if err == nil {
		t.Fatalf("expected error for empty report")
	}
}
