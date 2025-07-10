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

	// Store diagnostics and timestamp
	c.diagnosticsMux.Lock()
	c.diagnostics[relativePath] = diagnostics
	c.diagnosticsTimestamps[relativePath] = time.Now()
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

	// Wait for diagnostics to stabilize rather than using fixed timeout
	const stabilityWindow = 200 * time.Millisecond
	const maxWait = 5 * time.Second
	const pollInterval = 50 * time.Millisecond
	const minWaitForNonEmpty = 3 * time.Second // Don't accept empty diagnostics as stable for first 3 seconds

	startTime := time.Now()
	var lastTimestamp time.Time
	var diagnostics []Diagnostic

	for {
		// Check if we've exceeded maximum wait time
		if time.Since(startTime) > maxWait {
			c.logger.Debug("diagnostics wait timeout reached", "relativePath", relativePath, "duration", time.Since(startTime))
			break
		}

		// Get current diagnostics and timestamp
		c.diagnosticsMux.RLock()
		currentDiagnostics, exists := c.diagnostics[relativePath]
		currentTimestamp, hasTimestamp := c.diagnosticsTimestamps[relativePath]
		c.diagnosticsMux.RUnlock()

		if !exists || !hasTimestamp {
			// Wait for first diagnostics to arrive
			time.Sleep(pollInterval)
			continue
		}

		// New timestamp detected, update and continue waiting
		if currentTimestamp.After(lastTimestamp) {
			lastTimestamp = currentTimestamp
			c.logger.Debug("diagnostics updated", "relativePath", relativePath, "count", len(currentDiagnostics))
			time.Sleep(pollInterval)
			continue
		}

		// Check if diagnostics have been stable for the stability window
		if time.Since(currentTimestamp) < stabilityWindow {
			time.Sleep(pollInterval)
			continue
		}

		// Don't accept empty diagnostics as stable within first 3 seconds
		// This gives gopls time to complete its analysis
		if len(currentDiagnostics) == 0 && time.Since(startTime) < minWaitForNonEmpty {
			c.logger.Debug("ignoring empty diagnostics too early", "relativePath", relativePath,
				"duration", time.Since(startTime), "minWait", minWaitForNonEmpty)
			time.Sleep(pollInterval)
			continue
		}

		// Diagnostics are stable and ready
		diagnostics = currentDiagnostics
		c.logger.Debug("diagnostics stabilized", "relativePath", relativePath,
			"duration", time.Since(startTime), "count", len(diagnostics))
		break
	}

	// Return the diagnostics we found, or empty slice if none
	if diagnostics == nil {
		return []Diagnostic{}, nil
	}

	return diagnostics, nil
}
