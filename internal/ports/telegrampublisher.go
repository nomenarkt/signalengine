package ports

import "context"

// TelegramPublisher publishes messages to Telegram.
type TelegramPublisher interface {
	// PublishMessages sends the provided messages as Telegram alerts.
	PublishMessages(ctx context.Context, msgs []string) error
}
