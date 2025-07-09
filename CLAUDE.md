# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project for implementing a Model Context Protocol (MCP) server for gopls (Go language server). The server supports both HTTP and stdio transports to provide Go language server capabilities to MCP clients like Claude.

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
- **Current State**: Fully functional MCP server with gopls integration, Docker support, and CI/CD pipeline
- **Purpose**: MCP server for gopls integration
- **Transport**: HTTP and stdio transports (MCP specification compliant)
- **Dependencies**: `github.com/modelcontextprotocol/go-sdk`
- **Deployment**: Docker Hub (`megagrindstone/gopls-mcp`) with multi-platform support
- **CI/CD**: GitHub Actions with comprehensive quality gates and automated Docker builds

### Key Components

- **main.go**: Multi-transport server setup with HTTP and stdio transport support and MCP server initialization
- **client.go**: Complete gopls LSP client with lifecycle management and core LSP operations
- **mcp.go**: MCP (Model Context Protocol) tools and handlers integrated with goplsClient
- **logger.go**: Structured logging initialization with slog and environment variable configuration
- **Dockerfile**: Multi-stage Docker build with gopls installation and security hardening
- **.dockerignore**: Docker context optimization for faster builds
- **.github/workflows/release.yaml**: Release-triggered CI/CD pipeline with quality gates and automated Docker publishing
- **.golangci.yaml**: Comprehensive linting configuration for code quality

## Testing

### Testing Strategy

The project uses Go's standard testing package with comprehensive tests:

- **client_integration_test.go**: Comprehensive integration tests for goplsClient (goToDefinition, findReferences, getHover)
- **main_test.go**: Tests application layer components like argument parsing and HTTP server setup

### Test Coverage

- **LSP Client Layer**: Complete gopls integration testing with real workspace scenarios
- **MCP Integration**: All MCP tools tested through the gopls client interface
- **Application Layer**: Command-line parsing, HTTP server setup, and component creation

### Running Tests

```bash
# Run all tests with verbose output and no caching (recommended)
go test ./... -v -count=1 -p 1

# Run tests with coverage
go test ./... -cover

# Run specific test file
go test -v client_integration_test.go
```

### Available MCP Tools

1. **go_to_definition**: Navigate to symbol definitions
2. **find_references**: Find all references to a symbol
3. **get_hover_info**: Get documentation and type information

### MCP Architecture (mcp.go)

The MCP layer is cleanly separated in `mcp.go` and follows Go best practices:

**Architecture Pattern:**

- **`mcpTools` struct**: Wraps `goplsClient` for MCP functionality
- **Value receivers**: All methods use `(m mcpTools)` following Go guidelines for small structs
- **Direct path handling**: Accepts workspace-relative paths directly for simplicity
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
# Start server with HTTP transport (default)
./gopls-mcp -workspace /path/to/go/project
./gopls-mcp -workspace /path/to/go/project -transport http

# Start server with stdio transport
./gopls-mcp -workspace /path/to/go/project -transport stdio

# Or build and run with go
go run . -workspace /path/to/go/project -transport http
go run . -workspace /path/to/go/project -transport stdio

# With logging configuration
LOG_LEVEL=DEBUG ./gopls-mcp -workspace /path/to/go/project
LOG_FORMAT=json LOG_LEVEL=WARN ./gopls-mcp -workspace /path/to/go/project
```

#### Docker

```bash
# Quick start with Docker Hub image
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# With custom workspace path
docker run -v /path/to/go/project:/custom/path -p 8080:8080 megagrindstone/gopls-mcp:latest -workspace /custom/path

# Local development with built image
docker build -t gopls-mcp .
docker run -v /path/to/go/project:/workspace -p 8080:8080 gopls-mcp

# With logging configuration
docker run -e LOG_LEVEL=DEBUG -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
docker run -e LOG_FORMAT=json -e LOG_LEVEL=INFO -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
```

The HTTP transport server will start on port 8080 at `http://localhost:8080`. The stdio transport communicates via standard input/output.

### MCP Tool Examples

#### Go to Definition

```json
{
  "name": "go_to_definition",
  "arguments": {
    "path": "main.go",
    "line": 10,
    "character": 5
  }
}
```

#### Find References

```json
{
  "name": "find_references",
  "arguments": {
    "path": "pkg/client.go",
    "line": 10,
    "character": 5,
    "includeDeclaration": true
  }
}
```

#### Get Hover Information

```json
{
  "name": "get_hover_info",
  "arguments": {
    "path": "mcp.go",
    "line": 10,
    "character": 5
  }
}
```

## Configuration

### Command-line Flags

- **-workspace**: Required command-line flag to set the Go workspace path
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
2. **gopls Integration**: Subprocess management with LSP communication
3. **MCP Tools**: Structured tools for Go language server features
4. **Graceful Shutdown**: Proper cleanup of gopls processes

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
