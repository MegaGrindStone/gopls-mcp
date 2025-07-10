# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project for implementing a Model Context Protocol (MCP) server for gopls (Go language server). The server supports both HTTP and stdio transports to provide Go language server capabilities to MCP clients like Claude with full multi-workspace support.

## 🚨 MANDATORY Go Coding Guidelines

**CRITICAL**: ALL Go coding patterns from `~/.claude/CLAUDE.md` are equally mandatory for this project - no exceptions.

### Key Reminder

This project requires strict adherence to ALL guidelines including:

- Early return patterns (reduce nesting)
- Use `any` instead of `interface{}`
- Start with unexported identifiers by default
- Prefer non-pointer structs when passing to/from functions  
- Follow Go naming conventions (avoid stuttering)
- AND ALL OTHER patterns in the global guidelines

**📖 Complete Guidelines**: See `~/.claude/CLAUDE.md` for full Go development guidelines.

## Development Commands

### Standard Go Commands

```bash
# Run the application
go run main.go

# Build the application
go build -o gopls-mcp

# Run tests
go test ./...

# Run tests with verbose output and no caching
go test ./... -v -count=1 -p 1

# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Tidy dependencies
go mod tidy

# Run linter (if golangci-lint is installed)
golangci-lint run ./...
```

### Docker Commands

```bash
# Build Docker image locally
docker build -t gopls-mcp .

# Run with Docker (mount your Go project)
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Run from Docker Hub
docker pull megagrindstone/gopls-mcp:latest
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Run with custom workspace path
docker run -v /path/to/go/project:/custom/path -p 8080:8080 megagrindstone/gopls-mcp:latest -workspace /custom/path
```

### GitHub CLI Commands

```bash
# Create a new release with tag
gh release create v1.0.0 --generate-notes

# Create a release with manual notes
gh release create v1.0.0 --notes "Release notes here"

# Upload release assets (binaries)
gh release upload v1.0.0 ./gopls-mcp

# List existing releases
gh release list

# View specific release
gh release view v1.0.0

# Delete a release
gh release delete v1.0.0
```

## Architecture Notes

- **Module**: `github.com/MegaGrindStone/gopls-mcp`
- **Go Version**: 1.24.4
- **Current State**: Enhanced MCP server with comprehensive gopls integration, 14 powerful tools, full multi-workspace support, Docker support, and CI/CD pipeline
- **Purpose**: Full-featured MCP server for gopls integration with AI assistants supporting multiple Go workspaces simultaneously
- **Transport**: HTTP and stdio transports (MCP specification compliant)
- **Dependencies**: `github.com/modelcontextprotocol/go-sdk`
- **Deployment**: Docker Hub (`megagrindstone/gopls-mcp`) with multi-platform support
- **CI/CD**: GitHub Actions with comprehensive quality gates and automated Docker builds
- **Tools**: 14 comprehensive gopls tools covering workspace management, navigation, diagnostics, code assistance, and maintenance

### Key Components

#### Core Application Files

- **main.go**: Multi-transport server setup with HTTP and stdio transport support and MCP server initialization
- **mcp.go**: MCP (Model Context Protocol) tools and handlers integrated with goplsClient
- **logger.go**: Structured logging initialization with slog and environment variable configuration

#### Modular LSP Client Architecture

The LSP client functionality has been refactored into 8 focused, single-responsibility files:

- **client.go** (526 lines): Core gopls client lifecycle management, LSP communication infrastructure, file management, and message routing
- **types.go** (160 lines): All LSP data types, constants, and structures (Position, Range, Location, Diagnostic, etc.)
- **parsing.go** (136 lines): Common parsing utilities for LSP responses (locations, ranges, hover content)
- **diagnostic.go** (117 lines): Diagnostic handling, publishDiagnostics notifications, and diagnostic retrieval
- **navigation.go** (277 lines): Navigation and reference tools (goToDefinition, findReferences, getHover, getTypeDefinition, findImplementations)
- **symbols.go** (207 lines): Symbol discovery tools (getDocumentSymbols, getWorkspaceSymbols, symbol parsing)
- **completion.go** (265 lines): Code assistance tools (getSignatureHelp, getCompletions, parameter information)
- **formatting.go** (317 lines): Code maintenance tools (formatDocument, organizeImports, getInlayHints, text edits)

