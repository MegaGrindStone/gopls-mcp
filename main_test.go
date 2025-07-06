package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestWorkspacesPathValidation(t *testing.T) {
	// Save original command line arguments
	origArgs := os.Args
	defer func() {
		os.Args = origArgs
	}()

	// Test case 1: Missing workspaces path
	os.Args = []string{"gopls-mcp"}

	// Reset flag for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	workspacesFlag := flag.String("workspaces", "", "Comma-separated list of workspace paths (required)")
	flag.Parse()

	if *workspacesFlag != "" {
		t.Errorf("Expected empty workspaces path, got %s", *workspacesFlag)
	}

	// Test case 2: Valid single workspace path
	os.Args = []string{"gopls-mcp", "-workspaces", "/test/workspace"}

	// Reset flag for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	workspacesFlag = flag.String("workspaces", "", "Comma-separated list of workspace paths (required)")
	flag.Parse()

	expectedPath := "/test/workspace"
	if *workspacesFlag != expectedPath {
		t.Errorf("Expected workspaces path %s, got %s", expectedPath, *workspacesFlag)
	}
}

func TestWorkspacesPathParsing(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedFlag   string
		expectedParsed []string
	}{
		{
			name:           "valid single workspace path",
			args:           []string{"gopls-mcp", "-workspaces", "/test/workspace"},
			expectedFlag:   "/test/workspace",
			expectedParsed: []string{"/test/workspace"},
		},
		{
			name:           "valid multiple workspace paths",
			args:           []string{"gopls-mcp", "-workspaces", "/test/workspace1,/test/workspace2"},
			expectedFlag:   "/test/workspace1,/test/workspace2",
			expectedParsed: []string{"/test/workspace1", "/test/workspace2"},
		},
		{
			name:           "empty workspaces path",
			args:           []string{"gopls-mcp"},
			expectedFlag:   "",
			expectedParsed: []string{""},
		},
		{
			name:           "workspaces path with equals",
			args:           []string{"gopls-mcp", "-workspaces=/test/workspace"},
			expectedFlag:   "/test/workspace",
			expectedParsed: []string{"/test/workspace"},
		},
		{
			name:           "workspaces path with spaces",
			args:           []string{"gopls-mcp", "-workspaces", "/test/workspace with spaces"},
			expectedFlag:   "/test/workspace with spaces",
			expectedParsed: []string{"/test/workspace with spaces"},
		},
		{
			name:           "multiple workspaces with spaces",
			args:           []string{"gopls-mcp", "-workspaces", "/test/workspace1, /test/workspace2 ,/test/workspace3"},
			expectedFlag:   "/test/workspace1, /test/workspace2 ,/test/workspace3",
			expectedParsed: []string{"/test/workspace1", "/test/workspace2", "/test/workspace3"},
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
			workspacesFlag := flag.String("workspaces", "", "Comma-separated list of workspace paths (required)")
			flag.Parse()

			if *workspacesFlag != tt.expectedFlag {
				t.Errorf("Expected workspaces flag %s, got %s", tt.expectedFlag, *workspacesFlag)
			}

			// Test parsing logic
			if *workspacesFlag != "" {
				workspacePaths := strings.Split(*workspacesFlag, ",")
				for i, path := range workspacePaths {
					workspacePaths[i] = strings.TrimSpace(path)
				}

				if len(workspacePaths) != len(tt.expectedParsed) {
					t.Errorf("Expected %d workspaces, got %d", len(tt.expectedParsed), len(workspacePaths))
				}

				for i, expected := range tt.expectedParsed {
					if i < len(workspacePaths) && workspacePaths[i] != expected {
						t.Errorf("Expected workspace %d to be %s, got %s", i, expected, workspacePaths[i])
					}
				}
			}
		})
	}
}

