package delivery

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/nomenarkt/signalengine/internal/entity"
)

// FormatSignals converts trade signals into formatted strings suitable for
// Telegram notifications. Each signal is represented as a multi-line message.
func FormatSignals(signals []entity.Signal) []string {
	if len(signals) == 0 {
		return nil
	}

	out := make([]string, 0, len(signals))
	for _, s := range signals {
		symbol := strings.ToUpper(s.Symbol)

		confidence := int(math.Round((s.Confidence*100)/5) * 5)
		if confidence > 100 {
			confidence = 100
		} else if confidence < 0 {
			confidence = 0
		}

		minutes := int(s.TTL.Round(time.Minute) / time.Minute)
		if minutes < 1 {
			minutes = 1
		}

		msg := fmt.Sprintf(
			"⚡ Signal: %s\n📈 Direction: %s\n🎯 Confidence: %d%%\n⏱️ Expires in: %dm",
			symbol,
			strings.ToUpper(s.Direction),
			confidence,
			minutes,
		)
		out = append(out, msg)
	}
	return out
}
