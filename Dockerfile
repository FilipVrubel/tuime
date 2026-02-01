# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies (gcc and musl-dev needed for CGO/SQLite)
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

COPY go.mod ./
# go.sum is optional, only copy if it exists
COPY go.su[m] ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (CGO needed for modernc.org/sqlite)
# Add build flags for smaller binary
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o tuime .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 tuime && \
    adduser -D -u 1000 -G tuime tuime

# Set working directory
WORKDIR /home/tuime

# Copy binary from builder
COPY --from=builder /app/tuime /usr/local/bin/tuime

# Create directories for data and config
RUN mkdir -p /home/tuime/.local/share/tuime /home/tuime/.config/tuime && \
    chown -R tuime:tuime /home/tuime

# Switch to non-root user
USER tuime

# Set environment variables
ENV HOME=/home/tuime

# Volume for persistent data
VOLUME ["/home/tuime/.local/share/tuime", "/home/tuime/.config/tuime"]

# Healthcheck (verify binary is accessible)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["sh", "-c", "test -x /usr/local/bin/tuime || exit 1"]

# Entry point
ENTRYPOINT ["tuime"]
