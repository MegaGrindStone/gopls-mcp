package main

import (
	"context"
	"fmt"
	"log"
)

// Position represents a position in a document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Range represents a range in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
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

// GoToDefinition sends a textDocument/definition request to gopls.
func (m *Manager) GoToDefinition(_ context.Context, uri string, line, character int) ([]Location, error) {
	log.Printf("GoToDefinition called: uri=%s, line=%d, char=%d", uri, line, character)

	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := m.ensureFileOpen(uri); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      m.nextRequestID(),
		"method":  "textDocument/definition",
		"params": TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position: Position{
				Line:      line,
				Character: character,
			},
		},
	}

	response, err := m.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	// Extract locations from response
	return parseLocationsFromResponse(response)
}

// FindReferences sends a textDocument/references request to gopls.
func (m *Manager) FindReferences(
	_ context.Context,
	uri string,
	line, character int,
	includeDeclaration bool,
) ([]Location, error) {
	log.Printf("FindReferences called: uri=%s, line=%d, char=%d, includeDecl=%v", uri, line, character, includeDeclaration)

	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := m.ensureFileOpen(uri); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      m.nextRequestID(),
		"method":  "textDocument/references",
		"params": ReferenceParams{
			TextDocumentPositionParams: TextDocumentPositionParams{
				TextDocument: TextDocumentIdentifier{URI: uri},
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

	response, err := m.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	// Extract locations from response
	return parseLocationsFromResponse(response)
}

// GetHover sends a textDocument/hover request to gopls.
func (m *Manager) GetHover(_ context.Context, uri string, line, character int) (*Hover, error) {
	log.Printf("GetHover called: uri=%s, line=%d, char=%d", uri, line, character)

	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Ensure file is open in gopls
	if err := m.ensureFileOpen(uri); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      m.nextRequestID(),
		"method":  "textDocument/hover",
		"params": TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position: Position{
				Line:      line,
				Character: character,
			},
		},
	}

	response, err := m.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get hover info: %w", err)
	}

	// Extract hover information from response
	return parseHoverFromResponse(response)
}

// parseRange parses a range from a map.
func parseRange(rangeMap map[string]any) Range {
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

// parseLocationFromMap parses a single location from a map.
func parseLocationFromMap(locMap map[string]any) Location {
	var location Location
	if locURI, uriOk := locMap["uri"].(string); uriOk {
		location.URI = locURI
	}
	if rangeMap, rangeMapOk := locMap["range"].(map[string]any); rangeMapOk {
		location.Range = parseRange(rangeMap)
	}
	return location
}

// parseLocationsFromResponse extracts locations from LSP response.
func parseLocationsFromResponse(response map[string]any) ([]Location, error) {
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
			location := parseLocationFromMap(locMap)
			locs = append(locs, location)
		}
	}
	return locs, nil
}

// parseHoverContents parses hover contents from any type.
func parseHoverContents(contents any) []string {
	var result []string
	if contentList, contentListOk := contents.([]any); contentListOk {
		for _, content := range contentList {
			if contentStr, contentStrOk := content.(string); contentStrOk {
				result = append(result, contentStr)
			}
		}
	}
	return result
}

// parseHoverFromResponse extracts hover information from LSP response.
func parseHoverFromResponse(response map[string]any) (*Hover, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid hover response format")
	}

	hoverMap, hoverMapOk := result.(map[string]any)
	if !hoverMapOk {
		return nil, fmt.Errorf("invalid hover response format")
	}

	var hover Hover
	if contents, contentsOk := hoverMap["contents"]; contentsOk {
		hover.Contents = parseHoverContents(contents)
	}
	if rangeMap, rangeMapOk := hoverMap["range"].(map[string]any); rangeMapOk {
		rng := parseRange(rangeMap)
		hover.Range = &rng
	}
	return &hover, nil
}
