# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ynab-mcp-server .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 ynab && \
    adduser -D -u 1000 -G ynab ynab

WORKDIR /home/ynab

# Copy binary from builder
COPY --from=builder /app/ynab-mcp-server /usr/local/bin/ynab-mcp-server

# Copy config example
COPY config.json.example /home/ynab/

# Change ownership
RUN chown -R ynab:ynab /home/ynab

# Switch to non-root user
USER ynab

# Expose HTTP port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default to HTTP mode
ENTRYPOINT ["ynab-mcp-server"]
CMD ["serve", "--transport=http", "--port=8080"]
