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
)

const (
	fileScheme = "file"
)

// LSP types for gopls communication

// Position represents a position in a document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// TextDocumentPositionParams represents parameters for text document position requests.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// ReferenceParams represents parameters for find references requests.
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

// ReferenceContext represents context for find references requests.
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// Hover represents hover information.
type Hover struct {
	Contents []string `json:"contents"`
	Range    *Range   `json:"range,omitempty"`
}

// goplsClient manages a gopls subprocess and handles basic LSP communication.
type goplsClient struct {
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	workspacePath string
	mu            sync.RWMutex
	running       bool
	logger        *slog.Logger
	requestID     int
	requestIDMux  sync.Mutex
	responses     map[int]chan map[string]any
	responsesMux  sync.Mutex
	openFiles     map[string]bool
	openFilesMux  sync.RWMutex
}

// newClient creates a new gopls client with the specified workspace path.
func newClient(workspacePath string, logger *slog.Logger) *goplsClient {
	c := &goplsClient{
		workspacePath: workspacePath,
		logger:        logger,
		responses:     make(map[int]chan map[string]any),
		openFiles:     make(map[string]bool),
	}
	c.logger.Debug("created new gopls client", "workspacePath", workspacePath)
	return c
}

// start starts the gopls subprocess and initializes it.
func (c *goplsClient) start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("gopls is already running")
	}

	// Validate workspace path exists and is a directory
	info, err := os.Stat(c.workspacePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("workspace path does not exist: %s", c.workspacePath)
		}
		return fmt.Errorf("failed to access workspace path %s: %w", c.workspacePath, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("workspace path is not a directory: %s", c.workspacePath)
	}

	// Start gopls process
	c.cmd = exec.CommandContext(ctx, "gopls")

	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gopls: %w", err)
	}

	c.running = true
	c.logger.Info("gopls process started", "pid", c.cmd.Process.Pid)

	// Monitor stderr for gopls errors
	go c.monitorStderr()

	// Start message reader for LSP responses
	go c.messageReader()

	// Give gopls a moment to start
	time.Sleep(100 * time.Millisecond)

	// Initialize gopls
	if err := c.initialize(); err != nil {
		c.logger.Error("gopls initialization failed", "error", err)
		_ = c.stop()
		return fmt.Errorf("failed to initialize gopls: %w", err)
	}

	c.logger.Info("gopls client started successfully")
	return nil
}

// stop stops the gopls subprocess.
func (c *goplsClient) stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	var err error
	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.stdout != nil {
		_ = c.stdout.Close()
	}
	if c.stderr != nil {
		_ = c.stderr.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		err = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}

	c.running = false
	c.cmd = nil
	c.stdin = nil
	c.stdout = nil
	c.stderr = nil

	c.logger.Info("gopls client stopped")
	return err
}

// isRunning returns true if gopls is currently running.
func (c *goplsClient) isRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// initialize sends the LSP initialize request to gopls.
func (c *goplsClient) initialize() error {
	c.logger.Info("initializing gopls", "workspacePath", c.workspacePath)

	requestID := c.nextRequestID()
	initRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  "initialize",
		"params": map[string]any{
			"processId": os.Getpid(),
			"rootUri":   fmt.Sprintf("file://%s", c.workspacePath),
			"capabilities": map[string]any{
				"textDocument": map[string]any{
					"hover": map[string]any{
						"contentFormat": []string{"markdown", "plaintext"},
					},
					"definition": map[string]any{
						"linkSupport": true,
					},
					"references": map[string]any{},
				},
				"workspace": map[string]any{
					"workspaceFolders": true,
				},
			},
		},
	}

	// Send initialize request
	if err := c.sendRequest(initRequest); err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	// Send initialized notification
	initializedNotification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]any{},
	}

	if err := c.sendRequest(initializedNotification); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	c.logger.Info("gopls initialized successfully")
	return nil
}