#### Test Files

- **mcp_test.go** (687 lines): Comprehensive mock-based unit tests for MCP layer with interface abstraction and type conversion testing
- **mcp_integration_test.go** (421 lines): End-to-end integration tests for MCP tools with real gopls instances and multi-workspace testing
- **client_integration_test.go** (1622 lines): LSP client layer integration tests for gopls communication and workspace management
- **main_test.go** (345 lines): Application layer tests for command-line parsing, HTTP server setup, and component creation

#### Infrastructure Files

- **Dockerfile**: Multi-stage Docker build with gopls installation and security hardening
- **.dockerignore**: Docker context optimization for faster builds
- **.github/workflows/release.yaml**: Release-triggered CI/CD pipeline with quality gates and automated Docker publishing
- **.golangci.yaml**: Comprehensive linting configuration for code quality

#### Architectural Patterns

**Separation of Concerns**: Each file focuses on a specific domain of LSP functionality, making the codebase more maintainable and easier to navigate.

**Shared Client Instance**: All LSP functionality operates through methods on the central `goplsClient` struct, ensuring consistent state management and communication patterns.

**Domain-Driven Organization**: Files are organized by functional domains (navigation, diagnostics, completion, etc.) rather than technical layers, improving discoverability.

**Consistent Error Handling**: All LSP methods follow the same error handling patterns with proper context and early returns.

### Multi-Workspace Architecture

**Design Philosophy**: The system supports multiple workspaces through a clean separation of concerns:

- **Single-Workspace Client**: Each `goplsClient` manages exactly one workspace with its own gopls process
- **Multi-Client Orchestration**: The `mcpTools` struct manages a map of workspace paths to `goplsClient` instances
- **Explicit Workspace Selection**: All MCP tools require a `workspace` parameter for clear, unambiguous routing
- **Workspace Discovery**: The `list_workspaces` tool provides workspace enumeration for clients

**Benefits**:

- **Process Isolation**: Each workspace gets its own dedicated gopls process for reliability
- **Parallel Processing**: Multiple gopls instances can work simultaneously across workspaces
- **Clear Routing**: No path resolution ambiguity - workspace is explicitly specified
- **Zero Risk**: Existing single-workspace `goplsClient` code remains unchanged

## Testing

### Testing Strategy

The project uses Go's standard testing package with comprehensive multi-layer testing:

#### **MCP Layer Testing**

- **mcp_test.go**: Mock-based unit tests for MCP layer with interface abstraction (`goplsClientInterface`)
- **mcp_integration_test.go**: End-to-end integration tests with real gopls instances for all 14 MCP tools

#### **LSP Client Layer Testing**  

- **client_integration_test.go**: Comprehensive integration tests for goplsClient communication and workspace management

#### **Application Layer Testing**

- **main_test.go**: Command-line parsing, HTTP server setup, and component creation

### Test Architecture Patterns

#### **Mock Interface Pattern** (`mcp_test.go`)

- **Interface Abstraction**: `goplsClientInterface` defines testable contract for LSP operations
- **Test Isolation**: `testMCPTools` wrapper isolates MCP logic from LSP implementation details
- **Method Tracking**: Mock clients track method calls for verification
- **Error Injection**: Configurable error responses for comprehensive error handling tests

#### **Integration Test Patterns** (`mcp_integration_test.go`)

- **Real gopls Integration**: Tests with actual gopls processes for authentic behavior
- **Multi-Workspace Testing**: Validates workspace routing and client management
- **JSON Response Parsing**: Helper functions for MCP response validation
- **Common Test Fixtures**: Reusable workspace creation and cleanup utilities

### Test Coverage

- **MCP Layer**: Complete coverage of all 14 MCP tools with unit and integration tests
- **Multi-Workspace Support**: Full testing of workspace routing and client management  
- **Type Conversions**: Comprehensive testing of LSP↔MCP data structure conversions
- **Error Handling**: Extensive error scenario coverage and validation
- **LSP Client Layer**: Complete gopls integration testing with real workspace scenarios
- **Application Layer**: Command-line parsing, HTTP server setup, and component creation

### Running Tests

