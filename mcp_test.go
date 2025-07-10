package main

import (
	"fmt"
	"testing"
)

// Mock interfaces and implementations for testing

// goplsClientInterface defines the interface for gopls client operations.
type goplsClientInterface interface {
	isRunning() bool
	goToDefinition(path string, line, character int) ([]Location, error)
	findReferences(path string, line, character int, includeDeclaration bool) ([]Location, error)
	getHover(path string, line, character int) (*Hover, error)
	getDiagnostics(path string) ([]Diagnostic, error)
	getDocumentSymbols(path string) ([]DocumentSymbol, error)
	getWorkspaceSymbols(query string) ([]SymbolInformation, error)
	getSignatureHelp(path string, line, character int) (*SignatureHelp, error)
	getCompletions(path string, line, character int) (*CompletionList, error)
	getTypeDefinition(path string, line, character int) ([]Location, error)
	findImplementations(path string, line, character int) ([]Location, error)
	formatDocument(path string) ([]TextEdit, error)
	organizeImports(path string) ([]TextEdit, error)
	getInlayHints(path string, startLine, startChar, endLine, endChar int) ([]InlayHint, error)
}

// mockGoplsClient implements the goplsClientInterface for testing.
type mockGoplsClient struct {
	running bool
	// Method call tracking
	goToDefinitionCalled      bool
	findReferencesCalled      bool
	getHoverCalled            bool
	getDiagnosticsCalled      bool
	getDocumentSymbolsCalled  bool
	getWorkspaceSymbolsCalled bool
	getSignatureHelpCalled    bool
	getCompletionsCalled      bool
	getTypeDefinitionCalled   bool
	findImplementationsCalled bool
	formatDocumentCalled      bool
	organizeImportsCalled     bool
	getInlayHintsCalled       bool

	// Mock responses
	mockLocations        []Location
	mockHover            *Hover
	mockDiagnostics      []Diagnostic
	mockDocumentSymbols  []DocumentSymbol
	mockWorkspaceSymbols []SymbolInformation
	mockSignatureHelp    *SignatureHelp
	mockCompletions      *CompletionList
	mockTextEdits        []TextEdit
	mockInlayHints       []InlayHint

	// Error responses
	shouldError  bool
	errorMessage string
}

func newMockGoplsClient(running bool) *mockGoplsClient {
	return &mockGoplsClient{
		running: running,
		// Default mock responses
		mockLocations: []Location{
			{
				URI: "test.go",
				Range: Range{
					Start: Position{Line: 10, Character: 5},
					End:   Position{Line: 10, Character: 15},
				},
			},
		},
		mockHover: &Hover{
			Contents: []string{"func testFunction() int"},
			Range: &Range{
				Start: Position{Line: 10, Character: 5},
				End:   Position{Line: 10, Character: 15},
			},
		},
		mockDiagnostics: []Diagnostic{
			{
				Range: Range{
					Start: Position{Line: 5, Character: 0},
					End:   Position{Line: 5, Character: 10},
				},
				Severity: 1,
				Code:     "test-error",
				Source:   "gopls",
				Message:  "test diagnostic message",
			},
		},
		mockDocumentSymbols: []DocumentSymbol{
			{
				Name: "testFunction",
				Kind: 12, // Function
				Range: Range{
					Start: Position{Line: 10, Character: 0},
					End:   Position{Line: 15, Character: 1},
				},
				SelectionRange: Range{
					Start: Position{Line: 10, Character: 5},
					End:   Position{Line: 10, Character: 17},
				},
			},
		},
		mockWorkspaceSymbols: []SymbolInformation{
			{
				Name: "TestStruct",
				Kind: 23, // Struct
				Location: Location{
					URI: "test.go",
					Range: Range{
						Start: Position{Line: 20, Character: 0},
						End:   Position{Line: 25, Character: 1},
					},
				},
			},
		},
		mockSignatureHelp: &SignatureHelp{
			Signatures: []SignatureInformation{
				{
					Label:         "testFunction(input int) int",
					Documentation: "Test function documentation",
					Parameters: []ParameterInformation{
						{
							Label:         "input int",
							Documentation: "Input parameter",
						},
					},
				},
			},
			ActiveSignature: 0,
			ActiveParameter: 0,
		},
		mockCompletions: &CompletionList{
			IsIncomplete: false,
			Items: []CompletionItem{
				{
					Label:  "testFunction",
					Kind:   3, // Function
					Detail: "func() int",
				},
			},
		},
		mockTextEdits: []TextEdit{
			{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 10},
				},
				NewText: "formatted text",
			},
		},
		mockInlayHints: []InlayHint{
			{
				Position: Position{Line: 10, Character: 15},
				Label:    "int",
				Kind:     1, // Type
			},
		},
	}
}