// sendRequest sends a JSON-RPC request to gopls.
func (c *goplsClient) sendRequest(request map[string]any) error {
	if !c.running {
		return fmt.Errorf("gopls is not running")
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// LSP uses Content-Length header format
	message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	if method, ok := request["method"].(string); ok {
		c.logger.Debug("sending LSP request", "method", method, "id", request["id"])
	}

	_, err = c.stdin.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// nextRequestID generates the next request ID for LSP communication.
func (c *goplsClient) nextRequestID() int {
	c.requestIDMux.Lock()
	defer c.requestIDMux.Unlock()
	c.requestID++
	return c.requestID
}

// monitorStderr monitors stderr output from gopls.
func (c *goplsClient) monitorStderr() {
	c.logger.Debug("starting stderr monitor")
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		c.logger.Debug("gopls stderr", "output", line)

		// Log errors and warnings at higher level
		lowerLine := strings.ToLower(line)
		if c.isErrorLine(lowerLine) {
			c.logger.Error("gopls stderr error", "output", line)
		} else if strings.Contains(lowerLine, "warn") {
			c.logger.Warn("gopls stderr warning", "output", line)
		}
	}
	if err := scanner.Err(); err != nil {
		c.logger.Error("error reading gopls stderr", "error", err)
	}
	c.logger.Debug("stderr monitor exited")
}

// isErrorLine checks if a log line contains error indicators.
func (c *goplsClient) isErrorLine(lowerLine string) bool {
	return strings.Contains(lowerLine, "error") ||
		strings.Contains(lowerLine, "panic") ||
		strings.Contains(lowerLine, "fatal")
}

// messageReader continuously reads messages from gopls stdout.
func (c *goplsClient) messageReader() {
	c.logger.Debug("starting message reader")
	reader := bufio.NewReader(c.stdout)

	for c.isRunning() {
		message, err := c.readLSPMessage(reader)
		if err != nil {
			if c.isRunning() {
				c.logger.Error("error reading LSP message", "error", err)
			}
			return
		}

		c.handleLSPMessage(message)
	}
	c.logger.Debug("message reader exited")
}

// readLSPMessage reads a single LSP message from the reader.
func (c *goplsClient) readLSPMessage(reader *bufio.Reader) (map[string]any, error) {
	// Read headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header line: %w", err)
		}

		// Trim line ending
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

	// Get content length
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	var contentLength int
	if _, err := fmt.Sscanf(contentLengthStr, "%d", &contentLength); err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
	}

	// Read content
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, content); err != nil {
		return nil, fmt.Errorf("failed to read message content: %w", err)
	}

	// Parse JSON
	var message map[string]any
	if err := json.Unmarshal(content, &message); err != nil {
		return nil, fmt.Errorf("failed to parse LSP message: %w", err)
	}

	return message, nil
}

// handleLSPMessage routes an LSP message to the appropriate handler.
func (c *goplsClient) handleLSPMessage(message map[string]any) {
	id, hasID := message["id"]
	method, hasMethod := message["method"]

	c.logger.Debug("handleLSPMessage", "hasID", hasID, "hasMethod", hasMethod, "id", id, "method", method)

	switch {
	case hasID && !hasMethod:
		// This is a response to our request
		if idFloat, isFloat := id.(float64); isFloat {
			c.logger.Debug("routing response", "requestID", int(idFloat))
			c.routeResponse(int(idFloat), message)
		} else {
			c.logger.Debug("response with non-float ID", "id", id, "type", fmt.Sprintf("%T", id))
		}
	case hasMethod:
		// This is a notification or request from server
		c.logger.Debug("received notification/request from gopls", "method", method)
	default:
		c.logger.Debug("unhandled LSP message", "message", message)
	}
}

// routeResponse routes a response to the appropriate request handler.
func (c *goplsClient) routeResponse(id int, response map[string]any) {
	c.responsesMux.Lock()
	ch, ok := c.responses[id]
	if ok {
		delete(c.responses, id)
	}
	c.responsesMux.Unlock()

	if ok {
		select {
		case ch <- response:
			// Response delivered
		default:
			c.logger.Warn("response channel full", "requestID", id)
		}
	} else {
		c.logger.Warn("received response for unknown request ID", "requestID", id)
	}
}

