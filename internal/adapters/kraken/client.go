package kraken

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-exercise/internal/domain"
	"go-exercise/internal/ports"
)

// KrakenClient implements the External port for Kraken API
type KrakenClient struct {
	baseURL    string
	httpClient *http.Client
}

// KrakenTickerResponse represents the response from Kraken API
type KrakenTickerResponse struct {
	Error  []string                    `json:"error"`
	Result map[string]KrakenTickerData `json:"result"`
}

// KrakenTickerData represents ticker data for a pair
type KrakenTickerData struct {
	C []string `json:"c"` // c[0] = last trade closed price
}

// NewKrakenClient creates a new Kraken client
func NewKrakenClient(baseURL string) ports.External {
	if baseURL == "" {
		baseURL = "https://api.kraken.com/0/public"
	}
	return &KrakenClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// pairToKrakenSymbol converts domain pair to Kraken symbol for API request
func pairToKrakenSymbol(pair domain.Pair) string {
	mapping := map[string]string{
		domain.BTCUSD: "XBTUSD",
		domain.BTCCHF: "XBTCHF",
		domain.BTCEUR: "XBTEUR",
	}
	if symbol, ok := mapping[pair.Value()]; ok {
		return symbol
	}
	return pair.Value()
}

// findKrakenSymbolInResult searches for a symbol in the result map
// Kraken sometimes returns symbols with different formats (e.g., "XXBTZUSD" instead of "XBTUSD")
func findKrakenSymbolInResult(result map[string]KrakenTickerData, requestedSymbol string) (KrakenTickerData, string, bool) {
	// Try exact match first
	if tickerData, ok := result[requestedSymbol]; ok {
		return tickerData, requestedSymbol, true
	}

	// Common variants for XBT symbols
	if len(requestedSymbol) >= 3 && requestedSymbol[:3] == "XBT" {
		suffix := requestedSymbol[3:]
		variants := []string{
			"XXBTZ" + suffix, // Most common: XBTUSD -> XXBTZUSD
			"XXBT" + suffix,
		}
		for _, variant := range variants {
			if tickerData, ok := result[variant]; ok {
				return tickerData, variant, true
			}
		}
	}

	// If only one result exists, use it
	if len(result) == 1 {
		for symbol, tickerData := range result {
			return tickerData, symbol, true
		}
	}

	return KrakenTickerData{}, "", false
}

// GetTicker retrieves ticker information for a single pair
func (k *KrakenClient) GetTicker(pair domain.Pair) (domain.LTP, error) {
	ltps, err := k.GetTickers([]domain.Pair{pair})
	if err != nil {
		return domain.LTP{}, err
	}
	if len(ltps) == 0 {
		return domain.LTP{}, fmt.Errorf("no data returned for pair %s", pair.Value())
	}
	return ltps[0], nil
}

// GetTickers retrieves ticker information for multiple pairs
func (k *KrakenClient) GetTickers(pairs []domain.Pair) ([]domain.LTP, error) {
	if len(pairs) == 0 {
		return nil, fmt.Errorf("no pairs provided")
	}

	// Convert pairs to Kraken symbols
	symbols := make([]string, len(pairs))
	for i, pair := range pairs {
		symbols[i] = pairToKrakenSymbol(pair)
	}

	// Build URL with comma-separated symbols
	pairParam := strings.Join(symbols, ",")
	url := fmt.Sprintf("%s/Ticker?pair=%s", k.baseURL, pairParam)

	resp, err := k.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call Kraken API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kraken API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tickerResp KrakenTickerResponse
	if err := json.Unmarshal(body, &tickerResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(tickerResp.Error) > 0 {
		return nil, fmt.Errorf("kraken API error: %v", tickerResp.Error)
	}

	// Map response to domain LTPs
	result := make([]domain.LTP, 0, len(pairs))
	for _, pair := range pairs {
		requestedSymbol := pairToKrakenSymbol(pair)
		tickerData, foundSymbol, ok := findKrakenSymbolInResult(tickerResp.Result, requestedSymbol)
		if !ok {
			return nil, fmt.Errorf("no data found for symbol %s (tried %s and variants)", pair.Value(), requestedSymbol)
		}

		if len(tickerData.C) == 0 || tickerData.C[0] == "" {
			return nil, fmt.Errorf("invalid ticker data for symbol %s (found as %s)", pair.Value(), foundSymbol)
		}

		amount, err := strconv.ParseFloat(tickerData.C[0], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount for %s (found as %s): %w", pair.Value(), foundSymbol, err)
		}

		result = append(result, domain.LTP{
			Pair:   pair,
			Amount: amount,
		})
	}

	return result, nil
}
