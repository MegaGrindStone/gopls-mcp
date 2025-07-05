package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Manager manages a gopls subprocess and handles LSP communication.
type Manager struct {
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	requestIDMux  sync.Mutex
	requestID     int
	workspacePath string
	mu            sync.RWMutex
	running       bool
}

// NewManager creates a new gopls manager with the specified workspace path.
func NewManager(workspacePath string) *Manager {
	return &Manager{
		workspacePath: workspacePath,
	}
}

// Start starts the gopls subprocess.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("gopls is already running")
	}

	// Start gopls process
	m.cmd = exec.CommandContext(ctx, "gopls", "serve")

	var err error
	m.stdin, err = m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	m.stdout, err = m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	m.stderr, err = m.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gopls: %w", err)
	}

	m.running = true

	// Initialize gopls with workspace
	if err := m.initialize(ctx); err != nil {
		_ = m.Stop()
		return fmt.Errorf("failed to initialize gopls: %w", err)
	}

	return nil
}

// Stop stops the gopls subprocess.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	var err error
	if m.stdin != nil {
		_ = m.stdin.Close()
	}
	if m.stdout != nil {
		_ = m.stdout.Close()
	}
	if m.stderr != nil {
		_ = m.stderr.Close()
	}

	if m.cmd != nil && m.cmd.Process != nil {
		err = m.cmd.Process.Kill()
		_ = m.cmd.Wait()
	}

	m.running = false
	m.cmd = nil
	m.stdin = nil
	m.stdout = nil
	m.stderr = nil

	return err
}

// IsRunning returns true if gopls is currently running.
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// nextRequestID generates the next request ID for LSP communication.
func (m *Manager) nextRequestID() int {
	m.requestIDMux.Lock()
	defer m.requestIDMux.Unlock()
	m.requestID++
	return m.requestID
}

// initialize sends the LSP initialize request to gopls.
func (m *Manager) initialize(_ context.Context) error {
	initRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      m.nextRequestID(),
		"method":  "initialize",
		"params": map[string]any{
			"processId": nil,
			"rootUri":   fmt.Sprintf("file://%s", m.workspacePath),
			"capabilities": map[string]any{
				"textDocument": map[string]any{
					"hover": map[string]any{
						"contentFormat": []string{"markdown", "plaintext"},
					},
					"definition": map[string]any{
						"linkSupport": true,
					},
					"references":      map[string]any{},
					"documentSymbol":  map[string]any{},
					"workspaceSymbol": map[string]any{},
				},
				"workspace": map[string]any{
					"workspaceFolders": true,
				},
			},
		},
	}

	return m.sendRequest(initRequest)
}

