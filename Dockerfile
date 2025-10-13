# ---------- Build Stage ----------
ARG go_base_image=golang:1.25-alpine
FROM ${go_base_image} AS builder

# Working directory for sources
WORKDIR /src

# Redirect Go cache to a writable path (important for OpenShift)
ENV GOCACHE=/tmp/go-build
ENV GOMODCACHE=/tmp/go-mod

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
ENV GO_BUILDFLAGS="-trimpath -ldflags=-s -ldflags=-w"
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build $GO_BUILDFLAGS -o api .

# --------- Runtime Stage ---------
FROM alpine AS runner

# Working directory for runtime
WORKDIR /app

# Install tzdata for timezone handling, curl for Debezium setup
RUN apk add --no-cache tzdata curl
ENV TZ=Asia/Kolkata

# Copy binaries from builder
COPY --from=builder /src/api ./bin/api
COPY --from=builder /src/scripts ./opt/scripts

# Set environment variables to use a writable cache directory
ENV GOCACHE=/app/.cache/go-build
ENV GOMODCACHE=/app/.cache/mod
ENV XDG_CACHE_HOME=/app/.cache

# Ensure proper permissions for OpenShift arbitrary UID
RUN mkdir -p /app/.cache /app/logs && \
    chgrp -R 0 /app && \
    chmod -R g=u /app && \
    chmod -R g+w ./bin