func (m *mockGoplsClient) isRunning() bool {
	return m.running
}

func (m *mockGoplsClient) goToDefinition(_ string, _, _ int) ([]Location, error) {
	m.goToDefinitionCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockLocations, nil
}

func (m *mockGoplsClient) findReferences(_ string, _, _ int, _ bool) ([]Location, error) {
	m.findReferencesCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockLocations, nil
}

func (m *mockGoplsClient) getHover(_ string, _, _ int) (*Hover, error) {
	m.getHoverCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockHover, nil
}

func (m *mockGoplsClient) getDiagnostics(_ string) ([]Diagnostic, error) {
	m.getDiagnosticsCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockDiagnostics, nil
}

func (m *mockGoplsClient) getDocumentSymbols(_ string) ([]DocumentSymbol, error) {
	m.getDocumentSymbolsCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockDocumentSymbols, nil
}

func (m *mockGoplsClient) getWorkspaceSymbols(_ string) ([]SymbolInformation, error) {
	m.getWorkspaceSymbolsCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockWorkspaceSymbols, nil
}

func (m *mockGoplsClient) getSignatureHelp(_ string, _, _ int) (*SignatureHelp, error) {
	m.getSignatureHelpCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockSignatureHelp, nil
}

func (m *mockGoplsClient) getCompletions(_ string, _, _ int) (*CompletionList, error) {
	m.getCompletionsCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockCompletions, nil
}

func (m *mockGoplsClient) getTypeDefinition(_ string, _, _ int) ([]Location, error) {
	m.getTypeDefinitionCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockLocations, nil
}

func (m *mockGoplsClient) findImplementations(_ string, _, _ int) ([]Location, error) {
	m.findImplementationsCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockLocations, nil
}

func (m *mockGoplsClient) formatDocument(_ string) ([]TextEdit, error) {
	m.formatDocumentCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockTextEdits, nil
}

func (m *mockGoplsClient) organizeImports(_ string) ([]TextEdit, error) {
	m.organizeImportsCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockTextEdits, nil
}

func (m *mockGoplsClient) getInlayHints(_ string, _, _, _, _ int) ([]InlayHint, error) {
	m.getInlayHintsCalled = true
	if m.shouldError {
		return nil, &mockError{m.errorMessage}
	}
	return m.mockInlayHints, nil
}

// mockError implements error interface for testing.
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

// Test utilities

// testMCPTools is a test-specific version of mcpTools that uses interfaces.
type testMCPTools struct {
	clients map[string]goplsClientInterface
}

// newTestMCPTools creates a new test MCP tools instance.
func newTestMCPTools(clients map[string]goplsClientInterface) testMCPTools {
	return testMCPTools{
		clients: clients,
	}
}

// getClient returns the goplsClient for the specified workspace (test version).
func (m testMCPTools) getClient(workspace string) (goplsClientInterface, error) {
	client, exists := m.clients[workspace]
	if !exists {
		return nil, fmt.Errorf("workspace not found: %s", workspace)
	}
	if !client.isRunning() {
		return nil, fmt.Errorf("gopls is not running for workspace: %s", workspace)
	}
	return client, nil
}

// Add the same conversion methods as mcpTools.
func (m testMCPTools) convertLocationsToResults(locations []Location) []LocationResult {
	results := make([]LocationResult, len(locations))
	for i, loc := range locations {
		results[i] = LocationResult{
			URI:          loc.URI,
			Line:         loc.Range.Start.Line,
			Character:    loc.Range.Start.Character,
			EndLine:      loc.Range.End.Line,
			EndCharacter: loc.Range.End.Character,
		}
	}
	return results
}

func (m testMCPTools) convertLocationToResult(location Location) LocationResult {
	return LocationResult{
		URI:          location.URI,
		Line:         location.Range.Start.Line,
		Character:    location.Range.Start.Character,
		EndLine:      location.Range.End.Line,
		EndCharacter: location.Range.End.Character,
	}
}

func (m testMCPTools) convertDocumentSymbolToResult(symbol DocumentSymbol) DocumentSymbolResult {
	children := make([]DocumentSymbolResult, len(symbol.Children))
	for i, child := range symbol.Children {
		children[i] = m.convertDocumentSymbolToResult(child)
	}

	return DocumentSymbolResult{
		Name:       symbol.Name,
		Detail:     symbol.Detail,
		Kind:       symbol.Kind,
		Deprecated: symbol.Deprecated,
		Range: LocationResult{
			URI:          "",
			Line:         symbol.Range.Start.Line,
			Character:    symbol.Range.Start.Character,
			EndLine:      symbol.Range.End.Line,
			EndCharacter: symbol.Range.End.Character,
		},
		SelectionRange: LocationResult{
			URI:          "",
			Line:         symbol.SelectionRange.Start.Line,
			Character:    symbol.SelectionRange.Start.Character,
			EndLine:      symbol.SelectionRange.End.Line,
			EndCharacter: symbol.SelectionRange.End.Character,
		},
		Children: children,
	}
}

