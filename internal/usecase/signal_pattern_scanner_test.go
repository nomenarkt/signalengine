package usecase

import (
	"context"
	"testing"

	"github.com/nomenarkt/signalengine/internal/ports"
	"github.com/nomenarkt/signalengine/internal/testutils"
)

func TestScanSignalPatterns(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		data    func() ([]ports.Candle, []float64, []float64, []float64)
		want    int
		wantErr bool
	}{
		{
			name: "distinct",
			data: testutils.MakeScannerDistinctData,
			want: 3,
		},
		{
			name: "duplicates",
			data: testutils.MakeScannerDuplicateData,
			want: 1,
		},
		{
			name: "empty",
			data: func() ([]ports.Candle, []float64, []float64, []float64) {
				return nil, nil, nil, nil
			},
			wantErr: true,
		},
		{
			name: "mismatched lengths",
			data: func() ([]ports.Candle, []float64, []float64, []float64) {
				return make([]ports.Candle, 10), make([]float64, 10), make([]float64, 9), make([]float64, 10)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			candles, rsi, ema8, ema21 := tt.data()
			sigs, err := ScanSignalPatterns(ctx, nil, "EURUSD", candles, rsi, ema8, ema21)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if sigs != nil {
					t.Fatalf("expected nil signals, got %v", sigs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(sigs) != tt.want {
				t.Fatalf("expected %d signals, got %d", tt.want, len(sigs))
			}
		})
	}
}
