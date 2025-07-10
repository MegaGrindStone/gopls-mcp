package main

import (
	"fmt"
)

// getDocumentSymbols sends a textDocument/documentSymbol request to gopls.
func (c *goplsClient) getDocumentSymbols(relativePath string) ([]DocumentSymbol, error) {
	c.logger.Debug("getDocumentSymbols called", "relativePath", relativePath)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/documentSymbol request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/documentSymbol",
		"params": map[string]any{
			"textDocument": map[string]any{
				"uri": fileURI,
			},
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get document symbols: %w", err)
	}

	// Parse response to get symbols
	symbols, err := c.parseDocumentSymbolsFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document symbols: %w", err)
	}

	return symbols, nil
}

// parseDocumentSymbolsFromResponse extracts document symbols from LSP response.
func (c *goplsClient) parseDocumentSymbolsFromResponse(response map[string]any) ([]DocumentSymbol, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	if result == nil {
		return []DocumentSymbol{}, nil
	}

	symbolsData, symbolsOk := result.([]any)
	if !symbolsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	var symbols []DocumentSymbol
	for _, symbolData := range symbolsData {
		if symbolMap, ok := symbolData.(map[string]any); ok {
			symbol := c.parseDocumentSymbol(symbolMap)
			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}

// parseDocumentSymbol parses a document symbol from a map.
func (c *goplsClient) parseDocumentSymbol(symbolMap map[string]any) DocumentSymbol {
	var symbol DocumentSymbol

	if name, ok := symbolMap["name"].(string); ok {
		symbol.Name = name
	}

	if detail, ok := symbolMap["detail"].(string); ok {
		symbol.Detail = detail
	}

	if kind, ok := symbolMap["kind"].(float64); ok {
		symbol.Kind = int(kind)
	}

	if deprecated, ok := symbolMap["deprecated"].(bool); ok {
		symbol.Deprecated = deprecated
	}

	if rangeMap, ok := symbolMap["range"].(map[string]any); ok {
		symbol.Range = c.parseRange(rangeMap)
	}

	if selectionRangeMap, ok := symbolMap["selectionRange"].(map[string]any); ok {
		symbol.SelectionRange = c.parseRange(selectionRangeMap)
	}

	if childrenData, ok := symbolMap["children"].([]any); ok {
		for _, childData := range childrenData {
			if childMap, childOk := childData.(map[string]any); childOk {
				child := c.parseDocumentSymbol(childMap)
				symbol.Children = append(symbol.Children, child)
			}
		}
	}

	return symbol
}

// getWorkspaceSymbols sends a workspace/symbol request to gopls.
func (c *goplsClient) getWorkspaceSymbols(query string) ([]SymbolInformation, error) {
	c.logger.Debug("getWorkspaceSymbols called", "query", query)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Create workspace/symbol request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "workspace/symbol",
		"params": map[string]any{
			"query": query,
		},
	}

	// Send request and wait for response
	response, err := c.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace symbols: %w", err)
	}

	// Parse response to get symbols
	symbols, err := c.parseWorkspaceSymbolsFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workspace symbols: %w", err)
	}

	// Convert URIs back to relative paths
	for i := range symbols {
		symbols[i].Location.URI = c.uriToRelativePath(symbols[i].Location.URI)
	}

	return symbols, nil
}

// parseWorkspaceSymbolsFromResponse extracts workspace symbols from LSP response.
func (c *goplsClient) parseWorkspaceSymbolsFromResponse(response map[string]any) ([]SymbolInformation, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	if result == nil {
		return []SymbolInformation{}, nil
	}

	symbolsData, symbolsOk := result.([]any)
	if !symbolsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	var symbols []SymbolInformation
	for _, symbolData := range symbolsData {
		if symbolMap, ok := symbolData.(map[string]any); ok {
			symbol := c.parseSymbolInformation(symbolMap)
			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}

// parseSymbolInformation parses a symbol information from a map.
func (c *goplsClient) parseSymbolInformation(symbolMap map[string]any) SymbolInformation {
	var symbol SymbolInformation

	if name, ok := symbolMap["name"].(string); ok {
		symbol.Name = name
	}

	if kind, ok := symbolMap["kind"].(float64); ok {
		symbol.Kind = int(kind)
	}

	if deprecated, ok := symbolMap["deprecated"].(bool); ok {
		symbol.Deprecated = deprecated
	}

	if location, ok := symbolMap["location"].(map[string]any); ok {
		symbol.Location = c.parseLocationFromMap(location)
	}

	if containerName, ok := symbolMap["containerName"].(string); ok {
		symbol.ContainerName = containerName
	}

	return symbol
}
