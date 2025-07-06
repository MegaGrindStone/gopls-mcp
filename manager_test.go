package main

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewManager(t *testing.T) {
	workspacePath := "/test/workspace"
	logger := newTestLogger()
	manager := NewManager(workspacePath, logger)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.workspacePath != workspacePath {
		t.Errorf("NewManager() workspacePath = %v, want %v", manager.workspacePath, workspacePath)
	}

	if manager.running {
		t.Error("NewManager() created manager with running = true, want false")
	}

	if manager.requestID != 0 {
		t.Errorf("NewManager() created manager with requestID = %v, want 0", manager.requestID)
	}
}

func TestManagerIsRunning(t *testing.T) {
	logger := newTestLogger()
	manager := NewManager("/test/workspace", logger)

	// Initially not running
	if manager.IsRunning() {
		t.Error("IsRunning() = true, want false for new manager")
	}

	// Simulate setting running to true
	manager.mu.Lock()
	manager.running = true
	manager.mu.Unlock()

	if !manager.IsRunning() {
		t.Error("IsRunning() = false, want true after setting running = true")
	}

	// Set back to false
	manager.mu.Lock()
	manager.running = false
	manager.mu.Unlock()

	if manager.IsRunning() {
		t.Error("IsRunning() = true, want false after setting running = false")
	}
}

func TestManagerNextRequestID(t *testing.T) {
	logger := newTestLogger()
	manager := NewManager("/test/workspace", logger)

	// Test sequential ID generation
	id1 := manager.nextRequestID()
	id2 := manager.nextRequestID()
	id3 := manager.nextRequestID()

	if id1 != 1 {
		t.Errorf("nextRequestID() first call = %v, want 1", id1)
	}
	if id2 != 2 {
		t.Errorf("nextRequestID() second call = %v, want 2", id2)
	}
	if id3 != 3 {
		t.Errorf("nextRequestID() third call = %v, want 3", id3)
	}
}

func TestManagerNextRequestIDConcurrency(t *testing.T) {
	logger := newTestLogger()
	manager := NewManager("/test/workspace", logger)
	const numGoroutines = 100
	const numCalls = 100

	var wg sync.WaitGroup
	ids := make([]int, numGoroutines*numCalls)
	var mu sync.Mutex

	// Launch multiple goroutines to test thread safety
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				id := manager.nextRequestID()
				mu.Lock()
				ids[goroutineID*numCalls+j] = id
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Check that all IDs are unique
	idSet := make(map[int]bool)
	for _, id := range ids {
		if idSet[id] {
			t.Errorf("nextRequestID() generated duplicate ID: %v", id)
		}
		idSet[id] = true
	}

	// Check that we got the expected number of unique IDs
	if len(idSet) != numGoroutines*numCalls {
		t.Errorf("nextRequestID() generated %v unique IDs, want %v", len(idSet), numGoroutines*numCalls)
	}
}

func TestManagerStop(t *testing.T) {
	logger := newTestLogger()
	manager := NewManager("/test/workspace", logger)

	// Test stopping when not running
	err := manager.Stop()
	if err != nil {
		t.Errorf("Stop() on non-running manager returned error: %v", err)
	}

	// Test stopping when running (without actually starting gopls)
	manager.mu.Lock()
	manager.running = true
	manager.mu.Unlock()

	err = manager.Stop()
	if err != nil {
		t.Errorf("Stop() on running manager returned error: %v", err)
	}

	if manager.IsRunning() {
		t.Error("Stop() did not set running to false")
	}
}

