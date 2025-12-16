package ports

import "go-exercise/internal/domain"

// Repository defines the interface for LTP storage/cache
type Repository interface {
	// GetLTP retrieves a cached LTP for a given pair
	GetLTP(pair domain.Pair) (*domain.CachedLTP, bool)
	// SetLTP stores an LTP in the cache
	SetLTP(pair domain.Pair, ltp domain.LTP)
	// Clear removes all cached data
	Clear()
}

