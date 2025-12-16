package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-exercise/internal/adapters/http/dto"
	"go-exercise/internal/domain"
	"go-exercise/internal/ports/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_GetLTP_Endpoint(t *testing.T) {
	t.Run("success - all pairs", func(t *testing.T) {
		// Arrange
		ltpService := new(mocks.LTPService)
		handler := NewHandler(ltpService)
		router := SetupRouter(handler)

		btcUSD, _ := domain.NewPair(domain.BTCUSD)
		btcCHF, _ := domain.NewPair(domain.BTCCHF)
		btcEUR, _ := domain.NewPair(domain.BTCEUR)

		expectedLTPs := []domain.LTP{
			{Pair: btcCHF, Amount: 49000.12},
			{Pair: btcEUR, Amount: 50000.12},
			{Pair: btcUSD, Amount: 52000.12},
		}

		ltpService.On("GetLTPs", "").Return(expectedLTPs, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp", nil)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.LTPResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.LTP, 3)

		ltpService.AssertExpectations(t)
	})

	t.Run("success - single pair", func(t *testing.T) {
		// Arrange
		ltpService := new(mocks.LTPService)
		handler := NewHandler(ltpService)
		router := SetupRouter(handler)

		btcUSD, _ := domain.NewPair(domain.BTCUSD)
		expectedLTPs := []domain.LTP{
			{Pair: btcUSD, Amount: 52000.12},
		}

		ltpService.On("GetLTPs", "BTC/USD").Return(expectedLTPs, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD", nil)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.LTPResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.LTP, 1)
		assert.Equal(t, "BTC/USD", response.LTP[0].Pair)
		assert.Equal(t, 52000.12, response.LTP[0].Amount)

		ltpService.AssertExpectations(t)
	})

	t.Run("success - multiple pairs", func(t *testing.T) {
		// Arrange
		ltpService := new(mocks.LTPService)
		handler := NewHandler(ltpService)
		router := SetupRouter(handler)

		btcUSD, _ := domain.NewPair(domain.BTCUSD)
		btcEUR, _ := domain.NewPair(domain.BTCEUR)

		expectedLTPs := []domain.LTP{
			{Pair: btcEUR, Amount: 50000.12},
			{Pair: btcUSD, Amount: 52000.12},
		}

		ltpService.On("GetLTPs", "BTC/USD,BTC/EUR").Return(expectedLTPs, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD,BTC/EUR", nil)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.LTPResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.LTP, 2)

		ltpService.AssertExpectations(t)
	})

	t.Run("error - invalid pair", func(t *testing.T) {
		// Arrange
		ltpService := new(mocks.LTPService)
		handler := NewHandler(ltpService)
		router := SetupRouter(handler)

		ltpService.On("GetLTPs", "BTC/INVALID").Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/INVALID", nil)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Error)

		ltpService.AssertExpectations(t)
	})

	t.Run("error - empty pairs query param", func(t *testing.T) {
		// Arrange
		ltpService := new(mocks.LTPService)
		handler := NewHandler(ltpService)
		router := SetupRouter(handler)

		btcUSD, _ := domain.NewPair(domain.BTCUSD)
		btcCHF, _ := domain.NewPair(domain.BTCCHF)
		btcEUR, _ := domain.NewPair(domain.BTCEUR)

		expectedLTPs := []domain.LTP{
			{Pair: btcCHF, Amount: 49000.12},
			{Pair: btcEUR, Amount: 50000.12},
			{Pair: btcUSD, Amount: 52000.12},
		}

		ltpService.On("GetLTPs", "").Return(expectedLTPs, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=", nil)
		rec := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.LTPResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.LTP, 3)

		ltpService.AssertExpectations(t)
	})
}

func TestRouter_Health_Endpoint(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)
	router := SetupRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])

	ltpService.AssertNotCalled(t, "GetLTPs")
}

func TestRouter_NotFound(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)
	router := SetupRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	ltpService.AssertNotCalled(t, "GetLTPs")
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)
	router := SetupRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ltp", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)

	ltpService.AssertNotCalled(t, "GetLTPs")
}

func TestRouter_CORS_Headers(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)
	router := SetupRouter(handler)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	expectedLTPs := []domain.LTP{
		{Pair: btcUSD, Amount: 52000.12},
	}

	ltpService.On("GetLTPs", "BTC/USD").Return(expectedLTPs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	// CORS middleware should add headers
	assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))

	ltpService.AssertExpectations(t)
}