func TestManagerLSPMethodsWhenNotRunning(t *testing.T) {
	logger := newTestLogger()
	manager := NewManager("/test/workspace", logger)
	ctx := context.Background()

	// Test GoToDefinition when not running
	_, err := manager.GoToDefinition(ctx, "file:///test.go", 10, 5)
	if err == nil {
		t.Error("GoToDefinition() on non-running manager should return error")
	}

	// Test FindReferences when not running
	_, err = manager.FindReferences(ctx, "file:///test.go", 10, 5, true)
	if err == nil {
		t.Error("FindReferences() on non-running manager should return error")
	}

	// Test GetHover when not running
	_, err = manager.GetHover(ctx, "file:///test.go", 10, 5)
	if err == nil {
		t.Error("GetHover() on non-running manager should return error")
	}
}

func TestWorkspaceManagerMCPToolHandlersWhenNotRunning(t *testing.T) {
	logger := newTestLogger()
	workspaces := []string{"/test/workspace1", "/test/workspace2"}
	workspaceManager := NewWorkspaceManager(workspaces, logger)
	ctx := context.Background()

	// Test HandleGoToDefinition when not running
	params := &mcp.CallToolParamsFor[GoToDefinitionParams]{
		Arguments: GoToDefinitionParams{
			Workspace: "/test/workspace1",
			URI:       "file:///test.go",
			Line:      10,
			Character: 5,
		},
	}
	_, err := workspaceManager.HandleGoToDefinition(ctx, nil, params)
	if err == nil {
		t.Error("HandleGoToDefinition() on non-running workspace manager should return error")
	}

	// Test HandleFindReferences when not running
	refParams := &mcp.CallToolParamsFor[FindReferencesParams]{
		Arguments: FindReferencesParams{
			Workspace:          "/test/workspace1",
			URI:                "file:///test.go",
			Line:               10,
			Character:          5,
			IncludeDeclaration: true,
		},
	}
	_, err = workspaceManager.HandleFindReferences(ctx, nil, refParams)
	if err == nil {
		t.Error("HandleFindReferences() on non-running workspace manager should return error")
	}

	// Test HandleGetHover when not running
	hoverParams := &mcp.CallToolParamsFor[GetHoverParams]{
		Arguments: GetHoverParams{
			Workspace: "/test/workspace1",
			URI:       "file:///test.go",
			Line:      10,
			Character: 5,
		},
	}
	_, err = workspaceManager.HandleGetHover(ctx, nil, hoverParams)
	if err == nil {
		t.Error("HandleGetHover() on non-running workspace manager should return error")
	}

	// Test HandleListWorkspaces
	listParams := &mcp.CallToolParamsFor[ListWorkspacesParams]{
		Arguments: ListWorkspacesParams{},
	}
	result, err := workspaceManager.HandleListWorkspaces(ctx, nil, listParams)
	if err != nil {
		t.Errorf("HandleListWorkspaces() returned error: %v", err)
	}
	if result == nil {
		t.Error("HandleListWorkspaces() returned nil result")
	}

	// Test with nonexistent workspace
	badParams := &mcp.CallToolParamsFor[GoToDefinitionParams]{
		Arguments: GoToDefinitionParams{
			Workspace: "/nonexistent/workspace",
			URI:       "file:///test.go",
			Line:      10,
			Character: 5,
		},
	}
	_, err = workspaceManager.HandleGoToDefinition(ctx, nil, badParams)
	if err == nil {
		t.Error("HandleGoToDefinition() with nonexistent workspace should return error")
	}
}

func TestWorkspaceManagerCreateTools(t *testing.T) {
	logger := newTestLogger()
	workspaces := []string{"/test/workspace1", "/test/workspace2"}
	workspaceManager := NewWorkspaceManager(workspaces, logger)

	// Test CreateGoToDefinitionTool
	tool := workspaceManager.CreateGoToDefinitionTool()
	if tool == nil {
		t.Error("CreateGoToDefinitionTool() returned nil")
	}

	// Test CreateFindReferencesTool
	tool = workspaceManager.CreateFindReferencesTool()
	if tool == nil {
		t.Error("CreateFindReferencesTool() returned nil")
	}

	// Test CreateGetHoverTool
	tool = workspaceManager.CreateGetHoverTool()
	if tool == nil {
		t.Error("CreateGetHoverTool() returned nil")
	}

	// Test CreateListWorkspacesTool
	tool = workspaceManager.CreateListWorkspacesTool()
	if tool == nil {
		t.Error("CreateListWorkspacesTool() returned nil")
	}
}

