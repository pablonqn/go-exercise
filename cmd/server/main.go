package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"go-exercise/internal/adapters/cache"
	httphandler "go-exercise/internal/adapters/http"
	"go-exercise/internal/adapters/kraken"
	"go-exercise/internal/application/service"
	
	_ "go-exercise/docs" // Swagger documentation
)

// @title Bitcoin LTP API
// @version 1.0
// @description API for retrieving Last Traded Price of Bitcoin for currency pairs (BTC/USD, BTC/CHF, BTC/EUR)
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	// Initialize adapters
	krakenClient := kraken.NewKrakenClient("")
	cacheRepo := cache.NewInMemoryCache()

	// Initialize application service
	ltpService := service.NewLTPService(cacheRepo, krakenClient)

	// Initialize HTTP handler
	handler := httphandler.NewHandler(ltpService)

	// Setup router
	var e *echo.Echo = httphandler.SetupRouter(handler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server in a goroutine
	go func() {
		if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Server exited")
}

