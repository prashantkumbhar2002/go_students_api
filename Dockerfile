# Stage 1: Builder - Build the Go application
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy go mod files first (better layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
# CGO is required for SQLite
# -ldflags for smaller binary: strip debug info
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o students-api \
    ./cmd/go_students_api

# Verify the binary
RUN chmod +x students-api && \
    ./students-api --help 2>&1 | head -1 || echo "Binary built successfully"

# ============================================
# Stage 2: Runtime - Minimal production image
FROM alpine:3.19

# Add security: create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Install runtime dependencies only
RUN apk add --no-cache \
    ca-certificates \
    sqlite-libs \
    tzdata && \
    rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p /app/config /var/lib/students_api && \
    chown -R appuser:appuser /app /var/lib/students_api

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /build/students-api /app/

# Copy configuration files
COPY --chown=appuser:appuser config/ /app/config/

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Set environment variables
ENV CONFIG_PATH=/app/config/production.yml

# Run the application
CMD ["/app/students-api"]

