# gopls-mcp

[![GitHub release](https://img.shields.io/github/release/MegaGrindStone/gopls-mcp.svg)](https://github.com/MegaGrindStone/gopls-mcp/releases)
[![Docker Hub](https://img.shields.io/docker/pulls/megagrindstone/gopls-mcp)](https://hub.docker.com/r/megagrindstone/gopls-mcp)
[![CI](https://github.com/MegaGrindStone/gopls-mcp/actions/workflows/release.yaml/badge.svg)](https://github.com/MegaGrindStone/gopls-mcp/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Model Context Protocol (MCP) server that integrates [gopls](https://pkg.go.dev/golang.org/x/tools/gopls) (Go language server) with MCP-compatible hosts like Claude Code, Claude Desktop, and VS Code.

I've developed this project in collaboration with [Claude Code](https://claude.ai/code) to provide seamless Go language server capabilities through the MCP protocol, enabling powerful Go development assistance in AI-powered coding environments.

## Important Note

This project uses the [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk) as its core MCP library. Since the go-sdk is currently in early development (v0.1.0) and not yet stable, this project will not have stable releases until the go-sdk reaches stability. Please consider this when using gopls-mcp in production environments.

## Features

This MCP server provides **14 comprehensive Go development tools** organized across 6 categories, with full **multi-workspace support**:

### 🏢 Workspace Management Tools (1)

- **📋 List Workspaces** - Discover and enumerate all configured Go workspaces

### 🎯 Core Navigation Tools (3)

- **🎯 Go to Definition** - Navigate to symbol definitions across your Go workspace
- **🔍 Find References** - Locate all references to functions, variables, and types
- **📖 Hover Information** - Get documentation, type information, and signatures

### 🔍 Diagnostic and Analysis Tools (3)

- **🚨 Get Diagnostics** - Get compilation errors, warnings, and diagnostics for Go files
- **📄 Document Symbols** - Get outline of symbols (functions, types, etc.) defined in Go files
- **🔎 Workspace Symbols** - Search for symbols across the entire Go workspace/project

### 💡 Code Assistance Tools (2)

- **✍️ Signature Help** - Get function signature help and parameter information
- **🤖 Code Completions** - Get intelligent code completion suggestions

### 🧭 Advanced Navigation Tools (2)

- **🏷️ Type Definition** - Navigate to the type definition of symbols
- **🔗 Find Implementations** - Find all implementations of interfaces or methods

### 🛠️ Code Maintenance Tools (3)

- **✨ Format Document** - Format Go source files according to gofmt standards
- **📦 Organize Imports** - Organize and clean up import statements
- **💭 Inlay Hints** - Get inlay hints for implicit parameter names and type information

All tools work with your existing Go workspaces, support **multiple workspaces simultaneously**, and leverage gopls for accurate, fast results.

## Installation

### Claude Code Integration (Recommended)

The easiest way to use gopls-mcp is with Claude Code using Streamable HTTP transport:

1. **Start the server** with your Go workspace(s):

   ```bash
   # Single workspace using Docker (recommended)
   docker run -d -v /path/to/your/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
   
   # Multiple workspaces using Docker
   docker run -d \
     -v /path/to/project1:/workspace1 \
     -v /path/to/project2:/workspace2 \
     -v /path/to/project3:/workspace3 \
     -p 8080:8080 \
     megagrindstone/gopls-mcp:latest -workspace /workspace1,/workspace2,/workspace3
   
   # Or build from source (single workspace)
   go build -o gopls-mcp
   ./gopls-mcp -workspace /path/to/your/go/project
   
   # Multiple workspaces from source
   ./gopls-mcp -workspace /path/to/project1,/path/to/project2,/path/to/project3
   ```

2. **Add to Claude Code**:

   ```bash
   claude mcp add --transport http gopls-mcp http://localhost:8080
   ```

3. **Start using** - The Go tools will be available in your Claude Code conversations!

### Claude Desktop Integration

For Claude Desktop, add this to your MCP settings:

#### Single Workspace

```json
{
  "mcpServers": {
    "gopls-mcp": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-v", "/path/to/your/go/project:/workspace",
        "-p", "8080:8080",
        "megagrindstone/gopls-mcp:latest"
      ]
    }
  }
}
```

#### Multiple Workspaces

```json
{
  "mcpServers": {
    "gopls-mcp": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-v", "/path/to/project1:/workspace1",
        "-v", "/path/to/project2:/workspace2",
        "-v", "/path/to/project3:/workspace3",
        "-p", "8080:8080",
        "megagrindstone/gopls-mcp:latest",
        "-workspace", "/workspace1,/workspace2,/workspace3"
      ]
    }
  }
}
```

### Other MCP Hosts

For VS Code and other MCP hosts, use the Streamable HTTP transport:

```json
{
  "servers": {
    "gopls-mcp": {
      "type": "http",
      "url": "http://localhost:8080"
    }
  }
}
```

## Quick Start

### Prerequisites

- Go workspace with valid `go.mod` file
- Docker (recommended) or Go 1.24+ installed

### Step-by-Step Setup

1. **Clone or download** a Go project to work with
2. **Start gopls-mcp** server:

   ```bash
   docker run -d -v $(pwd):/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
   ```

3. **Integrate with Claude Code**:

   ```bash
   claude mcp add --transport http gopls-mcp http://localhost:8080
   ```

4. **Test the integration** - Ask Claude Code to help with your Go code!

## Usage Examples

Once integrated, you can use these tools naturally in your conversations. With multi-workspace support, you can work across multiple Go projects simultaneously:

### Workspace Management

```
"What workspaces are available?"
"List all configured Go projects"
```

### Core Navigation Tools

```
"Where is the `ProcessRequest` function defined in project1?"
"Show me all places where `UserService` is used across all workspaces"
"What does the `http.Client` struct contain?"
```

### Diagnostic and Analysis Tools

```
"Are there any compilation errors in main.go?"
"Show me all functions and types defined in client.go"
"Find all symbols named 'Handler' across the workspace"
```

### Code Assistance Tools

```
"What parameters does the `log.Printf` function take?"
"Show me code completion suggestions for this position"
```

### Advanced Navigation Tools

```
"What's the type definition of this variable?"
"Find all implementations of the Writer interface"
```

### Code Maintenance Tools

```
"Format this Go file according to gofmt standards"
"Clean up and organize the imports in this file"
"Show me type hints for this code range"
```

The MCP server will automatically use the appropriate tool based on your requests and provide accurate information from your Go workspace(s). All tools support workspace-specific operations when working with multiple projects.

## Configuration

### Server Options

- **`-workspace`** (required): Path(s) to your Go workspace directory(ies)
  - Single workspace: `-workspace /path/to/project`
  - Multiple workspaces: `-workspace /project1,/project2,/project3`
  - Supports spaces in paths: `-workspace "/path with spaces/project1,/project2"`
  
  **⚠️ Memory Usage Notice**: Each workspace uses approximately **300MB of RAM** as it runs a dedicated gopls process. When using multiple workspaces, plan accordingly:
  - 1 workspace: ~300MB RAM
  - 5 workspaces: ~1.5GB RAM
  - 10 workspaces: ~3GB RAM
  
- **`-transport`** (optional): Transport type, accepts 'http' or 'stdio' (defaults to 'http')
- **Port**: Fixed at 8080 (Streamable HTTP transport only)

### Transport Options

#### Streamable HTTP Transport (Default)

```bash
# Single workspace with Streamable HTTP transport (default)
./gopls-mcp -workspace /path/to/go/project
./gopls-mcp -workspace /path/to/go/project -transport http

# Multiple workspaces with Streamable HTTP transport
./gopls-mcp -workspace /project1,/project2,/project3
./gopls-mcp -workspace /project1,/project2 -transport http

# Docker with single workspace (Streamable HTTP transport)
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Docker with multiple workspaces (Streamable HTTP transport)
docker run \
  -v /project1:/workspace1 \
  -v /project2:/workspace2 \
  -v /project3:/workspace3 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspace /workspace1,/workspace2,/workspace3
```

#### Stdio Transport

```bash
# Single workspace with stdio transport
./gopls-mcp -workspace /path/to/go/project -transport stdio

# Multiple workspaces with stdio transport
./gopls-mcp -workspace /project1,/project2,/project3 -transport stdio

# Docker with single workspace (stdio transport)
docker run -i -v /path/to/go/project:/workspace megagrindstone/gopls-mcp:latest -transport stdio

# Docker with multiple workspaces (stdio transport)
docker run -i \
  -v /project1:/workspace1 \
  -v /project2:/workspace2 \
  -v /project3:/workspace3 \
  megagrindstone/gopls-mcp:latest -workspace /workspace1,/workspace2,/workspace3 -transport stdio
```

For Claude Desktop with stdio transport:

```json
{
  "mcpServers": {
    "gopls-mcp": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-v", "/path/to/your/go/project:/workspace",
        "megagrindstone/gopls-mcp:latest",
        "-transport", "stdio"
      ]
    }
  }
}
```

### Logging Configuration

The server supports structured logging with configurable levels and formats via environment variables:

#### Environment Variables

- **`LOG_LEVEL`**: Set logging level (DEBUG, INFO, WARN, ERROR) - defaults to INFO
- **`LOG_FORMAT`**: Set log output format (text, json) - defaults to text

#### Usage Examples

```bash
# Native with custom logging
LOG_LEVEL=DEBUG ./gopls-mcp -workspace /path/to/go/project
LOG_FORMAT=json LOG_LEVEL=WARN ./gopls-mcp -workspace /path/to/go/project

# Docker with logging configuration
docker run -e LOG_LEVEL=DEBUG -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
docker run -e LOG_FORMAT=json -e LOG_LEVEL=INFO -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
```

### Workspace Requirements

Your Go workspace must contain:

- Valid `go.mod` file
- Proper Go module structure
- Accessible Go source files

The server will automatically initialize gopls with your workspace and maintain the language server connection throughout the session.

## Docker Deployment

### Using Docker Hub Images

I maintain pre-built Docker images on Docker Hub with multi-platform support:

```bash
# Latest stable release
docker pull megagrindstone/gopls-mcp:latest

# Specific version
docker pull megagrindstone/gopls-mcp:v0.3.0

# Run with single Go project
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# Run with multiple Go projects
docker run \
  -v /path/to/project1:/workspace1 \
  -v /path/to/project2:/workspace2 \
  -v /path/to/project3:/workspace3 \
  -p 8080:8080 \
  megagrindstone/gopls-mcp:latest -workspace /workspace1,/workspace2,/workspace3
```

### Available Tags

- `latest` - Latest stable release
- `v*` - Semantic version tags (e.g., `v0.3.0`)
- `main` - Latest development build

### Multi-Platform Support

Docker images support:

- `linux/amd64` - Intel/AMD 64-bit
- `linux/arm64` - ARM 64-bit (Apple Silicon, ARM servers)

## Development Setup

### Prerequisites

- Go 1.24+
- Docker (optional)
- golangci-lint (optional, for linting)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/MegaGrindStone/gopls-mcp.git
cd gopls-mcp

# Build the binary
go build -o gopls-mcp

# Run with single Go workspace
./gopls-mcp -workspace /path/to/your/go/project

# Run with multiple Go workspaces
./gopls-mcp -workspace /path/to/project1,/path/to/project2,/path/to/project3
```

### Testing

```bash
# Run tests with verbose output
go test ./... -v -count=1 -p 1

# Run linter
golangci-lint run ./...

# Format code
go fmt ./...
```

### Docker Development

```bash
# Build Docker image
docker build -t gopls-mcp .

# Test locally with single workspace
docker run -v /path/to/test/project:/workspace -p 8080:8080 gopls-mcp

# Test locally with multiple workspaces
docker run \
  -v /path/to/test/project1:/workspace1 \
  -v /path/to/test/project2:/workspace2 \
  -p 8080:8080 \
  gopls-mcp -workspace /workspace1,/workspace2
```

## API Reference

### MCP Tools

All tools require a `workspace` parameter to specify which workspace to operate on. Use relative paths within the workspace.

#### 🏢 Workspace Management Tools

##### list_workspaces

List all available Go workspaces configured in the server.

**Parameters:** None

**Example:**

```json
{
  "name": "list_workspaces",
  "arguments": {}
}
```

#### 🎯 Core Navigation Tools

##### go_to_definition

Navigate to symbol definitions in your Go workspace.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file (e.g., `main.go`, `pkg/client.go`)
- `line` (number): Line number (1-based)
- `character` (number): Character position (0-based)

**Example:**

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

##### find_references

Find all references to a symbol across your Go workspace.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file
- `line` (number): Line number (1-based)
- `character` (number): Character position (0-based)
- `includeDeclaration` (boolean): Include the declaration in results

**Example:**

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

##### get_hover_info

Get documentation and type information for symbols.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file
- `line` (number): Line number (1-based)
- `character` (number): Character position (0-based)

**Example:**

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

#### 🔍 Diagnostic and Analysis Tools

##### get_diagnostics

Get compilation errors, warnings, and diagnostics for a Go file.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file

**Example:**

```json
{
  "name": "get_diagnostics",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go"
  }
}
```

##### get_document_symbols

Get outline of symbols (functions, types, etc.) defined in a Go file.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file

**Example:**

```json
{
  "name": "get_document_symbols",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "client.go"
  }
}
```

##### get_workspace_symbols

Search for symbols across the entire Go workspace/project.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `query` (string): Symbol name to search for

**Example:**

```json
{
  "name": "get_workspace_symbols",
  "arguments": {
    "workspace": "/path/to/workspace",
    "query": "Client"
  }
}
```

#### 💡 Code Assistance Tools

##### get_signature_help

Get function signature help (parameter information) at the specified position.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file
- `line` (number): Line number (1-based)
- `character` (number): Character position (0-based)

**Example:**

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

##### get_completions

Get code completion suggestions at the specified position.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file
- `line` (number): Line number (1-based)
- `character` (number): Character position (0-based)

**Example:**

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

#### 🧭 Advanced Navigation Tools

##### get_type_definition

Navigate to the type definition of a symbol at the specified position.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file
- `line` (number): Line number (1-based)
- `character` (number): Character position (0-based)

**Example:**

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

##### find_implementations

Find all implementations of an interface or method at the specified position.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file
- `line` (number): Line number (1-based)
- `character` (number): Character position (0-based)

**Example:**

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

#### 🛠️ Code Maintenance Tools

##### format_document

Format a Go source file according to gofmt standards.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file

**Example:**

```json
{
  "name": "format_document",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "main.go"
  }
}
```

##### organize_imports

Organize and clean up import statements in a Go file.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file

**Example:**

```json
{
  "name": "organize_imports",
  "arguments": {
    "workspace": "/path/to/workspace",
    "path": "client.go"
  }
}
```

##### get_inlay_hints

Get inlay hints (implicit parameter names, type information) for a range in a Go file.

**Parameters:**

- `workspace` (string): Workspace path to use for this request
- `path` (string): Relative path to Go file
- `startLine` (number): Start line number (1-based)
- `startChar` (number): Start character position (0-based)
- `endLine` (number): End line number (1-based)
- `endChar` (number): End character position (0-based)

**Example:**

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

## Troubleshooting

### Common Issues

**"workspace flag is required"**

- Ensure you provide the `-workspace` flag when running the server
- The workspace path must contain a valid Go module

**"Failed to start gopls"**

- Check that gopls is available in your PATH (automatically handled in Docker)
- Ensure your Go workspace has a valid `go.mod` file
- Verify the workspace path is accessible

**"Connection refused"**

- Ensure the server is running on port 8080
- Check that no other service is using the port
- Verify the MCP server is accessible at `http://localhost:8080`

### Debug Mode

To debug issues, run the server with verbose logging:

```bash
# Native with debug logging
LOG_LEVEL=DEBUG ./gopls-mcp -workspace /path/to/project

# Docker with debug logging (Streamable HTTP)
docker run -e LOG_LEVEL=DEBUG -v /path/to/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest

# View Docker container logs
docker logs <container-id>

# JSON format logging for easier parsing
docker run -e LOG_FORMAT=json -e LOG_LEVEL=DEBUG -v /path/to/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
```

## Contributing

I welcome contributions to improve gopls-mcp! This project has been developed in collaboration with Claude Code, and I'm excited to see how the community can enhance it further.

### Development Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite: `go test ./... -v -count=1 -p 1`
6. Run the linter: `golangci-lint run ./...`
7. Submit a pull request

### Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Follow the existing patterns in the codebase
- Add tests for new functionality

### Reporting Issues

Please report issues on the [GitHub issue tracker](https://github.com/MegaGrindStone/gopls-mcp/issues) with:

- Your operating system and Go version
- Steps to reproduce the issue
- Expected vs actual behavior
- Relevant logs or error messages

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **Claude Code** - This project was developed in collaboration with [Claude Code](https://claude.ai/code), Anthropic's AI-powered coding assistant
- **gopls team** - For creating the excellent Go language server
- **Model Context Protocol** - For providing the framework that enables this integration
- **Go community** - For building the tools and ecosystem that make this possible

---

**Author:** Gerard Adam  
**Collaboration:** Developed with Claude Code  
**Repository:** <https://github.com/MegaGrindStone/gopls-mcp>  
**Docker Hub:** <https://hub.docker.com/r/megagrindstone/gopls-mcp>