```bash
# Run all tests with verbose output and no caching (recommended)
go test ./... -v -count=1 -p 1

# Run tests with coverage
go test ./... -cover

# Run specific test layers
go test -run "Test.*MCP" ./...              # MCP layer tests only
go test -run "TestGoplsClient" ./...        # LSP client tests only  
go test -v mcp_test.go                      # MCP unit tests only
go test -v mcp_integration_test.go          # MCP integration tests only
```

### Available MCP Tools

#### Workspace Management Tools

1. **list_workspaces**: List all available Go workspaces configured in the server

#### Core Navigation Tools

2. **go_to_definition**: Navigate to symbol definitions
3. **find_references**: Find all references to a symbol
4. **get_hover_info**: Get documentation and type information

#### Diagnostic and Analysis Tools

5. **get_diagnostics**: Get compilation errors, warnings, and diagnostics for a Go file
6. **get_document_symbols**: Get outline of symbols (functions, types, etc.) defined in a Go file
7. **get_workspace_symbols**: Search for symbols across the entire Go workspace/project

#### Code Assistance Tools

8. **get_signature_help**: Get function signature help (parameter information) at the specified position
9. **get_completions**: Get code completion suggestions at the specified position

#### Advanced Navigation Tools

10. **get_type_definition**: Navigate to the type definition of a symbol at the specified position
11. **find_implementations**: Find all implementations of an interface or method at the specified position

#### Code Maintenance Tools

12. **format_document**: Format a Go source file according to gofmt standards
13. **organize_imports**: Organize and clean up import statements in a Go file
14. **get_inlay_hints**: Get inlay hints (implicit parameter names, type information) for a range in a Go file

### MCP Architecture (mcp.go)

The MCP layer is cleanly separated in `mcp.go` and follows Go best practices:

**Architecture Pattern:**

- **`mcpTools` struct**: Wraps multiple `goplsClient` instances for multi-workspace MCP functionality
- **Value receivers**: All methods use `(m mcpTools)` following Go guidelines for small structs
- **Explicit workspace routing**: All tools require workspace parameter for clear client selection
- **Client validation**: Built-in workspace validation and client routing via `getClient()` helper
- **Error handling**: Proper error propagation and context

**Key Components:**

- **Parameter types**: `GoToDefinitionParams`, `FindReferencesParams`, `GetHoverParams`
- **Result types**: `GoToDefinitionResult`, `FindReferencesResult`, `GetHoverResult`
- **Handler methods**: `HandleGoToDefinition`, `HandleFindReferences`, `HandleGetHover`
- **Tool creators**: `CreateGoToDefinitionTool`, `CreateFindReferencesTool`, `CreateGetHoverTool`
- **Setup function**: `setupMCPServer` for main.go integration

**Go Best Practices Followed:**

- **Non-pointer structs**: Returns `mcpTools` value, not `*mcpTools` pointer
- **Value receivers**: No unnecessary pointer passing for small structs
- **Early returns**: Proper error handling with early returns
- **Simple path handling**: Direct workspace-relative path processing
- **Consistent naming**: Follows Go naming conventions

## Usage Examples

### Starting the Server

#### Native Go

```bash
# Single workspace (traditional usage)
./gopls-mcp -workspace /path/to/go/project
./gopls-mcp -workspace /path/to/go/project -transport http

# Multiple workspaces (new multi-workspace support)
./gopls-mcp -workspace /project1,/project2,/project3
./gopls-mcp -workspace "/path/with spaces/project1,/project2"

# Start server with stdio transport
./gopls-mcp -workspace /path/to/go/project -transport stdio

# Or build and run with go
go run . -workspace /path/to/go/project -transport http
go run . -workspace /project1,/project2 -transport stdio

# With logging configuration
LOG_LEVEL=DEBUG ./gopls-mcp -workspace /project1,/project2
LOG_FORMAT=json LOG_LEVEL=WARN ./gopls-mcp -workspace /path/to/go/project
```

#### Docker

