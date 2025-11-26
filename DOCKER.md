# Docker Build and Run Commands

## Build the Docker image
```bash
docker build -t go-base:latest .
```

## Build with BuildKit (recommended for caching)
```bash
DOCKER_BUILDKIT=1 docker build -t go-base:latest .
```

## Run the container
```bash
docker run -d \
  --name go-base-app \
  -p 8080:8080 \
  --env-file .env \
  go-base:latest
```

## Using Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

## Multi-platform build (for ARM/AMD64)
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t go-base:latest \
  --push .
```

## Image size optimization
The Dockerfile uses:
- Multi-stage build (builder + distroless)
- Go build caching with BuildKit
- Distroless base image (~20MB vs ~800MB with full Go image)
- Static binary compilation (CGO_ENABLED=0)
- Build flags: -ldflags="-w -s" (strip debug info)

Expected final image size: **~25-30MB**
