package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newDebugLogger creates a logger with DEBUG level for integration tests.
func newDebugLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}

// requireGopls skips the test if gopls is not available in PATH.
func requireGopls(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls not found in PATH, skipping integration test")
	}
}

// createTempGoWorkspace creates a temporary directory with a valid Go module.
func createTempGoWorkspace(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "gopls-test-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	// Create go.mod file
	goModContent := `module test-workspace

go 1.21
`
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create simple main.go file
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	result := testFunction()
	fmt.Println("Result:", result)
}

// testFunction is a simple function for testing gopls features
func testFunction() int {
	return 42
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create main.go: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to cleanup temp directory %s: %v", tempDir, err)
		}
	}

	return tempDir, cleanup
}

func TestGoplsClientLifecycle(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	// Test initial state
	if client.isRunning() {
		t.Error("client should not be running initially")
	}

	// Test start
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}

	// Test running state
	if !client.isRunning() {
		t.Error("client should be running after start")
	}

	// Test stop
	if err := client.stop(); err != nil {
		t.Errorf("failed to stop client: %v", err)
	}

	// Test final state
	if client.isRunning() {
		t.Error("client should not be running after stop")
	}
}

func TestGoplsClientDoubleStart(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)
	defer func() { _ = client.stop() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First start should succeed
	if err := client.start(ctx); err != nil {
		t.Fatalf("first start failed: %v", err)
	}

	// Second start should fail
	if err := client.start(ctx); err == nil {
		t.Error("second start should have failed")
	}
}

func TestGoplsClientProcessTracking(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Verify gopls process is actually running
	if client.cmd == nil {
		t.Fatal("client.cmd should not be nil after start")
	}

	if client.cmd.Process == nil {
		t.Fatal("client.cmd.Process should not be nil after start")
	}

	pid := client.cmd.Process.Pid
	if pid <= 0 {
		t.Errorf("invalid PID: %d", pid)
	}

	t.Logf("gopls process started with PID: %d", pid)

	// Verify process is actually alive
	if client.cmd.ProcessState != nil {
		t.Errorf("process should still be running, but ProcessState is: %v", client.cmd.ProcessState)
	}
}

