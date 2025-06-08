package entity

import "time"

// Signal represents a binary trade setup.
type Signal struct {
	Symbol     string
	Direction  string // "UP" or "DOWN"
	Confidence float64
	TTL        time.Duration
}