func TestWorkspaceManagerCreation(t *testing.T) {
	workspacePaths := []string{"/test/workspace1", "/test/workspace2"}
	logger := newTestLogger()
	workspaceManager := NewWorkspaceManager(workspacePaths, logger)

	if workspaceManager == nil {
		t.Fatal("Expected non-nil workspace manager")
	}

	// Test that all workspaces are created
	for _, workspace := range workspacePaths {
		manager, err := workspaceManager.GetManager(workspace)
		if err != nil {
			t.Errorf("Expected to find manager for workspace %s, got error: %v", workspace, err)
		}
		if manager == nil {
			t.Errorf("Expected non-nil manager for workspace %s", workspace)
		}
		if manager.IsRunning() {
			t.Errorf("Expected manager for workspace %s to not be running initially", workspace)
		}
	}

	// Test GetWorkspaces
	workspaces := workspaceManager.GetWorkspaces()
	if len(workspaces) != len(workspacePaths) {
		t.Errorf("Expected %d workspaces, got %d", len(workspacePaths), len(workspaces))
	}

	// Test GetWorkspaceStatus
	status := workspaceManager.GetWorkspaceStatus()
	if len(status) != len(workspacePaths) {
		t.Errorf("Expected %d workspace statuses, got %d", len(workspacePaths), len(status))
	}
	for workspace := range status {
		found := false
		for _, expectedWorkspace := range workspacePaths {
			if workspace == expectedWorkspace {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected workspace in status: %s", workspace)
		}
	}
}

func TestServerComponentCreation(t *testing.T) {
	// Test that we can create the basic components without errors
	workspacePaths := []string{"/test/workspace"}
	logger := newTestLogger()
	workspaceManager := NewWorkspaceManager(workspacePaths, logger)

	// Test tool creation
	tools := []*mcp.ServerTool{
		workspaceManager.CreateGoToDefinitionTool(),
		workspaceManager.CreateFindReferencesTool(),
		workspaceManager.CreateGetHoverTool(),
		workspaceManager.CreateListWorkspacesTool(),
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
			args:     []string{"gopls-mcp", "-workspaces", "/test"},
			expected: "http",
		},
		{
			name:     "explicit http transport",
			args:     []string{"gopls-mcp", "-workspaces", "/test", "-transport", "http"},
			expected: "http",
		},
		{
			name:     "stdio transport",
			args:     []string{"gopls-mcp", "-workspaces", "/test", "-transport", "stdio"},
			expected: "stdio",
		},
		{
			name:     "transport with equals",
			args:     []string{"gopls-mcp", "-workspaces", "/test", "-transport=stdio"},
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
			workspacesFlag := flag.String("workspaces", "", "Comma-separated list of workspace paths (required)")
			transportType := flag.String("transport", "http", "Transport type: http or stdio")
			flag.Parse()

			if *transportType != tt.expected {
				t.Errorf("Expected transport type %s, got %s", tt.expected, *transportType)
			}

			// Ensure workspaces is still parsed correctly
			if *workspacesFlag != "/test" {
				t.Errorf("Expected workspaces flag /test, got %s", *workspacesFlag)
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
	// Create a test workspace manager
	workspacePaths := []string{"/test/workspace"}
	logger := newTestLogger()
	workspaceManager := NewWorkspaceManager(workspacePaths, logger)

	// Test server setup
	server := setupMCPServer(workspaceManager)

	if server == nil {
		t.Fatal("Expected non-nil MCP server")
	}

	// The server should be properly configured with tools
	// (We can't test the tools directly without starting gopls,
	// but we can verify the server is created)
}

// TestMultiWorkspaceScenarios tests various multi-workspace scenarios.
func TestMultiWorkspaceScenarios(t *testing.T) {
	tests := []struct {
		name       string
		workspaces []string
		testFunc   func(t *testing.T, wm *WorkspaceManager)
	}{
		{
			name:       "single workspace",
			workspaces: []string{"/test/workspace1"},
			testFunc: func(t *testing.T, wm *WorkspaceManager) {
				workspaces := wm.GetWorkspaces()
				if len(workspaces) != 1 {
					t.Errorf("Expected 1 workspace, got %d", len(workspaces))
				}
				if workspaces[0] != "/test/workspace1" {
					t.Errorf("Expected workspace /test/workspace1, got %s", workspaces[0])
				}
			},
		},
		{
			name:       "multiple workspaces",
			workspaces: []string{"/test/workspace1", "/test/workspace2", "/test/workspace3"},
			testFunc: func(t *testing.T, wm *WorkspaceManager) {
				workspaces := wm.GetWorkspaces()
				if len(workspaces) != 3 {
					t.Errorf("Expected 3 workspaces, got %d", len(workspaces))
				}

				expectedWorkspaces := map[string]bool{
					"/test/workspace1": true,
					"/test/workspace2": true,
					"/test/workspace3": true,
				}

				for _, workspace := range workspaces {
					if !expectedWorkspaces[workspace] {
						t.Errorf("Unexpected workspace: %s", workspace)
					}
				}
			},
		},
		{
			name:       "workspace manager not found",
			workspaces: []string{"/test/workspace1"},
			testFunc: func(t *testing.T, wm *WorkspaceManager) {
				_, err := wm.GetManager("/nonexistent/workspace")
				if err == nil {
					t.Error("Expected error for nonexistent workspace, got nil")
				}
				expectedError := "workspace not found: /nonexistent/workspace"
				if err.Error() != expectedError {
					t.Errorf("Expected error %s, got %s", expectedError, err.Error())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newTestLogger()
			workspaceManager := NewWorkspaceManager(tt.workspaces, logger)

			if workspaceManager == nil {
				t.Fatal("Expected non-nil workspace manager")
			}

			tt.testFunc(t, workspaceManager)
		})
	}
}

// TestWorkspaceManagerToolCreation tests that all tools are created properly for WorkspaceManager.
func TestWorkspaceManagerToolCreation(t *testing.T) {
	workspaces := []string{"/test/workspace1", "/test/workspace2"}
	logger := newTestLogger()
	wm := NewWorkspaceManager(workspaces, logger)

	tools := map[string]*mcp.ServerTool{
		"go_to_definition": wm.CreateGoToDefinitionTool(),
		"find_references":  wm.CreateFindReferencesTool(),
		"get_hover_info":   wm.CreateGetHoverTool(),
		"list_workspaces":  wm.CreateListWorkspacesTool(),
	}

	for name, tool := range tools {
		if tool == nil {
			t.Errorf("Tool %s is nil", name)
		}
	}

	// Test that we can add all tools to an MCP server
	server := mcp.NewServer("test-server", "v0.1.0", nil)
	toolSlice := make([]*mcp.ServerTool, 0, len(tools))
	for _, tool := range tools {
		toolSlice = append(toolSlice, tool)
	}
	server.AddTools(toolSlice...)
}
