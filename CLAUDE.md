# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project for implementing a Model Context Protocol (MCP) server for gopls (Go language server). The server supports both HTTP and stdio transports to provide Go language server capabilities to MCP clients like Claude.

## Development Commands

### Standard Go Commands

```bash
# Run the application with multiple workspaces
go run main.go -workspaces /path/workspace1,/path/workspace2

# Run the application with single workspace
go run main.go -workspaces /path/workspace

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

# Run with Docker (single workspace)
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Run with multiple workspaces (mount multiple volumes)
docker run \
  -v /path/to/workspace1:/workspace1 \
  -v /path/to/workspace2:/workspace2 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspaces /workspace1,/workspace2

# Run from Docker Hub (single workspace)
docker pull megagrindstone/gopls-mcp:latest
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Run with custom workspace paths
docker run \
  -v /path/to/project1:/custom/path1 \
  -v /path/to/project2:/custom/path2 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspaces /custom/path1,/custom/path2
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
- **Current State**: Fully functional multi-workspace MCP server with gopls integration, Docker support, and CI/CD pipeline
- **Purpose**: Multi-workspace MCP server for gopls integration
- **Transport**: HTTP and stdio transports (MCP specification compliant)
- **Dependencies**: `github.com/modelcontextprotocol/go-sdk`
- **Deployment**: Docker Hub (`megagrindstone/gopls-mcp`) with multi-platform support
- **CI/CD**: GitHub Actions with comprehensive quality gates and automated Docker builds
- **Multi-Workspace**: Supports multiple Go workspaces simultaneously with dedicated gopls processes

### Key Components

- **main.go**: Multi-transport server setup with HTTP and stdio transport support and MCP server initialization
- **manager.go**: WorkspaceManager for multi-workspace coordination, gopls process management, LSP client, and MCP tool handlers
- **lsp.go**: LSP protocol types and gopls communication methods
- **logger.go**: Structured logging initialization with slog and environment variable configuration
- **test_helpers.go**: Testing utilities including reduced-verbosity logger for tests
- **Dockerfile**: Multi-stage Docker build with gopls installation and security hardening
- **.dockerignore**: Docker context optimization for faster builds
- **.github/workflows/release.yaml**: Release-triggered CI/CD pipeline with quality gates and automated Docker publishing
- **.golangci.yaml**: Comprehensive linting configuration for code quality

### Test Files

- **lsp_test.go**: Tests LSP protocol parsing functions, response handling, and error cases
- **manager_test.go**: Tests Manager lifecycle, thread safety, MCP tool handlers, and JSON marshaling
- **main_test.go**: Tests application layer components like argument parsing and HTTP server setup

### Available MCP Tools

1. **go_to_definition**: Navigate to symbol definitions
2. **find_references**: Find all references to a symbol
3. **get_hover_info**: Get documentation and type information
4. **list_workspaces**: List all available workspaces and their status

## Usage Examples

### Starting the Server

#### Native Go

```bash
# Start server with single workspace (HTTP transport - default)
./gopls-mcp -workspaces /path/to/go/project
./gopls-mcp -workspaces /path/to/go/project -transport http

# Start server with multiple workspaces
./gopls-mcp -workspaces /path/to/workspace1,/path/to/workspace2
./gopls-mcp -workspaces /path/to/workspace1,/path/to/workspace2 -transport http

# Start server with stdio transport
./gopls-mcp -workspaces /path/to/go/project -transport stdio

# Or build and run with go
go run . -workspaces /path/to/go/project -transport http
go run . -workspaces /path/to/workspace1,/path/to/workspace2 -transport stdio

# With logging configuration
LOG_LEVEL=DEBUG ./gopls-mcp -workspaces /path/to/go/project
LOG_FORMAT=json LOG_LEVEL=WARN ./gopls-mcp -workspaces /path/to/workspace1,/path/to/workspace2
```

#### Docker

```bash
# Quick start with Docker Hub image (single workspace)
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Multiple workspaces with Docker
docker run \
  -v /path/to/workspace1:/workspace1 \
  -v /path/to/workspace2:/workspace2 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspaces /workspace1,/workspace2

# With custom workspace paths
docker run \
  -v /path/to/project1:/custom/path1 \
  -v /path/to/project2:/custom/path2 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspaces /custom/path1,/custom/path2

# Local development with built image
docker build -t gopls-mcp .
docker run -v /path/to/go/project:/workspace -p 8080:8080 gopls-mcp

# With logging configuration (single workspace)
docker run -e LOG_LEVEL=DEBUG -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# With logging configuration (multiple workspaces)
docker run \
  -e LOG_FORMAT=json -e LOG_LEVEL=INFO \
  -v /path/to/workspace1:/workspace1 \
  -v /path/to/workspace2:/workspace2 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspaces /workspace1,/workspace2
