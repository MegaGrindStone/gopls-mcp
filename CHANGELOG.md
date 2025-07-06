# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Multi-Workspace Support**: Complete multi-workspace architecture allowing multiple Go projects in a single server instance
- **WorkspaceManager Component**: New component to coordinate multiple Manager instances with dedicated gopls processes per workspace
- **New MCP Tool**: `list_workspaces` tool to display all available workspaces and their running status
- **Phase 1 Core Navigation MCP Tools**: Three foundational navigation and discovery tools:
  - `get_document_symbols`: Lists all symbols (functions, types, variables, etc.) in a specific Go file
  - `search_workspace_symbols`: Searches for symbols across the entire Go workspace using fuzzy matching
  - `go_to_type_definition`: Navigates to the type definition of a symbol at the specified position
- **LSP Method Integration**: Added support for `textDocument/documentSymbol`, `workspace/symbol`, and `textDocument/typeDefinition` LSP methods
- **Symbol Type Definitions**: Complete LSP symbol type system including `SymbolKind` constants, `DocumentSymbol`, and `SymbolInformation` types
- **Enhanced MCP Response Types**: New structured response types for document symbols, workspace symbols, and type definitions
- **Workspace Routing**: Intelligent request routing based on workspace parameter in MCP tool calls
- **Enhanced Docker Support**: Multi-workspace Docker deployment with multiple volume mount examples
- **Comprehensive Testing Coverage**: 47 tests covering single workspace, multiple workspaces, workspace management, and new navigation tools

### Changed

- **BREAKING**: Replaced `-workspace` flag with `-workspaces` flag accepting comma-separated workspace paths
- **BREAKING**: All MCP tools now require mandatory `workspace` parameter for workspace routing:
  - `go_to_definition` requires `workspace` field
  - `find_references` requires `workspace` field  
  - `get_hover_info` requires `workspace` field
  - `get_document_symbols` requires `workspace` field
  - `search_workspace_symbols` requires `workspace` field
  - `go_to_type_definition` requires `workspace` field
- **MCP Server Enhancement**: Expanded from 4 to 7 total MCP tools for comprehensive Go code navigation and discovery
- **BREAKING**: No backward compatibility - old single-workspace usage patterns no longer supported
- Updated Dockerfile CMD to use `-workspaces` flag instead of `-workspace`
- Enhanced documentation with multi-workspace usage examples and Docker deployment patterns
- Updated command-line help and flag descriptions for multi-workspace usage

### Removed

- **BREAKING**: Removed `-workspace` flag (replaced by `-workspaces`)
- **BREAKING**: Removed automatic workspace discovery - explicit workspace specification now required

## [v0.2.1] - 2025-07-05

### Changed

- **Docker Image Optimization**: Reduced final Docker image size from 766MB to 121MB (84% reduction) by using alpine:latest base image and copying gopls binary from builder stage

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