// sendRequestAndWait sends a request and waits for the response.
func (c *goplsClient) sendRequestAndWait(request map[string]any) (map[string]any, error) {
	id, ok := request["id"].(int)
	if !ok {
		return nil, fmt.Errorf("request missing integer ID")
	}

	// Create response channel
	responseCh := make(chan map[string]any, 1)
	c.responsesMux.Lock()
	c.responses[id] = responseCh
	c.responsesMux.Unlock()

	// Send the request
	if err := c.sendRequest(request); err != nil {
		c.responsesMux.Lock()
		delete(c.responses, id)
		c.responsesMux.Unlock()
		return nil, err
	}

	// Wait for response with timeout
	c.logger.Debug("waiting for response", "requestID", id)
	select {
	case response := <-responseCh:
		c.logger.Debug("received response", "requestID", id)

		// Check for LSP error
		if errorField, hasError := response["error"]; hasError {
			errorMap, _ := errorField.(map[string]any)
			code, _ := errorMap["code"].(float64)
			message, _ := errorMap["message"].(string)
			return nil, fmt.Errorf("LSP error %d: %s", int(code), message)
		}
		return response, nil

	case <-time.After(30 * time.Second):
		c.responsesMux.Lock()
		delete(c.responses, id)
		c.responsesMux.Unlock()
		return nil, fmt.Errorf("timeout waiting for response to request %d", id)
	}
}

// relativePathToURI converts a workspace-relative path to a file:// URI.
func (c *goplsClient) relativePathToURI(relativePath string) string {
	absolutePath := filepath.Join(c.workspacePath, relativePath)
	return fmt.Sprintf("file://%s", absolutePath)
}

// uriToRelativePath converts a file:// URI to a workspace-relative path.
func (c *goplsClient) uriToRelativePath(uri string) string {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return uri // Return as-is if parsing fails
	}

	if parsedURI.Scheme != fileScheme {
		return uri // Return as-is if not a file URI
	}

	absolutePath := parsedURI.Path
	relativePath, err := filepath.Rel(c.workspacePath, absolutePath)
	if err != nil {
		return uri // Return as-is if can't make relative
	}

	return relativePath
}

// ensureFileOpen ensures a file is opened in gopls before making requests about it.
func (c *goplsClient) ensureFileOpen(relativePath string) error {
	// Check if file is already open
	c.openFilesMux.RLock()
	isOpen := c.openFiles[relativePath]
	c.openFilesMux.RUnlock()

	if isOpen {
		return nil // File already open
	}

	// Convert to absolute path and URI
	absolutePath := filepath.Join(c.workspacePath, relativePath)
	fileURI := c.relativePathToURI(relativePath)

	// Read file content
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", absolutePath, err)
	}

	// Determine language ID based on file extension
	languageID := "go" // Default to Go
	ext := filepath.Ext(relativePath)
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

	if err := c.sendRequest(didOpenNotification); err != nil {
		return fmt.Errorf("failed to send didOpen notification: %w", err)
	}

	// Mark file as open
	c.openFilesMux.Lock()
	c.openFiles[relativePath] = true
	c.openFilesMux.Unlock()

	c.logger.Debug("opened file in gopls", "relativePath", relativePath, "uri", fileURI)
	return nil
}

// goToDefinition sends a textDocument/definition request to gopls using relative paths.
func (c *goplsClient) goToDefinition(relativePath string, line, character int) ([]Location, error) {
	c.logger.Debug("goToDefinition called", "relativePath", relativePath, "line", line, "character", character)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/definition request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/definition",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": fileURI,
			},
			"position": map[string]any{
				"line":      line,
				"character": character,
			},
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	// Parse response to get locations
	locations, err := c.parseLocationsFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse locations: %w", err)
	}

	// Convert URIs back to relative paths
	for i := range locations {
		locations[i].URI = c.uriToRelativePath(locations[i].URI)
	}

	return locations, nil
}

// findReferences sends a textDocument/references request to gopls using relative paths.
func (c *goplsClient) findReferences(
	relativePath string, line, character int, includeDeclaration bool,
) ([]Location, error) {
	c.logger.Debug("findReferences called",
		"relativePath", relativePath,
		"line", line,
		"character", character,
		"includeDeclaration", includeDeclaration)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/references request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/references",
		"params": ReferenceParams{
			TextDocumentPositionParams: TextDocumentPositionParams{
				TextDocument: TextDocumentIdentifier{URI: fileURI},
				Position: Position{
					Line:      line,
					Character: character,
				},
			},
			Context: ReferenceContext{
				IncludeDeclaration: includeDeclaration,
			},
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	// Parse response to get locations
	locations, err := c.parseLocationsFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse locations: %w", err)
	}

	// Convert URIs back to relative paths
	for i := range locations {
		locations[i].URI = c.uriToRelativePath(locations[i].URI)
	}

	return locations, nil
}

