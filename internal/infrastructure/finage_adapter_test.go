package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	ports "github.com/nomenarkt/signalengine/internal/interface"
)

type failingDialer struct{}

func (f failingDialer) DialContext(ctx context.Context, url string, header http.Header) (*websocket.Conn, *http.Response, error) {
	return nil, nil, errors.New("failed")
}

type simpleDialer struct{}

func (d simpleDialer) DialContext(ctx context.Context, url string, header http.Header) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.DialContext(ctx, url, header)
}

func TestFinageAdapter_Subscribe(t *testing.T) {
	candle := ports.Candle{Symbol: "EURUSD", Close: 1.2}
	cases := []struct {
		name      string
		dialer    wsDialer
		wsHandler http.HandlerFunc
	}{
		{
			name:      "fallback to rest",
			dialer:    failingDialer{},
			wsHandler: nil,
		},
		{
			name:   "websocket success",
			dialer: simpleDialer{},
			wsHandler: func(w http.ResponseWriter, r *http.Request) {
				var upgrader websocket.Upgrader
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}
				defer conn.Close()
				b, _ := json.Marshal(candle)
				conn.WriteMessage(websocket.TextMessage, b)
				time.Sleep(10 * time.Millisecond)
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var srv *httptest.Server
			if tt.wsHandler != nil {
				srv = httptest.NewServer(tt.wsHandler)
			} else {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(candle)
				}))
			}
			defer srv.Close()

			wsURL := "ws" + srv.URL[len("http"):]
			adapter := NewFinageAdapter("key", wsURL, srv.URL, tt.dialer, srv.Client(), 10*time.Millisecond, 10*time.Millisecond, log.New())
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			ch, err := adapter.Subscribe(ctx, []string{"EURUSD"})
			if err != nil {
				t.Fatalf("subscribe error: %v", err)
			}
			select {
			case got, ok := <-ch:
				if !ok {
					t.Fatalf("channel closed early")
				}
				if got.Symbol != candle.Symbol {
					t.Fatalf("got %v want %v", got.Symbol, candle.Symbol)
				}
			case <-ctx.Done():
				t.Fatalf("timeout waiting for candle")
			}
		})
	}
}