func TestGoplsClientInvalidWorkspace(t *testing.T) {
	requireGopls(t)

	logger := newDebugLogger()
	// Use non-existent workspace
	client := newClient("/non/existent/workspace", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start should fail due to invalid workspace
	if err := client.start(ctx); err == nil {
		defer func() { _ = client.stop() }()
		t.Error("start should have failed with invalid workspace")
	}
}

func TestGoplsClientStopNotRunning(t *testing.T) {
	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	// Stop should not error when client is not running
	if err := client.stop(); err != nil {
		t.Errorf("stop should not error when not running: %v", err)
	}
}

func TestGoplsClientCleanupOnFailure(t *testing.T) {
	requireGopls(t)

	logger := newDebugLogger()
	// Use invalid workspace to trigger initialization failure
	client := newClient("/non/existent/workspace", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start should fail
	err := client.start(ctx)
	if err == nil {
		defer func() { _ = client.stop() }()
		t.Error("start should have failed")
		return
	}

	// Client should not be running after failed start
	if client.isRunning() {
		t.Error("client should not be running after failed start")
	}

	// Process should be cleaned up
	if client.cmd != nil && client.cmd.Process != nil && client.cmd.ProcessState == nil {
		t.Error("process should be cleaned up after failed start")
	}
}

func TestGoplsClientInitialization(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Verify that the client successfully initialized
	if !client.isRunning() {
		t.Error("client should be running after successful initialization")
	}

	// Verify gopls process is responsive
	if client.cmd == nil || client.cmd.Process == nil {
		t.Fatal("gopls process should be running")
	}

	// Give gopls a moment to fully initialize
	time.Sleep(1 * time.Second)

	// Verify process is still alive after initialization
	if client.cmd.ProcessState != nil {
		t.Errorf("gopls process should still be running after initialization, but got: %v", client.cmd.ProcessState)
	}
}

func TestGoplsClientStderrMonitoring(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give some time for stderr monitoring to work
	time.Sleep(2 * time.Second)

	// The test passes if no panics occur and gopls remains running
	if !client.isRunning() {
		t.Error("client should still be running")
	}

	// Note: This test mainly verifies that stderr monitoring doesn't crash
	// the client and that logging works properly. The actual stderr content
	// depends on gopls internal behavior.
}

func TestGoplsClientGoToDefinition(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize and load the workspace
	time.Sleep(3 * time.Second)

	// Test go to definition on "testFunction" call in main.go
	// The function call is on line 6 (0-based), character 11 (start of "testFunction")
	// result := testFunction()
	//           ^-- position 11
	locations, err := client.goToDefinition("main.go", 6, 11) // Position of "testFunction" call
	if err != nil {
		t.Fatalf("goToDefinition failed: %v", err)
	}

	// Verify exactly one definition location is returned
	if len(locations) != 1 {
		t.Fatalf("expected exactly 1 definition location, got %d", len(locations))
	}

	// Verify the location points to the exact function definition
	location := locations[0]
	if location.URI != "main.go" {
		t.Errorf("expected URI 'main.go', got '%s'", location.URI)
	}

	// The function definition should be exactly at line 11 (0-based), character 5
	// func testFunction() int {
	//      ^-- position 5
	expectedLine := 11
	expectedChar := 5
	if location.Range.Start.Line != expectedLine {
		t.Errorf("expected definition at line %d, got %d", expectedLine, location.Range.Start.Line)
	}
	if location.Range.Start.Character != expectedChar {
		t.Errorf("expected definition at character %d, got %d", expectedChar, location.Range.Start.Character)
	}

	t.Logf("Successfully found definition at %s:%d:%d (exact match)",
		location.URI, location.Range.Start.Line, location.Range.Start.Character)

	// Test go to definition on function definition itself (should return same location)
	defLocations, err := client.goToDefinition("main.go", 11, 5) // Position of "testFunction" definition
	if err != nil {
		t.Fatalf("goToDefinition on definition failed: %v", err)
	}

	if len(defLocations) != 1 {
		t.Fatalf("expected exactly 1 definition location from definition position, got %d", len(defLocations))
	}

	// Should return the same location when called on the definition itself
	defLocation := defLocations[0]
	if defLocation.URI != location.URI ||
		defLocation.Range.Start.Line != location.Range.Start.Line ||
		defLocation.Range.Start.Character != location.Range.Start.Character {
		t.Errorf("definition position should return same location as call position")
		t.Errorf("  from call: %s:%d:%d", location.URI, location.Range.Start.Line, location.Range.Start.Character)
		t.Errorf("  from def:  %s:%d:%d", defLocation.URI, defLocation.Range.Start.Line, defLocation.Range.Start.Character)
	}

	// Test go to definition on non-existent symbol (should error or return empty)
	_, err = client.goToDefinition("main.go", 0, 0) // Position with no symbol
	if err == nil {
		t.Log("goToDefinition on empty position succeeded (gopls behavior may vary)")
	} else {
		t.Logf("goToDefinition on empty position failed as expected: %v", err)
	}
}

func TestGoplsClientFindReferences(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize and load the workspace
	time.Sleep(3 * time.Second)

	// Test find references on "testFunction" definition in main.go
	// The function definition is on line 11 (0-based), character 5 (start of "testFunction")
	// func testFunction() int {
	//      ^-- position 5
	locationsWithDecl, err := client.findReferences("main.go", 11, 5, true) // Include declaration
	if err != nil {
		t.Fatalf("findReferences with declaration failed: %v", err)
	}

	// Should find exactly 2 references: definition + call
	expectedWithDecl := 2
	if len(locationsWithDecl) != expectedWithDecl {
		t.Errorf("expected exactly %d references with declaration, got %d", expectedWithDecl, len(locationsWithDecl))
	}

	// Test find references without declaration
	locationsWithoutDecl, err := client.findReferences("main.go", 11, 5, false) // Exclude declaration
	if err != nil {
		t.Fatalf("findReferences without declaration failed: %v", err)
	}

	// Should find exactly 1 reference: call only
	expectedWithoutDecl := 1
	if len(locationsWithoutDecl) != expectedWithoutDecl {
		t.Errorf("expected exactly %d references without declaration, got %d", expectedWithoutDecl, len(locationsWithoutDecl))
	}

	// Verify that including declaration gives exactly one more result
	if len(locationsWithDecl) != len(locationsWithoutDecl)+1 {
		t.Errorf("expected declaration to add exactly 1 reference, got %d with declaration vs %d without",
			len(locationsWithDecl), len(locationsWithoutDecl))
	}

	// Verify all locations point to main.go
	for i, location := range locationsWithDecl {
		if location.URI != "main.go" {
			t.Errorf("location %d: expected URI 'main.go', got '%s'", i, location.URI)
		}
	}

	// Verify exact positions of references
	// Expected positions: definition at (11,5) and call at (6,11)
	expectedPositions := []struct {
		line int
		char int
		desc string
	}{
		{11, 5, "definition"},
		{6, 11, "call"},
	}

	// Verify exact positions of references exist in the results
	// (order might vary so we check if all expected positions are present)
	actualPositions := make(map[string]bool)
	for _, loc := range locationsWithDecl {
		key := fmt.Sprintf("%d:%d", loc.Range.Start.Line, loc.Range.Start.Character)
		actualPositions[key] = true
	}

	// Verify each expected position is found
	for _, expectedPos := range expectedPositions {
		key := fmt.Sprintf("%d:%d", expectedPos.line, expectedPos.char)
		if !actualPositions[key] {
			t.Errorf("missing expected reference (%s) at line %d, character %d",
				expectedPos.desc, expectedPos.line, expectedPos.char)
		} else {
			t.Logf("Found expected reference (%s): %s:%d:%d ✓", expectedPos.desc,
				"main.go", expectedPos.line, expectedPos.char)
		}
	}

	// Test find references on the function call in main.go
	// The function call is on line 6 (0-based), character 11 (start of "testFunction")
	// result := testFunction()
	//           ^-- position 11
	callLocations, err := client.findReferences("main.go", 6, 11, true) // Position of "testFunction" call
	if err != nil {
		t.Fatalf("findReferences on call failed: %v", err)
	}

	// Should return same results as finding references from definition
	if len(callLocations) != len(locationsWithDecl) {
		t.Errorf("expected same number of references from call position (%d) as from definition (%d)",
			len(callLocations), len(locationsWithDecl))
	}

	// Verify positions match (order might differ but content should be same)
	if len(callLocations) == len(locationsWithDecl) {
		foundPositions := make(map[string]bool)
		for _, loc := range callLocations {
			key := fmt.Sprintf("%d:%d", loc.Range.Start.Line, loc.Range.Start.Character)
			foundPositions[key] = true
		}

		for _, expectedPos := range expectedPositions {
			key := fmt.Sprintf("%d:%d", expectedPos.line, expectedPos.char)
			if !foundPositions[key] {
				t.Errorf("missing expected reference position %s in call results", key)
			}
		}
	}

	t.Logf("Successfully found %d references with declaration, %d without declaration",
		len(locationsWithDecl), len(locationsWithoutDecl))
}

// testHoverOnFunction tests hover functionality on function calls and definitions.
func testHoverOnFunction(t *testing.T, client *goplsClient) {
	t.Helper()

	// Test hover on "testFunction" call in main.go
	// The function call is on line 6 (0-based), character 11 (start of "testFunction")
	// result := testFunction()
	//           ^-- position 11
	hoverCall, err := client.getHover("main.go", 6, 11) // Position of "testFunction" call
	if err != nil {
		t.Fatalf("getHover on function call failed: %v", err)
	}

	if hoverCall == nil {
		t.Fatal("expected hover info, got nil")
	}

	if len(hoverCall.Contents) == 0 {
		t.Fatal("expected hover contents, got empty")
	}

	// Verify hover contains function signature
	foundFunctionSignature := false
	for _, content := range hoverCall.Contents {
		if strings.Contains(content, "func testFunction() int") {
			foundFunctionSignature = true
			break
		}
	}
	if !foundFunctionSignature {
		t.Errorf("expected hover to contain function signature 'func testFunction() int', got: %+v", hoverCall.Contents)
	}

	t.Logf("Hover on function call: %+v", hoverCall.Contents)

	// Test hover on "testFunction" definition in main.go
	// The function definition is on line 11 (0-based), character 5 (start of "testFunction")
	// func testFunction() int {
	//      ^-- position 5
	hoverDef, err := client.getHover("main.go", 11, 5) // Position of "testFunction" definition
	if err != nil {
		t.Fatalf("getHover on function definition failed: %v", err)
	}

	if hoverDef == nil {
		t.Fatal("expected hover info, got nil")
	}

	if len(hoverDef.Contents) == 0 {
		t.Fatal("expected hover contents, got empty")
	}

	// Verify hover contains function signature (should be same as call)
	foundFunctionSignature = false
	for _, content := range hoverDef.Contents {
		if strings.Contains(content, "func testFunction() int") {
			foundFunctionSignature = true
			break
		}
	}
	if !foundFunctionSignature {
		t.Errorf("expected hover to contain function signature 'func testFunction() int', got: %+v", hoverDef.Contents)
	}

	t.Logf("Hover on function definition: %+v", hoverDef.Contents)
}

// testHoverOnPackage tests hover functionality on package references.
func testHoverOnPackage(t *testing.T, client *goplsClient) {
	t.Helper()

	// Test hover on "fmt.Println" call in main.go
	// The fmt.Println call is on line 5 (0-based), character 1 (start of "fmt")
	// fmt.Println("Hello, World!")
	// ^-- position 1
	hoverFmt, err := client.getHover("main.go", 5, 1) // Position of "fmt" in fmt.Println
	if err != nil {
		t.Fatalf("getHover on fmt failed: %v", err)
	}

	if hoverFmt == nil {
		t.Fatal("expected hover info, got nil")
	}

	if len(hoverFmt.Contents) == 0 {
		t.Fatal("expected hover contents, got empty")
	}

	// Verify hover contains package information
	foundPackageInfo := false
	for _, content := range hoverFmt.Contents {
		if strings.Contains(content, "package fmt") || strings.Contains(content, "fmt") {
			foundPackageInfo = true
			break
		}
	}
	if !foundPackageInfo {
		t.Errorf("expected hover to contain package information about 'fmt', got: %+v", hoverFmt.Contents)
	}

	t.Logf("Hover on fmt: %+v", hoverFmt.Contents)
}

// testHoverOnVariableAndLiteral tests hover functionality on variables and literals.
func testHoverOnVariableAndLiteral(t *testing.T, client *goplsClient) {
	t.Helper()

	// Test hover on "result" variable in main.go
	// The result variable is on line 7 (0-based), character 19 (position of "result")
	// fmt.Println("Result:", result)
	//                       ^-- position 19
	hoverResult, err := client.getHover("main.go", 7, 19) // Position of "result" variable
	if err != nil {
		t.Fatalf("getHover on result variable failed: %v", err)
	}

	switch {
	case hoverResult == nil:
		t.Log("No hover info for result variable (gopls behavior may vary)")
	case len(hoverResult.Contents) == 0:
		t.Log("Empty hover contents for result variable (gopls behavior may vary)")
	default:
		// Verify hover contains variable type information if present
		foundVariableType := false
		for _, content := range hoverResult.Contents {
			if strings.Contains(content, "int") || strings.Contains(content, "result") {
				foundVariableType = true
				break
			}
		}
		if !foundVariableType {
			t.Logf("Hover for result variable doesn't contain expected type info, got: %+v", hoverResult.Contents)
		} else {
			t.Logf("Hover on result variable: %+v", hoverResult.Contents)
		}
	}

	// Test hover on return value (42) in testFunction
	// The return value is on line 12 (0-based), character 8 (start of "42")
	// return 42
	//        ^-- position 8
	hoverReturnValue, err := client.getHover("main.go", 12, 8) // Position of "42"
	if err != nil {
		t.Fatalf("getHover on return value failed: %v", err)
	}

	switch {
	case hoverReturnValue == nil:
		t.Log("No hover info for return value (gopls behavior may vary)")
	case len(hoverReturnValue.Contents) == 0:
		t.Log("Empty hover contents for return value (gopls behavior may vary)")
	default:
		// Verify hover contains numeric literal information if present
		foundNumericInfo := false
		for _, content := range hoverReturnValue.Contents {
			if strings.Contains(content, "42") || strings.Contains(content, "int") {
				foundNumericInfo = true
				break
			}
		}
		if !foundNumericInfo {
			t.Logf("Hover for return value doesn't contain expected numeric info, got: %+v", hoverReturnValue.Contents)
		} else {
			t.Logf("Hover on return value: %+v", hoverReturnValue.Contents)
		}
	}
}

// testHoverErrorCases tests hover functionality on invalid positions and files.
func testHoverErrorCases(t *testing.T, client *goplsClient) {
	t.Helper()

	// Test hover on invalid position (may error or return nil/empty)
	hoverInvalid, err := client.getHover("main.go", 100, 100) // Invalid position
	if err != nil {
		// This is acceptable - gopls may return an error for invalid positions
		t.Logf("getHover on invalid position returned error (expected): %v", err)
	} else {
		// This is also acceptable - gopls might return nil or empty hover for invalid positions
		t.Logf("Hover on invalid position: %+v", hoverInvalid)
	}

	// Test hover on non-existent file (should error)
	_, err = client.getHover("nonexistent.go", 0, 0)
	if err == nil {
		t.Error("getHover on non-existent file should have failed")
	}
}

func TestGoplsClientGetHover(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize and load the workspace
	time.Sleep(3 * time.Second)

	// Run test suites
	testHoverOnFunction(t, client)
	testHoverOnPackage(t, client)
	testHoverOnVariableAndLiteral(t, client)
	testHoverErrorCases(t, client)

	t.Logf("getHover tests completed successfully")
}
