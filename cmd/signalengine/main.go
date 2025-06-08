package main

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/nomenarkt/signalengine/internal/infrastructure"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	logger := log.New()
	apiKey := os.Getenv("FINAGE_API_KEY")
	wsURL := envOrDefault("FINAGE_WS_URL", "wss://api.finage.co.uk/agg")
	restURL := envOrDefault("FINAGE_REST_URL", "https://api.finage.co.uk/last")
	healthInterval, _ := time.ParseDuration(envOrDefault("FINAGE_HEALTH_INTERVAL", "30s"))
	restInterval, _ := time.ParseDuration(envOrDefault("FINAGE_REST_INTERVAL", "5s"))

	feed := infrastructure.NewFinageAdapter(apiKey, wsURL, restURL, nil, nil, healthInterval, restInterval, logger)
	ctx := context.Background()
	ch, err := feed.Subscribe(ctx, []string{"EURUSD"})
	if err != nil {
		logger.WithError(err).Fatal("subscribe failed")
	}
	for candle := range ch {
		logger.WithFields(log.Fields{"symbol": candle.Symbol, "close": candle.Close}).Info("candle")
	}
}
