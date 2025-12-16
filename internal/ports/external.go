package ports

import "go-exercise/internal/domain"

// External defines the interface for external service clients (e.g., Kraken API)
type External interface {
	// GetTicker retrieves the ticker information for a given pair
	GetTicker(pair domain.Pair) (domain.LTP, error)
	// GetTickers retrieves ticker information for multiple pairs
	GetTickers(pairs []domain.Pair) ([]domain.LTP, error)
}

