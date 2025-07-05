# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

- **Docker Image Optimization**: Reduced final Docker image size from 766MB to 121MB (84% reduction) by using alpine:latest base image and copying gopls binary from builder stage

### Fixed

### Removed

## [v0.2.0] - 2025-07-05

### Added

- **Structured Logging with slog**: Replaced standard log package with slog for structured, contextual logging
- **Logging Configuration**: Environment variable support for log level (`LOG_LEVEL`) and format (`LOG_FORMAT`)
- **Stdio Transport Support**: Added `-transport` flag supporting both 'http' and 'stdio' transports for full MCP specification compliance
- **LSP Communication Improvements**: Continuous message reader, proper header parsing, request-response correlation, and timeout protection
- **File Management**: Automatic file opening in gopls with `textDocument/didOpen` before making requests
- **Workspace Readiness Tracking**: Monitors gopls notifications and waits for "Finished loading packages" before allowing requests
- **Enhanced Error Handling**: Comprehensive error handling for LSP communication failures
- **Timeout Management**: Extended LSP request timeout to 60 seconds with progress logging for large codebases
- **Transport Tests**: Comprehensive test coverage for transport flag parsing and validation

### Changed

- **BREAKING**: Replaced SSE transport with streamable HTTP transport for MCP specification compliance
- Updated connection endpoint from `http://localhost:8080/sse` to `http://localhost:8080`
- Updated client connection from `--transport sse` to `--transport http`
- Simplified HTTP server setup by removing mux routing in favor of direct handler
- Updated Docker health check to use root path instead of `/sse` endpoint

## [v0.1.0] - 2025-07-05

### Added

- Initial release of gopls-mcp MCP server
- MCP server implementation with SSE transport over HTTP
- gopls process management and LSP communication
- Three MCP tools:
  - `go_to_definition`: Navigate to symbol definitions
  - `find_references`: Find all references to symbols
  - `get_hover_info`: Get documentation and type information
- HTTP server with graceful shutdown (port 8080)
- Command-line interface with required `-workspace` flag
- Docker support with multi-stage builds and security hardening
- Multi-platform Docker images (linux/amd64, linux/arm64)
- GitHub Actions CI/CD pipeline with quality gates
- Comprehensive test suite (26 unit tests)
- golangci-lint configuration and code formatting checks
- Docker Hub publishing to `megagrindstone/gopls-mcp`

### Technical Details

- Go 1.24.4
- Dependency: `github.com/modelcontextprotocol/go-sdk v0.1.0`
- SSE endpoint: `/sse`
- Docker health checks and non-root user execution

