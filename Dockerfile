# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install swag for Swagger documentation generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/server/main.go -o docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./server"]

