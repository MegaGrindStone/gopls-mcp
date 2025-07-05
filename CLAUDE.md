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

## Architecture Notes

- **Module**: `github.com/MegaGrindStone/gopls-mcp`
- **Go Version**: 1.24.4
- **Current State**: Fully functional MCP server with gopls integration
- **Purpose**: MCP server for gopls integration
- **Transport**: SSE (Server-Sent Events) over HTTP
- **Dependencies**: `github.com/modelcontextprotocol/go-sdk`

### Key Components

- **main.go**: HTTP server setup with SSE transport and MCP server initialization
- **manager.go**: gopls process management, LSP client, and MCP tool handlers
- **lsp.go**: LSP protocol types and gopls communication methods

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

```bash
# Start server with specific workspace path (required)
./gopls-mcp -workspace /path/to/go/project

# Or build and run with go
go run . -workspace /path/to/go/project
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

- Follow standard Go project structure as the codebase grows
- Implement proper error handling for MCP communication
- Consider adding configuration files for gopls integration settings
- Add appropriate logging for debugging MCP interactions
- Write tests for new functionality using the established patterns
- Run tests and linter before committing changes
