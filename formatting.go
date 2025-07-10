package main

import (
	"fmt"
)

// formatDocument sends a textDocument/formatting request to gopls.
func (c *goplsClient) formatDocument(relativePath string) ([]TextEdit, error) {
	c.logger.Debug("formatDocument called", "relativePath", relativePath)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/formatting request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/formatting",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": fileURI,
			},
			"options": map[string]any{
				"tabSize":      4,
				"insertSpaces": false,
			},
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to format document: %w", err)
	}

	// Parse response to get text edits
	textEdits, err := c.parseTextEditsFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text edits: %w", err)
	}

	return textEdits, nil
}

// organizeImports sends a source.organizeImports code action request to gopls.
func (c *goplsClient) organizeImports(relativePath string) ([]TextEdit, error) {
	c.logger.Debug("organizeImports called", "relativePath", relativePath)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/codeAction request for organizing imports
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/codeAction",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": fileURI,
			},
			"range": map[string]any{
				"start": map[string]any{
					"line":      0,
					"character": 0,
				},
				"end": map[string]any{
					"line":      0,
					"character": 0,
				},
			},
			"context": map[string]any{
				"only": []string{"source.organizeImports"},
			},
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to organize imports: %w", err)
	}

	// Parse response to get workspace edit
	workspaceEdit, err := c.parseWorkspaceEditFromCodeActionResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workspace edit: %w", err)
	}

	// Extract text edits for the current file
	if changes, ok := workspaceEdit.Changes[fileURI]; ok {
		return changes, nil
	}

	return []TextEdit{}, nil
}

// getInlayHints sends a textDocument/inlayHint request to gopls.
func (c *goplsClient) getInlayHints(relativePath string, startLine, startChar, endLine, endChar int) (
	[]InlayHint, error) {
	c.logger.Debug("getInlayHints called",
		"relativePath", relativePath, "startLine", startLine,
		"startChar", startChar, "endLine", endLine, "endChar", endChar)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/inlayHint request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/inlayHint",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": fileURI,
			},
			"range": map[string]any{
				"start": map[string]any{
					"line":      startLine,
					"character": startChar,
				},
				"end": map[string]any{
					"line":      endLine,
					"character": endChar,
				},
			},
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get inlay hints: %w", err)
	}

	// Parse response to get inlay hints
	inlayHints, err := c.parseInlayHintsFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inlay hints: %w", err)
	}

	return inlayHints, nil
}

// parseTextEditsFromResponse extracts text edits from LSP response.
func (c *goplsClient) parseTextEditsFromResponse(response map[string]any) ([]TextEdit, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	if result == nil {
		return []TextEdit{}, nil
	}

	editsData, editsOk := result.([]any)
	if !editsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	var textEdits []TextEdit
	for _, editData := range editsData {
		if editMap, ok := editData.(map[string]any); ok {
			textEdit := c.parseTextEdit(editMap)
			textEdits = append(textEdits, textEdit)
		}
	}

	return textEdits, nil
}

// parseTextEdit parses a text edit from a map.
func (c *goplsClient) parseTextEdit(editMap map[string]any) TextEdit {
	var textEdit TextEdit

	if rangeMap, ok := editMap["range"].(map[string]any); ok {
		textEdit.Range = c.parseRange(rangeMap)
	}

	if newText, ok := editMap["newText"].(string); ok {
		textEdit.NewText = newText
	}

	return textEdit
}

// parseWorkspaceEditFromCodeActionResponse extracts workspace edit from code action response.
func (c *goplsClient) parseWorkspaceEditFromCodeActionResponse(response map[string]any) (*WorkspaceEdit, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	if result == nil {
		return &WorkspaceEdit{Changes: make(map[string][]TextEdit)}, nil
	}

	actionsData, actionsOk := result.([]any)
	if !actionsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	for _, actionData := range actionsData {
		if actionMap, ok := actionData.(map[string]any); ok {
			if editMap, editOk := actionMap["edit"].(map[string]any); editOk {
				return c.parseWorkspaceEdit(editMap), nil
			}
		}
	}

	return &WorkspaceEdit{Changes: make(map[string][]TextEdit)}, nil
}

// parseWorkspaceEdit parses a workspace edit from a map.
func (c *goplsClient) parseWorkspaceEdit(editMap map[string]any) *WorkspaceEdit {
	workspaceEdit := &WorkspaceEdit{
		Changes: make(map[string][]TextEdit),
	}

	if changesMap, ok := editMap["changes"].(map[string]any); ok {
		for uri, changesData := range changesMap {
			if editsData, editsOk := changesData.([]any); editsOk {
				var textEdits []TextEdit
				for _, editData := range editsData {
					if editDataMap, editDataOk := editData.(map[string]any); editDataOk {
						textEdit := c.parseTextEdit(editDataMap)
						textEdits = append(textEdits, textEdit)
					}
				}
				workspaceEdit.Changes[uri] = textEdits
			}
		}
	}

	return workspaceEdit
}

// parseInlayHintsFromResponse extracts inlay hints from LSP response.
func (c *goplsClient) parseInlayHintsFromResponse(response map[string]any) ([]InlayHint, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	if result == nil {
		return []InlayHint{}, nil
	}

	hintsData, hintsOk := result.([]any)
	if !hintsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	var inlayHints []InlayHint
	for _, hintData := range hintsData {
		if hintMap, ok := hintData.(map[string]any); ok {
			inlayHint := c.parseInlayHint(hintMap)
			inlayHints = append(inlayHints, inlayHint)
		}
	}

	return inlayHints, nil
}

// parseInlayHint parses an inlay hint from a map.
func (c *goplsClient) parseInlayHint(hintMap map[string]any) InlayHint {
	var inlayHint InlayHint

	if positionMap, ok := hintMap["position"].(map[string]any); ok {
		if line, lineOk := positionMap["line"].(float64); lineOk {
			inlayHint.Position.Line = int(line)
		}
		if char, charOk := positionMap["character"].(float64); charOk {
			inlayHint.Position.Character = int(char)
		}
	}

	if label, ok := hintMap["label"].(string); ok {
		inlayHint.Label = label
	}

	if kind, ok := hintMap["kind"].(float64); ok {
		inlayHint.Kind = int(kind)
	}

	if tooltip, ok := hintMap["tooltip"].(string); ok {
		inlayHint.Tooltip = tooltip
	}

	return inlayHint
}
