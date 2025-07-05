# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

