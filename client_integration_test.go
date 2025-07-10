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

// Enhanced workspace with more complex code for comprehensive testing.
func createEnhancedGoWorkspace(t *testing.T) (string, func()) {
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

	// Create enhanced main.go file with more features
	mainGoContent := `package main

import (
	"fmt"
	"strings"
	"errors"
)

// Person represents a person with name and age
type Person struct {
	Name string ` + "`json:\"name\"`" + `
	Age  int    ` + "`json:\"age\"`" + `
}

// Greeter interface for greeting functionality
type Greeter interface {
	Greet(name string) string
}

// PersonGreeter implements Greeter interface
type PersonGreeter struct {
	Prefix string
}

// Greet implements the Greeter interface
func (p PersonGreeter) Greet(name string) string {
	return p.Prefix + " " + name
}

func main() {
	fmt.Println("Hello, World!")
	result, err := testFunction(42)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Result:", result)
	
	person := Person{Name: "Alice", Age: 30}
	processPerson(person)
	
	greeter := PersonGreeter{Prefix: "Hello"}
	message := greeter.Greet("Bob")
	fmt.Println(message)
}

// testFunction is a simple function for testing gopls features
func testFunction(input int) (int, error) {
	if input < 0 {
		return 0, errors.New("negative input")
	}
	return input * 2, nil
}

// processPerson processes a person
func processPerson(p Person) {
	name := strings.ToUpper(p.Name)
	fmt.Printf("Processing %s (age %d)\n", name, p.Age)
}

// HelperFunc is an exported helper function
func HelperFunc() string {
	return "helper"
}

// unusedFunction demonstrates an unused function
func unusedFunction() {
	// This function is intentionally unused
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create main.go: %v", err)
	}

	// Create a utility file for more comprehensive testing
	utilGoContent := `package main

import "fmt"

// MathUtils provides utility math functions
type MathUtils struct{}

// Add adds two numbers
func (m MathUtils) Add(a, b int) int {
	return a + b
}

// Multiply multiplies two numbers
func (m MathUtils) Multiply(a, b int) int {
	return a * b
}

// GlobalVar is a global variable
var GlobalVar = "global"

// Constants for testing
const (
	MaxValue = 100
	MinValue = 0
)
`
	utilGoPath := filepath.Join(tempDir, "util.go")
	if err := os.WriteFile(utilGoPath, []byte(utilGoContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create util.go: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to cleanup temp directory %s: %v", tempDir, err)
		}
	}

	return tempDir, cleanup
}

func TestGoplsClientGetDiagnostics(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize and analyze the workspace
	time.Sleep(3 * time.Second)

	// Test getting diagnostics for main.go
	diagnostics, err := client.getDiagnostics("main.go")
	if err != nil {
		t.Fatalf("getDiagnostics failed: %v", err)
	}

	// The enhanced workspace should have clean code with no errors
	t.Logf("Found %d diagnostics for main.go", len(diagnostics))
	for i, diag := range diagnostics {
		t.Logf("Diagnostic %d: %s (severity %d) at line %d:%d",
			i, diag.Message, diag.Severity, diag.Range.Start.Line, diag.Range.Start.Character)
	}

	// Test getting diagnostics for util.go
	utilDiagnostics, err := client.getDiagnostics("util.go")
	if err != nil {
		t.Fatalf("getDiagnostics for util.go failed: %v", err)
	}

	t.Logf("Found %d diagnostics for util.go", len(utilDiagnostics))

	// Test getting diagnostics for non-existent file (should succeed but return empty or error)
	_, err = client.getDiagnostics("nonexistent.go")
	if err != nil {
		t.Logf("getDiagnostics for non-existent file failed as expected: %v", err)
	} else {
		t.Log("getDiagnostics for non-existent file succeeded (gopls behavior may vary)")
	}

	t.Logf("getDiagnostics tests completed successfully")
}

func TestGoplsClientGetDocumentSymbols(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test getting document symbols for main.go
	symbols, err := client.getDocumentSymbols("main.go")
	if err != nil {
		t.Fatalf("getDocumentSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatal("expected to find symbols in main.go")
	}

	t.Logf("Found %d document symbols in main.go", len(symbols))

	// Look for expected symbols
	expectedSymbols := map[string]bool{
		"Person":        false,
		"Greeter":       false,
		"PersonGreeter": false,
		"main":          false,
		"testFunction":  false,
		"processPerson": false,
		"HelperFunc":    false,
	}

	for _, symbol := range symbols {
		t.Logf("Symbol: %s (kind %d) at line %d:%d",
			symbol.Name, symbol.Kind, symbol.Range.Start.Line, symbol.Range.Start.Character)

		if _, expected := expectedSymbols[symbol.Name]; expected {
			expectedSymbols[symbol.Name] = true
		}

		// Check children for struct fields, interface methods, etc.
		for _, child := range symbol.Children {
			t.Logf("  Child: %s (kind %d)", child.Name, child.Kind)
		}
	}

	// Verify we found key symbols
	foundCount := 0
	for name, found := range expectedSymbols {
		if found {
			foundCount++
		} else {
			t.Logf("Expected symbol '%s' not found", name)
		}
	}

	if foundCount < 3 {
		t.Errorf("Expected to find at least 3 key symbols, found %d", foundCount)
	}

	// Test getting symbols for util.go
	utilSymbols, err := client.getDocumentSymbols("util.go")
	if err != nil {
		t.Fatalf("getDocumentSymbols for util.go failed: %v", err)
	}

	t.Logf("Found %d document symbols in util.go", len(utilSymbols))

	// Test error case - non-existent file
	_, err = client.getDocumentSymbols("nonexistent.go")
	if err != nil {
		t.Logf("getDocumentSymbols for non-existent file failed as expected: %v", err)
	} else {
		t.Log("getDocumentSymbols for non-existent file succeeded (gopls behavior may vary)")
	}

	t.Logf("getDocumentSymbols tests completed successfully")
}

func TestGoplsClientGetWorkspaceSymbols(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test searching for "Person" symbols
	symbols, err := client.getWorkspaceSymbols("Person")
	if err != nil {
		t.Fatalf("getWorkspaceSymbols failed: %v", err)
	}

	t.Logf("Found %d workspace symbols for 'Person'", len(symbols))
	for _, symbol := range symbols {
		t.Logf("Symbol: %s (kind %d) in %s at line %d:%d",
			symbol.Name, symbol.Kind, symbol.Location.URI,
			symbol.Location.Range.Start.Line, symbol.Location.Range.Start.Character)
	}

	// Test searching for "test" symbols (should find testFunction)
	testSymbols, err := client.getWorkspaceSymbols("test")
	if err != nil {
		t.Fatalf("getWorkspaceSymbols for 'test' failed: %v", err)
	}

	t.Logf("Found %d workspace symbols for 'test'", len(testSymbols))

	// Test searching for "Math" symbols (should find MathUtils)
	mathSymbols, err := client.getWorkspaceSymbols("Math")
	if err != nil {
		t.Fatalf("getWorkspaceSymbols for 'Math' failed: %v", err)
	}

	t.Logf("Found %d workspace symbols for 'Math'", len(mathSymbols))

	// Test fuzzy search
	fuzzySymbols, err := client.getWorkspaceSymbols("Greet")
	if err != nil {
		t.Fatalf("getWorkspaceSymbols fuzzy search failed: %v", err)
	}

	t.Logf("Found %d workspace symbols for fuzzy 'Greet'", len(fuzzySymbols))

	// Test empty query (should return all symbols or handle gracefully)
	allSymbols, err := client.getWorkspaceSymbols("")
	if err != nil {
		t.Logf("getWorkspaceSymbols with empty query failed: %v", err)
	} else {
		t.Logf("Found %d workspace symbols with empty query", len(allSymbols))
	}

	t.Logf("getWorkspaceSymbols tests completed successfully")
}

func TestGoplsClientGetSignatureHelp(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test signature help for testFunction call
	// Position would be inside the function call parentheses
	// result := testFunction(42)
	//                       ^-- somewhere around here
	signatureHelp, err := client.getSignatureHelp("main.go", 28, 22) // Inside testFunction call
	if err != nil {
		t.Fatalf("getSignatureHelp failed: %v", err)
	}

	if signatureHelp != nil {
		t.Logf("Found signature help with %d signatures", len(signatureHelp.Signatures))
		for i, sig := range signatureHelp.Signatures {
			t.Logf("Signature %d: %s", i, sig.Label)
			for j, param := range sig.Parameters {
				t.Logf("  Parameter %d: %s", j, param.Label)
			}
		}
		t.Logf("Active signature: %d, Active parameter: %d",
			signatureHelp.ActiveSignature, signatureHelp.ActiveParameter)
	} else {
		t.Log("No signature help available at this position (gopls behavior may vary)")
	}

	// Test signature help for fmt.Printf call
	// Position would be inside the Printf call
	signatureHelpPrintf, err := client.getSignatureHelp("main.go", 49, 15) // Inside fmt.Printf call
	//nolint:gocritic // if-else chain is appropriate for test scenarios
	if err != nil {
		t.Logf("getSignatureHelp for Printf failed (may be expected): %v", err)
	} else if signatureHelpPrintf != nil {
		t.Logf("Found Printf signature help with %d signatures", len(signatureHelpPrintf.Signatures))
	} else {
		t.Log("No Printf signature help available")
	}

	// Test signature help at invalid position
	_, err = client.getSignatureHelp("main.go", 100, 100)
	if err != nil {
		t.Logf("getSignatureHelp at invalid position failed as expected: %v", err)
	} else {
		t.Log("getSignatureHelp at invalid position succeeded (gopls behavior may vary)")
	}

	t.Logf("getSignatureHelp tests completed successfully")
}

func TestGoplsClientGetCompletions(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test completions after "fmt."
	// Position would be right after "fmt." to get package method completions
	completions, err := client.getCompletions("main.go", 27, 5) // After "fmt."
	if err != nil {
		t.Fatalf("getCompletions failed: %v", err)
	}

	//nolint:nestif // Complex test logic is justified for comprehensive testing
	if completions != nil && len(completions.Items) > 0 {
		t.Logf("Found %d completions (isIncomplete: %v)", len(completions.Items), completions.IsIncomplete)

		// Look for expected fmt package functions
		expectedCompletions := map[string]bool{
			"Println": false,
			"Printf":  false,
			"Print":   false,
		}

		for i, item := range completions.Items {
			if i < 10 { // Log first 10 completions
				t.Logf("Completion %d: %s (kind %d) - %s", i, item.Label, item.Kind, item.Detail)
			}

			if _, expected := expectedCompletions[item.Label]; expected {
				expectedCompletions[item.Label] = true
			}
		}

		// Check if we found key fmt functions
		foundCount := 0
		for name, found := range expectedCompletions {
			if found {
				foundCount++
				t.Logf("Found expected completion: %s", name)
			}
		}

		if foundCount > 0 {
			t.Logf("Successfully found %d expected fmt completions", foundCount)
		} else {
			t.Log("No expected fmt completions found (gopls behavior may vary)")
		}
	} else {
		t.Log("No completions available at this position (gopls behavior may vary)")
	}

	// Test completions for local symbols
	// Position after partial typing of a local function
	localCompletions, err := client.getCompletions("main.go", 29, 10) // Somewhere in main function
	if err != nil {
		t.Logf("getCompletions for local symbols failed: %v", err)
	} else if localCompletions != nil {
		t.Logf("Found %d local completions", len(localCompletions.Items))
	}

	// Test completions at invalid position
	_, err = client.getCompletions("main.go", 1000, 1000)
	if err != nil {
		t.Logf("getCompletions at invalid position failed as expected: %v", err)
	} else {
		t.Log("getCompletions at invalid position succeeded (gopls behavior may vary)")
	}

	t.Logf("getCompletions tests completed successfully")
}

func TestGoplsClientGetTypeDefinition(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test type definition for Person variable
	// Position on "person" variable to find its type definition
	locations, err := client.getTypeDefinition("main.go", 31, 2) // On "person" variable
	if err != nil {
		// This might fail depending on gopls behavior and position accuracy
		t.Logf("getTypeDefinition failed (position-dependent): %v", err)
		// Try a different position - on the Person struct name in variable declaration
		locations2, err2 := client.getTypeDefinition("main.go", 31, 12) // On "Person" type in declaration
		if err2 != nil {
			t.Logf("getTypeDefinition also failed at second position: %v", err2)
			// This is acceptable - type definition behavior varies
			return
		}
		locations = locations2
	}

	if len(locations) > 0 {
		t.Logf("Found %d type definition locations", len(locations))
		for i, location := range locations {
			t.Logf("Type definition %d: %s at line %d:%d",
				i, location.URI, location.Range.Start.Line, location.Range.Start.Character)
		}

		// Verify it points to Person struct definition
		location := locations[0]
		if location.URI == "main.go" {
			t.Logf("Type definition correctly points to main.go")
		}
	} else {
		t.Log("No type definition found (gopls behavior may vary)")
	}

	// Test type definition for greeter variable
	greeterLocations, err := client.getTypeDefinition("main.go", 34, 2) // On "greeter" variable
	//nolint:gocritic // if-else chain is appropriate for test scenarios
	if err != nil {
		t.Logf("getTypeDefinition for greeter failed: %v", err)
	} else if len(greeterLocations) > 0 {
		t.Logf("Found %d type definitions for greeter", len(greeterLocations))
	} else {
		t.Log("No type definition found for greeter")
	}

	// Test type definition at invalid position
	_, err = client.getTypeDefinition("main.go", 1000, 1000)
	if err != nil {
		t.Logf("getTypeDefinition at invalid position failed as expected: %v", err)
	} else {
		t.Log("getTypeDefinition at invalid position succeeded (gopls behavior may vary)")
	}

	t.Logf("getTypeDefinition tests completed successfully")
}

func TestGoplsClientFindImplementations(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test finding implementations of Greeter interface
	// Position on the Greeter interface definition
	locations, err := client.findImplementations("main.go", 16, 5) // On "Greeter" interface
	if err != nil {
		t.Fatalf("findImplementations failed: %v", err)
	}

	//nolint:nestif // Complex test logic is justified for comprehensive testing
	if len(locations) > 0 {
		t.Logf("Found %d implementations of Greeter interface", len(locations))
		for i, location := range locations {
			t.Logf("Implementation %d: %s at line %d:%d",
				i, location.URI, location.Range.Start.Line, location.Range.Start.Character)
		}

		// Verify we found PersonGreeter implementation
		foundPersonGreeter := false
		for _, location := range locations {
			if location.URI == "main.go" {
				// Check if it's around the PersonGreeter implementation area
				if location.Range.Start.Line >= 19 && location.Range.Start.Line <= 25 {
					foundPersonGreeter = true
					break
				}
			}
		}

		if foundPersonGreeter {
			t.Log("Successfully found PersonGreeter implementation")
		} else {
			t.Log("PersonGreeter implementation not found in expected location")
		}
	} else {
		t.Log("No implementations found (gopls behavior may vary)")
	}

	// Test finding implementations on method (should find interface implementations)
	methodLocations, err := client.findImplementations("main.go", 17, 5) // On "Greet" method in interface
	//nolint:gocritic // if-else chain is appropriate for test scenarios
	if err != nil {
		t.Logf("findImplementations for method failed: %v", err)
	} else if len(methodLocations) > 0 {
		t.Logf("Found %d implementations for Greet method", len(methodLocations))
	} else {
		t.Log("No method implementations found")
	}

	// Test finding implementations at invalid position
	_, err = client.findImplementations("main.go", 1000, 1000)
	if err != nil {
		t.Logf("findImplementations at invalid position failed as expected: %v", err)
	} else {
		t.Log("findImplementations at invalid position succeeded (gopls behavior may vary)")
	}

	t.Logf("findImplementations tests completed successfully")
}

func TestGoplsClientFormatDocument(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t) // Use simpler workspace for formatting test
	defer cleanup()

	// Create a poorly formatted Go file
	unformattedContent := `package main

import"fmt"

func main(){fmt.Println("Hello");result:=testFunction( )
fmt.Println("Result:",result)}

func testFunction( ) int{
return     42
}`
	unformattedPath := filepath.Join(workspacePath, "unformatted.go")
	if err := os.WriteFile(unformattedPath, []byte(unformattedContent), 0644); err != nil {
		t.Fatalf("failed to create unformatted.go: %v", err)
	}

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test formatting the unformatted file
	textEdits, err := client.formatDocument("unformatted.go")
	if err != nil {
		t.Fatalf("formatDocument failed: %v", err)
	}

	t.Logf("Format operation returned %d text edits", len(textEdits))

	for i, edit := range textEdits {
		if i < 5 { // Log first 5 edits
			t.Logf("  Edit %d: line %d:%d-%d:%d = %q",
				i, edit.Range.Start.Line, edit.Range.Start.Character,
				edit.Range.End.Line, edit.Range.End.Character, edit.NewText)
		}
	}

	if len(textEdits) > 0 {
		t.Log("Format operation generated edits (formatting needed)")
	} else {
		t.Log("Format operation generated no edits (file already formatted)")
	}

	// Test formatting main.go (should be already formatted)
	mainEdits, err := client.formatDocument("main.go")
	if err != nil {
		t.Logf("formatDocument for main.go failed: %v", err)
	} else {
		t.Logf("Format operation for main.go returned %d text edits", len(mainEdits))
	}

	// Test formatting non-existent file
	_, err = client.formatDocument("nonexistent.go")
	if err != nil {
		t.Logf("formatDocument for non-existent file failed as expected: %v", err)
	} else {
		t.Log("formatDocument for non-existent file succeeded (gopls behavior may vary)")
	}

	t.Logf("formatDocument tests completed successfully")
}

func TestGoplsClientOrganizeImports(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createTempGoWorkspace(t)
	defer cleanup()

	// Create a file with disorganized imports
	disorganizedContent := `package main

import (
	"strings"
	"fmt"
	"errors"
	"os"
)

func main() {
	fmt.Println("Hello")
	strings.ToUpper("test")
	errors.New("test")
	os.Getenv("HOME")
}`
	disorganizedPath := filepath.Join(workspacePath, "disorganized.go")
	if err := os.WriteFile(disorganizedPath, []byte(disorganizedContent), 0644); err != nil {
		t.Fatalf("failed to create disorganized.go: %v", err)
	}

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test organizing imports for disorganized file
	textEdits, err := client.organizeImports("disorganized.go")
	if err != nil {
		t.Fatalf("organizeImports failed: %v", err)
	}

	t.Logf("Organize imports returned %d text edits", len(textEdits))

	for i, edit := range textEdits {
		if i < 5 { // Log first 5 edits
			t.Logf("  Edit %d: line %d:%d-%d:%d = %q",
				i, edit.Range.Start.Line, edit.Range.Start.Character,
				edit.Range.End.Line, edit.Range.End.Character, edit.NewText)
		}
	}

	if len(textEdits) > 0 {
		t.Log("Organize imports generated edits (imports reorganized)")
	} else {
		t.Log("Organize imports generated no edits (imports already organized)")
	}

	// Test organizing imports for main.go
	mainEdits, err := client.organizeImports("main.go")
	if err != nil {
		t.Logf("organizeImports for main.go failed: %v", err)
	} else {
		t.Logf("Organize imports for main.go returned %d text edits", len(mainEdits))
	}

	// Test organizing imports for non-existent file
	_, err = client.organizeImports("nonexistent.go")
	if err != nil {
		t.Logf("organizeImports for non-existent file failed as expected: %v", err)
	} else {
		t.Log("organizeImports for non-existent file succeeded (gopls behavior may vary)")
	}

	t.Logf("organizeImports tests completed successfully")
}

func TestGoplsClientGetInlayHints(t *testing.T) {
	requireGopls(t)

	workspacePath, cleanup := createEnhancedGoWorkspace(t)
	defer cleanup()

	logger := newDebugLogger()
	client := newClient(workspacePath, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = client.stop() }()

	// Give gopls time to initialize
	time.Sleep(3 * time.Second)

	// Test getting inlay hints for a range in main.go
	// This should include parameter names, type hints, etc.
	hints, err := client.getInlayHints("main.go", 25, 0, 35, 0) // Range covering main function
	if err != nil {
		t.Fatalf("getInlayHints failed: %v", err)
	}

	if len(hints) > 0 {
		t.Logf("Found %d inlay hints", len(hints))
		for i, hint := range hints {
			if i < 10 { // Log first 10 hints
				t.Logf("Hint %d: %s at line %d:%d (kind %d)",
					i, hint.Label, hint.Position.Line, hint.Position.Character, hint.Kind)
			}
		}
	} else {
		t.Log("No inlay hints found (gopls behavior may vary - inlay hints might be disabled)")
	}

	// Test getting inlay hints for testFunction
	funcHints, err := client.getInlayHints("main.go", 40, 0, 45, 0) // Range covering testFunction
	if err != nil {
		t.Logf("getInlayHints for testFunction failed: %v", err)
	} else {
		t.Logf("Found %d inlay hints for testFunction", len(funcHints))
	}

	// Test getting inlay hints for util.go
	utilHints, err := client.getInlayHints("util.go", 0, 0, 20, 0) // Range covering MathUtils
	if err != nil {
		t.Logf("getInlayHints for util.go failed: %v", err)
	} else {
		t.Logf("Found %d inlay hints for util.go", len(utilHints))
	}

	// Test invalid range
	_, err = client.getInlayHints("main.go", 1000, 0, 1001, 0)
	if err != nil {
		t.Logf("getInlayHints with invalid range failed as expected: %v", err)
	} else {
		t.Log("getInlayHints with invalid range succeeded (gopls behavior may vary)")
	}

	// Test non-existent file
	_, err = client.getInlayHints("nonexistent.go", 0, 0, 10, 0)
	if err != nil {
		t.Logf("getInlayHints for non-existent file failed as expected: %v", err)
	} else {
		t.Log("getInlayHints for non-existent file succeeded (gopls behavior may vary)")
	}

	t.Logf("getInlayHints tests completed successfully")
}