func (m testMCPTools) convertTextEditsToResults(textEdits []TextEdit) []TextEditResult {
	results := make([]TextEditResult, len(textEdits))
	for i, edit := range textEdits {
		results[i] = TextEditResult{
			Range: LocationResult{
				URI:          "",
				Line:         edit.Range.Start.Line,
				Character:    edit.Range.Start.Character,
				EndLine:      edit.Range.End.Line,
				EndCharacter: edit.Range.End.Character,
			},
			NewText: edit.NewText,
		}
	}
	return results
}

func (m testMCPTools) convertInlayHintsToResults(inlayHints []InlayHint) []InlayHintResult {
	results := make([]InlayHintResult, len(inlayHints))
	for i, hint := range inlayHints {
		results[i] = InlayHintResult{
			Position: LocationResult{
				URI:          "",
				Line:         hint.Position.Line,
				Character:    hint.Position.Character,
				EndLine:      hint.Position.Line,
				EndCharacter: hint.Position.Character,
			},
			Label:   hint.Label,
			Kind:    hint.Kind,
			Tooltip: hint.Tooltip,
		}
	}
	return results
}

func createTestMCPTools() testMCPTools {
	mockClient := newMockGoplsClient(true)
	clients := map[string]goplsClientInterface{
		"/test/workspace": mockClient,
	}
	return newTestMCPTools(clients)
}

// Tests for testMCPTools struct

func TestNewTestMCPTools(t *testing.T) {
	mockClient := newMockGoplsClient(true)
	clients := map[string]goplsClientInterface{
		"/test/workspace": mockClient,
	}

	tools := newTestMCPTools(clients)

	if tools.clients == nil {
		t.Fatal("Expected non-nil clients map")
	}

	if len(tools.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(tools.clients))
	}

	if tools.clients["/test/workspace"] != mockClient {
		t.Error("Client not properly stored in map")
	}
}

func TestGetClient(t *testing.T) {
	tools := createTestMCPTools()

	// Test valid workspace
	client, err := tools.getClient("/test/workspace")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client")
	}

	// Test invalid workspace
	_, err = tools.getClient("/invalid/workspace")
	if err == nil {
		t.Error("Expected error for invalid workspace")
	}

	// Test workspace with non-running client
	mockClient := newMockGoplsClient(false) // not running
	clients := map[string]goplsClientInterface{
		"/stopped/workspace": mockClient,
	}
	stoppedTools := newTestMCPTools(clients)

	_, err = stoppedTools.getClient("/stopped/workspace")
	if err == nil {
		t.Error("Expected error for non-running client")
	}
}

// Tests for conversion methods

func TestConvertLocationsToResults(t *testing.T) {
	tools := createTestMCPTools()

	locations := []Location{
		{
			URI: "test.go",
			Range: Range{
				Start: Position{Line: 10, Character: 5},
				End:   Position{Line: 10, Character: 15},
			},
		},
		{
			URI: "other.go",
			Range: Range{
				Start: Position{Line: 20, Character: 0},
				End:   Position{Line: 20, Character: 10},
			},
		},
	}

	results := tools.convertLocationsToResults(locations)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test first location
	if results[0].URI != "test.go" {
		t.Errorf("Expected URI 'test.go', got '%s'", results[0].URI)
	}
	if results[0].Line != 10 {
		t.Errorf("Expected line 10, got %d", results[0].Line)
	}
	if results[0].Character != 5 {
		t.Errorf("Expected character 5, got %d", results[0].Character)
	}
	if results[0].EndLine != 10 {
		t.Errorf("Expected end line 10, got %d", results[0].EndLine)
	}
	if results[0].EndCharacter != 15 {
		t.Errorf("Expected end character 15, got %d", results[0].EndCharacter)
	}
}