func TestGoToDefinitionParams(t *testing.T) {
	params := GoToDefinitionParams{
		Workspace: "/test/workspace",
		URI:       "file:///test.go",
		Line:      10,
		Character: 5,
	}

	// Test JSON marshaling
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal(GoToDefinitionParams) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GoToDefinitionParams
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GoToDefinitionParams) error = %v", err)
	}

	if unmarshaled != params {
		t.Errorf("JSON roundtrip failed: got %v, want %v", unmarshaled, params)
	}
}

func TestFindReferencesParams(t *testing.T) {
	params := FindReferencesParams{
		Workspace:          "/test/workspace",
		URI:                "file:///test.go",
		Line:               10,
		Character:          5,
		IncludeDeclaration: true,
	}

	// Test JSON marshaling
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal(FindReferencesParams) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled FindReferencesParams
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(FindReferencesParams) error = %v", err)
	}

	if unmarshaled != params {
		t.Errorf("JSON roundtrip failed: got %v, want %v", unmarshaled, params)
	}
}

func TestGetHoverParams(t *testing.T) {
	params := GetHoverParams{
		Workspace: "/test/workspace",
		URI:       "file:///test.go",
		Line:      10,
		Character: 5,
	}

	// Test JSON marshaling
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal(GetHoverParams) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GetHoverParams
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GetHoverParams) error = %v", err)
	}

	if unmarshaled != params {
		t.Errorf("JSON roundtrip failed: got %v, want %v", unmarshaled, params)
	}
}