// getHover sends a textDocument/hover request to gopls using relative paths.
func (c *goplsClient) getHover(relativePath string, line, character int) (*Hover, error) {
	c.logger.Debug("getHover called", "relativePath", relativePath, "line", line, "character", character)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/hover request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/hover",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": fileURI,
			},
			"position": map[string]any{
				"line":      line,
				"character": character,
			},
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get hover info: %w", err)
	}

	// Parse response to get hover info
	hover, err := c.parseHoverFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hover info: %w", err)
	}

	return hover, nil
}

// parseLocationsFromResponse extracts locations from LSP response.
func (c *goplsClient) parseLocationsFromResponse(response map[string]any) ([]Location, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	locations, locationsOk := result.([]any)
	if !locationsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	var locs []Location
	for _, loc := range locations {
		if locMap, locMapOk := loc.(map[string]any); locMapOk {
			location := c.parseLocationFromMap(locMap)
			locs = append(locs, location)
		}
	}
	return locs, nil
}

// parseLocationFromMap parses a single location from a map.
func (c *goplsClient) parseLocationFromMap(locMap map[string]any) Location {
	var location Location
	if locURI, uriOk := locMap["uri"].(string); uriOk {
		location.URI = locURI
	}
	if rangeMap, rangeMapOk := locMap["range"].(map[string]any); rangeMapOk {
		location.Range = c.parseRange(rangeMap)
	}
	return location
}

// parseRange parses a range from a map.
func (c *goplsClient) parseRange(rangeMap map[string]any) Range {
	var rng Range
	if startMap, startMapOk := rangeMap["start"].(map[string]any); startMapOk {
		if line, lineOk := startMap["line"].(float64); lineOk {
			rng.Start.Line = int(line)
		}
		if char, charOk := startMap["character"].(float64); charOk {
			rng.Start.Character = int(char)
		}
	}
	if endMap, endMapOk := rangeMap["end"].(map[string]any); endMapOk {
		if line, lineOk := endMap["line"].(float64); lineOk {
			rng.End.Line = int(line)
		}
		if char, charOk := endMap["character"].(float64); charOk {
			rng.End.Character = int(char)
		}
	}
	return rng
}

// parseHoverContents parses hover contents from any type.
func (c *goplsClient) parseHoverContents(contents any) []string {
	var result []string

	// Handle string directly
	if contentStr, contentStrOk := contents.(string); contentStrOk {
		result = append(result, contentStr)
		return result
	}

	// Handle single MarkupContent object
	if contentMap, contentMapOk := contents.(map[string]any); contentMapOk {
		if _, kindOk := contentMap["kind"].(string); kindOk {
			if value, valueOk := contentMap["value"].(string); valueOk {
				result = append(result, value)
			}
		}
		return result
	}

	// Handle array of contents
	if contentList, contentListOk := contents.([]any); contentListOk {
		for _, content := range contentList {
			if contentStr, contentStrOk := content.(string); contentStrOk {
				result = append(result, contentStr)
				continue
			}

			contentMap, contentMapOk := content.(map[string]any)
			if !contentMapOk {
				continue
			}

			// Handle MarkupContent format
			if _, kindOk := contentMap["kind"].(string); !kindOk {
				continue
			}

			if value, valueOk := contentMap["value"].(string); valueOk {
				result = append(result, value)
			}
		}
	}

	return result
}

// parseHoverFromResponse extracts hover information from LSP response.
func (c *goplsClient) parseHoverFromResponse(response map[string]any) (*Hover, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid hover response format")
	}

	// Handle null result (no hover info available)
	if result == nil {
		return &Hover{Contents: []string{}}, nil
	}

	hoverMap, hoverMapOk := result.(map[string]any)
	if !hoverMapOk {
		return nil, fmt.Errorf("invalid hover response format")
	}

	var hover Hover
	if contents, contentsOk := hoverMap["contents"]; contentsOk {
		hover.Contents = c.parseHoverContents(contents)
	}
	if rangeMap, rangeMapOk := hoverMap["range"].(map[string]any); rangeMapOk {
		rng := c.parseRange(rangeMap)
		hover.Range = &rng
	}
	return &hover, nil
}
