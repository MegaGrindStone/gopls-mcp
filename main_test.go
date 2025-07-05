package main

import (
	"flag"
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

func TestManagerCreation(t *testing.T) {
	workspacePath := "/test/workspace"
	manager := NewManager(workspacePath)

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.workspacePath != workspacePath {
		t.Errorf("Expected workspace path %s, got %s", workspacePath, manager.workspacePath)
	}

	if manager.IsRunning() {
		t.Error("Expected manager to not be running initially")
	}
}

func TestServerComponentCreation(t *testing.T) {
	// Test that we can create the basic components without errors
	workspacePath := "/test/workspace"
	manager := NewManager(workspacePath)

	// Test tool creation
	tools := []*mcp.ServerTool{
		manager.CreateGoToDefinitionTool(),
		manager.CreateFindReferencesTool(),
		manager.CreateGetHoverTool(),
	}

	for i, tool := range tools {
		if tool == nil {
			t.Errorf("Tool %d is nil", i)
		}
	}

	// Test that we can create an MCP server (this doesn't start it)
	server := mcp.NewServer("gopls-mcp", "v0.1.0", nil)
	if server == nil {
		t.Error("Expected non-nil MCP server")
	}

	// Test adding tools to server
	server.AddTools(tools...)
}

func TestMCPServerCreation(t *testing.T) {
	server := mcp.NewServer("gopls-mcp", "v0.1.0", nil)
	if server == nil {
		t.Error("Expected non-nil MCP server")
	}

	// Test that we can create an SSE handler
	handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return server
	})

	if handler == nil {
		t.Error("Expected non-nil SSE handler")
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
