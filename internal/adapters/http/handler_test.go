package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-exercise/internal/adapters/http/dto"
	"go-exercise/internal/domain"
	"go-exercise/internal/ports/mocks"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	ltpService := new(mocks.LTPService)

	handler := NewHandler(ltpService)

	assert.NotNil(t, handler)
	assert.Equal(t, ltpService, handler.ltpService)
}

func TestHandler_GetLTP_Success_AllPairs(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcCHF, _ := domain.NewPair(domain.BTCCHF)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	expectedLTPs := []domain.LTP{
		{Pair: btcCHF, Amount: 49000.12},
		{Pair: btcEUR, Amount: 50000.12},
		{Pair: btcUSD, Amount: 52000.12},
	}

	ltpService.On("GetLTPs", "").Return(expectedLTPs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetLTP(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.LTPResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.LTP, 3)
	assert.Equal(t, "BTC/CHF", response.LTP[0].Pair)
	assert.Equal(t, 49000.12, response.LTP[0].Amount)
	assert.Equal(t, "BTC/EUR", response.LTP[1].Pair)
	assert.Equal(t, 50000.12, response.LTP[1].Amount)
	assert.Equal(t, "BTC/USD", response.LTP[2].Pair)
	assert.Equal(t, 52000.12, response.LTP[2].Amount)

	ltpService.AssertExpectations(t)
}

func TestHandler_GetLTP_Success_SinglePair(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	expectedLTPs := []domain.LTP{
		{Pair: btcUSD, Amount: 52000.12},
	}

	ltpService.On("GetLTPs", "BTC/USD").Return(expectedLTPs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetLTP(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.LTPResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.LTP, 1)
	assert.Equal(t, "BTC/USD", response.LTP[0].Pair)
	assert.Equal(t, 52000.12, response.LTP[0].Amount)

	ltpService.AssertExpectations(t)
}

func TestHandler_GetLTP_Success_MultiplePairs(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	expectedLTPs := []domain.LTP{
		{Pair: btcEUR, Amount: 50000.12},
		{Pair: btcUSD, Amount: 52000.12},
	}

	ltpService.On("GetLTPs", "BTC/USD,BTC/EUR").Return(expectedLTPs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD,BTC/EUR", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetLTP(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.LTPResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.LTP, 2)

	ltpService.AssertExpectations(t)
}

func TestHandler_GetLTP_ServiceError_ReturnsBadRequest(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)

	expectedError := errors.New("invalid pair: BTC/INVALID")
	ltpService.On("GetLTPs", "BTC/INVALID").Return(nil, expectedError)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/INVALID", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetLTP(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response dto.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Error, "invalid pair")

	ltpService.AssertExpectations(t)
}

func TestHandler_GetLTP_EmptyPairsQueryParam(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcCHF, _ := domain.NewPair(domain.BTCCHF)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	expectedLTPs := []domain.LTP{
		{Pair: btcCHF, Amount: 49000.12},
		{Pair: btcEUR, Amount: 50000.12},
		{Pair: btcUSD, Amount: 52000.12},
	}

	ltpService.On("GetLTPs", "").Return(expectedLTPs, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.GetLTP(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.LTPResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.LTP, 3)

	ltpService.AssertExpectations(t)
}

func TestHandler_Health_Success(t *testing.T) {
	// Arrange
	ltpService := new(mocks.LTPService)
	handler := NewHandler(ltpService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Act
	err := handler.Health(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])

	ltpService.AssertNotCalled(t, "GetLTPs")
}
