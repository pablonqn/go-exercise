package dto

// LTPItem represents a single LTP item in the response
// @Description Single Last Traded Price item
type LTPItem struct {
	Pair   string  `json:"pair" example:"BTC/USD"`   // Currency pair
	Amount float64 `json:"amount" example:"52000.12"` // Last traded price amount
}

// LTPResponse represents the API response structure
// @Description Response containing list of Last Traded Prices
type LTPResponse struct {
	LTP []LTPItem `json:"ltp"` // List of LTP items
}

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	Error string `json:"error" example:"invalid pair: BTC/INVALID"` // Error message
}

