package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go-exercise/internal/adapters/http/dto"
	"go-exercise/internal/ports"
)

// Handler handles HTTP requests
type Handler struct {
	ltpService ports.LTPService
}

// NewHandler creates a new HTTP handler
func NewHandler(ltpService ports.LTPService) *Handler {
	return &Handler{
		ltpService: ltpService,
	}
}

// GetLTP handles GET /api/v1/ltp
// @Summary Get Last Traded Price
// @Description Get LTP for BTC currency pairs (BTC/USD, BTC/CHF, BTC/EUR). If no pairs are specified, returns all pairs.
// @Tags ltp
// @Accept json
// @Produce json
// @Param pairs query string false "Currency pairs (comma-separated, e.g., BTC/USD,BTC/EUR)"
// @Success 200 {object} dto.LTPResponse "Successfully retrieved LTP data"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/ltp [get]
func (h *Handler) GetLTP(c echo.Context) error {
	pairsStr := c.QueryParam("pairs")

	ltps, err := h.ltpService.GetLTPs(pairsStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
	}

	// Convert domain LTPs to DTOs
	ltpItems := make([]dto.LTPItem, len(ltps))
	for i, ltp := range ltps {
		ltpItems[i] = dto.LTPItem{
			Pair:   ltp.Pair.Value(),
			Amount: ltp.Amount,
		}
	}

	return c.JSON(http.StatusOK, dto.LTPResponse{
		LTP: ltpItems,
	})
}

// Health handles GET /health
// @Summary Health check
// @Description Health check endpoint
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Service is healthy"
// @Router /health [get]
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

