package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Manager manages a gopls subprocess and handles LSP communication.
type Manager struct {
	cmd               *exec.Cmd
	stdin             io.WriteCloser
	stdout            io.ReadCloser
	stderr            io.ReadCloser
	requestIDMux      sync.Mutex
	requestID         int
	workspacePath     string
	mu                sync.RWMutex
	running           bool
	responses         map[int]chan map[string]any
	responsesMux      sync.Mutex
	openFiles         map[string]bool
	openFilesMux      sync.RWMutex
	workspaceReady    bool
	workspaceReadyMux sync.RWMutex
	logger            *slog.Logger
}

// NewManager creates a new gopls manager with the specified workspace path.
func NewManager(workspacePath string, logger *slog.Logger) *Manager {
	return &Manager{
		workspacePath: workspacePath,
		responses:     make(map[int]chan map[string]any),
		openFiles:     make(map[string]bool),
		logger:        logger,
	}
}

// Start starts the gopls subprocess.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("gopls is already running")
	}

	// Start gopls process (it defaults to stdio mode)
	m.cmd = exec.CommandContext(ctx, "gopls")

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

	// Monitor stderr for gopls errors and logging
	go func() {
		scanner := bufio.NewScanner(m.stderr)
		for scanner.Scan() {
			m.logger.Debug("gopls stderr", "output", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			m.logger.Error("error reading gopls stderr", "error", err)
		}
	}()

	// Start the continuous message reader
	go m.messageReader()

	// Give gopls a moment to start
	time.Sleep(100 * time.Millisecond)

	// Initialize gopls with workspace
	if err := m.initialize(ctx); err != nil {
		_ = m.Stop()
		return fmt.Errorf("failed to initialize gopls: %w", err)
	}

	return nil
}

// messageReader continuously reads messages from gopls stdout.
func (m *Manager) messageReader() {
	reader := bufio.NewReader(m.stdout)

	for m.running {
		message, err := m.readLSPMessage(reader)
		if err != nil {
			if m.running {
				m.logger.Error("error reading LSP message", "error", err)
			}
			return
		}

		m.handleLSPMessage(message)
	}
}

// readLSPMessage reads a single LSP message from the reader.
func (m *Manager) readLSPMessage(reader *bufio.Reader) (map[string]any, error) {
	// Read headers
	headers, err := m.readLSPHeaders(reader)
	if err != nil {
		return nil, err
	}

	// Get content length
	contentLength, err := m.getContentLength(headers)
	if err != nil {
		return nil, err
	}

	// Read the content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, content); err != nil {
		return nil, fmt.Errorf("failed to read message content: %w", err)
	}

	// Parse the JSON
	var message map[string]any
	if err := json.Unmarshal(content, &message); err != nil {
		return nil, fmt.Errorf("failed to parse LSP message: %w", err)
	}

	return message, nil
}

// readLSPHeaders reads LSP message headers until an empty line.
func (m *Manager) readLSPHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header line: %w", err)
		}

		// Trim the line ending
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		// Empty line marks end of headers
		if line == "" {
			break
		}

		// Parse header
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	return headers, nil
}

// getContentLength extracts and validates the Content-Length header.
func (m *Manager) getContentLength(headers map[string]string) (int, error) {
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return 0, fmt.Errorf("missing Content-Length header")
	}

	var contentLength int
	if _, err := fmt.Sscanf(contentLengthStr, "%d", &contentLength); err != nil {
		return 0, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
	}

	return contentLength, nil
}

