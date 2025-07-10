# Single-stage build using golang:1.24.4-alpine
FROM golang:1.24.4-alpine

# Install dependencies
RUN apk add --no-cache ca-certificates git

# Set working directory for build
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gopls-mcp .

# Install gopls
RUN go install golang.org/x/tools/gopls@latest

# Create non-root user
RUN addgroup -g 1001 -S gopls && \
    adduser -S gopls -u 1001 -G gopls

# Set working directory for runtime
WORKDIR /workspace

# Copy binary to standard location
RUN cp /app/gopls-mcp /usr/local/bin/gopls-mcp

# Change ownership of the binaries
RUN chown gopls:gopls /usr/local/bin/gopls-mcp /go/bin/gopls

# Switch to non-root user
USER gopls

# Expose port
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/gopls-mcp"]
CMD ["-workspace", "/workspace"]