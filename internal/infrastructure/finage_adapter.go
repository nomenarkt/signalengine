package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/gorilla/websocket"

	"github.com/nomenarkt/signalengine/internal/ports"
)

// FinageAdapter implements the MarketFeedAdapter using the Finage WebSocket API.
type FinageAdapter struct {
	apiKey     string
	baseURL    string
	logger     *slog.Logger
	dialer     *websocket.Dialer
	now        func() time.Time
	staleAfter time.Duration
	backoff    ports.BackoffStrategy
}

// ExponentialBackoff implements a simple exponential backoff strategy.
type ExponentialBackoff struct {
	Base time.Duration
	Max  time.Duration
}

// Next returns the next backoff duration for the given retry attempt.
func (e ExponentialBackoff) Next(retry int) time.Duration {
	d := e.Base << retry
	if d > e.Max {
		return e.Max
	}
	return d
}

// NewFinageAdapter initializes a FinageAdapter with the FINAGE_API_KEY
// environment variable. The provided logger will be used for structured logging.
// Optionally a custom websocket.Dialer can be supplied; otherwise the
// websocket.DefaultDialer is used.
func NewFinageAdapter(logger *slog.Logger, dialer *websocket.Dialer, backoff ports.BackoffStrategy) *FinageAdapter {
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	if backoff == nil {
		backoff = ExponentialBackoff{Base: time.Second, Max: 30 * time.Second}
	}
	return &FinageAdapter{
		apiKey:     os.Getenv("FINAGE_API_KEY"),
		baseURL:    "wss://api.finage.co.uk/agg/forex",
		logger:     logger,
		dialer:     dialer,
		now:        time.Now,
		staleAfter: 30 * time.Second,
		backoff:    backoff,
	}
}

// finageCandle models the JSON payload from Finage.
type finageCandle struct {
	Symbol    string  `json:"s"`
	Timestamp int64   `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    float64 `json:"v"`
}

// StreamCandles connects to Finage and streams candles for the given symbols.
// The returned channel closes when the context is canceled or the connection
// fails after retries.
func (a *FinageAdapter) StreamCandles(ctx context.Context, symbols []string) (<-chan ports.Candle, error) {
	if a.apiKey == "" {
		return nil, errors.New("missing FINAGE_API_KEY")
	}
	if len(symbols) == 0 {
		return nil, errors.New("no symbols provided")
	}

	out := make(chan ports.Candle)
	go a.run(ctx, symbols, out)
	return out, nil
}

func (a *FinageAdapter) run(ctx context.Context, symbols []string, out chan ports.Candle) {
	defer close(out)

	lastTS := make(map[string]time.Time)
	var mu sync.Mutex

	retries := 0

	for {
		if ctx.Err() != nil {
			return
		}

		u := url.URL{Scheme: "wss", Host: "api.finage.co.uk", Path: "/agg/forex", RawQuery: "apikey=" + url.QueryEscape(a.apiKey)}
		a.logger.With("url", u.String()).InfoContext(ctx, "connecting to Finage")

		conn, _, err := a.dialer.DialContext(ctx, u.String(), nil)
		if err != nil {
			a.logger.ErrorContext(ctx, "connection failed", "error", err)
			wait := a.backoff.Next(retries)
			retries++
			if !sleep(ctx, wait) {
				return
			}
			continue
		}

		retries = 0
		if err := a.subscribe(conn, symbols); err != nil {
			a.logger.ErrorContext(ctx, "subscription failed", "error", err)
			conn.Close()
			wait := a.backoff.Next(retries)
			retries++
			if !sleep(ctx, wait) {
				return
			}
			continue
		}

		lastRecv := a.now()

		for {
			if ctx.Err() != nil {
				conn.Close()
				return
			}

			conn.SetReadDeadline(time.Now().Add(15 * time.Second))
			_, message, err := conn.ReadMessage()
			if err != nil {
				a.logger.ErrorContext(ctx, "read error", "error", err)
				conn.Close()
				break
			}

			var fc finageCandle
			if err := json.Unmarshal(message, &fc); err != nil {
				a.logger.ErrorContext(ctx, "decode error", "error", err)
				continue
			}

			ts := time.Unix(0, fc.Timestamp*int64(time.Millisecond))
			if ts.IsZero() || fc.Symbol == "" || (fc.Open == 0 && fc.Close == 0 && fc.High == 0 && fc.Low == 0) {
				continue
			}
			if time.Since(ts) > 5*time.Second {
				continue
			}

			mu.Lock()
			if prev, ok := lastTS[fc.Symbol]; ok && prev.Equal(ts) {
				mu.Unlock()
				continue
			}
			lastTS[fc.Symbol] = ts
			mu.Unlock()

			c := ports.Candle{
				Symbol: fc.Symbol,
				Time:   ts,
				Open:   fc.Open,
				High:   fc.High,
				Low:    fc.Low,
				Close:  fc.Close,
				Volume: fc.Volume,
			}

			select {
			case out <- c:
			case <-ctx.Done():
				conn.Close()
				return
			}

			if a.now().Sub(lastRecv) > a.staleAfter {
				a.logger.WarnContext(ctx, "stale stream detected, reconnecting")
				conn.Close()
				break
			}

			lastRecv = a.now()
		}

	}
}

func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (a *FinageAdapter) subscribe(conn *websocket.Conn, symbols []string) error {
	msg := map[string]any{
		"action":  "subscribe",
		"symbols": strings.Join(symbols, ","),
	}
	return conn.WriteJSON(msg)
}
