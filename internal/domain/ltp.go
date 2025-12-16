package domain

import (
	"time"
)

// LTP represents a Last Traded Price entity
type LTP struct {
	Pair   Pair
	Amount float64
}

// CachedLTP represents an LTP with timestamp for cache management
type CachedLTP struct {
	LTP       LTP
	Timestamp time.Time
}

// IsExpired checks if the cached LTP has expired (older than 1 minute)
func (c *CachedLTP) IsExpired() bool {
	return time.Since(c.Timestamp) > time.Minute
}

// NewCachedLTP creates a new CachedLTP with current timestamp
func NewCachedLTP(ltp LTP) *CachedLTP {
	return &CachedLTP{
		LTP:       ltp,
		Timestamp: time.Now(),
	}
}

