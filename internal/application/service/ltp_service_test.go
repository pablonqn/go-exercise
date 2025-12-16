package service

import (
	"errors"
	"testing"

	"go-exercise/internal/domain"
	"go-exercise/internal/ports/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewLTPService(t *testing.T) {
	repo := new(mocks.Repository)
	external := new(mocks.External)

	service := NewLTPService(repo, external)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.repository)
	assert.Equal(t, external, service.external)
}

func TestLTPService_GetLTPs_EmptyPairs_ReturnsAllPairs(t *testing.T) {
	// Arrange
	repo := new(mocks.Repository)
	external := new(mocks.External)
	service := NewLTPService(repo, external)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcCHF, _ := domain.NewPair(domain.BTCCHF)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	expectedLTPs := []domain.LTP{
		{Pair: btcUSD, Amount: 52000.12},
		{Pair: btcCHF, Amount: 49000.12},
		{Pair: btcEUR, Amount: 50000.12},
	}

	// Mock repository - no cached data
	repo.On("GetLTP", btcUSD).Return((*domain.CachedLTP)(nil), false)
	repo.On("GetLTP", btcCHF).Return((*domain.CachedLTP)(nil), false)
	repo.On("GetLTP", btcEUR).Return((*domain.CachedLTP)(nil), false)

	// Mock external service
	external.On("GetTickers", mock.MatchedBy(func(pairs []domain.Pair) bool {
		return len(pairs) == 3
	})).Return(expectedLTPs, nil)

	// Mock repository SetLTP calls
	repo.On("SetLTP", btcUSD, expectedLTPs[0]).Return()
	repo.On("SetLTP", btcCHF, expectedLTPs[1]).Return()
	repo.On("SetLTP", btcEUR, expectedLTPs[2]).Return()

	// Act
	result, err := service.GetLTPs("")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, btcCHF.Value(), result[0].Pair.Value()) // Sorted by pair name
	assert.Equal(t, btcEUR.Value(), result[1].Pair.Value())
	assert.Equal(t, btcUSD.Value(), result[2].Pair.Value())

	repo.AssertExpectations(t)
	external.AssertExpectations(t)
}

func TestLTPService_GetLTPs_SinglePair_FromCache(t *testing.T) {
	// Arrange
	repo := new(mocks.Repository)
	external := new(mocks.External)
	service := NewLTPService(repo, external)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	cachedLTP := domain.NewCachedLTP(domain.LTP{Pair: btcUSD, Amount: 52000.12})

	// Mock repository - cached data found
	repo.On("GetLTP", btcUSD).Return(cachedLTP, true)

	// Act
	result, err := service.GetLTPs("BTC/USD")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, btcUSD.Value(), result[0].Pair.Value())
	assert.Equal(t, 52000.12, result[0].Amount)

	repo.AssertExpectations(t)
	external.AssertNotCalled(t, "GetTickers")
}

func TestLTPService_GetLTPs_SinglePair_FromExternal(t *testing.T) {
	// Arrange
	repo := new(mocks.Repository)
	external := new(mocks.External)
	service := NewLTPService(repo, external)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	expectedLTP := domain.LTP{Pair: btcUSD, Amount: 52000.12}

	// Mock repository - no cached data
	repo.On("GetLTP", btcUSD).Return((*domain.CachedLTP)(nil), false)

	// Mock external service
	external.On("GetTickers", []domain.Pair{btcUSD}).Return([]domain.LTP{expectedLTP}, nil)

	// Mock repository SetLTP call
	repo.On("SetLTP", btcUSD, expectedLTP).Return()

	// Act
	result, err := service.GetLTPs("BTC/USD")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, btcUSD.Value(), result[0].Pair.Value())
	assert.Equal(t, 52000.12, result[0].Amount)

	repo.AssertExpectations(t)
	external.AssertExpectations(t)
}

func TestLTPService_GetLTPs_MultiplePairs_MixedCache(t *testing.T) {
	// Arrange
	repo := new(mocks.Repository)
	external := new(mocks.External)
	service := NewLTPService(repo, external)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	cachedLTP := domain.NewCachedLTP(domain.LTP{Pair: btcUSD, Amount: 52000.12})
	expectedLTP := domain.LTP{Pair: btcEUR, Amount: 50000.12}

	// Mock repository - one cached, one not
	repo.On("GetLTP", btcUSD).Return(cachedLTP, true)
	repo.On("GetLTP", btcEUR).Return((*domain.CachedLTP)(nil), false)

	// Mock external service for missing pair
	external.On("GetTickers", []domain.Pair{btcEUR}).Return([]domain.LTP{expectedLTP}, nil)

	// Mock repository SetLTP call
	repo.On("SetLTP", btcEUR, expectedLTP).Return()

	// Act
	result, err := service.GetLTPs("BTC/USD,BTC/EUR")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Results should be sorted by pair name
	assert.Equal(t, btcEUR.Value(), result[0].Pair.Value())
	assert.Equal(t, btcUSD.Value(), result[1].Pair.Value())

	repo.AssertExpectations(t)
	external.AssertExpectations(t)
}

func TestLTPService_GetLTPs_InvalidPair_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.Repository)
	external := new(mocks.External)
	service := NewLTPService(repo, external)

	// Act
	result, err := service.GetLTPs("BTC/INVALID")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid pair")

	repo.AssertNotCalled(t, "GetLTP")
	external.AssertNotCalled(t, "GetTickers")
}

func TestLTPService_GetLTPs_ExternalServiceError_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.Repository)
	external := new(mocks.External)
	service := NewLTPService(repo, external)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	expectedError := errors.New("external service unavailable")

	// Mock repository - no cached data
	repo.On("GetLTP", btcUSD).Return((*domain.CachedLTP)(nil), false)

	// Mock external service error
	external.On("GetTickers", []domain.Pair{btcUSD}).Return(nil, expectedError)

	// Act
	result, err := service.GetLTPs("BTC/USD")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch from external service")

	repo.AssertExpectations(t)
	external.AssertExpectations(t)
}

func TestLTPService_GetLTPs_ResultsAreSorted(t *testing.T) {
	// Arrange
	repo := new(mocks.Repository)
	external := new(mocks.External)
	service := NewLTPService(repo, external)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcCHF, _ := domain.NewPair(domain.BTCCHF)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	// Mock repository - no cached data
	repo.On("GetLTP", btcUSD).Return((*domain.CachedLTP)(nil), false)
	repo.On("GetLTP", btcCHF).Return((*domain.CachedLTP)(nil), false)
	repo.On("GetLTP", btcEUR).Return((*domain.CachedLTP)(nil), false)

	// Mock external service - return in unsorted order
	expectedLTPs := []domain.LTP{
		{Pair: btcUSD, Amount: 52000.12},
		{Pair: btcEUR, Amount: 50000.12},
		{Pair: btcCHF, Amount: 49000.12},
	}

	external.On("GetTickers", mock.MatchedBy(func(pairs []domain.Pair) bool {
		return len(pairs) == 3
	})).Return(expectedLTPs, nil)

	// Mock repository SetLTP calls
	repo.On("SetLTP", mock.Anything, mock.Anything).Return()

	// Act
	result, err := service.GetLTPs("BTC/USD,BTC/CHF,BTC/EUR")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	// Verify sorted order: BTC/CHF < BTC/EUR < BTC/USD
	assert.Equal(t, btcCHF.Value(), result[0].Pair.Value())
	assert.Equal(t, btcEUR.Value(), result[1].Pair.Value())
	assert.Equal(t, btcUSD.Value(), result[2].Pair.Value())

	repo.AssertExpectations(t)
	external.AssertExpectations(t)
}
