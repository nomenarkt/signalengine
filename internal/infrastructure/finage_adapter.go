package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	ports "github.com/nomenarkt/signalengine/internal/interface"
)

// wsDialer abstracts websocket dialing for testability.
type wsDialer interface {
	DialContext(ctx context.Context, urlStr string, reqHeader http.Header) (*websocket.Conn, *http.Response, error)
}

// FinageAdapter implements ports.MarketFeedPort.
type FinageAdapter struct {
	apiKey         string
	wsURL          string
	restURL        string
	dialer         wsDialer
	client         *http.Client
	healthInterval time.Duration
	restInterval   time.Duration
	log            *log.Entry

	mu   sync.Mutex
	conn *websocket.Conn
}

// NewFinageAdapter creates a new Finage market feed adapter.
func NewFinageAdapter(apiKey, wsURL, restURL string, dialer wsDialer, client *http.Client, healthInterval, restInterval time.Duration, logger *log.Logger) *FinageAdapter {
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &FinageAdapter{
		apiKey:         apiKey,
		wsURL:          wsURL,
		restURL:        restURL,
		dialer:         dialer,
		client:         client,
		healthInterval: healthInterval,
		restInterval:   restInterval,
		log:            logger.WithField("component", "finage_adapter"),
	}
}

// Subscribe implements ports.MarketFeedPort.
func (f *FinageAdapter) Subscribe(ctx context.Context, symbols []string) (<-chan ports.Candle, error) {
	if len(symbols) == 0 {
		return nil, errors.New("no symbols provided")
	}
	out := make(chan ports.Candle)
	go f.run(ctx, out, symbols)
	return out, nil
}

func (f *FinageAdapter) run(ctx context.Context, out chan<- ports.Candle, symbols []string) {
	defer close(out)
	for {
		if ctx.Err() != nil {
			return
		}
		wsURL := fmt.Sprintf("%s?apikey=%s&symbols=%s", f.wsURL, f.apiKey, strings.Join(symbols, ","))
		conn, _, err := f.dialer.DialContext(ctx, wsURL, nil)
		if err != nil {
			f.log.WithError(err).Warn("websocket dial failed, using REST fallback")
			f.pollREST(ctx, out, symbols)
			continue
		}

		f.mu.Lock()
		f.conn = conn
		f.mu.Unlock()

		errChan := make(chan error, 1)
		go f.readWS(ctx, conn, out, errChan)

		ticker := time.NewTicker(f.healthInterval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				conn.Close()
				return
			case err := <-errChan:
				ticker.Stop()
				conn.Close()
				f.log.WithError(err).Warn("websocket error, reconnecting")
				f.pollREST(ctx, out, symbols)
				goto reconnect
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					ticker.Stop()
					conn.Close()
					f.log.WithError(err).Warn("ping failed, reconnecting")
					f.pollREST(ctx, out, symbols)
					goto reconnect
				}
			}
		}
	reconnect:
		time.Sleep(time.Second)
	}
}

func (f *FinageAdapter) readWS(ctx context.Context, conn *websocket.Conn, out chan<- ports.Candle, errChan chan<- error) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			errChan <- err
			return
		}
		var c ports.Candle
		if err := json.Unmarshal(data, &c); err != nil {
			f.log.WithError(err).Warn("failed to decode message")
			continue
		}
		select {
		case out <- c:
		case <-ctx.Done():
			return
		}
	}
}

func (f *FinageAdapter) pollREST(ctx context.Context, out chan<- ports.Candle, symbols []string) {
	ticker := time.NewTicker(f.restInterval)
	defer ticker.Stop()
	for {
		for _, s := range symbols {
			url := fmt.Sprintf("%s?apikey=%s&symbol=%s", f.restURL, f.apiKey, s)
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			resp, err := f.client.Do(req)
			if err != nil {
				f.log.WithError(err).Error("rest request failed")
				continue
			}
			var c ports.Candle
			if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
				f.log.WithError(err).Error("rest decode failed")
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			select {
			case out <- c:
			case <-ctx.Done():
				return
			}
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}
