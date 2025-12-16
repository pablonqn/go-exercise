package service

import (
	"fmt"
	"sort"

	"go-exercise/internal/domain"
	"go-exercise/internal/ports"
)

// LTPService handles the business logic for LTP operations
// It implements ports.LTPService interface
type LTPService struct {
	repository ports.Repository
	external   ports.External
}

// Ensure LTPService implements ports.LTPService interface
var _ ports.LTPService = (*LTPService)(nil)

// NewLTPService creates a new LTP service
func NewLTPService(repository ports.Repository, external ports.External) *LTPService {
	return &LTPService{
		repository: repository,
		external:   external,
	}
}

// GetLTPs retrieves LTPs for the requested pairs
// If pairs is empty, returns all valid pairs
func (s *LTPService) GetLTPs(pairsStr string) ([]domain.LTP, error) {
	pairs, err := domain.ParsePairs(pairsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid pairs: %w", err)
	}

	// Use map to track which pairs we need to fetch
	ltpMap := make(map[string]domain.LTP)
	var pairsToFetch []domain.Pair

	for _, pair := range pairs {
		cached, found := s.repository.GetLTP(pair)
		if found && cached != nil {
			ltpMap[pair.Value()] = cached.LTP
		} else {
			pairsToFetch = append(pairsToFetch, pair)
		}
	}

	if len(pairsToFetch) > 0 {
		ltps, err := s.external.GetTickers(pairsToFetch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch from external service: %w", err)
		}

		for _, ltp := range ltps {
			s.repository.SetLTP(ltp.Pair, ltp)
			ltpMap[ltp.Pair.Value()] = ltp
		}
	}

	// Build result slice maintaining original order
	result := make([]domain.LTP, 0, len(pairs))
	for _, pair := range pairs {
		if ltp, ok := ltpMap[pair.Value()]; ok {
			result = append(result, ltp)
		}
	}

	// Sort by pair name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Pair.Value() < result[j].Pair.Value()
	})

	return result, nil
}