```bash
# Single workspace with Docker Hub image
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Multiple workspaces (mount each workspace separately)
docker run \
  -v /project1:/workspace1 \
  -v /project2:/workspace2 \
  -v /project3:/workspace3 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspace /workspace1,/workspace2,/workspace3

# With custom workspace paths
docker run -v /path/to/go/project:/custom/path -p 8080:8080 megagrindstone/gopls-mcp:latest -workspace /custom/path

# Local development with built image
docker build -t gopls-mcp .
docker run -v /path/to/go/project:/workspace -p 8080:8080 gopls-mcp

# Multi-workspace with logging configuration
docker run \
  -e LOG_LEVEL=DEBUG \
  -v /project1:/workspace1 \
  -v /project2:/workspace2 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspace /workspace1,/workspace2
```

The HTTP transport server will start on port 8080 at `http://localhost:8080`. The stdio transport communicates via standard input/output.

### MCP Tool Examples

#### Workspace Management Tools

##### List Workspaces

```json
{
  "name": "list_workspaces",
  "arguments": {}
}
```

#### Core Navigation Tools

##### Go to Definition

```json
{
  "name": "go_to_definition",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go",
    "line": 10,
    "character": 5
  }
}
```

##### Find References

```json
{
  "name": "find_references",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "pkg/client.go",
    "line": 10,
    "character": 5,
    "includeDeclaration": true
  }
}
```

##### Get Hover Information

```json
{
  "name": "get_hover_info",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "mcp.go",
    "line": 10,
    "character": 5
  }
}
```

#### Diagnostic and Analysis Tools

##### Get Diagnostics

```json
{
  "name": "get_diagnostics",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go"
  }
}
```

##### Get Document Symbols

```json
{
  "name": "get_document_symbols",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "client.go"
  }
}
```

##### Get Workspace Symbols

```json
{
  "name": "get_workspace_symbols",
  "arguments": {
    "workspace": "/path/to/workspace",
    "query": "Client"
  }
}
```

#### Code Assistance Tools

##### Get Signature Help

```json
{
  "name": "get_signature_help",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go",
    "line": 15,
    "character": 20
  }
}
```

##### Get Completions

```json
{
  "name": "get_completions",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go",
    "line": 8,
    "character": 5
  }
}
```

#### Advanced Navigation Tools

##### Get Type Definition

```json
{
  "name": "get_type_definition",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "client.go",
    "line": 25,
    "character": 10
  }
}
```

##### Find Implementations

```json
{
  "name": "find_implementations",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "interfaces.go",
    "line": 12,
    "character": 8
  }
}
```

#### Code Maintenance Tools

##### Format Document

```json
{
  "name": "format_document",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go"
  }
}
```

##### Organize Imports

```json
{
  "name": "organize_imports",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "client.go"
  }
}
```

##### Get Inlay Hints

```json
{
  "name": "get_inlay_hints",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go",
    "startLine": 10,
    "startChar": 0,
    "endLine": 20,
    "endChar": 50
  }
}
```

## Configuration

### Command-line Flags

- **-workspace**: Required command-line flag to set Go workspace path(s). Accepts single path or comma-separated list for multiple workspaces
- **-transport**: Transport type, accepts 'http' or 'stdio' (defaults to 'http')

### Environment Variables

- **LOG_LEVEL**: Set logging level (DEBUG, INFO, WARN, ERROR) - defaults to INFO
- **LOG_FORMAT**: Set log output format (text, json) - defaults to text

### Transport Details

- **HTTP Transport**: Port 8080 (streamable HTTP transport)
- **Stdio Transport**: Uses standard input/output for communication

## Docker Deployment

### Docker Hub

The project is automatically built and published to Docker Hub at `megagrindstone/gopls-mcp`.

**Available Tags:**

- `latest` - Latest stable release
- `v*` - Semantic version tags (e.g., `v0.1.0`)
- `*.*` - Version without 'v' prefix (e.g., `0.1.0`, `0.1`, `0`)
- `main` - Latest main branch build
- `main-<sha>` - Specific commit builds

**Multi-platform Support:**

- `linux/amd64` - Intel/AMD 64-bit
- `linux/arm64` - ARM 64-bit (Apple Silicon, ARM servers)

### Docker Image Details

- **Base Image**: `golang:1.24.4-alpine`
- **Runtime**: Multi-stage build for minimal size
- **User**: Non-root user (`gopls`) for security
- **Volumes**: `/workspace` for Go project mounting
- **Health Check**: HTTP endpoint monitoring
- **Dependencies**: gopls pre-installed

