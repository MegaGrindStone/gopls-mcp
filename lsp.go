package main

import (
	"context"
	"fmt"
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
	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
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

	if err := m.sendRequest(request); err != nil {
		return nil, fmt.Errorf("failed to send definition request: %w", err)
	}

	response, err := m.readResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read definition response: %w", err)
	}

	// Extract locations from response
	if result, ok := response["result"]; ok {
		if locations, ok := result.([]any); ok {
			var locs []Location
			for _, loc := range locations {
				if locMap, ok := loc.(map[string]any); ok {
					var location Location
					if uri, ok := locMap["uri"].(string); ok {
						location.URI = uri
					}
					if rangeMap, ok := locMap["range"].(map[string]any); ok {
						location.Range = parseRange(rangeMap)
					}
					locs = append(locs, location)
				}
			}
			return locs, nil
		}
	}

	return nil, fmt.Errorf("invalid definition response format")
}

// FindReferences sends a textDocument/references request to gopls.
func (m *Manager) FindReferences(_ context.Context, uri string, line, character int, includeDeclaration bool) ([]Location, error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
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

	if err := m.sendRequest(request); err != nil {
		return nil, fmt.Errorf("failed to send references request: %w", err)
	}

	response, err := m.readResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read references response: %w", err)
	}

	// Extract locations from response
	if result, ok := response["result"]; ok {
		if locations, ok := result.([]any); ok {
			var locs []Location
			for _, loc := range locations {
				if locMap, ok := loc.(map[string]any); ok {
					var location Location
					if uri, ok := locMap["uri"].(string); ok {
						location.URI = uri
					}
					if rangeMap, ok := locMap["range"].(map[string]any); ok {
						location.Range = parseRange(rangeMap)
					}
					locs = append(locs, location)
				}
			}
			return locs, nil
		}
	}

	return nil, fmt.Errorf("invalid references response format")
}

// GetHover sends a textDocument/hover request to gopls.
func (m *Manager) GetHover(_ context.Context, uri string, line, character int) (*Hover, error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
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

	if err := m.sendRequest(request); err != nil {
		return nil, fmt.Errorf("failed to send hover request: %w", err)
	}

	response, err := m.readResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read hover response: %w", err)
	}

	// Extract hover information from response
	if result, ok := response["result"]; ok {
		if hoverMap, ok := result.(map[string]any); ok {
			var hover Hover
			if contents, ok := hoverMap["contents"]; ok {
				if contentList, ok := contents.([]any); ok {
					for _, content := range contentList {
						if contentStr, ok := content.(string); ok {
							hover.Contents = append(hover.Contents, contentStr)
						}
					}
				}
			}
			if rangeMap, ok := hoverMap["range"].(map[string]any); ok {
				rng := parseRange(rangeMap)
				hover.Range = &rng
			}
			return &hover, nil
		}
	}

	return nil, fmt.Errorf("invalid hover response format")
}

// parseRange parses a range from a map.
func parseRange(rangeMap map[string]any) Range {
	var rng Range
	if startMap, ok := rangeMap["start"].(map[string]any); ok {
		if line, ok := startMap["line"].(float64); ok {
			rng.Start.Line = int(line)
		}
		if char, ok := startMap["character"].(float64); ok {
			rng.Start.Character = int(char)
		}
	}
	if endMap, ok := rangeMap["end"].(map[string]any); ok {
		if line, ok := endMap["line"].(float64); ok {
			rng.End.Line = int(line)
		}
		if char, ok := endMap["character"].(float64); ok {
			rng.End.Character = int(char)
		}
	}
	return rng
}
