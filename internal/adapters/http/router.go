package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	
	_ "go-exercise/docs" // Swagger documentation
)

// SetupRouter configures the Echo router with routes and middleware
func SetupRouter(handler *Handler) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	api := e.Group("/api/v1")
	api.GET("/ltp", handler.GetLTP)

	// Health check
	e.GET("/health", handler.Health)

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

