# Build stage
FROM golang:1.24.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gopls-mcp .

# Runtime stage
FROM golang:1.24.4-alpine AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates git

# Install gopls
RUN go install golang.org/x/tools/gopls@latest

# Create non-root user
RUN addgroup -g 1001 -S gopls && \
    adduser -S gopls -u 1001 -G gopls

# Set working directory
WORKDIR /workspace

# Copy binary from builder stage
COPY --from=builder /app/gopls-mcp /usr/local/bin/gopls-mcp

# Change ownership of the binary
RUN chown gopls:gopls /usr/local/bin/gopls-mcp

# Switch to non-root user
USER gopls

# Expose port
EXPOSE 8080

# Set default workspace
ENV WORKSPACE_PATH=/workspace

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/sse || exit 1

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/gopls-mcp"]
CMD ["-workspace", "/workspace"]