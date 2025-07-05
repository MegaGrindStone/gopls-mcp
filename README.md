# gopls-mcp

[![GitHub release](https://img.shields.io/github/release/MegaGrindStone/gopls-mcp.svg)](https://github.com/MegaGrindStone/gopls-mcp/releases)
[![Docker Hub](https://img.shields.io/docker/pulls/megagrindstone/gopls-mcp)](https://hub.docker.com/r/megagrindstone/gopls-mcp)
[![CI](https://github.com/MegaGrindStone/gopls-mcp/workflows/Release/badge.svg)](https://github.com/MegaGrindStone/gopls-mcp/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Model Context Protocol (MCP) server that integrates [gopls](https://pkg.go.dev/golang.org/x/tools/gopls) (Go language server) with MCP-compatible hosts like Claude Code, Claude Desktop, and VS Code.

I've developed this project in collaboration with [Claude Code](https://claude.ai/code) to provide seamless Go language server capabilities through the MCP protocol, enabling powerful Go development assistance in AI-powered coding environments.

## Features

This MCP server provides three essential Go development tools:

- **🎯 Go to Definition** - Navigate to symbol definitions across your Go workspace
- **🔍 Find References** - Locate all references to functions, variables, and types
- **📖 Hover Information** - Get documentation, type information, and signatures

All tools work with your existing Go workspace and leverage gopls for accurate, fast results.

## Installation

### Claude Code Integration (Recommended)

The easiest way to use gopls-mcp is with Claude Code using SSE transport:

1. **Start the server** with your Go workspace:

   ```bash
   # Using Docker (recommended)
   docker run -d -v /path/to/your/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
   
   # Or build from source
   go build -o gopls-mcp
   ./gopls-mcp -workspace /path/to/your/go/project
   ```

2. **Add to Claude Code**:

   ```bash
   claude mcp add --transport sse gopls-mcp http://localhost:8080/sse
   ```

3. **Start using** - The Go tools will be available in your Claude Code conversations!

### Claude Desktop Integration

For Claude Desktop, add this to your MCP settings:

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

### Other MCP Hosts

For VS Code and other MCP hosts, use the HTTP transport:

```json
{
  "servers": {
    "gopls-mcp": {
      "type": "http",
      "url": "http://localhost:8080/sse"
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
   claude mcp add --transport sse gopls-mcp http://localhost:8080/sse
   ```

4. **Test the integration** - Ask Claude Code to help with your Go code!

## Usage Examples

Once integrated, you can use these tools naturally in your conversations:

### Go to Definition

```
"Where is the `ProcessRequest` function defined?"
```

### Find References

```
"Show me all places where `UserService` is used"
```

### Hover Information

```
"What does the `http.Client` struct contain?"
```

The MCP server will automatically use the appropriate tool based on your requests and provide accurate information from your Go workspace.

## Configuration

### Server Options

- **`-workspace`** (required): Path to your Go workspace directory
- **Port**: Fixed at 8080 (SSE endpoint at `/sse`)

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
docker pull megagrindstone/gopls-mcp:v0.1.0

# Run with your Go project
docker run -v /path/to/go/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
```

### Available Tags

- `latest` - Latest stable release
- `v*` - Semantic version tags (e.g., `v0.1.0`)
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

# Run with your Go workspace
./gopls-mcp -workspace /path/to/your/go/project
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

# Test locally
docker run -v /path/to/test/project:/workspace -p 8080:8080 gopls-mcp
```

## API Reference

### MCP Tools

#### go_to_definition

Navigate to symbol definitions in your Go workspace.

**Parameters:**

- `uri` (string): File URI (e.g., `file:///path/to/file.go`)
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

**Example:**

```json
{
  "name": "go_to_definition",
  "arguments": {
    "uri": "file:///workspace/main.go",
    "line": 10,
    "character": 5
  }
}
```

#### find_references

Find all references to a symbol across your Go workspace.

**Parameters:**

- `uri` (string): File URI
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)
- `includeDeclaration` (boolean): Include the declaration in results

**Example:**

```json
{
  "name": "find_references",
  "arguments": {
    "uri": "file:///workspace/main.go",
    "line": 10,
    "character": 5,
    "includeDeclaration": true
  }
}
```

#### get_hover_info

Get documentation and type information for symbols.

**Parameters:**

- `uri` (string): File URI
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

**Example:**

```json
{
  "name": "get_hover_info",
  "arguments": {
    "uri": "file:///workspace/main.go",
    "line": 10,
    "character": 5
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
- Verify the SSE endpoint is accessible at `http://localhost:8080/sse`

### Debug Mode

To debug issues, run the server with verbose logging:

```bash
# Native
./gopls-mcp -workspace /path/to/project

# Docker with logs
docker run -v /path/to/project:/workspace -p 8080:8080 megagrindstone/gopls-mcp:latest
docker logs <container-id>
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

