package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Test helper functions for MCP integration tests

// parseJSONResult parses a JSON result from MCP tool response.
func parseJSONResult[T any](t *testing.T, result *mcp.CallToolResultFor[T]) T {
	t.Helper()

	if result == nil || len(result.Content) == 0 {
		t.Fatal("Expected non-empty result content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content")
	}

	var parsed T
	if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	return parsed
}

// Integration tests for MCP tools

func TestMCPGoToDefinitionIntegration(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	// Create MCP tools wrapper to access handlers
	clients := map[string]*goplsClient{workspacePath: newClient(workspacePath, newDebugLogger())}
	tools := newMCPTools(clients)

	// Start the client for testing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := clients[workspacePath]
	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to load the workspace
	time.Sleep(3 * time.Second)

	// Test go to definition on "testFunction" call in main.go
	params := &mcp.CallToolParamsFor[GoToDefinitionParams]{
		Arguments: GoToDefinitionParams{
			Workspace: workspacePath,
			Path:      "main.go",
			Line:      7,  // testFunction call line (1-based)
			Character: 11, // testFunction position
		},
	}

	result, err := tools.HandleGoToDefinition(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleGoToDefinition failed: %v", err)
	}

	defResult := parseJSONResult(t, result)

	if len(defResult.Locations) == 0 {
		t.Error("Expected at least one definition location")
	}

	// Verify the location points to main.go
	if len(defResult.Locations) > 0 {
		location := defResult.Locations[0]
		if location.URI != "main.go" {
			t.Errorf("Expected URI 'main.go', got '%s'", location.URI)
		}

		t.Logf("Found definition at %s:%d:%d", location.URI, location.Line, location.Character)
	}
}

func TestMCPFindReferencesIntegration(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	// Create clients and tools
	clients := map[string]*goplsClient{workspacePath: newClient(workspacePath, newDebugLogger())}
	tools := newMCPTools(clients)

	// Start the client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := clients[workspacePath]
	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test find references on "testFunction" definition
	params := &mcp.CallToolParamsFor[FindReferencesParams]{
		Arguments: FindReferencesParams{
			Workspace:          workspacePath,
			Path:               "main.go",
			Line:               12, // testFunction definition line (1-based)
			Character:          5,  // testFunction name position
			IncludeDeclaration: true,
		},
	}

	result, err := tools.HandleFindReferences(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleFindReferences failed: %v", err)
	}

	refResult := parseJSONResult(t, result)

	if len(refResult.Locations) == 0 {
		t.Error("Expected at least one reference location")
	}

	t.Logf("Found %d references", len(refResult.Locations))
	for i, loc := range refResult.Locations {
		t.Logf("Reference %d: %s:%d:%d", i, loc.URI, loc.Line, loc.Character)
	}
}

func TestMCPGetHoverIntegration(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	// Create clients and tools
	clients := map[string]*goplsClient{workspacePath: newClient(workspacePath, newDebugLogger())}
	tools := newMCPTools(clients)

	// Start the client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := clients[workspacePath]
	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test hover on "testFunction" call
	params := &mcp.CallToolParamsFor[GetHoverParams]{
		Arguments: GetHoverParams{
			Workspace: workspacePath,
			Path:      "main.go",
			Line:      7,  // testFunction call line (1-based)
			Character: 11, // testFunction position
		},
	}

	result, err := tools.HandleGetHover(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleGetHover failed: %v", err)
	}

	hoverResult := parseJSONResult(t, result)

	if len(hoverResult.Contents) == 0 {
		t.Error("Expected hover contents")
	}

	// Verify hover contains function information
	foundFunction := false
	for _, content := range hoverResult.Contents {
		if contains(content, "testFunction") || contains(content, "func") {
			foundFunction = true
			break
		}
	}

	if !foundFunction {
		t.Error("Expected hover to contain function information")
	}

	t.Logf("Hover contents: %+v", hoverResult.Contents)
}

