package kraken

import (
	"encoding/json"
	"testing"

	"go-exercise/internal/domain"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKrakenClient(t *testing.T) {
	t.Run("with empty baseURL uses default", func(t *testing.T) {
		client := NewKrakenClient("")
		krakenClient, ok := client.(*KrakenClient)
		require.True(t, ok)
		assert.Equal(t, "https://api.kraken.com/0/public", krakenClient.baseURL)
		assert.NotNil(t, krakenClient.httpClient)
	})

	t.Run("with custom baseURL uses provided", func(t *testing.T) {
		customURL := "http://localhost:8080"
		client := NewKrakenClient(customURL)
		krakenClient, ok := client.(*KrakenClient)
		require.True(t, ok)
		assert.Equal(t, customURL, krakenClient.baseURL)
	})
}

func TestPairToKrakenSymbol(t *testing.T) {
	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcCHF, _ := domain.NewPair(domain.BTCCHF)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	tests := []struct {
		name     string
		pair     domain.Pair
		expected string
	}{
		{"BTC/USD", btcUSD, "XBTUSD"},
		{"BTC/CHF", btcCHF, "XBTCHF"},
		{"BTC/EUR", btcEUR, "XBTEUR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pairToKrakenSymbol(tt.pair)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindKrakenSymbolInResult(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		result := map[string]KrakenTickerData{
			"XBTUSD": {C: []string{"50000.12"}},
		}
		tickerData, symbol, ok := findKrakenSymbolInResult(result, "XBTUSD")
		assert.True(t, ok)
		assert.Equal(t, "XBTUSD", symbol)
		assert.Equal(t, "50000.12", tickerData.C[0])
	})

	t.Run("variant match XXBTZUSD", func(t *testing.T) {
		result := map[string]KrakenTickerData{
			"XXBTZUSD": {C: []string{"50000.12"}},
		}
		tickerData, symbol, ok := findKrakenSymbolInResult(result, "XBTUSD")
		assert.True(t, ok)
		assert.Equal(t, "XXBTZUSD", symbol)
		assert.Equal(t, "50000.12", tickerData.C[0])
	})

	t.Run("single result fallback", func(t *testing.T) {
		result := map[string]KrakenTickerData{
			"UNKNOWN": {C: []string{"50000.12"}},
		}
		tickerData, symbol, ok := findKrakenSymbolInResult(result, "XBTUSD")
		assert.True(t, ok)
		assert.Equal(t, "UNKNOWN", symbol)
		assert.Equal(t, "50000.12", tickerData.C[0])
	})

	t.Run("not found - multiple results", func(t *testing.T) {
		result := map[string]KrakenTickerData{
			"OTHER1": {C: []string{"50000.12"}},
			"OTHER2": {C: []string{"60000.12"}},
		}
		_, _, ok := findKrakenSymbolInResult(result, "XBTUSD")
		assert.False(t, ok)
	})
}

func TestKrakenClient_GetTicker_Success(t *testing.T) {
	defer gock.Off()

	// Setup gock mock
	response := KrakenTickerResponse{
		Error: []string{},
		Result: map[string]KrakenTickerData{
			"XXBTZUSD": {C: []string{"52000.12"}},
		},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	ltp, err := client.GetTicker(pair)

	require.NoError(t, err)
	assert.Equal(t, pair, ltp.Pair)
	assert.Equal(t, 52000.12, ltp.Amount)
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTicker_NoData(t *testing.T) {
	defer gock.Off()

	response := KrakenTickerResponse{
		Error:  []string{},
		Result: map[string]KrakenTickerData{},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	_, err := client.GetTicker(pair)

	assert.Error(t, err)
	// GetTicker calls GetTickers, which will return "no data found for symbol" error
	assert.Contains(t, err.Error(), "no data")
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_Success_SinglePair(t *testing.T) {
	defer gock.Off()

	response := KrakenTickerResponse{
		Error: []string{},
		Result: map[string]KrakenTickerData{
			"XXBTZUSD": {C: []string{"52000.12"}},
		},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	ltps, err := client.GetTickers([]domain.Pair{pair})

	require.NoError(t, err)
	require.Len(t, ltps, 1)
	assert.Equal(t, pair, ltps[0].Pair)
	assert.Equal(t, 52000.12, ltps[0].Amount)
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_Success_MultiplePairs(t *testing.T) {
	defer gock.Off()

	response := KrakenTickerResponse{
		Error: []string{},
		Result: map[string]KrakenTickerData{
			"XXBTZUSD": {C: []string{"52000.12"}},
			"XXBTZEUR": {C: []string{"50000.12"}},
		},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD,XBTEUR").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	btcUSD, _ := domain.NewPair(domain.BTCUSD)
	btcEUR, _ := domain.NewPair(domain.BTCEUR)

	ltps, err := client.GetTickers([]domain.Pair{btcUSD, btcEUR})

	require.NoError(t, err)
	require.Len(t, ltps, 2)

	// Check that both pairs are present
	pairs := make(map[string]float64)
	for _, ltp := range ltps {
		pairs[ltp.Pair.Value()] = ltp.Amount
	}
	assert.Equal(t, 52000.12, pairs[domain.BTCUSD])
	assert.Equal(t, 50000.12, pairs[domain.BTCEUR])
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_EmptyPairs(t *testing.T) {
	client := NewKrakenClient("http://localhost").(*KrakenClient)

	_, err := client.GetTickers([]domain.Pair{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pairs provided")
}

func TestKrakenClient_GetTickers_HTTPError(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(500).
		BodyString("Internal Server Error")

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	_, err := client.GetTickers([]domain.Pair{pair})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_InvalidJSON(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		BodyString("invalid json")

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	_, err := client.GetTickers([]domain.Pair{pair})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_APIError(t *testing.T) {
	defer gock.Off()

	response := KrakenTickerResponse{
		Error:  []string{"EGeneral:Invalid arguments"},
		Result: map[string]KrakenTickerData{},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	_, err := client.GetTickers([]domain.Pair{pair})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kraken API error")
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_NoDataForSymbol(t *testing.T) {
	defer gock.Off()

	response := KrakenTickerResponse{
		Error: []string{},
		Result: map[string]KrakenTickerData{
			"OTHER1": {C: []string{"100.0"}},
			"OTHER2": {C: []string{"200.0"}},
		},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	_, err := client.GetTickers([]domain.Pair{pair})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data found for symbol")
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_InvalidTickerData(t *testing.T) {
	defer gock.Off()

	response := KrakenTickerResponse{
		Error: []string{},
		Result: map[string]KrakenTickerData{
			"XXBTZUSD": {C: []string{}}, // Empty price
		},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	_, err := client.GetTickers([]domain.Pair{pair})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ticker data")
	assert.True(t, gock.IsDone())
}

func TestKrakenClient_GetTickers_InvalidAmount(t *testing.T) {
	defer gock.Off()

	response := KrakenTickerResponse{
		Error: []string{},
		Result: map[string]KrakenTickerData{
			"XXBTZUSD": {C: []string{"invalid"}},
		},
	}
	responseBody, _ := json.Marshal(response)

	gock.New("https://api.kraken.com").
		Get("/0/public/Ticker").
		MatchParam("pair", "XBTUSD").
		Reply(200).
		JSON(responseBody)

	client := NewKrakenClient("").(*KrakenClient)
	pair, _ := domain.NewPair(domain.BTCUSD)

	_, err := client.GetTickers([]domain.Pair{pair})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse amount")
	assert.True(t, gock.IsDone())
}