func TestLocationResult(t *testing.T) {
	result := LocationResult{
		URI:          "file:///test.go",
		Line:         10,
		Character:    5,
		EndLine:      10,
		EndCharacter: 15,
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(LocationResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled LocationResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(LocationResult) error = %v", err)
	}

	if unmarshaled != result {
		t.Errorf("JSON roundtrip failed: got %v, want %v", unmarshaled, result)
	}
}

func TestGoToDefinitionResult(t *testing.T) {
	result := GoToDefinitionResult{
		Locations: []LocationResult{
			{
				URI:          "file:///test.go",
				Line:         10,
				Character:    5,
				EndLine:      10,
				EndCharacter: 15,
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(GoToDefinitionResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GoToDefinitionResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GoToDefinitionResult) error = %v", err)
	}

	if len(unmarshaled.Locations) != len(result.Locations) {
		t.Errorf("JSON roundtrip failed: got %d locations, want %d", len(unmarshaled.Locations), len(result.Locations))
	}
}

func TestFindReferencesResult(t *testing.T) {
	result := FindReferencesResult{
		Locations: []LocationResult{
			{
				URI:          "file:///test.go",
				Line:         10,
				Character:    5,
				EndLine:      10,
				EndCharacter: 15,
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(FindReferencesResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled FindReferencesResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(FindReferencesResult) error = %v", err)
	}

	if len(unmarshaled.Locations) != len(result.Locations) {
		t.Errorf("JSON roundtrip failed: got %d locations, want %d", len(unmarshaled.Locations), len(result.Locations))
	}
}

func TestGetHoverResult(t *testing.T) {
	result := GetHoverResult{
		Contents: []string{"func example()", "Example function"},
		HasRange: true,
		Range: &LocationResult{
			URI:          "file:///test.go",
			Line:         10,
			Character:    5,
			EndLine:      10,
			EndCharacter: 15,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(GetHoverResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GetHoverResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GetHoverResult) error = %v", err)
	}

	if len(unmarshaled.Contents) != len(result.Contents) {
		t.Errorf("JSON roundtrip failed: got %d contents, want %d", len(unmarshaled.Contents), len(result.Contents))
	}

	if unmarshaled.HasRange != result.HasRange {
		t.Errorf("JSON roundtrip failed: got HasRange=%v, want %v", unmarshaled.HasRange, result.HasRange)
	}
}

func TestGetDocumentSymbolsParams(t *testing.T) {
	params := GetDocumentSymbolsParams{
		Workspace: "/test/workspace",
		URI:       "file:///test.go",
	}

	// Test JSON marshaling
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal(GetDocumentSymbolsParams) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GetDocumentSymbolsParams
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GetDocumentSymbolsParams) error = %v", err)
	}

	if unmarshaled != params {
		t.Errorf("JSON roundtrip failed: got %v, want %v", unmarshaled, params)
	}
}

func TestSearchWorkspaceSymbolsParams(t *testing.T) {
	params := SearchWorkspaceSymbolsParams{
		Workspace: "/test/workspace",
		Query:     "TestFunction",
	}

	// Test JSON marshaling
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal(SearchWorkspaceSymbolsParams) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SearchWorkspaceSymbolsParams
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(SearchWorkspaceSymbolsParams) error = %v", err)
	}

	if unmarshaled != params {
		t.Errorf("JSON roundtrip failed: got %v, want %v", unmarshaled, params)
	}
}

func TestGoToTypeDefinitionParams(t *testing.T) {
	params := GoToTypeDefinitionParams{
		Workspace: "/test/workspace",
		URI:       "file:///test.go",
		Line:      10,
		Character: 5,
	}

	// Test JSON marshaling
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal(GoToTypeDefinitionParams) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GoToTypeDefinitionParams
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GoToTypeDefinitionParams) error = %v", err)
	}

	if unmarshaled != params {
		t.Errorf("JSON roundtrip failed: got %v, want %v", unmarshaled, params)
	}
}

func TestDocumentSymbolResult(t *testing.T) {
	result := DocumentSymbolResult{
		Name:       "TestFunction",
		Detail:     "func TestFunction()",
		Kind:       12, // Function
		Deprecated: false,
		Range: LocationResult{
			URI:          "file:///test.go",
			Line:         10,
			Character:    0,
			EndLine:      15,
			EndCharacter: 1,
		},
		SelectionRange: LocationResult{
			URI:          "file:///test.go",
			Line:         10,
			Character:    5,
			EndLine:      10,
			EndCharacter: 17,
		},
		Children: []DocumentSymbolResult{
			{
				Name: "LocalVar",
				Kind: 13, // Variable
				Range: LocationResult{
					URI:          "file:///test.go",
					Line:         11,
					Character:    4,
					EndLine:      11,
					EndCharacter: 12,
				},
				SelectionRange: LocationResult{
					URI:          "file:///test.go",
					Line:         11,
					Character:    4,
					EndLine:      11,
					EndCharacter: 12,
				},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(DocumentSymbolResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled DocumentSymbolResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(DocumentSymbolResult) error = %v", err)
	}

	if len(unmarshaled.Children) != len(result.Children) {
		t.Errorf("JSON roundtrip failed: got %d children, want %d", len(unmarshaled.Children), len(result.Children))
	}
}

func TestGetDocumentSymbolsResult(t *testing.T) {
	result := GetDocumentSymbolsResult{
		Symbols: []DocumentSymbolResult{
			{
				Name: "TestFunction",
				Kind: 12, // Function
				Range: LocationResult{
					URI:          "file:///test.go",
					Line:         10,
					Character:    0,
					EndLine:      15,
					EndCharacter: 1,
				},
				SelectionRange: LocationResult{
					URI:          "file:///test.go",
					Line:         10,
					Character:    5,
					EndLine:      10,
					EndCharacter: 17,
				},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(GetDocumentSymbolsResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GetDocumentSymbolsResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GetDocumentSymbolsResult) error = %v", err)
	}

	if len(unmarshaled.Symbols) != len(result.Symbols) {
		t.Errorf("JSON roundtrip failed: got %d symbols, want %d", len(unmarshaled.Symbols), len(result.Symbols))
	}
}

func TestWorkspaceSymbolResult(t *testing.T) {
	result := WorkspaceSymbolResult{
		Name:       "TestStruct",
		Kind:       23, // Struct
		Deprecated: false,
		Location: LocationResult{
			URI:          "file:///test.go",
			Line:         5,
			Character:    0,
			EndLine:      10,
			EndCharacter: 1,
		},
		ContainerName: "main",
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(WorkspaceSymbolResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled WorkspaceSymbolResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(WorkspaceSymbolResult) error = %v", err)
	}

	if unmarshaled.Name != result.Name {
		t.Errorf("JSON roundtrip failed: got name=%v, want %v", unmarshaled.Name, result.Name)
	}
	if unmarshaled.ContainerName != result.ContainerName {
		t.Errorf("JSON roundtrip failed: got containerName=%v, want %v", unmarshaled.ContainerName, result.ContainerName)
	}
}

func TestSearchWorkspaceSymbolsResult(t *testing.T) {
	result := SearchWorkspaceSymbolsResult{
		Symbols: []WorkspaceSymbolResult{
			{
				Name: "TestStruct",
				Kind: 23, // Struct
				Location: LocationResult{
					URI:          "file:///test.go",
					Line:         5,
					Character:    0,
					EndLine:      10,
					EndCharacter: 1,
				},
				ContainerName: "main",
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(SearchWorkspaceSymbolsResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SearchWorkspaceSymbolsResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(SearchWorkspaceSymbolsResult) error = %v", err)
	}

	if len(unmarshaled.Symbols) != len(result.Symbols) {
		t.Errorf("JSON roundtrip failed: got %d symbols, want %d", len(unmarshaled.Symbols), len(result.Symbols))
	}
}

func TestGoToTypeDefinitionResult(t *testing.T) {
	result := GoToTypeDefinitionResult{
		Locations: []LocationResult{
			{
				URI:          "file:///test.go",
				Line:         10,
				Character:    5,
				EndLine:      10,
				EndCharacter: 15,
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("json.Marshal(GoToTypeDefinitionResult) error = %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled GoToTypeDefinitionResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("json.Unmarshal(GoToTypeDefinitionResult) error = %v", err)
	}

	if len(unmarshaled.Locations) != len(result.Locations) {
		t.Errorf("JSON roundtrip failed: got %d locations, want %d", len(unmarshaled.Locations), len(result.Locations))
	}
}

func TestWorkspaceManagerMCPNewToolHandlersWhenNotRunning(t *testing.T) {
	logger := newTestLogger()
	workspaces := []string{"/test/workspace1", "/test/workspace2"}
	workspaceManager := NewWorkspaceManager(workspaces, logger)
	ctx := context.Background()

	// Test HandleGetDocumentSymbols when not running
	docParams := &mcp.CallToolParamsFor[GetDocumentSymbolsParams]{
		Arguments: GetDocumentSymbolsParams{
			Workspace: "/test/workspace1",
			URI:       "file:///test.go",
		},
	}
	_, err := workspaceManager.HandleGetDocumentSymbols(ctx, nil, docParams)
	if err == nil {
		t.Error("HandleGetDocumentSymbols() on non-running workspace manager should return error")
	}

	// Test HandleSearchWorkspaceSymbols when not running
	searchParams := &mcp.CallToolParamsFor[SearchWorkspaceSymbolsParams]{
		Arguments: SearchWorkspaceSymbolsParams{
			Workspace: "/test/workspace1",
			Query:     "TestFunction",
		},
	}
	_, err = workspaceManager.HandleSearchWorkspaceSymbols(ctx, nil, searchParams)
	if err == nil {
		t.Error("HandleSearchWorkspaceSymbols() on non-running workspace manager should return error")
	}

	// Test HandleGoToTypeDefinition when not running
	typeParams := &mcp.CallToolParamsFor[GoToTypeDefinitionParams]{
		Arguments: GoToTypeDefinitionParams{
			Workspace: "/test/workspace1",
			URI:       "file:///test.go",
			Line:      10,
			Character: 5,
		},
	}
	_, err = workspaceManager.HandleGoToTypeDefinition(ctx, nil, typeParams)
	if err == nil {
		t.Error("HandleGoToTypeDefinition() on non-running workspace manager should return error")
	}

	// Test with nonexistent workspace
	badDocParams := &mcp.CallToolParamsFor[GetDocumentSymbolsParams]{
		Arguments: GetDocumentSymbolsParams{
			Workspace: "/nonexistent/workspace",
			URI:       "file:///test.go",
		},
	}
	_, err = workspaceManager.HandleGetDocumentSymbols(ctx, nil, badDocParams)
	if err == nil {
		t.Error("HandleGetDocumentSymbols() with nonexistent workspace should return error")
	}

	badSearchParams := &mcp.CallToolParamsFor[SearchWorkspaceSymbolsParams]{
		Arguments: SearchWorkspaceSymbolsParams{
			Workspace: "/nonexistent/workspace",
			Query:     "TestFunction",
		},
	}
	_, err = workspaceManager.HandleSearchWorkspaceSymbols(ctx, nil, badSearchParams)
	if err == nil {
		t.Error("HandleSearchWorkspaceSymbols() with nonexistent workspace should return error")
	}

	badTypeParams := &mcp.CallToolParamsFor[GoToTypeDefinitionParams]{
		Arguments: GoToTypeDefinitionParams{
			Workspace: "/nonexistent/workspace",
			URI:       "file:///test.go",
			Line:      10,
			Character: 5,
		},
	}
	_, err = workspaceManager.HandleGoToTypeDefinition(ctx, nil, badTypeParams)
	if err == nil {
		t.Error("HandleGoToTypeDefinition() with nonexistent workspace should return error")
	}
}

func TestWorkspaceManagerCreateNewTools(t *testing.T) {
	logger := newTestLogger()
	workspaces := []string{"/test/workspace1", "/test/workspace2"}
	workspaceManager := NewWorkspaceManager(workspaces, logger)

	// Test CreateGetDocumentSymbolsTool
	tool := workspaceManager.CreateGetDocumentSymbolsTool()
	if tool == nil {
		t.Error("CreateGetDocumentSymbolsTool() returned nil")
	}

	// Test CreateSearchWorkspaceSymbolsTool
	tool = workspaceManager.CreateSearchWorkspaceSymbolsTool()
	if tool == nil {
		t.Error("CreateSearchWorkspaceSymbolsTool() returned nil")
	}

	// Test CreateGoToTypeDefinitionTool
	tool = workspaceManager.CreateGoToTypeDefinitionTool()
	if tool == nil {
		t.Error("CreateGoToTypeDefinitionTool() returned nil")
	}
}

func TestManagerNewLSPMethodsWhenNotRunning(t *testing.T) {
	logger := newTestLogger()
	manager := NewManager("/test/workspace", logger)
	ctx := context.Background()

	// Test GetDocumentSymbols when not running
	_, err := manager.GetDocumentSymbols(ctx, "file:///test.go")
	if err == nil {
		t.Error("GetDocumentSymbols() on non-running manager should return error")
	}

	// Test SearchWorkspaceSymbols when not running
	_, err = manager.SearchWorkspaceSymbols(ctx, "TestFunction")
	if err == nil {
		t.Error("SearchWorkspaceSymbols() on non-running manager should return error")
	}

	// Test GoToTypeDefinition when not running
	_, err = manager.GoToTypeDefinition(ctx, "file:///test.go", 10, 5)
	if err == nil {
		t.Error("GoToTypeDefinition() on non-running manager should return error")
	}
}
