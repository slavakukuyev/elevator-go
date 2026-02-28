# syntax=docker/dockerfile:1
# Stage 1: Build the Go binary
FROM golang:1.25-alpine3.21 AS builder

# Install git and ca-certificates for dependency management
RUN apk add --no-cache git ca-certificates && \
    update-ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./

# Download all dependencies with optimized settings for faster builds
# Use Go proxy for faster dependency resolution in CI
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy only necessary source files
COPY cmd/ cmd/
COPY internal/ internal/
COPY metrics/ metrics/
COPY configs/production.env configs/

# Build the Go app with optimization flags and build cache
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/out/elevator-go \
    ./cmd/server

# Stage 2: Create lightweight runtime image with shell
FROM alpine:3.21

# Add metadata labels
LABEL maintainer="Slava Kukuyev"
LABEL description="Elevator Go Application - Runtime"
LABEL version="1.0"

# Install essential debugging and runtime tools
RUN apk add --no-cache \
    ca-certificates \
    curl \
    wget \
    netcat-openbsd \
    procps \
    htop \
    tree \
    jq \
    tzdata && \
    update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 appgroup && \
    adduser -D -s /bin/ash -u 1001 -G appgroup appuser

# Set working directory
WORKDIR /app

# Copy the pre-built binary with proper ownership
COPY --from=builder --chown=appuser:appgroup /app/out/elevator-go /app/elevator-go

# Copy production config with proper ownership
COPY --from=builder --chown=appuser:appgroup /app/configs/production.env /app/configs/production.env

# Make binary executable
RUN chmod +x /app/elevator-go

# Create necessary directories with proper permissions
RUN mkdir -p /tmp/elevator && \
    chown -R appuser:appgroup /tmp/elevator

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 6660

# Health check using curl (now available in the image)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:6660/v1/health/live || exit 1

# Command to run the executable
ENTRYPOINT ["/app/elevator-go"]