## LSP Communication Architecture

The system implements robust LSP communication with gopls using the following components:

### Message Handling

- **messageReader()**: Continuous goroutine that reads all LSP messages from gopls stdout
- **sendRequestAndWait()**: Sends LSP requests and waits for correlated responses with 60-second timeout  
- **routeResponse()**: Routes responses to waiting request handlers by request ID
- **readLSPMessage()**: Reads complete LSP messages with proper header parsing (`\r\n` line endings)
- **handleLSPMessage()**: Routes messages by type (response/request/notification)

### File Management

The system automatically manages file context for gopls:

- **ensureFileOpen()**: Opens files in gopls via `textDocument/didOpen` before making requests
- **File Tracking**: Maintains registry of open files to avoid duplicate notifications
- **Content Reading**: Reads file content from disk with automatic language detection
- **Language Detection**: Supports Go files (`.go`), `go.mod`, and `go.sum` files

### Workspace Readiness

Ensures reliable operation with large codebases:

- **Readiness Tracking**: Monitors `window/showMessage` and `$/progress` notifications from gopls
- **Initialization Blocking**: Waits for "Finished loading packages" before allowing LSP requests
- **Timeout Management**: 60-second timeout for LSP operations with progress logging every 10 seconds
- **Error Handling**: Comprehensive error handling for LSP communication failures

## CI/CD Pipeline

### GitHub Actions Workflow

Automated release pipeline (`.github/workflows/release.yaml`) that:

1. **Quality Gates** (runs on GitHub releases):
   - Go tests: `go test ./... -v -count=1 -p 1`
   - Code formatting: `gofmt` validation
   - Code quality: `golangci-lint` with comprehensive rules
   - Dependencies: `go mod verify`

2. **Docker Build & Push** (release only):
   - Multi-platform builds (amd64/arm64)
   - Automated push to Docker Hub
   - Semantic versioning tags
   - Build caching for performance

### Required GitHub Secrets

- `DOCKERHUB_USERNAME`: Docker Hub username
- `DOCKERHUB_TOKEN`: Docker Hub access token

### Development Workflow

1. **Local Development**: Standard Go commands + local Docker testing
2. **Pull Request**: Manual testing and code review
3. **Release Creation**: Triggers full pipeline + Docker Hub push
4. **Version Tags**: Automated creation of versioned Docker images (`v1.0.0`, etc.)

## MCP Development Context

This project implements a Model Context Protocol server that interfaces with gopls using:

1. **Multi-Transport Support**: HTTP and stdio transports for MCP specification compliance
2. **Multi-Workspace Support**: Full support for multiple Go workspaces with explicit workspace selection
3. **gopls Integration**: Subprocess management with LSP communication (one gopls process per workspace)
4. **MCP Tools**: 14 structured tools for Go language server features including workspace management
5. **Graceful Shutdown**: Proper cleanup of all gopls processes across workspaces

## Development Guidelines

### Documentation Strategy

- **CLAUDE.md**: Document the current state of the application (architecture, commands, usage)
- **CHANGELOG.md**: Record all changes chronologically under "Unreleased" section during development
- **Avoid Historical Sections**: Do not add dated fix descriptions to CLAUDE.md (e.g., "LSP Communication Fix (2025-07-05)")
- **When Releasing**: Move "Unreleased" changes in CHANGELOG.md to the new version section

### Code Quality

- Follow standard Go project structure as the codebase grows
- Implement proper error handling for MCP communication
- Consider adding configuration files for gopls integration settings
- Add appropriate logging for debugging MCP interactions
- Write tests for new functionality using the established patterns
- Run tests and linter before committing changes

### Testing Guidelines

**Follow Established Testing Patterns**:

- **MCP Layer**: Use mock interface pattern (`goplsClientInterface`) for unit tests and real gopls integration for end-to-end tests
- **Interface Abstraction**: Create testable interfaces for external dependencies (gopls client)
- **Test Isolation**: Use wrapper structs (`testMCPTools`) to isolate logic under test
- **Error Scenarios**: Include comprehensive error injection and validation tests
- **JSON Validation**: Test MCP response parsing and structure validation
- **Multi-Workspace**: Include workspace routing tests for new MCP functionality

**Test Organization**:

