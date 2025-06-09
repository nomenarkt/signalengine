package delivery

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nomenarkt/signalengine/internal/usecase"
)

// ExportBacktestReport writes the provided BacktestReport to path in the given
// format. If format is empty, it is inferred from the file extension. Supported
// formats are "json" and "csv".
func ExportBacktestReport(rep usecase.BacktestReport, path, format string) (err error) {
	if len(rep.Results) == 0 {
		return fmt.Errorf("export report: empty results")
	}

	if format == "" {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".json":
			format = "json"
		case ".csv":
			format = "csv"
		default:
			return fmt.Errorf("export report: unknown format for %s", path)
		}
	}
	format = strings.ToLower(format)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("export report: %w", err)
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	switch format {
	case "json":
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			return fmt.Errorf("export report: %w", err)
		}
	case "csv":
		w := csv.NewWriter(f)
		header := []string{"symbol", "direction", "entry_time", "expiry_time", "outcome", "reason"}
		if err := w.Write(header); err != nil {
			return fmt.Errorf("export report: %w", err)
		}
		for _, r := range rep.Results {
			row := []string{
				r.Symbol,
				r.Direction,
				r.EntryTime.Format(time.RFC3339),
				r.ExpiryTime.Format(time.RFC3339),
				r.Outcome,
				r.Reason,
			}
			if err := w.Write(row); err != nil {
				return fmt.Errorf("export report: %w", err)
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			return fmt.Errorf("export report: %w", err)
		}
	default:
		return fmt.Errorf("export report: unsupported format %s", format)
	}
	return nil
}
