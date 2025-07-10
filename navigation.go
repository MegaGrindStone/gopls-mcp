package main

import (
	"fmt"
)

// goToDefinition sends a textDocument/definition request to gopls using relative paths.
//
//nolint:dupl // LSP methods follow similar request/response patterns
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
//
//nolint:dupl // LSP methods follow similar request/response patterns
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

// getTypeDefinition sends a textDocument/typeDefinition request to gopls.
//
//nolint:dupl // LSP methods follow similar request/response patterns
func (c *goplsClient) getTypeDefinition(relativePath string, line, character int) ([]Location, error) {
	c.logger.Debug("getTypeDefinition called", "relativePath", relativePath, "line", line, "character", character)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/typeDefinition request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/typeDefinition",
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
		return nil, fmt.Errorf("failed to get type definition: %w", err)
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

// findImplementations sends a textDocument/implementation request to gopls.
//
//nolint:dupl // LSP methods follow similar request/response patterns
func (c *goplsClient) findImplementations(relativePath string, line, character int) ([]Location, error) {
	c.logger.Debug("findImplementations called", "relativePath", relativePath, "line", line, "character", character)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/implementation request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/implementation",
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
		return nil, fmt.Errorf("failed to find implementations: %w", err)
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
