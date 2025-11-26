# Docker Health Check Implementation Guide

## Current Dockerfile Health Check

The Dockerfile now includes:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/server", "-health-check"] || exit 1
```

## Implementation Options

### Option 1: Add Health Check Flag to Your Application (Recommended)

Add a simple health check flag to your `main.go`:

```go
package main

import (
    "flag"
    "fmt"
    "net/http"
    "os"
)

func main() {
    // Health check flag
    healthCheck := flag.Bool("health-check", false, "Run health check")
    flag.Parse()

    if *healthCheck {
        // Simple HTTP health check
        resp, err := http.Get("http://localhost:8080/health")
        if err != nil || resp.StatusCode != 200 {
            os.Exit(1)
        }
        os.Exit(0)
    }

    // Your normal application code here...
}
```

And add a health endpoint to your application:

```go
// In your router setup
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
    })
})
```

### Option 2: Use grpc_health_probe (For gRPC Apps)

If you're using gRPC, add grpc_health_probe to the builder stage:

```dockerfile
# In builder stage
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.24 && \
    wget -qO/bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# Copy to final stage
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/bin/grpc_health_probe", "-addr=:9090"]
```

### Option 3: External Health Check (Docker Compose)

Remove HEALTHCHECK from Dockerfile and use docker-compose.yml:

```yaml
services:
  app:
    image: go-base:latest
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
```

### Option 4: Simple TCP Check (No Application Changes)

For a basic TCP port check without HTTP:

```dockerfile
# Use base image with shell for healthcheck
FROM gcr.io/distroless/base-debian12:nonroot

# Health check - simple TCP check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/bin/sh", "-c", "timeout 1 bash -c '</dev/tcp/localhost/8080' || exit 1"]
```

**Note**: This requires using `base-debian12` instead of `static-debian12` (adds ~10MB).

## Recommended Approach

**For your go-base application**, I recommend **Option 1**:

1. Add `/health` endpoint to your Gin router
2. Add `-health-check` flag support to main.go
3. Keep the current Dockerfile HEALTHCHECK

This provides:
- ✅ Minimal overhead
- ✅ Application-level health verification
- ✅ Works with distroless static image
- ✅ Can check database connectivity, etc.

## Health Check Parameters

```dockerfile
HEALTHCHECK --interval=30s    # Check every 30 seconds
            --timeout=3s      # Timeout after 3 seconds
            --start-period=5s # Grace period on startup
            --retries=3       # Mark unhealthy after 3 failures
```

## Testing Health Check

```bash
# Build image
docker build -t go-base:latest .

# Run container
docker run -d --name test-app go-base:latest

# Check health status
docker inspect --format='{{.State.Health.Status}}' test-app

# View health check logs
docker inspect --format='{{json .State.Health}}' test-app | jq
```

## Health Status Values

- `starting` - During start_period
- `healthy` - Health check passing
- `unhealthy` - Health check failing (after retries)
