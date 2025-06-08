package infrastructure

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
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

type wsHandler func(*websocket.Conn)

func newWSServer(t *testing.T, handlers ...wsHandler) (*httptest.Server, *websocket.Dialer) {
	t.Helper()

	var (
		mu       sync.Mutex
		idx      int
		upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if idx >= len(handlers) {
			mu.Unlock()
			t.Fatalf("unexpected connection")
			return
		}
		h := handlers[idx]
		idx++
		mu.Unlock()

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade error: %v", err)
		}
		h(c)
	}))

	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, srv.Listener.Addr().String())
		},
	}

	return srv, dialer
}

func candleMsg(sym string, ts time.Time) []byte {
	b, _ := json.Marshal(finageCandle{
		Symbol:    sym,
		Timestamp: ts.UnixMilli(),
		Open:      1,
		High:      1,
		Low:       1,
		Close:     1,
		Volume:    1,
	})
	return b
}

func TestFinageAdapter_StreamCandles(t *testing.T) {
	t.Setenv("FINAGE_API_KEY", "test")

	now := time.Now()

	newAdapter := func(d *websocket.Dialer) *FinageAdapter {
		return NewFinageAdapter(slog.New(slog.NewTextHandler(io.Discard, nil)), d)
	}

	t.Run("normal streaming", func(t *testing.T) {
		handler := func(c *websocket.Conn) {
			defer c.Close()
			c.ReadMessage() // subscription
			msgs := [][]byte{
				candleMsg("EURUSD", now),
				candleMsg("EURUSD", now.Add(time.Second)),
			}
			for _, m := range msgs {
				if err := c.WriteMessage(websocket.TextMessage, m); err != nil {
					t.Fatalf("write message: %v", err)
				}
			}
			time.Sleep(10 * time.Millisecond)
		}

		srv, dialer := newWSServer(t, handler)
		defer srv.Close()

		a := newAdapter(dialer)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		ch, err := a.StreamCandles(ctx, []string{"EURUSD"})
		if err != nil {
			t.Fatalf("stream: %v", err)
		}

		for i := 0; i < 2; i++ {
			select {
			case <-time.After(time.Second):
				t.Fatalf("timeout waiting for candle %d", i)
			case _, ok := <-ch:
				if !ok {
					t.Fatalf("channel closed early")
				}
			}
		}
	})

	t.Run("reconnect after disconnection", func(t *testing.T) {
		first := func(c *websocket.Conn) {
			defer c.Close()
			c.ReadMessage()
			_ = c.WriteMessage(websocket.TextMessage, candleMsg("EURUSD", now))
		}

		second := func(c *websocket.Conn) {
			defer c.Close()
			c.ReadMessage()
			_ = c.WriteMessage(websocket.TextMessage, candleMsg("EURUSD", now.Add(time.Second)))
		}

		srv, dialer := newWSServer(t, first, second)
		defer srv.Close()

		a := newAdapter(dialer)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ch, err := a.StreamCandles(ctx, []string{"EURUSD"})
		if err != nil {
			t.Fatalf("stream: %v", err)
		}

		for i := 0; i < 2; i++ {
			select {
			case <-time.After(2 * time.Second):
				t.Fatalf("timeout waiting for candle %d", i)
			case _, ok := <-ch:
				if !ok {
					t.Fatalf("channel closed early")
				}
			}
		}
	})

	t.Run("early context cancelation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		a := newAdapter(websocket.DefaultDialer)
		ch, err := a.StreamCandles(ctx, []string{"EURUSD"})
		if err != nil {
			t.Fatalf("stream: %v", err)
		}

		if _, ok := <-ch; ok {
			t.Fatalf("expected closed channel")
		}
	})
}