// sendRequest sends a JSON-RPC request to gopls.
func (m *Manager) sendRequest(request map[string]any) error {
	if !m.running {
		return fmt.Errorf("gopls is not running")
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// LSP uses Content-Length header format
	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	_, err = m.stdin.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// readResponse reads a JSON-RPC response from gopls.
func (m *Manager) readResponse() (map[string]any, error) {
	if !m.running {
		return nil, fmt.Errorf("gopls is not running")
	}

	scanner := bufio.NewScanner(m.stdout)

	// Read Content-Length header
	var contentLength int
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		if _, err := fmt.Sscanf(line, "Content-Length: %d", &contentLength); err == nil {
			break
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no content-length header found")
	}

	// Read the JSON content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(m.stdout, content); err != nil {
		return nil, fmt.Errorf("failed to read response content: %w", err)
	}

	var response map[string]any
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

// MCP tool parameter types

// GoToDefinitionParams represents parameters for go to definition requests.
type GoToDefinitionParams struct {
	URI       string `json:"uri"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// FindReferencesParams represents parameters for find references requests.
type FindReferencesParams struct {
	URI                string `json:"uri"`
	Line               int    `json:"line"`
	Character          int    `json:"character"`
	IncludeDeclaration bool   `json:"includeDeclaration"`
}

// GetHoverParams represents parameters for get hover info requests.
type GetHoverParams struct {
	URI       string `json:"uri"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// LocationResult represents a location result.
type LocationResult struct {
	URI          string `json:"uri"`
	Line         int    `json:"line"`
	Character    int    `json:"character"`
	EndLine      int    `json:"endLine"`
	EndCharacter int    `json:"endCharacter"`
}

// GoToDefinitionResult represents the result of a go to definition request.
type GoToDefinitionResult struct {
	Locations []LocationResult `json:"locations"`
}

// FindReferencesResult represents the result of a find references request.
type FindReferencesResult struct {
	Locations []LocationResult `json:"locations"`
}

// GetHoverResult represents the result of a get hover request.
type GetHoverResult struct {
	Contents []string        `json:"contents"`
	HasRange bool            `json:"hasRange"`
	Range    *LocationResult `json:"range,omitempty"`
}

// MCP tool handlers

// HandleGoToDefinition handles go to definition requests.
func (m *Manager) HandleGoToDefinition(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[GoToDefinitionParams]) (*mcp.CallToolResultFor[GoToDefinitionResult], error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	locations, err := m.GoToDefinition(ctx, params.Arguments.URI, params.Arguments.Line, params.Arguments.Character)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	result := GoToDefinitionResult{
		Locations: make([]LocationResult, len(locations)),
	}

	for i, loc := range locations {
		result.Locations[i] = LocationResult{
			URI:          loc.URI,
			Line:         loc.Range.Start.Line,
			Character:    loc.Range.Start.Character,
			EndLine:      loc.Range.End.Line,
			EndCharacter: loc.Range.End.Character,
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GoToDefinitionResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleFindReferences handles find references requests.
func (m *Manager) HandleFindReferences(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[FindReferencesParams]) (*mcp.CallToolResultFor[FindReferencesResult], error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	locations, err := m.FindReferences(ctx, params.Arguments.URI, params.Arguments.Line, params.Arguments.Character, params.Arguments.IncludeDeclaration)
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	result := FindReferencesResult{
		Locations: make([]LocationResult, len(locations)),
	}

	for i, loc := range locations {
		result.Locations[i] = LocationResult{
			URI:          loc.URI,
			Line:         loc.Range.Start.Line,
			Character:    loc.Range.Start.Character,
			EndLine:      loc.Range.End.Line,
			EndCharacter: loc.Range.End.Character,
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[FindReferencesResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetHover handles get hover info requests.
func (m *Manager) HandleGetHover(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[GetHoverParams]) (*mcp.CallToolResultFor[GetHoverResult], error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	hover, err := m.GetHover(ctx, params.Arguments.URI, params.Arguments.Line, params.Arguments.Character)
	if err != nil {
		return nil, fmt.Errorf("failed to get hover info: %w", err)
	}

	result := GetHoverResult{
		Contents: hover.Contents,
		HasRange: hover.Range != nil,
	}

	if hover.Range != nil {
		result.Range = &LocationResult{
			URI:          params.Arguments.URI,
			Line:         hover.Range.Start.Line,
			Character:    hover.Range.Start.Character,
			EndLine:      hover.Range.End.Line,
			EndCharacter: hover.Range.End.Character,
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetHoverResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// MCP tool creation methods

// CreateGoToDefinitionTool creates the go to definition MCP tool.
func (m *Manager) CreateGoToDefinitionTool() *mcp.ServerTool {
	return mcp.NewServerTool[GoToDefinitionParams, GoToDefinitionResult](
		"go_to_definition",
		"Navigate to the definition of a symbol at the specified position in a Go file",
		m.HandleGoToDefinition,
		mcp.Input(
			mcp.Property("uri", mcp.Description("File URI (e.g., file:///path/to/file.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateFindReferencesTool creates the find references MCP tool.
func (m *Manager) CreateFindReferencesTool() *mcp.ServerTool {
	return mcp.NewServerTool[FindReferencesParams, FindReferencesResult](
		"find_references",
		"Find all references to a symbol at the specified position in a Go file",
		m.HandleFindReferences,
		mcp.Input(
			mcp.Property("uri", mcp.Description("File URI (e.g., file:///path/to/file.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
			mcp.Property("includeDeclaration", mcp.Description("Include declaration in results"), mcp.Required(false)),
		),
	)
}

// CreateGetHoverTool creates the get hover info MCP tool.
func (m *Manager) CreateGetHoverTool() *mcp.ServerTool {
	return mcp.NewServerTool[GetHoverParams, GetHoverResult](
		"get_hover_info",
		"Get hover information (documentation, type info) for a symbol at the specified position",
		m.HandleGetHover,
		mcp.Input(
			mcp.Property("uri", mcp.Description("File URI (e.g., file:///path/to/file.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}
