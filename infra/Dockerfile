# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with caching
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath -o /app/server .

# Final stage - distroless
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/server /app/server

# Copy any required files (adjust as needed)
# COPY --from=builder /build/private_key.pem /app/private_key.pem
# COPY --from=builder /build/public_key.pem /app/public_key.pem

# Set working directory
WORKDIR /app

# Expose port (adjust to your app's port)
EXPOSE 8080

# Use non-root user (distroless nonroot user)
USER nonroot:nonroot

# Health check - checks if the app is responding on the health endpoint
# Note: Distroless images don't have curl/wget, so we use a simple TCP check
# For HTTP health checks, consider using a sidecar or external monitoring
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/server", "-health-check"] || exit 1

# Run the application
ENTRYPOINT ["/app/server"]
