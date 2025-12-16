package cache

import (
	"sync"

	"go-exercise/internal/domain"
	"go-exercise/internal/ports"
)

// InMemoryCache implements the Repository port using in-memory storage
type InMemoryCache struct {
	mu    sync.RWMutex
	store map[string]*domain.CachedLTP
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() ports.Repository {
	return &InMemoryCache{
		store: make(map[string]*domain.CachedLTP),
	}
}

// GetLTP retrieves a cached LTP for a given pair
func (c *InMemoryCache) GetLTP(pair domain.Pair) (*domain.CachedLTP, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.store[pair.Value()]
	if !exists {
		return nil, false
	}

	// Check if expired
	if cached.IsExpired() {
		return nil, false
	}

	return cached, true
}

// SetLTP stores an LTP in the cache
func (c *InMemoryCache) SetLTP(pair domain.Pair, ltp domain.LTP) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[pair.Value()] = domain.NewCachedLTP(ltp)
}

// Clear removes all cached data
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]*domain.CachedLTP)
}

