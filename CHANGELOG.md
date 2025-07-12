# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **BREAKING**: Line numbering converted from 0-based to 1-based for all MCP tools to eliminate AI assistant off-by-one errors

## [v0.3.2] - 2025-07-10

### Fixed

- **Unnecessary Type Arguments**: Removed redundant type arguments from MCP tools and test functions for cleaner Go code
- **Diagnostic Timing Bug**: Fixed inconsistent diagnostic results on first call by preventing premature acceptance of empty diagnostics
- **Docker "No active builds" Error**: Fixed gopls module analysis failure in Docker containers by including full Go toolchain

### Improved

- **Mutex Placement**: Reorganized goplsClient struct to follow mutex placement guidelines for better concurrency clarity

## [v0.3.1] - 2025-07-10

### Fixed

- **Workspace Recognition Issue**: Fixed "No active builds" error by adding missing `workspaceFolders` parameter to LSP initialize request

## [v0.3.0] - 2025-07-10

### Added

- **Multi-Workspace Support**: Full support for multiple Go workspaces with comma-separated workspace paths and explicit workspace selection
- **Expanded MCP Tool Suite**: Increased from 3 to 14 comprehensive MCP tools covering workspace management, navigation, diagnostics, code assistance, and maintenance
- **Modular LSP Client Architecture**: Split LSP functionality into 8 focused, single-responsibility files for better maintainability
- **Comprehensive Testing**: Added complete MCP test coverage with mock-based unit tests and real gopls integration tests

### Changed

- **BREAKING**: All MCP tools now require `workspace` parameter for explicit workspace selection and routing
- **Multi-Client Architecture**: `mcpTools` struct now manages multiple `goplsClient` instances instead of single client
- **Major Architecture Refactor**: Restructured codebase for better separation of concerns and maintainability

## [v0.2.1] - 2025-07-05

### Changed

- **Docker Image Optimization**: Reduced Docker image size from 766MB to 121MB (84% reduction)

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

### Changed

- **BREAKING**: Replaced SSE transport with streamable HTTP transport for MCP specification compliance
- Updated connection endpoint from `http://localhost:8080/sse` to `http://localhost:8080`
- Updated client connection from `--transport sse` to `--transport http`

## [v0.1.0] - 2025-07-05

### Added

- Initial release of gopls-mcp MCP server
- MCP server implementation with SSE transport over HTTP
- gopls process management and LSP communication
- Three MCP tools: `go_to_definition`, `find_references`, `get_hover_info`
- HTTP server with graceful shutdown (port 8080)
- Command-line interface with required `-workspace` flag
- Docker support with multi-stage builds and security hardening
- Multi-platform Docker images (linux/amd64, linux/arm64)
- GitHub Actions CI/CD pipeline with quality gates
- Comprehensive test suite and golangci-lint configuration
- Docker Hub publishing to `megagrindstone/gopls-mcp`