// ensureFileOpen ensures a file is opened in gopls before making requests about it.
func (m *Manager) ensureFileOpen(fileURI string) error {
	// Wait for workspace to be ready before making any requests
	if err := m.waitForWorkspaceReady(30 * time.Second); err != nil {
		return fmt.Errorf("workspace not ready: %w", err)
	}

	// Check if file is already open
	m.openFilesMux.RLock()
	isOpen := m.openFiles[fileURI]
	m.openFilesMux.RUnlock()

	if isOpen {
		return nil // File already open
	}

	// Parse URI to get file path
	parsedURI, err := url.Parse(fileURI)
	if err != nil {
		return fmt.Errorf("invalid file URI: %w", err)
	}

	if parsedURI.Scheme != "file" {
		return fmt.Errorf("unsupported URI scheme: %s", parsedURI.Scheme)
	}

	filePath := parsedURI.Path

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Determine language ID based on file extension
	languageID := "go" // Default to Go
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		languageID = "go"
	case ".mod":
		languageID = "go.mod"
	case ".sum":
		languageID = "go.sum"
	}

	// Send textDocument/didOpen notification
	didOpenNotification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "textDocument/didOpen",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri":        fileURI,
				"languageId": languageID,
				"version":    1,
				"text":       string(content),
			},
		},
	}

	if err := m.sendRequest(didOpenNotification); err != nil {
		return fmt.Errorf("failed to send didOpen notification: %w", err)
	}

	// Mark file as open
	m.openFilesMux.Lock()
	m.openFiles[fileURI] = true
	m.openFilesMux.Unlock()

	m.logger.Info("opened file in gopls", "uri", fileURI)
	return nil
}

// isWorkspaceReady returns true if gopls has finished loading packages.
func (m *Manager) isWorkspaceReady() bool {
	m.workspaceReadyMux.RLock()
	defer m.workspaceReadyMux.RUnlock()
	return m.workspaceReady
}

// setWorkspaceReady marks the workspace as ready.
func (m *Manager) setWorkspaceReady() {
	m.workspaceReadyMux.Lock()
	defer m.workspaceReadyMux.Unlock()
	if !m.workspaceReady {
		m.workspaceReady = true
		m.logger.Info("workspace marked as ready for LSP requests")
	}
}

// waitForWorkspaceReady waits until gopls has finished loading packages.
func (m *Manager) waitForWorkspaceReady(timeout time.Duration) error {
	if m.isWorkspaceReady() {
		return nil
	}

	m.logger.Info("waiting for gopls to finish loading packages")

	start := time.Now()
	for time.Since(start) < timeout {
		if m.isWorkspaceReady() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for gopls workspace to be ready")
}

// handleLSPMessage routes an LSP message to the appropriate handler.
func (m *Manager) handleLSPMessage(message map[string]any) {
	id, hasID := message["id"]
	method, hasMethod := message["method"]

	switch {
	case hasID && !hasMethod:
		// This is a response to our request
		if idFloat, isFloat := id.(float64); isFloat {
			m.routeResponse(int(idFloat), message)
		}
	case hasID && hasMethod:
		// This is a request from the server (we don't handle these yet)
		m.logger.Debug("received request from gopls", "message", message)
	case hasMethod:
		// This is a notification
		m.logger.Debug("received notification from gopls", "message", message)
		if methodStr, ok := method.(string); ok {
			m.handleNotification(methodStr, message)
		}
	}
}

// handleNotification processes notifications from gopls.
func (m *Manager) handleNotification(method string, message map[string]any) {
	switch method {
	case "window/showMessage":
		m.handleShowMessage(message)
	case "$/progress":
		m.handleProgress(message)
	}
}

// handleShowMessage processes window/showMessage notifications.
func (m *Manager) handleShowMessage(message map[string]any) {
	params, paramsOK := message["params"].(map[string]any)
	if !paramsOK {
		return
	}

	messageText, msgOK := params["message"].(string)
	if !msgOK {
		return
	}

	if strings.Contains(messageText, "Finished loading packages") {
		m.setWorkspaceReady()
	}
}

// handleProgress processes $/progress notifications.
func (m *Manager) handleProgress(message map[string]any) {
	params, paramsOK := message["params"].(map[string]any)
	if !paramsOK {
		return
	}

	value, valueOK := params["value"].(map[string]any)
	if !valueOK {
		return
	}

	kind, kindOK := value["kind"].(string)
	if !kindOK || kind != "end" {
		return
	}

	messageText, msgOK := value["message"].(string)
	if !msgOK {
		return
	}

	if strings.Contains(messageText, "Finished loading packages") {
		m.setWorkspaceReady()
	}
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
	requestID := m.nextRequestID()
	initRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  "initialize",
		"params": map[string]any{
			"processId": os.Getpid(), // Use actual process ID instead of nil
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

	// Send initialize request and wait for response
	m.logger.Debug("sending initialize request", "requestID", requestID)
	response, err := m.sendRequestAndWait(initRequest)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	m.logger.Debug("received initialize response", "response", response)

	// Log server capabilities for debugging
	if result, ok := response["result"].(map[string]any); ok {
		if capabilities, capOk := result["capabilities"]; capOk {
			m.logger.Debug("gopls server capabilities", "capabilities", capabilities)
		}
	}

	// Send initialized notification (no response expected)
	initializedNotification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]any{},
	}

	if err := m.sendRequest(initializedNotification); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	m.logger.Info("gopls initialized successfully")
	return nil
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

	// Log outgoing request for debugging
	if method, ok := request["method"].(string); ok {
		m.logger.Debug("sending LSP request", "method", method, "id", request["id"])
		m.logger.Debug("request data", "data", string(data))
	}

	// LSP uses Content-Length header format
	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)
	m.logger.Debug("full message being sent", "message", message)

	_, err = m.stdin.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	// Ensure the data is flushed
	if flusher, ok := m.stdin.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			m.logger.Warn("failed to flush stdin", "error", err)
		}
	}

	return nil
}

