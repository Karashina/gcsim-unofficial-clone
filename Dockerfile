# Multi-stage Dockerfile for gcsim-unofficial-clone targeting Debian 12 (amd64)
# Stage 1: build the Go binary
FROM golang:1.20 AS builder

WORKDIR /src
# Download modules first for cache
COPY go.mod go.sum ./
RUN go mod download

# Copy repository
COPY . .

# Optionally generate mappings during build (uncomment if desired)
# RUN go run ./cmd/generate-webui-mappings

# Build a statically linked linux/amd64 binary for portability
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/go/pkg/mod \
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /out/gcsim-webui ./cmd/gcsim-webui

# Stage 2: minimal runtime image (Debian 12 slim)
FROM debian:12-slim

# Install ca-certificates for HTTPS calls
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary and web UI assets (if server expects a local webui folder)
COPY --from=builder /out/gcsim-webui /app/gcsim-webui
COPY --from=builder /src/webui /app/webui

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/gcsim-webui"]
