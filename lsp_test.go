package main

import (
	"reflect"
	"testing"
)

func TestParseRange(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected Range
	}{
		{
			name: "valid range",
			input: map[string]any{
				"start": map[string]any{
					"line":      float64(10),
					"character": float64(5),
				},
				"end": map[string]any{
					"line":      float64(10),
					"character": float64(15),
				},
			},
			expected: Range{
				Start: Position{Line: 10, Character: 5},
				End:   Position{Line: 10, Character: 15},
			},
		},
		{
			name:  "empty range",
			input: map[string]any{},
			expected: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 0},
			},
		},
		{
			name: "missing start",
			input: map[string]any{
				"end": map[string]any{
					"line":      float64(10),
					"character": float64(15),
				},
			},
			expected: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 10, Character: 15},
			},
		},
		{
			name: "missing end",
			input: map[string]any{
				"start": map[string]any{
					"line":      float64(10),
					"character": float64(5),
				},
			},
			expected: Range{
				Start: Position{Line: 10, Character: 5},
				End:   Position{Line: 0, Character: 0},
			},
		},
		{
			name: "invalid types",
			input: map[string]any{
				"start": map[string]any{
					"line":      "not a number",
					"character": float64(5),
				},
				"end": map[string]any{
					"line":      float64(10),
					"character": "not a number",
				},
			},
			expected: Range{
				Start: Position{Line: 0, Character: 5},
				End:   Position{Line: 10, Character: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRange(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseRange() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseLocationFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected Location
	}{
		{
			name: "valid location",
			input: map[string]any{
				"uri": "file:///test.go",
				"range": map[string]any{
					"start": map[string]any{
						"line":      float64(10),
						"character": float64(5),
					},
					"end": map[string]any{
						"line":      float64(10),
						"character": float64(15),
					},
				},
			},
			expected: Location{
				URI: "file:///test.go",
				Range: Range{
					Start: Position{Line: 10, Character: 5},
					End:   Position{Line: 10, Character: 15},
				},
			},
		},
		{
			name:  "empty location",
			input: map[string]any{},
			expected: Location{
				URI: "",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			},
		},
		{
			name: "missing uri",
			input: map[string]any{
				"range": map[string]any{
					"start": map[string]any{
						"line":      float64(10),
						"character": float64(5),
					},
					"end": map[string]any{
						"line":      float64(10),
						"character": float64(15),
					},
				},
			},
			expected: Location{
				URI: "",
				Range: Range{
					Start: Position{Line: 10, Character: 5},
					End:   Position{Line: 10, Character: 15},
				},
			},
		},
		{
			name: "missing range",
			input: map[string]any{
				"uri": "file:///test.go",
			},
			expected: Location{
				URI: "file:///test.go",
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 0},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLocationFromMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseLocationFromMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseLocationsFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []Location
		wantErr  bool
	}{
		{
			name: "valid locations response",
			input: map[string]any{
				"result": []any{
					map[string]any{
						"uri": "file:///test.go",
						"range": map[string]any{
							"start": map[string]any{
								"line":      float64(10),
								"character": float64(5),
							},
							"end": map[string]any{
								"line":      float64(10),
								"character": float64(15),
							},
						},
					},
					map[string]any{
						"uri": "file:///test2.go",
						"range": map[string]any{
							"start": map[string]any{
								"line":      float64(20),
								"character": float64(10),
							},
							"end": map[string]any{
								"line":      float64(20),
								"character": float64(20),
							},
						},
					},
				},
			},
			expected: []Location{
				{
					URI: "file:///test.go",
					Range: Range{
						Start: Position{Line: 10, Character: 5},
						End:   Position{Line: 10, Character: 15},
					},
				},
				{
					URI: "file:///test2.go",
					Range: Range{
						Start: Position{Line: 20, Character: 10},
						End:   Position{Line: 20, Character: 20},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty locations response",
			input: map[string]any{
				"result": []any{},
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "missing result",
			input:    map[string]any{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid result type",
			input: map[string]any{
				"result": "not an array",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid location in array",
			input: map[string]any{
				"result": []any{
					map[string]any{
						"uri": "file:///test.go",
						"range": map[string]any{
							"start": map[string]any{
								"line":      float64(10),
								"character": float64(5),
							},
							"end": map[string]any{
								"line":      float64(10),
								"character": float64(15),
							},
						},
					},
					"not a map",
				},
			},
			expected: []Location{
				{
					URI: "file:///test.go",
					Range: Range{
						Start: Position{Line: 10, Character: 5},
						End:   Position{Line: 10, Character: 15},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLocationsFromResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLocationsFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseLocationsFromResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseHoverContents(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
	}{
		{
			name:     "valid string array",
			input:    []any{"content1", "content2", "content3"},
			expected: []string{"content1", "content2", "content3"},
		},
		{
			name:     "empty array",
			input:    []any{},
			expected: nil,
		},
		{
			name:     "mixed types in array",
			input:    []any{"content1", 123, "content2", true, "content3"},
			expected: []string{"content1", "content2", "content3"},
		},
		{
			name:     "not an array",
			input:    "single string",
			expected: nil,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHoverContents(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseHoverContents() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseHoverFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected *Hover
		wantErr  bool
	}{
		{
			name: "valid hover response with range",
			input: map[string]any{
				"result": map[string]any{
					"contents": []any{"func example()", "Example function documentation"},
					"range": map[string]any{
						"start": map[string]any{
							"line":      float64(10),
							"character": float64(5),
						},
						"end": map[string]any{
							"line":      float64(10),
							"character": float64(15),
						},
					},
				},
			},
			expected: &Hover{
				Contents: []string{"func example()", "Example function documentation"},
				Range: &Range{
					Start: Position{Line: 10, Character: 5},
					End:   Position{Line: 10, Character: 15},
				},
			},
			wantErr: false,
		},
		{
			name: "valid hover response without range",
			input: map[string]any{
				"result": map[string]any{
					"contents": []any{"func example()", "Example function documentation"},
				},
			},
			expected: &Hover{
				Contents: []string{"func example()", "Example function documentation"},
				Range:    nil,
			},
			wantErr: false,
		},
		{
			name: "empty hover response",
			input: map[string]any{
				"result": map[string]any{},
			},
			expected: &Hover{
				Contents: nil,
				Range:    nil,
			},
			wantErr: false,
		},
		{
			name:     "missing result",
			input:    map[string]any{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid result type",
			input: map[string]any{
				"result": "not a map",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseHoverFromResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHoverFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseHoverFromResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseDocumentSymbolFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected DocumentSymbol
	}{
		{
			name: "valid document symbol with children",
			input: map[string]any{
				"name":       "TestFunction",
				"detail":     "func TestFunction()",
				"kind":       float64(12), // Function
				"deprecated": false,
				"range": map[string]any{
					"start": map[string]any{
						"line":      float64(10),
						"character": float64(0),
					},
					"end": map[string]any{
						"line":      float64(15),
						"character": float64(1),
					},
				},
				"selectionRange": map[string]any{
					"start": map[string]any{
						"line":      float64(10),
						"character": float64(5),
					},
					"end": map[string]any{
						"line":      float64(10),
						"character": float64(17),
					},
				},
				"children": []any{
					map[string]any{
						"name": "LocalVar",
						"kind": float64(13), // Variable
						"range": map[string]any{
							"start": map[string]any{
								"line":      float64(11),
								"character": float64(4),
							},
							"end": map[string]any{
								"line":      float64(11),
								"character": float64(12),
							},
						},
						"selectionRange": map[string]any{
							"start": map[string]any{
								"line":      float64(11),
								"character": float64(4),
							},
							"end": map[string]any{
								"line":      float64(11),
								"character": float64(12),
							},
						},
					},
				},
			},
			expected: DocumentSymbol{
				Name:       "TestFunction",
				Detail:     "func TestFunction()",
				Kind:       SymbolKindFunction,
				Deprecated: false,
				Range: Range{
					Start: Position{Line: 10, Character: 0},
					End:   Position{Line: 15, Character: 1},
				},
				SelectionRange: Range{
					Start: Position{Line: 10, Character: 5},
					End:   Position{Line: 10, Character: 17},
				},
				Children: []DocumentSymbol{
					{
						Name:       "LocalVar",
						Kind:       SymbolKindVariable,
						Deprecated: false,
						Range: Range{
							Start: Position{Line: 11, Character: 4},
							End:   Position{Line: 11, Character: 12},
						},
						SelectionRange: Range{
							Start: Position{Line: 11, Character: 4},
							End:   Position{Line: 11, Character: 12},
						},
					},
				},
			},
		},
		{
			name:  "empty symbol",
			input: map[string]any{},
			expected: DocumentSymbol{
				Name:           "",
				Kind:           0,
				Deprecated:     false,
				Range:          Range{},
				SelectionRange: Range{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDocumentSymbolFromMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseDocumentSymbolFromMap() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseDocumentSymbolsFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []DocumentSymbol
		wantErr  bool
	}{
		{
			name: "valid document symbols response",
			input: map[string]any{
				"result": []any{
					map[string]any{
						"name": "TestFunction",
						"kind": float64(12), // Function
						"range": map[string]any{
							"start": map[string]any{
								"line":      float64(10),
								"character": float64(0),
							},
							"end": map[string]any{
								"line":      float64(15),
								"character": float64(1),
							},
						},
						"selectionRange": map[string]any{
							"start": map[string]any{
								"line":      float64(10),
								"character": float64(5),
							},
							"end": map[string]any{
								"line":      float64(10),
								"character": float64(17),
							},
						},
					},
				},
			},
			expected: []DocumentSymbol{
				{
					Name: "TestFunction",
					Kind: SymbolKindFunction,
					Range: Range{
						Start: Position{Line: 10, Character: 0},
						End:   Position{Line: 15, Character: 1},
					},
					SelectionRange: Range{
						Start: Position{Line: 10, Character: 5},
						End:   Position{Line: 10, Character: 17},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty response",
			input: map[string]any{
				"result": []any{},
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "missing result",
			input:    map[string]any{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid result type",
			input: map[string]any{
				"result": "not an array",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDocumentSymbolsFromResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDocumentSymbolsFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseDocumentSymbolsFromResponse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseSymbolInformationFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected SymbolInformation
	}{
		{
			name: "valid symbol information",
			input: map[string]any{
				"name":       "TestStruct",
				"kind":       float64(23), // Struct
				"deprecated": false,
				"location": map[string]any{
					"uri": "file:///test.go",
					"range": map[string]any{
						"start": map[string]any{
							"line":      float64(5),
							"character": float64(0),
						},
						"end": map[string]any{
							"line":      float64(10),
							"character": float64(1),
						},
					},
				},
				"containerName": "main",
			},
			expected: SymbolInformation{
				Name:       "TestStruct",
				Kind:       SymbolKindStruct,
				Deprecated: false,
				Location: Location{
					URI: "file:///test.go",
					Range: Range{
						Start: Position{Line: 5, Character: 0},
						End:   Position{Line: 10, Character: 1},
					},
				},
				ContainerName: "main",
			},
		},
		{
			name:  "empty symbol information",
			input: map[string]any{},
			expected: SymbolInformation{
				Name:          "",
				Kind:          0,
				Deprecated:    false,
				Location:      Location{},
				ContainerName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSymbolInformationFromMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseSymbolInformationFromMap() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseWorkspaceSymbolsFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []SymbolInformation
		wantErr  bool
	}{
		{
			name: "valid workspace symbols response",
			input: map[string]any{
				"result": []any{
					map[string]any{
						"name": "TestStruct",
						"kind": float64(23), // Struct
						"location": map[string]any{
							"uri": "file:///test.go",
							"range": map[string]any{
								"start": map[string]any{
									"line":      float64(5),
									"character": float64(0),
								},
								"end": map[string]any{
									"line":      float64(10),
									"character": float64(1),
								},
							},
						},
						"containerName": "main",
					},
					map[string]any{
						"name": "TestFunction",
						"kind": float64(12), // Function
						"location": map[string]any{
							"uri": "file:///test2.go",
							"range": map[string]any{
								"start": map[string]any{
									"line":      float64(15),
									"character": float64(0),
								},
								"end": map[string]any{
									"line":      float64(20),
									"character": float64(1),
								},
							},
						},
					},
				},
			},
			expected: []SymbolInformation{
				{
					Name: "TestStruct",
					Kind: SymbolKindStruct,
					Location: Location{
						URI: "file:///test.go",
						Range: Range{
							Start: Position{Line: 5, Character: 0},
							End:   Position{Line: 10, Character: 1},
						},
					},
					ContainerName: "main",
				},
				{
					Name: "TestFunction",
					Kind: SymbolKindFunction,
					Location: Location{
						URI: "file:///test2.go",
						Range: Range{
							Start: Position{Line: 15, Character: 0},
							End:   Position{Line: 20, Character: 1},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty response",
			input: map[string]any{
				"result": []any{},
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "missing result",
			input:    map[string]any{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid result type",
			input: map[string]any{
				"result": "not an array",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseWorkspaceSymbolsFromResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseWorkspaceSymbolsFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseWorkspaceSymbolsFromResponse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}