// routeResponse routes a response to the appropriate request handler.
func (m *Manager) routeResponse(id int, response map[string]any) {
	m.responsesMux.Lock()
	ch, ok := m.responses[id]
	if ok {
		delete(m.responses, id)
	}
	m.responsesMux.Unlock()

	if ok {
		select {
		case ch <- response:
			// Response delivered
		default:
			m.logger.Warn("response channel full", "requestID", id)
		}
	} else {
		m.logger.Warn("received response for unknown request ID", "requestID", id)
	}
}

// sendRequestAndWait sends a request and waits for the response.
func (m *Manager) sendRequestAndWait(request map[string]any) (map[string]any, error) {
	id, ok := request["id"].(int)
	if !ok {
		return nil, fmt.Errorf("request missing integer ID")
	}

	// Create response channel
	responseCh := make(chan map[string]any, 1)
	m.responsesMux.Lock()
	m.responses[id] = responseCh
	m.responsesMux.Unlock()

	// Send the request
	if err := m.sendRequest(request); err != nil {
		m.responsesMux.Lock()
		delete(m.responses, id)
		m.responsesMux.Unlock()
		return nil, err
	}

	// Wait for response with timeout (60s for large codebases)
	m.logger.Debug("waiting for response", "requestID", id)
	startTime := time.Now()

	// Progress ticker to show we're still waiting
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case response := <-responseCh:
			elapsed := time.Since(startTime)
			m.logger.Debug("received response", "requestID", id, "elapsed", elapsed)

			// Check for LSP error
			if errorField, hasError := response["error"]; hasError {
				errorMap, _ := errorField.(map[string]any)
				code, _ := errorMap["code"].(float64)
				message, _ := errorMap["message"].(string)
				return nil, fmt.Errorf("LSP error %d: %s", int(code), message)
			}
			return response, nil

		case <-ticker.C:
			elapsed := time.Since(startTime)
			m.logger.Info("still waiting for response", "requestID", id, "elapsed", elapsed)

		case <-time.After(60 * time.Second):
			m.responsesMux.Lock()
			delete(m.responses, id)
			m.responsesMux.Unlock()
			elapsed := time.Since(startTime)
			msg := "timeout waiting for response to request %d after %v (gopls may be processing large codebase)"
			return nil, fmt.Errorf(msg, id, elapsed)
		}
	}
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
func (m *Manager) HandleGoToDefinition(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GoToDefinitionParams],
) (*mcp.CallToolResultFor[GoToDefinitionResult], error) {
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
func (m *Manager) HandleFindReferences(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[FindReferencesParams],
) (*mcp.CallToolResultFor[FindReferencesResult], error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	locations, err := m.FindReferences(
		ctx,
		params.Arguments.URI,
		params.Arguments.Line,
		params.Arguments.Character,
		params.Arguments.IncludeDeclaration,
	)
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
func (m *Manager) HandleGetHover(
	ctx context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetHoverParams],
) (*mcp.CallToolResultFor[GetHoverResult], error) {
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
