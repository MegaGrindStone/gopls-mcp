package main

import (
	"fmt"
	"time"
)

// handlePublishDiagnostics handles publishDiagnostics notifications from gopls.
func (c *goplsClient) handlePublishDiagnostics(message map[string]any) {
	params, ok := message["params"].(map[string]any)
	if !ok {
		c.logger.Debug("invalid publishDiagnostics params")
		return
	}

	uri, ok := params["uri"].(string)
	if !ok {
		c.logger.Debug("invalid publishDiagnostics uri")
		return
	}

	diagnosticsData, ok := params["diagnostics"].([]any)
	if !ok {
		c.logger.Debug("invalid publishDiagnostics diagnostics")
		return
	}

	// Convert URI to relative path
	relativePath := c.uriToRelativePath(uri)

	// Parse diagnostics
	var diagnostics []Diagnostic
	for _, diagData := range diagnosticsData {
		if diagMap, diagOk := diagData.(map[string]any); diagOk {
			diagnostic := c.parseDiagnostic(diagMap)
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	// Store diagnostics
	c.diagnosticsMux.Lock()
	c.diagnostics[relativePath] = diagnostics
	c.diagnosticsMux.Unlock()

	c.logger.Debug("stored diagnostics", "relativePath", relativePath, "count", len(diagnostics))
}

// parseDiagnostic parses a diagnostic from a map.
func (c *goplsClient) parseDiagnostic(diagMap map[string]any) Diagnostic {
	var diagnostic Diagnostic

	if rangeMap, ok := diagMap["range"].(map[string]any); ok {
		diagnostic.Range = c.parseRange(rangeMap)
	}

	if severity, ok := diagMap["severity"].(float64); ok {
		diagnostic.Severity = DiagnosticSeverity(int(severity))
	}

	if code, ok := diagMap["code"].(string); ok {
		diagnostic.Code = code
	}

	if source, ok := diagMap["source"].(string); ok {
		diagnostic.Source = source
	}

	if message, ok := diagMap["message"].(string); ok {
		diagnostic.Message = message
	}

	//nolint:nestif // Complex JSON parsing requires nested type assertions
	if relatedInfoList, ok := diagMap["relatedInformation"].([]any); ok {
		for _, relatedData := range relatedInfoList {
			if relatedMap, relatedOk := relatedData.(map[string]any); relatedOk {
				var relatedInfo DiagnosticRelatedInformation
				if location, locOk := relatedMap["location"].(map[string]any); locOk {
					relatedInfo.Location = c.parseLocationFromMap(location)
				}
				if message, msgOk := relatedMap["message"].(string); msgOk {
					relatedInfo.Message = message
				}
				diagnostic.RelatedInfo = append(diagnostic.RelatedInfo, relatedInfo)
			}
		}
	}

	return diagnostic
}

// getDiagnostics returns diagnostics for a specific file.
func (c *goplsClient) getDiagnostics(relativePath string) ([]Diagnostic, error) {
	c.logger.Debug("getDiagnostics called", "relativePath", relativePath)

	if !c.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls to trigger diagnostics
	if err := c.ensureFileOpen(relativePath); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Give gopls a moment to process the file and send diagnostics
	time.Sleep(100 * time.Millisecond)

	// Get diagnostics from stored cache
	c.diagnosticsMux.RLock()
	diagnostics, exists := c.diagnostics[relativePath]
	c.diagnosticsMux.RUnlock()

	if !exists {
		return []Diagnostic{}, nil
	}

	return diagnostics, nil
}
