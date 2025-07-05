# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project for implementing a Model Context Protocol (MCP) server for gopls (Go language server). The server uses SSE (Server-Sent Events) transport to provide Go language server capabilities to MCP clients like Claude.

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
- **Transport**: SSE (Server-Sent Events) over HTTP
- **Dependencies**: `github.com/modelcontextprotocol/go-sdk`
- **Deployment**: Docker Hub (`megagrindstone/gopls-mcp`) with multi-platform support
- **CI/CD**: GitHub Actions with comprehensive quality gates and automated Docker builds

### Key Components

- **main.go**: HTTP server setup with SSE transport and MCP server initialization
- **manager.go**: gopls process management, LSP client, and MCP tool handlers
- **lsp.go**: LSP protocol types and gopls communication methods
- **Dockerfile**: Multi-stage Docker build with gopls installation and security hardening
- **.dockerignore**: Docker context optimization for faster builds
- **.github/workflows/docker.yml**: CI/CD pipeline with quality gates and automated Docker publishing
- **.golangci.yaml**: Comprehensive linting configuration for code quality

### Test Files

- **lsp_test.go**: Tests LSP protocol parsing functions, response handling, and error cases
- **manager_test.go**: Tests Manager lifecycle, thread safety, MCP tool handlers, and JSON marshaling
- **main_test.go**: Tests application layer components like argument parsing and HTTP server setup

### Available MCP Tools

1. **go_to_definition**: Navigate to symbol definitions
2. **find_references**: Find all references to a symbol
3. **get_hover_info**: Get documentation and type information

## Usage Examples

### Starting the Server

#### Native Go

```bash
# Start server with specific workspace path (required)
./gopls-mcp -workspace /path/to/go/project

# Or build and run with go
go run . -workspace /path/to/go/project
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
```

The server will start on port 8080 with SSE endpoint at `http://localhost:8080/sse`.

### MCP Tool Examples

#### Go to Definition

```json
{
  "name": "go_to_definition",
  "arguments": {
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
    "uri": "file:///path/to/file.go",
    "line": 10,
    "character": 5
  }
}
```

## Configuration

- **-workspace**: Required command-line flag to set the Go workspace path
- **Default Port**: 8080 (hardcoded)
- **SSE Endpoint**: `/sse`

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

## CI/CD Pipeline

### GitHub Actions Workflow

Automated pipeline (`.github/workflows/docker.yml`) that:

1. **Quality Gates** (runs on all pushes/PRs):
   - Go tests: `go test ./... -v -count=1 -p 1`
   - Code formatting: `gofmt` validation
   - Code quality: `golangci-lint` with comprehensive rules
   - Dependencies: `go mod verify`

2. **Docker Build & Push** (main branch and tags only):
   - Multi-platform builds (amd64/arm64)
   - Automated push to Docker Hub
   - Smart tagging strategy
   - Build caching for performance

### Required GitHub Secrets

- `DOCKERHUB_USERNAME`: Docker Hub username
- `DOCKERHUB_TOKEN`: Docker Hub access token

### Development Workflow

1. **Local Development**: Standard Go commands + local Docker testing
2. **Pull Request**: Triggers quality gates (tests, linting)
3. **Main Branch**: Triggers full pipeline + Docker Hub push
4. **Version Tags**: Creates versioned Docker images (`v1.0.0`, etc.)

## MCP Development Context

This project implements a Model Context Protocol server that interfaces with gopls using:

1. **SSE Transport**: HTTP-based communication suitable for web clients
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
- **26 Total Tests**: All major functions and methods tested with success/error paths

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

2. **Create Release**:
   - Use semantic versioning (e.g., `v1.0.0`)
   - Create release with GitHub CLI: `gh release create v1.0.0 --generate-notes`
   - Upload binary: `gh release upload v1.0.0 ./gopls-mcp`

3. **Post-Release**:
   - GitHub Actions will automatically build and push Docker images
   - Docker Hub will have new tags: `v1.0.0`, `latest`
   - Release notes will be auto-generated from commit history

4. **Monitor Release**:
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

# Common fix: Docker tag issues in .github/workflows/docker.yml
# Remove problematic tag patterns like "type=sha,prefix={{branch}}-"
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
