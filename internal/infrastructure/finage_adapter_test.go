package infrastructure

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewFinageAdapter_Dialer(t *testing.T) {
	t.Parallel()

	custom := &websocket.Dialer{HandshakeTimeout: 5 * time.Second}

	tests := []struct {
		name   string
		dialer *websocket.Dialer
		expect *websocket.Dialer
	}{
		{name: "default", dialer: nil, expect: websocket.DefaultDialer},
		{name: "custom", dialer: custom, expect: custom},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := NewFinageAdapter(slog.New(slog.NewTextHandler(io.Discard, nil)), tt.dialer)
			if a.dialer != tt.expect {
				t.Fatalf("expected %v, got %v", tt.expect, a.dialer)
			}
		})
	}
}