```

The HTTP transport server will start on port 8080 at `http://localhost:8080`. The stdio transport communicates via standard input/output.

### MCP Tool Examples

#### Go to Definition

```json
{
  "name": "go_to_definition",
  "arguments": {
    "workspace": "/path/to/workspace",
    "uri": "file:///path/to/file.go",
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
    "workspace": "/path/to/workspace",
    "uri": "file:///path/to/file.go",
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
    "workspace": "/path/to/workspace",
    "uri": "file:///path/to/file.go",
    "line": 10,
    "character": 5
  }
}
```

#### List Workspaces

```json
{
  "name": "list_workspaces",
  "arguments": {}
}
```

## Configuration

### Command-line Flags

- **-workspaces**: Required command-line flag to set Go workspace paths (comma-separated list)
- **-transport**: Transport type, accepts 'http' or 'stdio' (defaults to 'http')

### Environment Variables

- **LOG_LEVEL**: Set logging level (DEBUG, INFO, WARN, ERROR) - defaults to INFO
- **LOG_FORMAT**: Set log output format (text, json) - defaults to text

### Transport Details

- **HTTP Transport**: Port 8080 (streamable HTTP transport)
- **Stdio Transport**: Uses standard input/output for communication

## Multi-Workspace Architecture

The gopls-mcp server supports multiple Go workspaces simultaneously through the WorkspaceManager component:

### Key Features

- **Multiple Workspaces**: Support for multiple Go projects in a single server instance
- **Isolated gopls Processes**: Each workspace gets its own dedicated gopls process
- **Workspace Routing**: MCP tools route requests to the appropriate workspace based on the workspace parameter
- **Centralized Management**: Single HTTP/stdio endpoint manages all workspaces
- **Workspace Discovery**: Use the `list_workspaces` tool to see all available workspaces and their status

### WorkspaceManager Component

The `WorkspaceManager` component coordinates multiple `Manager` instances:

- **Workspace Creation**: Automatically creates a `Manager` for each specified workspace path
- **Lifecycle Management**: Starts/stops all workspace gopls processes together
- **Request Routing**: Routes MCP tool calls to the correct workspace Manager
- **Status Monitoring**: Tracks the running status of all workspace gopls processes
- **Error Handling**: Provides clear error messages for invalid workspace requests

### Usage Benefits

- **Reduced Docker Instances**: Handle multiple Go projects without deploying multiple containers
- **Resource Efficiency**: Shared HTTP server and MCP infrastructure across workspaces
- **Simplified Management**: Single endpoint for multiple Go projects
- **Workspace Isolation**: Each workspace maintains independent gopls state

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

### GitHub Actions Workflows

**Pull Request CI** (`.github/workflows/pr.yaml`):

1. **Quality Gates** (runs on PRs and pushes to main):
   - Go tests: `go test ./... -v -count=1 -p 1`
   - Code formatting: `gofmt` validation
   - Code quality: `golangci-lint` with comprehensive rules
   - Dependencies: `go mod verify`

**Release Pipeline** (`.github/workflows/release.yaml`):

1. **Quality Gates** (runs on GitHub releases):
   - Same quality gates as PR workflow
   - Ensures release readiness

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
2. **Pull Request**: Automated quality gates (tests, linting, formatting) + code review
3. **Release Creation**: Triggers full pipeline + Docker Hub push
4. **Version Tags**: Automated creation of versioned Docker images (`v1.0.0`, etc.)

## MCP Development Context

This project implements a Model Context Protocol server that interfaces with gopls using:

1. **Multi-Transport Support**: HTTP and stdio transports for MCP specification compliance
2. **gopls Integration**: Subprocess management with LSP communication
3. **MCP Tools**: Structured tools for Go language server features
4. **Graceful Shutdown**: Proper cleanup of gopls processes

## Testing

### Testing Strategy

The project uses Go's standard testing package with comprehensive unit tests covering all major functionality:

- **Unit Tests Only**: No integration tests to avoid external dependencies
- **Table-Driven Tests**: Multiple test cases with edge cases and error conditions
- **Thread Safety**: Concurrent testing for Manager request ID generation
- **Error Handling**: Comprehensive testing of error conditions and invalid inputs
- **JSON Marshaling**: Validation of MCP parameter and result serialization

### Test Coverage

- **LSP Protocol Layer**: Response parsing, type conversion, and error handling
- **Manager Component**: Lifecycle management, tool handlers, and thread safety
- **Application Layer**: Command-line parsing, HTTP server setup, and component creation
- **31 Total Tests**: All major functions and methods tested with success/error paths including transport functionality

### Running Tests

```bash
# Run all tests with verbose output and no caching (recommended)
go test ./... -v -count=1 -p 1

# Run tests with coverage
go test ./... -cover

# Run specific test file
go test -v lsp_test.go
```

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
