package delivery

import (
	"reflect"
	"testing"
	"time"

	"github.com/nomenarkt/signalengine/internal/entity"
)

func TestFormatSignals(t *testing.T) {
	tests := []struct {
		name  string
		input []entity.Signal
		want  []string
	}{
		{
			name:  "single signal",
			input: []entity.Signal{{Symbol: "eurusd", Direction: "up", Confidence: 0.83, TTL: 2 * time.Minute}},
			want:  []string{"⚡ Signal: EURUSD\n📈 Direction: UP\n🎯 Confidence: 85%\n⏱️ Expires in: 2m"},
		},
		{
			name: "multiple signals",
			input: []entity.Signal{
				{Symbol: "gbpusd", Direction: "down", Confidence: 0.7, TTL: 5 * time.Minute},
				{Symbol: "usdchf", Direction: "UP", Confidence: 0.92, TTL: 1 * time.Minute},
			},
			want: []string{
				"⚡ Signal: GBPUSD\n📈 Direction: DOWN\n🎯 Confidence: 70%\n⏱️ Expires in: 5m",
				"⚡ Signal: USDCHF\n📈 Direction: UP\n🎯 Confidence: 90%\n⏱️ Expires in: 1m",
			},
		},
		{
			name:  "empty",
			input: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSignals(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
