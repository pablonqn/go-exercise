package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Pair represents a currency pair value object
type Pair struct {
	value string
}

// Valid pairs
const (
	BTCUSD = "BTC/USD"
	BTCCHF = "BTC/CHF"
	BTCEUR = "BTC/EUR"
)

var validPairs = map[string]bool{
	BTCUSD: true,
	BTCCHF: true,
	BTCEUR: true,
}

// NewPair creates a new Pair value object
func NewPair(value string) (Pair, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if !validPairs[value] {
		return Pair{}, fmt.Errorf("invalid pair: %s. Valid pairs are: BTC/USD, BTC/CHF, BTC/EUR", value)
	}
	return Pair{value: value}, nil
}

// Value returns the string value of the pair
func (p Pair) Value() string {
	return p.value
}

// String implements the Stringer interface
func (p Pair) String() string {
	return p.value
}

// IsValid checks if a string is a valid pair
func IsValidPair(value string) bool {
	value = strings.ToUpper(strings.TrimSpace(value))
	return validPairs[value]
}

// ParsePairs parses a comma-separated string of pairs
func ParsePairs(pairsStr string) ([]Pair, error) {
	if pairsStr == "" {
		// Return all valid pairs if none specified
		return []Pair{
			Pair{value: BTCUSD},
			Pair{value: BTCCHF},
			Pair{value: BTCEUR},
		}, nil
	}

	pairs := strings.Split(pairsStr, ",")
	result := make([]Pair, 0, len(pairs))
	seen := make(map[string]bool)

	for _, p := range pairs {
		pair, err := NewPair(p)
		if err != nil {
			return nil, err
		}
		// Avoid duplicates
		if !seen[pair.Value()] {
			result = append(result, pair)
			seen[pair.Value()] = true
		}
	}

	if len(result) == 0 {
		return nil, errors.New("at least one valid pair must be specified")
	}

	return result, nil
}

