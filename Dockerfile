# Build stage
FROM golang:1.22-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make openssl

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o lenovo-console ./cmd/lenovo-console

# Generate self-signed certificates
RUN openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
    -days 365 -nodes -subj "/CN=localhost"

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S console && \
    adduser -u 1000 -S console -G console

# Set working directory
WORKDIR /home/console

# Copy binary and certificates from builder
COPY --from=builder /app/lenovo-console /usr/local/bin/
COPY --from=builder /app/server.crt /app/server.key ./

# Change ownership
RUN chown -R console:console /home/console

# Switch to non-root user
USER console

# Expose port (can be overridden)
EXPOSE 8443

# Set entrypoint
ENTRYPOINT ["lenovo-console"]

# Default CMD (can be overridden)
CMD ["--help"]