- **Unit Tests**: Fast, isolated tests with mocks for individual components
- **Integration Tests**: Real gopls integration for authentic behavior validation
- **Common Utilities**: Extract reusable test helpers (workspace creation, parsing, etc.)
- **Clear Naming**: Use descriptive test names that explain the scenario being tested

### Modular Architecture Guidelines

**Maintain Separation of Concerns**: When adding new LSP functionality, place it in the appropriate domain-specific file:

- **Navigation tools** → `navigation.go`
- **Diagnostic features** → `diagnostic.go`
- **Symbol operations** → `symbols.go`
- **Code assistance** → `completion.go`
- **Formatting/maintenance** → `formatting.go`
- **Core types** → `types.go`
- **Common parsing** → `parsing.go`
- **Infrastructure** → `client.go`

**File Size Guidelines**: Keep individual files focused and reasonably sized (100-400 lines). If a file grows beyond ~500 lines, consider further subdivision by functionality.

**Cross-File Dependencies**: All LSP functionality should operate through the central `goplsClient` struct. Avoid tight coupling between domain-specific files.

### Docker Development

- Test Docker builds locally before pushing: `docker build -t gopls-mcp .`
- Verify container functionality: `docker run -v /path/to/test/project:/workspace -p 8080:8080 gopls-mcp`
- Keep Dockerfile optimized for size and security
- Update .dockerignore when adding new file types

### CI/CD Best Practices

- All commits to main branch trigger Docker builds
- Use semantic versioning for releases: `git tag v1.0.0`
- Pull requests automatically run quality gates
- Monitor GitHub Actions for build failures
- Docker images are automatically pushed to `megagrindstone/gopls-mcp`
- Ensure golangci-lint config (.golangci.yaml) compatibility with CI version

### Development Workflow

1. **Local Development**: Use standard Go commands for fast iteration
2. **Testing**: Run `go test ./... -v -count=1 -p 1` and `golangci-lint run ./...`
3. **Docker Testing**: Build and test Docker image locally
4. **Pull Request**: Creates quality gate checks
5. **Merge to Main**: Triggers full CI/CD pipeline and Docker Hub push
6. **Version Release**: Tag with `v*` for versioned Docker images

### Release Process

When ready to create a release:

1. **Prepare Release**:
   - Update CHANGELOG.md with new version and release notes
   - Ensure all tests pass: `go test ./... -v -count=1 -p 1`
   - Run linter: `golangci-lint run ./...`
   - Build binary: `go build -o gopls-mcp`
   - Test Docker build: `docker build -t gopls-mcp .`

2. **Commit Changelog**:
   - Commit CHANGELOG.md changes: `git add CHANGELOG.md && git commit -m "Update CHANGELOG.md for v1.0.0"`
   - Push to main branch: `git push origin main`

3. **Create Release**:
   - Use semantic versioning (e.g., `v1.0.0`)
   - Create release with GitHub CLI: `gh release create v1.0.0 --generate-notes`
   - Upload binary: `gh release upload v1.0.0 ./gopls-mcp`

4. **Automated CI/CD**:
   - GitHub Actions automatically triggers on release publication
   - Quality gates run (tests, linting, formatting)
   - Docker images built and pushed to Docker Hub
   - Docker Hub will have new tags: `v1.0.0`, `latest`

5. **Monitor Release**:
   - Check GitHub Actions: `gh run list --limit 5`
   - View specific run: `gh run view <run-id>`
   - Check Docker Hub tags: `docker pull <image> --all-tags`
   - View release status: `gh release view <tag>`

### Troubleshooting Releases

**Failed GitHub Actions Build**:

```bash
# Check recent runs
gh run list --limit 5

# View failed run details
gh run view <run-id> --log-failed

# Common fix: Docker tag issues in .github/workflows/release.yaml
# Ensure proper semantic versioning patterns in tagging strategy
```

**Re-tagging Failed Release**:

```bash
# Delete tag locally and remotely
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Fix issues, commit, push

# Re-create tag
git tag v1.0.0
git push origin v1.0.0
```

**Docker Image Cleanup**:

```bash
# Remove all local images for cleanup
docker rmi $(docker images <image-name> -q)

# View available tags
docker images <image-name> --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}"
```
