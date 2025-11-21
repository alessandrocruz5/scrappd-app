# Multi-stage Dockerfile for Go API Service

# ============================================
# Stage 1: Base
# ============================================
FROM golang:1.21-alpine AS base

# Install build dependencies
RUN apk add --no-cache git make bash

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# ============================================
# Stage 2: Development
# ============================================
FROM base AS development

# Install Air for hot reloading
RUN go install github.com/cosmtrek/air@latest

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Use Air for hot reloading
CMD ["air", "-c", ".air.toml"]

# ============================================
# Stage 3: Builder (for production)
# ============================================
FROM base AS builder

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/bin/server \
    ./cmd/server

# ============================================
# Stage 4: Production
# ============================================
FROM alpine:3.19 AS production

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 scrappd && \
    adduser -D -u 1000 -G scrappd scrappd

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server

# Copy migrations (if running migrations from container)
COPY --from=builder /app/migrations /app/migrations

# Change ownership
RUN chown -R scrappd:scrappd /app

# Switch to non-root user
USER scrappd

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/server"]