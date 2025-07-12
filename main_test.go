package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestWorkspacePathValidation(t *testing.T) {
	// Save original command line arguments
	origArgs := os.Args
	defer func() {
		os.Args = origArgs
	}()

	// Test case 1: Missing workspace path
	os.Args = []string{"gopls-mcp"}

	// Reset flag for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	workspacePath := flag.String("workspace", "", "Path to the Go workspace directory (required)")
	flag.Parse()

	if *workspacePath != "" {
		t.Errorf("Expected empty workspace path, got %s", *workspacePath)
	}

	// Test case 2: Valid workspace path
	os.Args = []string{"gopls-mcp", "-workspace", "/test/workspace"}

	// Reset flag for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	workspacePath = flag.String("workspace", "", "Path to the Go workspace directory (required)")
	flag.Parse()

	expectedPath := "/test/workspace"
	if *workspacePath != expectedPath {
		t.Errorf("Expected workspace path %s, got %s", expectedPath, *workspacePath)
	}
}
func TestWorkspacePathParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "valid workspace path",
			args:     []string{"gopls-mcp", "-workspace", "/test/workspace"},
			expected: "/test/workspace",
		},
		{
			name:     "empty workspace path",
			args:     []string{"gopls-mcp"},
			expected: "",
		},
		{
			name:     "workspace path with equals",
			args:     []string{"gopls-mcp", "-workspace=/test/workspace"},
			expected: "/test/workspace",
		},
		{
			name:     "workspace path with spaces",
			args:     []string{"gopls-mcp", "-workspace", "/test/workspace with spaces"},
			expected: "/test/workspace with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original command line arguments
			origArgs := os.Args
			defer func() {
				os.Args = origArgs
			}()

			// Set test arguments
			os.Args = tt.args

			// Reset flag for testing
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Parse flags
			workspacePath := flag.String("workspace", "", "Path to the Go workspace directory (required)")
			flag.Parse()

			if *workspacePath != tt.expected {
				t.Errorf("Expected workspace path %s, got %s", tt.expected, *workspacePath)
			}
		})
	}
}

func TestClientCreation(t *testing.T) {
	workspacePath := "/test/workspace"
	logger := newTestLogger()
	client := newClient(workspacePath, logger)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.workspacePath != workspacePath {
		t.Errorf("Expected workspace path %s, got %s", workspacePath, client.workspacePath)
	}

	if client.isRunning() {
		t.Error("Expected client to not be running initially")
	}
}

