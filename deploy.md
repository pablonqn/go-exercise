# Bitcoin LTP API

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/go-exercise)](https://goreportcard.com/report/github.com/yourusername/go-exercise)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

REST API to retrieve the Last Traded Price of Bitcoin for currency pairs (BTC/USD, BTC/CHF, BTC/EUR).

## 🚀 Quick Start

```bash
# Clone and run
git clone <repo-url>
cd go-exercise
make run

# Test the API
curl http://localhost:8080/api/v1/ltp?pairs=BTC/USD
```

## Architecture

Hexagonal Architecture (Ports & Adapters):
- **Domain**: Entities and value objects
- **Application**: Business logic
- **Ports**: Interfaces
- **Adapters**: Implementations (HTTP, Kraken, Cache)

## Build

```bash
make build
```

## Run

```bash
make run
```

The server starts at `http://localhost:8080`

## Tests

### Unit tests
```bash
make test
```

### Integration tests (Container tests)
```bash
make inttest
```

This command will:
1. Build the Docker image (`bitcoin-ltp-api:test`)
2. Run integration tests using testcontainers

## Docker

### Build
```bash
make docker-build
```

### Run
```bash
make docker-run
```

## Available Commands

Run `make` or `make help` to see all available commands:

```bash
make help
```

Available commands:
- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run all unit tests
- `make inttest` - Run integration tests (requires Docker, builds image first)
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make swagger` - Generate Swagger documentation
- `make clean` - Clean build artifacts
- `make deps` - Install/update dependencies
- `make install-swag` - Install swag tool for Swagger

## API Endpoints

### GET `/api/v1/ltp`
Retrieves LTP for specified pairs or all pairs if none specified.

**Query params:**
- `pairs` (optional): Comma-separated pairs (e.g., `BTC/USD,BTC/EUR`)

**Example:**
```bash
curl http://localhost:8080/api/v1/ltp?pairs=BTC/USD
```

### GET `/health`
Health check endpoint.

### GET `/swagger/index.html`
Interactive API documentation.

## Project Structure

```
go-exercise/
├── cmd/server/          # Entry point
├── internal/
│   ├── domain/          # Domain entities
│   ├── application/     # Application services
│   ├── ports/           # Interfaces
│   └── adapters/        # Implementations (http, kraken, cache)
├── tests/               # Integration tests
└── docs/                # Swagger documentation
```

## Technologies

- Go 1.21+
- Echo (HTTP framework)
- Swagger (Documentation)
- Gock (HTTP mocking for tests)
- Testify (Testing and mocks)
- Testcontainers (Container-based integration tests)
