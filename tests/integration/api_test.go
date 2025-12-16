package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"go-exercise/internal/adapters/cache"
	httphandler "go-exercise/internal/adapters/http"
	"go-exercise/internal/application/service"
	"go-exercise/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExternalService is a mock implementation of the External port for testing
type mockExternalService struct {
	tickers map[string]domain.LTP
}

func newMockExternalService() *mockExternalService {
	return &mockExternalService{
		tickers: make(map[string]domain.LTP),
	}
}

func (m *mockExternalService) GetTicker(pair domain.Pair) (domain.LTP, error) {
	ltp, ok := m.tickers[pair.Value()]
	if !ok {
		return domain.LTP{}, assert.AnError
	}
	return ltp, nil
}

func (m *mockExternalService) GetTickers(pairs []domain.Pair) ([]domain.LTP, error) {
	result := make([]domain.LTP, 0, len(pairs))
	for _, pair := range pairs {
		if ltp, ok := m.tickers[pair.Value()]; ok {
			result = append(result, ltp)
		}
	}
	if len(result) == 0 {
		return nil, assert.AnError
	}
	return result, nil
}

func (m *mockExternalService) setTicker(pair domain.Pair, amount float64) {
	m.tickers[pair.Value()] = domain.LTP{
		Pair:   pair,
		Amount: amount,
	}
}

func setupTestServer() (*echo.Echo, *mockExternalService) {
	mockExternal := newMockExternalService()
	cacheRepo := cache.NewInMemoryCache()
	ltpService := service.NewLTPService(cacheRepo, mockExternal)
	handler := httphandler.NewHandler(ltpService)
	e := httphandler.SetupRouter(handler)
	return e, mockExternal
}

func TestGetLTP_AllPairs(t *testing.T) {
	e, mockExternal := setupTestServer()

	// Setup mock data
	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcCHF, _ := domain.NewPair(domain.BTCCHF)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	mockExternal.setTicker(btcUSD, 52000.12)
	mockExternal.setTicker(btcCHF, 49000.12)
	mockExternal.setTicker(btcEUR, 50000.12)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}

	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.LTP, 3)

	// Verify all pairs are present
	pairs := make(map[string]float64)
	for _, item := range response.LTP {
		pairs[item.Pair] = item.Amount
	}

	assert.Equal(t, 52000.12, pairs[domain.BTCUSD])
	assert.Equal(t, 49000.12, pairs[domain.BTCCHF])
	assert.Equal(t, 50000.12, pairs[domain.BTCEUR])
}

func TestGetLTP_SinglePair(t *testing.T) {
	e, mockExternal := setupTestServer()

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	mockExternal.setTicker(btcUSD, 52000.12)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}

	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.LTP, 1)
	assert.Equal(t, domain.BTCUSD, response.LTP[0].Pair)
	assert.Equal(t, 52000.12, response.LTP[0].Amount)
}

func TestGetLTP_MultiplePairs(t *testing.T) {
	e, mockExternal := setupTestServer()

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	mockExternal.setTicker(btcUSD, 52000.12)
	mockExternal.setTicker(btcEUR, 50000.12)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD,BTC/EUR", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}

	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.LTP, 2)
}

func TestGetLTP_InvalidPair(t *testing.T) {
	e, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/INVALID", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response struct {
		Error string `json:"error"`
	}

	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "invalid pair")
}

func TestGetLTP_CacheExpiration(t *testing.T) {
	e, mockExternal := setupTestServer()

	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	mockExternal.setTicker(btcUSD, 52000.12)

	// First request - should fetch from external
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second request - should use cache
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/ltp?pairs=BTC/USD", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Verify response structure
	var response struct {
		LTP []struct {
			Pair   string  `json:"pair"`
			Amount float64 `json:"amount"`
		} `json:"ltp"`
	}

	err := json.Unmarshal(rec2.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.LTP, 1)
	assert.Equal(t, 52000.12, response.LTP[0].Amount)
}

func TestHealthEndpoint(t *testing.T) {
	e, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

