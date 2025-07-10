package main

import (
	"fmt"
)

// getSignatureHelp sends a textDocument/signatureHelp request to gopls.
//
//nolint:dupl // LSP methods follow similar request/response patterns
func (c *goplsClient) getSignatureHelp(relativePath string, line, character int) (*SignatureHelp, error) {
	c.logger.Debug("getSignatureHelp called", "relativePath", relativePath, "line", line, "character", character)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/signatureHelp request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/signatureHelp",
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
		return nil, fmt.Errorf("failed to get signature help: %w", err)
	}

	// Parse response to get signature help
	signatureHelp, err := c.parseSignatureHelpFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature help: %w", err)
	}

	return signatureHelp, nil
}

// parseSignatureHelpFromResponse extracts signature help from LSP response.
func (c *goplsClient) parseSignatureHelpFromResponse(response map[string]any) (*SignatureHelp, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	if result == nil {
		return &SignatureHelp{}, nil
	}

	signatureHelpMap, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	var signatureHelp SignatureHelp

	if signaturesData, sigOk := signatureHelpMap["signatures"].([]any); sigOk {
		for _, sigData := range signaturesData {
			if sigMap, mapOk := sigData.(map[string]any); mapOk {
				signature := c.parseSignatureInformation(sigMap)
				signatureHelp.Signatures = append(signatureHelp.Signatures, signature)
			}
		}
	}

	if activeSignature, activeOk := signatureHelpMap["activeSignature"].(float64); activeOk {
		signatureHelp.ActiveSignature = int(activeSignature)
	}

	if activeParameter, paramOk := signatureHelpMap["activeParameter"].(float64); paramOk {
		signatureHelp.ActiveParameter = int(activeParameter)
	}

	return &signatureHelp, nil
}

// parseSignatureInformation parses signature information from a map.
func (c *goplsClient) parseSignatureInformation(sigMap map[string]any) SignatureInformation {
	var signature SignatureInformation

	if label, ok := sigMap["label"].(string); ok {
		signature.Label = label
	}

	if documentation, ok := sigMap["documentation"].(string); ok {
		signature.Documentation = documentation
	}

	if parametersData, ok := sigMap["parameters"].([]any); ok {
		for _, paramData := range parametersData {
			if paramMap, paramOk := paramData.(map[string]any); paramOk {
				parameter := c.parseParameterInformation(paramMap)
				signature.Parameters = append(signature.Parameters, parameter)
			}
		}
	}

	return signature
}

// parseParameterInformation parses parameter information from a map.
func (c *goplsClient) parseParameterInformation(paramMap map[string]any) ParameterInformation {
	var parameter ParameterInformation

	if label, ok := paramMap["label"].(string); ok {
		parameter.Label = label
	}

	if documentation, ok := paramMap["documentation"].(string); ok {
		parameter.Documentation = documentation
	}

	return parameter
}

// getCompletions sends a textDocument/completion request to gopls.
//
//nolint:dupl // LSP methods follow similar request/response patterns
func (c *goplsClient) getCompletions(relativePath string, line, character int) (*CompletionList, error) {
	c.logger.Debug("getCompletions called", "relativePath", relativePath, "line", line, "character", character)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Convert relative path to URI for LSP request
	fileURI := c.relativePathToURI(relativePath)

	// Create textDocument/completion request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextRequestID(),
		"method":  "textDocument/completion",
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
		return nil, fmt.Errorf("failed to get completions: %w", err)
	}

	// Parse response to get completions
	completions, err := c.parseCompletionListFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse completions: %w", err)
	}

	return completions, nil
}

// parseCompletionListFromResponse extracts completion list from LSP response.
func (c *goplsClient) parseCompletionListFromResponse(response map[string]any) (*CompletionList, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	if result == nil {
		return &CompletionList{}, nil
	}

	// Handle both CompletionList and CompletionItem[] formats
	//nolint:nestif // Complex JSON parsing requires nested type assertions
	if completionMap, ok := result.(map[string]any); ok {
		// This is a CompletionList
		var completionList CompletionList

		if isIncomplete, incompleteOk := completionMap["isIncomplete"].(bool); incompleteOk {
			completionList.IsIncomplete = isIncomplete
		}

		if itemsData, itemsOk := completionMap["items"].([]any); itemsOk {
			for _, itemData := range itemsData {
				if itemMap, itemMapOk := itemData.(map[string]any); itemMapOk {
					item := c.parseCompletionItem(itemMap)
					completionList.Items = append(completionList.Items, item)
				}
			}
		}

		return &completionList, nil
	} else if itemsData, itemsDataOk := result.([]any); itemsDataOk {
		// This is a CompletionItem[]
		var completionList CompletionList
		for _, itemData := range itemsData {
			if itemMap, itemMapOk := itemData.(map[string]any); itemMapOk {
				item := c.parseCompletionItem(itemMap)
				completionList.Items = append(completionList.Items, item)
			}
		}
		return &completionList, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

// parseCompletionItem parses a completion item from a map.
func (c *goplsClient) parseCompletionItem(itemMap map[string]any) CompletionItem {
	var item CompletionItem

	if label, ok := itemMap["label"].(string); ok {
		item.Label = label
	}

	if kind, ok := itemMap["kind"].(float64); ok {
		item.Kind = int(kind)
	}

	if detail, ok := itemMap["detail"].(string); ok {
		item.Detail = detail
	}

	if documentation, ok := itemMap["documentation"].(string); ok {
		item.Documentation = documentation
	}

	if insertText, ok := itemMap["insertText"].(string); ok {
		item.InsertText = insertText
	}

	if insertTextFormat, ok := itemMap["insertTextFormat"].(float64); ok {
		item.InsertTextFormat = int(insertTextFormat)
	}

	if sortText, ok := itemMap["sortText"].(string); ok {
		item.SortText = sortText
	}

	if filterText, ok := itemMap["filterText"].(string); ok {
		item.FilterText = filterText
	}

	return item
}