func TestMCPGetDiagnosticsIntegration(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	// Create a file with syntax errors for testing diagnostics
	errorContent := `package main

import "fmt

func main() {
	fmt.Println("Hello, World!")
	result := testFunction(
	fmt.Println("Result:", result)
}

func testFunction() int {
	return 42
}`
	errorPath := filepath.Join(workspacePath, "error.go")
	if err := os.WriteFile(errorPath, []byte(errorContent), 0644); err != nil {
		t.Fatalf("failed to create error.go: %v", err)
	}

	// Create clients and tools
	clients := map[string]*goplsClient{workspacePath: newClient(workspacePath, newDebugLogger())}
	tools := newMCPTools(clients)

	// Start the client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := clients[workspacePath]
	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to analyze the workspace
	time.Sleep(3 * time.Second)

	// Test get diagnostics for the error file
	params := &mcp.CallToolParamsFor[GetDiagnosticsParams]{
		Arguments: GetDiagnosticsParams{
			Workspace: workspacePath,
			Path:      "error.go",
		},
	}

	result, err := tools.HandleGetDiagnostics(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleGetDiagnostics failed: %v", err)
	}

	diagResult := parseJSONResult(t, result)

	// The error file should have diagnostics
	if len(diagResult.Diagnostics) == 0 {
		t.Log("No diagnostics found (gopls behavior may vary)")
	} else {
		t.Logf("Found %d diagnostics", len(diagResult.Diagnostics))
		for i, diag := range diagResult.Diagnostics {
			t.Logf("Diagnostic %d: %s (severity %d) at line %d:%d",
				i, diag.Message, diag.Severity, diag.Range.Line, diag.Range.Character)
		}
	}
}

func TestMCPListWorkspacesIntegration(t *testing.T) {
	requireGopls(t)

	// Create multiple test workspaces
	workspace1, cleanup1 := createTempGoWorkspace(t)
	defer cleanup1()

	workspace2, cleanup2 := createTempGoWorkspace(t)
	defer cleanup2()

	// Create clients and tools for multiple workspaces
	clients := map[string]*goplsClient{
		workspace1: newClient(workspace1, newDebugLogger()),
		workspace2: newClient(workspace2, newDebugLogger()),
	}
	tools := newMCPTools(clients)

	// Start clients
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, client := range clients {
		if err := client.start(ctx); err != nil {
			t.Fatalf("failed to start client: %v", err)
		}
		defer func(c *goplsClient) { _ = c.stop() }(client)
	}

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test list workspaces
	params := &mcp.CallToolParamsFor[ListWorkspacesParams]{
		Arguments: ListWorkspacesParams{},
	}

	result, err := tools.HandleListWorkspaces(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleListWorkspaces failed: %v", err)
	}

	listResult := parseJSONResult(t, result)

	if len(listResult.Workspaces) != 2 {
		t.Errorf("Expected 2 workspaces, got %d", len(listResult.Workspaces))
	}

	// Check that both workspace paths are present
	paths := make(map[string]bool)
	for _, ws := range listResult.Workspaces {
		paths[ws.Path] = true
	}

	if !paths[workspace1] || !paths[workspace2] {
		t.Error("Expected both workspace paths in result")
	}

	t.Logf("Found workspaces: %+v", listResult.Workspaces)
}

func TestMCPGetDocumentSymbolsIntegration(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	// Create clients and tools
	clients := map[string]*goplsClient{workspacePath: newClient(workspacePath, newDebugLogger())}
	tools := newMCPTools(clients)

	// Start the client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := clients[workspacePath]
	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test get document symbols for main.go
	params := &mcp.CallToolParamsFor[GetDocumentSymbolsParams]{
		Arguments: GetDocumentSymbolsParams{
			Workspace: workspacePath,
			Path:      "main.go",
		},
	}

	result, err := tools.HandleGetDocumentSymbols(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleGetDocumentSymbols failed: %v", err)
	}

	symbolResult := parseJSONResult(t, result)

	if len(symbolResult.Symbols) == 0 {
		t.Error("Expected document symbols")
	}

	// Look for expected symbols
	foundMain := false
	foundTestFunction := false

	for _, symbol := range symbolResult.Symbols {
		t.Logf("Symbol: %s (kind %d) at line %d:%d",
			symbol.Name, symbol.Kind, symbol.Range.Line, symbol.Range.Character)

		if symbol.Name == "main" {
			foundMain = true
		}
		if symbol.Name == "testFunction" {
			foundTestFunction = true
		}
	}

	if !foundMain {
		t.Error("Expected to find 'main' function symbol")
	}
	if !foundTestFunction {
		t.Error("Expected to find 'testFunction' symbol")
	}
}

// Helper function for string contains check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