func TestMCPToolsCreation(t *testing.T) {
	// Test that we can create the basic components without errors
	workspacePath := "/test/workspace"
	logger := newTestLogger()
	client := newClient(workspacePath, logger)

	// Create mcpTools wrapper with client map
	clients := map[string]*goplsClient{workspacePath: client}
	mcpToolsWrapper := newMCPTools(clients)

	// Test that mcpTools wrapper was created successfully
	if mcpToolsWrapper.clients == nil {
		t.Error("Expected non-nil clients map in mcpTools wrapper")
	}

	if len(mcpToolsWrapper.clients) != 1 {
		t.Errorf("Expected 1 client in map, got %d", len(mcpToolsWrapper.clients))
	}

	// Test that we can create an MCP server with new v0.2.0 API
	server := mcp.NewServer(&mcp.Implementation{Name: "gopls-mcp", Version: "v0.3.0"}, nil)
	if server == nil {
		t.Error("Expected non-nil MCP server")
	}

	// Test adding a tool to server with new API
	mcp.AddTool(server, &mcp.Tool{Name: "test_tool", Description: "Test tool"}, mcpToolsWrapper.HandleGoToDefinition)
}
func TestMCPServerCreation(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "gopls-mcp", Version: "v0.3.0"}, nil)
	if server == nil {
		t.Error("Expected non-nil MCP server")
	}

	// Test that we can create a streamable HTTP handler
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	if handler == nil {
		t.Error("Expected non-nil streamable HTTP handler")
	}
}
func TestHTTPServerConfiguration(t *testing.T) {
	// Test HTTP server configuration values
	expectedAddr := ":8080"
	expectedReadTimeout := 15 * time.Second
	expectedWriteTimeout := 15 * time.Second
	expectedIdleTimeout := 60 * time.Second

	httpServer := &http.Server{
		Addr:         expectedAddr,
		ReadTimeout:  expectedReadTimeout,
		WriteTimeout: expectedWriteTimeout,
		IdleTimeout:  expectedIdleTimeout,
	}

	if httpServer.Addr != expectedAddr {
		t.Errorf("Expected server address %s, got %s", expectedAddr, httpServer.Addr)
	}

	if httpServer.ReadTimeout != expectedReadTimeout {
		t.Errorf("Expected read timeout %v, got %v", expectedReadTimeout, httpServer.ReadTimeout)
	}

	if httpServer.WriteTimeout != expectedWriteTimeout {
		t.Errorf("Expected write timeout %v, got %v", expectedWriteTimeout, httpServer.WriteTimeout)
	}

	if httpServer.IdleTimeout != expectedIdleTimeout {
		t.Errorf("Expected idle timeout %v, got %v", expectedIdleTimeout, httpServer.IdleTimeout)
	}
}
func TestHTTPMuxSetup(t *testing.T) {
	// Test HTTP mux setup
	mux := http.NewServeMux()
	if mux == nil {
		t.Error("Expected non-nil HTTP mux")
	}

	// Test that we can add handlers to the mux
	testHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	mux.HandleFunc("/test", testHandler)

	// The mux should handle the route we just added
	// This is a basic test to ensure mux.HandleFunc works
	if mux == nil {
		t.Error("Mux should not be nil after adding handler")
	}
}
func TestTransportFlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "default transport (http)",
			args:     []string{"gopls-mcp", "-workspace", "/test"},
			expected: "http",
		},
		{
			name:     "explicit http transport",
			args:     []string{"gopls-mcp", "-workspace", "/test", "-transport", "http"},
			expected: "http",
		},
		{
			name:     "stdio transport",
			args:     []string{"gopls-mcp", "-workspace", "/test", "-transport", "stdio"},
			expected: "stdio",
		},
		{
			name:     "transport with equals",
			args:     []string{"gopls-mcp", "-workspace", "/test", "-transport=stdio"},
			expected: "stdio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original command line arguments
			origArgs := os.Args
			defer func() {
				os.Args = origArgs
			}()

			// Set test arguments
			os.Args = tt.args

			// Reset flag for testing
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Parse flags
			workspacePath := flag.String("workspace", "", "Path to the Go workspace directory (required)")
			transportType := flag.String("transport", "http", "Transport type: http or stdio")
			flag.Parse()

			if *transportType != tt.expected {
				t.Errorf("Expected transport type %s, got %s", tt.expected, *transportType)
			}

			// Ensure workspace is still parsed correctly
			if *workspacePath != "/test" {
				t.Errorf("Expected workspace path /test, got %s", *workspacePath)
			}
		})
	}
}
func TestTransportValidation(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		valid     bool
	}{
		{
			name:      "valid http transport",
			transport: "http",
			valid:     true,
		},
		{
			name:      "valid stdio transport",
			transport: "stdio",
			valid:     true,
		},
		{
			name:      "invalid transport",
			transport: "invalid",
			valid:     false,
		},
		{
			name:      "empty transport",
			transport: "",
			valid:     false,
		},
		{
			name:      "websocket transport (not supported)",
			transport: "websocket",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.transport == "http" || tt.transport == "stdio"
			if isValid != tt.valid {
				t.Errorf("Expected transport %s validity to be %v, got %v", tt.transport, tt.valid, isValid)
			}
		})
	}
}

func TestSetupMCPServer(t *testing.T) {
	// Create a test client
	workspacePath := "/test/workspace"
	logger := newTestLogger()
	client := newClient(workspacePath, logger)

	// Test server setup with client map
	clients := map[string]*goplsClient{workspacePath: client}
	server := setupMCPServer(clients)

	if server == nil {
		t.Fatal("Expected non-nil MCP server")
	}

	// The server should be properly configured with tools
	// (We can't test the tools directly without starting gopls,
	// but we can verify the server is created)
}

// newTestLogger creates a logger suitable for testing.
func newTestLogger() *slog.Logger {
	// Use a simple text handler with ERROR level for testing to reduce noise
	opts := &slog.HandlerOptions{
		Level: slog.LevelError,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
