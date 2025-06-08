package usecase

import "testing"

func TestCalcRSI(t *testing.T) {
	closes := []float64{1, 2, 1, 2, 1, 2, 1}
	want := []float64{0, 0, 50, 75, 37.5, 68.75, 34.375}
	got := CalcRSI(closes, 2)
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if diff := got[i] - want[i]; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("index %d: want %.5f got %.5f", i, want[i], got[i])
		}
	}
}