func TestConvertLocationToResult(t *testing.T) {
	tools := createTestMCPTools()

	location := Location{
		URI: "test.go",
		Range: Range{
			Start: Position{Line: 5, Character: 10},
			End:   Position{Line: 5, Character: 20},
		},
	}

	result := tools.convertLocationToResult(location)

	if result.URI != "test.go" {
		t.Errorf("Expected URI 'test.go', got '%s'", result.URI)
	}
	if result.Line != 5 {
		t.Errorf("Expected line 5, got %d", result.Line)
	}
	if result.Character != 10 {
		t.Errorf("Expected character 10, got %d", result.Character)
	}
	if result.EndLine != 5 {
		t.Errorf("Expected end line 5, got %d", result.EndLine)
	}
	if result.EndCharacter != 20 {
		t.Errorf("Expected end character 20, got %d", result.EndCharacter)
	}
}

func TestConvertDocumentSymbolToResult(t *testing.T) {
	tools := createTestMCPTools()

	symbol := DocumentSymbol{
		Name:   "TestFunction",
		Detail: "func() int",
		Kind:   12, // Function
		Range: Range{
			Start: Position{Line: 10, Character: 0},
			End:   Position{Line: 15, Character: 1},
		},
		SelectionRange: Range{
			Start: Position{Line: 10, Character: 5},
			End:   Position{Line: 10, Character: 17},
		},
		Children: []DocumentSymbol{
			{
				Name: "ChildSymbol",
				Kind: 13, // Variable
				Range: Range{
					Start: Position{Line: 12, Character: 4},
					End:   Position{Line: 12, Character: 15},
				},
				SelectionRange: Range{
					Start: Position{Line: 12, Character: 4},
					End:   Position{Line: 12, Character: 15},
				},
			},
		},
	}

	result := tools.convertDocumentSymbolToResult(symbol)

	if result.Name != "TestFunction" {
		t.Errorf("Expected name 'TestFunction', got '%s'", result.Name)
	}
	if result.Detail != "func() int" {
		t.Errorf("Expected detail 'func() int', got '%s'", result.Detail)
	}
	if result.Kind != 12 {
		t.Errorf("Expected kind 12, got %d", result.Kind)
	}
	if result.Range.Line != 10 {
		t.Errorf("Expected range line 10, got %d", result.Range.Line)
	}

	// Test children conversion
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(result.Children))
	}
	if result.Children[0].Name != "ChildSymbol" {
		t.Errorf("Expected child name 'ChildSymbol', got '%s'", result.Children[0].Name)
	}
}

func TestConvertTextEditsToResults(t *testing.T) {
	tools := createTestMCPTools()

	edits := []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 10},
			},
			NewText: "new text",
		},
		{
			Range: Range{
				Start: Position{Line: 1, Character: 5},
				End:   Position{Line: 1, Character: 15},
			},
			NewText: "other text",
		},
	}

	results := tools.convertTextEditsToResults(edits)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].NewText != "new text" {
		t.Errorf("Expected new text 'new text', got '%s'", results[0].NewText)
	}
	if results[0].Range.Line != 0 {
		t.Errorf("Expected line 0, got %d", results[0].Range.Line)
	}
	if results[1].NewText != "other text" {
		t.Errorf("Expected new text 'other text', got '%s'", results[1].NewText)
	}
}

func TestConvertInlayHintsToResults(t *testing.T) {
	tools := createTestMCPTools()

	hints := []InlayHint{
		{
			Position: Position{Line: 10, Character: 15},
			Label:    "int",
			Kind:     1, // Type
			Tooltip:  "Type hint",
		},
		{
			Position: Position{Line: 20, Character: 5},
			Label:    "param:",
			Kind:     2, // Parameter
		},
	}

	results := tools.convertInlayHintsToResults(hints)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].Label != "int" {
		t.Errorf("Expected label 'int', got '%s'", results[0].Label)
	}
	if results[0].Kind != 1 {
		t.Errorf("Expected kind 1, got %d", results[0].Kind)
	}
	if results[0].Tooltip != "Type hint" {
		t.Errorf("Expected tooltip 'Type hint', got '%s'", results[0].Tooltip)
	}
	if results[0].Position.Line != 10 {
		t.Errorf("Expected position line 10, got %d", results[0].Position.Line)
	}
}

// Tests for mock client method tracking

func TestMockClientMethodTracking(t *testing.T) {
	mockClient := newMockGoplsClient(true)

	// Test that methods are called and tracked
	_, _ = mockClient.goToDefinition("test.go", 10, 5)
	if !mockClient.goToDefinitionCalled {
		t.Error("Expected goToDefinitionCalled to be true")
	}

	_, _ = mockClient.findReferences("test.go", 10, 5, true)
	if !mockClient.findReferencesCalled {
		t.Error("Expected findReferencesCalled to be true")
	}

	_, _ = mockClient.getHover("test.go", 10, 5)
	if !mockClient.getHoverCalled {
		t.Error("Expected getHoverCalled to be true")
	}
}
