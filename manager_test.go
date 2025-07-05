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
	manager := NewManager(workspacePath)

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
	manager := NewManager("/test/workspace")

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
	manager := NewManager("/test/workspace")

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
	manager := NewManager("/test/workspace")
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
	manager := NewManager("/test/workspace")

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
	manager := NewManager("/test/workspace")
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

func TestManagerMCPToolHandlersWhenNotRunning(t *testing.T) {
	manager := NewManager("/test/workspace")
	ctx := context.Background()

	// Test HandleGoToDefinition when not running
	params := &mcp.CallToolParamsFor[GoToDefinitionParams]{
		Arguments: GoToDefinitionParams{
			URI:       "file:///test.go",
			Line:      10,
			Character: 5,
		},
	}
	_, err := manager.HandleGoToDefinition(ctx, nil, params)
	if err == nil {
		t.Error("HandleGoToDefinition() on non-running manager should return error")
	}

	// Test HandleFindReferences when not running
	refParams := &mcp.CallToolParamsFor[FindReferencesParams]{
		Arguments: FindReferencesParams{
			URI:                "file:///test.go",
			Line:               10,
			Character:          5,
			IncludeDeclaration: true,
		},
	}
	_, err = manager.HandleFindReferences(ctx, nil, refParams)
	if err == nil {
		t.Error("HandleFindReferences() on non-running manager should return error")
	}

	// Test HandleGetHover when not running
	hoverParams := &mcp.CallToolParamsFor[GetHoverParams]{
		Arguments: GetHoverParams{
			URI:       "file:///test.go",
			Line:      10,
			Character: 5,
		},
	}
	_, err = manager.HandleGetHover(ctx, nil, hoverParams)
	if err == nil {
		t.Error("HandleGetHover() on non-running manager should return error")
	}
}

func TestManagerCreateTools(t *testing.T) {
	manager := NewManager("/test/workspace")

	// Test CreateGoToDefinitionTool
	tool := manager.CreateGoToDefinitionTool()
	if tool == nil {
		t.Error("CreateGoToDefinitionTool() returned nil")
	}

	// Test CreateFindReferencesTool
	tool = manager.CreateFindReferencesTool()
	if tool == nil {
		t.Error("CreateFindReferencesTool() returned nil")
	}

	// Test CreateGetHoverTool
	tool = manager.CreateGetHoverTool()
	if tool == nil {
		t.Error("CreateGetHoverTool() returned nil")
	}
}

func TestGoToDefinitionParams(t *testing.T) {
	params := GoToDefinitionParams{
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
