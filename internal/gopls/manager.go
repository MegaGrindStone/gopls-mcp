package gopls

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Manager manages a gopls subprocess and handles LSP communication
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

// NewManager creates a new gopls manager with the specified workspace path
func NewManager(workspacePath string) *Manager {
	return &Manager{
		workspacePath: workspacePath,
	}
}

// Start starts the gopls subprocess
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

// Stop stops the gopls subprocess
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

// IsRunning returns true if gopls is currently running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// nextRequestID generates the next request ID for LSP communication
func (m *Manager) nextRequestID() int {
	m.requestIDMux.Lock()
	defer m.requestIDMux.Unlock()
	m.requestID++
	return m.requestID
}

// initialize sends the LSP initialize request to gopls
func (m *Manager) initialize(ctx context.Context) error {
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
					"references": map[string]any{},
					"documentSymbol": map[string]any{},
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

// sendRequest sends a JSON-RPC request to gopls
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

// readResponse reads a JSON-RPC response from gopls
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