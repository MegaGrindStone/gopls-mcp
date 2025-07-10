# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **Unnecessary Type Arguments**: Removed redundant type arguments from MCP tools and test functions to follow Go's type inference best practices
  - **Root Cause**: Explicit type arguments in `mcp.NewServerTool[ParamType, ResultType]` and `parseJSONResult[ResultType]` calls were unnecessary as Go can infer types from function signatures
  - **Solution**: Removed type arguments from 14 `mcp.NewServerTool` calls in `mcp.go` and 6 `parseJSONResult` calls in `mcp_integration_test.go`
  - **Impact**: Cleaner, more idiomatic Go code following type inference best practices; resolved all 20 "unnecessary type arguments" diagnostics warnings
  - **Files Modified**: `mcp.go` (14 fixes), `mcp_integration_test.go` (6 fixes)
  - **Testing**: All tests pass, linter reports 0 issues, functionality remains intact
- **Diagnostic Timing Bug**: Fixed inconsistent diagnostic results on first call to `get_diagnostics` by preventing premature acceptance of empty diagnostics
  - **Root Cause**: The stability mechanism was accepting empty diagnostics as "stable" after 200ms, but gopls takes ~2 seconds to complete analysis
  - **Solution**: Added `minWaitForNonEmpty = 3 * time.Second` constant to prevent accepting empty diagnostics within first 3 seconds
  - **Impact**: All `get_diagnostics` calls now consistently return complete diagnostics (14 for `mcp.go`)
  - **Files Modified**: `diagnostic.go:109` (constant), `diagnostic.go:135-138` (early rejection logic)
  - **Testing**: Verified with multiple consecutive calls returning 14 diagnostics consistently instead of 0 on first call
- **Docker "No active builds" Error**: Fixed gopls module analysis failure in Docker containers by including full Go toolchain
  - **Root Cause**: Docker runtime stage was missing Go toolchain required for gopls internal `go list` operations
  - **Solution**: Simplified Dockerfile to single-stage build using `golang:1.24.4-alpine` base image
  - **Impact**: Enables full gopls functionality in Docker environments with proper workspace module analysis
  - **Architecture**: Eliminated redundant multi-stage build pattern for cleaner, more efficient Docker builds
  - **Testing**: Verified with MCP tools returning actual Go diagnostics instead of "No active builds" errors

## [v0.3.1] - 2025-07-10

### Fixed

- **Workspace Recognition Issue**: Fixed "No active builds contain /workspace/*/main.go" error by adding missing `workspaceFolders` parameter to LSP initialize request
  - **Root Cause**: LSP initialize request declared `"workspaceFolders": true` capability but didn't provide actual workspace folders for gopls to track
  - **Solution**: Added `workspaceFolders` parameter with proper workspace URI and name mapping to LSP initialize request in `client.go:173-178`
  - **Impact**: Resolves workspace recognition issues in both local and containerized environments (Docker/Kubernetes)
  - **Testing**: All 70 tests pass without workspace recognition errors, confirmed working across single and multi-workspace configurations

## [v0.3.0] - 2025-07-10

### Added

- **Multi-Workspace Support**: Full support for multiple Go workspaces with explicit workspace selection
  - **Comma-Separated Workspace Paths**: Accept multiple workspaces via `-workspace /project1,/project2,/project3`
  - **Workspace Management Tool**: New `list_workspaces` tool to discover available workspaces
  - **Explicit Workspace Routing**: All tools now require `workspace` parameter for clear client selection
  - **Process Isolation**: Each workspace gets its own dedicated gopls process for reliability
  - **Parallel Processing**: Multiple gopls instances can work simultaneously across workspaces
  - **Docker Multi-Workspace**: Support for mounting and managing multiple workspaces in containers
- **Expanded MCP Tool Suite**: Increased from 3 to 14 comprehensive MCP tools across 6 categories:
  - **Workspace Management Tools** (1): `list_workspaces`
  - **Core Navigation Tools** (3): `go_to_definition`, `find_references`, `get_hover_info`
  - **Diagnostic and Analysis Tools** (3): `get_diagnostics`, `get_document_symbols`, `get_workspace_symbols`
  - **Code Assistance Tools** (2): `get_signature_help`, `get_completions`
  - **Advanced Navigation Tools** (2): `get_type_definition`, `find_implementations`
  - **Code Maintenance Tools** (3): `format_document`, `organize_imports`, `get_inlay_hints`
- **Modular LSP Client Architecture**: Split LSP functionality into 8 focused, single-responsibility files:
  - **client.go** (526 lines): Core gopls client lifecycle management and LSP communication infrastructure
  - **types.go** (160 lines): All LSP data types, constants, and structures
  - **parsing.go** (136 lines): Common parsing utilities for LSP responses
  - **diagnostic.go** (117 lines): Diagnostic handling and publishDiagnostics notifications
  - **navigation.go** (277 lines): Navigation and reference tools
  - **symbols.go** (207 lines): Symbol discovery tools
  - **completion.go** (265 lines): Code assistance tools
  - **formatting.go** (317 lines): Code maintenance tools
- **Comprehensive Integration Testing**: Enhanced testing with 1,621 lines of real-world validation scenarios
- **Complete MCP Test Coverage**: Added comprehensive testing infrastructure for MCP layer with 1,108 lines of new test code:
  - **mcp_test.go** (687 lines): Mock-based unit tests with interface abstraction (`goplsClientInterface`) for all 14 MCP tools
  - **mcp_integration_test.go** (421 lines): End-to-end integration tests with real gopls instances and multi-workspace validation
  - **Mock Interface Pattern**: Testable contracts with `testMCPTools` wrapper for test isolation and error injection
  - **Integration Test Patterns**: Real gopls integration, JSON response validation, and reusable test fixtures
  - **Multi-Layer Testing Strategy**: MCP Layer → LSP Client Layer → Application Layer testing architecture
  - **Type Conversion Testing**: Comprehensive validation of LSP↔MCP data structure conversions
  - **Error Scenario Coverage**: Extensive error injection and validation testing for robustness

### Changed

- **BREAKING**: All MCP tools now require `workspace` parameter for explicit workspace selection and routing
- **Multi-Client Architecture**: `mcpTools` struct now manages multiple `goplsClient` instances instead of single client
  - **Workspace Validation**: Added `getClient()` helper for workspace validation and client routing
  - **Client Orchestration**: Clean separation between single-workspace clients and multi-client management
  - **Command-Line Parsing**: Enhanced argument parsing with `parseAndValidateWorkspaces()` helper function
- **Major Architecture Refactor**: Restructured codebase for better separation of concerns and maintainability
  - **Improved Separation**: Domain-driven organization by functional areas (navigation, diagnostics, completion, etc.)
  - **Enhanced Testing**: Replaced unit tests with comprehensive integration tests for real-world validation
  - **Cleaner API**: MCP tools now use workspace-relative paths instead of URIs for simplified usage
  - **Better Documentation**: Updated `CLAUDE.md` with comprehensive multi-workspace architecture and usage examples
  - **Shared Client Instance**: All LSP functionality operates through methods on the central `goplsClient` struct
  - **Consistent Error Handling**: All LSP methods follow the same error handling patterns with proper context

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
