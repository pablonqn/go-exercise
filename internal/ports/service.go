package ports

import "go-exercise/internal/domain"

// LTPService defines the interface for LTP service operations
type LTPService interface {
	// GetLTPs retrieves LTPs for the requested pairs
	// If pairs is empty, returns all valid pairs
	GetLTPs(pairsStr string) ([]domain.LTP, error)
}

