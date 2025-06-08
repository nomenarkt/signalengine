package testutils

import (
	"context"
	"errors"

	"github.com/nomenarkt/signalengine/internal/ports"
)

// MockPublisher records published messages and can simulate transient failures.
type MockPublisher struct {
	FailFirst bool
	Calls     int
	Messages  [][]string
}

// PublishMessages appends messages and optionally returns an error on the first call.
func (m *MockPublisher) PublishMessages(ctx context.Context, msgs []string) error {
	m.Calls++
	m.Messages = append(m.Messages, msgs)
	if m.FailFirst && m.Calls == 1 {
		return errors.New("publish error")
	}
	return nil
}

var _ ports.TelegramPublisher = (*MockPublisher)(nil)
