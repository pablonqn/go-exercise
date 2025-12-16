# Bitcoin LTP API

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/go-exercise)](https://goreportcard.com/report/github.com/yourusername/go-exercise)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/yourusername/go-exercise)
[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/yourusername/go-exercise)

REST API to retrieve the Last Traded Price of Bitcoin for currency pairs (BTC/USD, BTC/CHF, BTC/EUR).

## 🚀 Quick Start

[▶️ **Run in Gitpod**](https://gitpod.io/#https://github.com/yourusername/go-exercise) | [▶️ **Run in Codespaces**](https://codespaces.new/yourusername/go-exercise)

```bash
# Clone and run
git clone <repo-url>
cd go-exercise
go run ./cmd/server

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
go build -o bin/server ./cmd/server
```

## Run

```bash
go run ./cmd/server
```

The server starts at `http://localhost:8080`

## Tests

### Unit tests
```bash
go test -v ./...
```

### Integration tests
```bash
go test -v ./tests/integration/...
```

## Docker

### Build
```bash
docker build -t bitcoin-ltp-api:latest .
```

### Run
```bash
docker run -p 8080:8080 bitcoin-ltp-api:latest
```

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
