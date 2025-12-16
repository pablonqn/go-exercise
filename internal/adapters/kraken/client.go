package kraken

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// findKrakenSymbolInResult searches for a symbol in the result map, trying different variants
// Kraken may return symbols with different formats (e.g., "XXBTZUSD" instead of "XBTUSD")
func findKrakenSymbolInResult(result map[string]KrakenTickerData, requestedSymbol string) (KrakenTickerData, string, bool) {
	// Try exact match first
	if tickerData, ok := result[requestedSymbol]; ok {
		return tickerData, requestedSymbol, true
	}

	// Build variants based on common Kraken symbol patterns
	variants := []string{}

	// If symbol starts with XBT, try XXBT variants (common pattern: XBTUSD -> XXBTZUSD)
	if len(requestedSymbol) >= 3 && requestedSymbol[:3] == "XBT" {
		suffix := requestedSymbol[3:] // Everything after XBT (e.g., "USD", "EUR", "CHF")
		variants = append(variants,
			"XXBTZ"+suffix,      // XBTUSD -> XXBTZUSD (most common variant)
			"XXBT"+suffix,       // XBTUSD -> XXBTUSD
			"X"+requestedSymbol, // XBTUSD -> XXBTUSD
		)
	} else if len(requestedSymbol) > 0 && requestedSymbol[0] == 'X' {
		// If starts with X, try XX variants
		variants = append(variants,
			"X"+requestedSymbol,      // XBTUSD -> XXBTUSD
			"XX"+requestedSymbol[1:], // XBTUSD -> XXBTUSD
		)
	} else {
		// Generic variants
		variants = append(variants,
			"X"+requestedSymbol,
		)
	}

	for _, variant := range variants {
		if tickerData, ok := result[variant]; ok {
			return tickerData, variant, true
		}
	}

	// If only one result and we requested one pair, use it
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

	// Build comma-separated list of symbols
	symbols := make([]string, len(pairs))
	for i, pair := range pairs {
		symbols[i] = pairToKrakenSymbol(pair)
	}

	url := fmt.Sprintf("%s/Ticker?pair=%s", k.baseURL, symbols[0])
	if len(symbols) > 1 {
		// Kraken accepts comma-separated pairs
		url = fmt.Sprintf("%s/Ticker?pair=%s", k.baseURL, joinStrings(symbols, ","))
	}

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

		var amount float64
		if _, err := fmt.Sscanf(tickerData.C[0], "%f", &amount); err != nil {
			return nil, fmt.Errorf("failed to parse amount for %s (found as %s): %w", pair.Value(), foundSymbol, err)
		}

		result = append(result, domain.LTP{
			Pair:   pair,
			Amount: amount,
		})
	}

	return result, nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
