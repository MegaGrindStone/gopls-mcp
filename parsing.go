package main

import (
	"fmt"
)

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
