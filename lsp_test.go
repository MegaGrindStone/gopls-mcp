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

func TestParseDiagnosticFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected Diagnostic
	}{
		{
			name: "valid diagnostic",
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
				"severity": float64(1),
				"code":     "unused",
				"source":   "gopls",
				"message":  "variable declared but not used",
				"tags":     []any{float64(1)},
			},
			expected: Diagnostic{
				Range: Range{
					Start: Position{Line: 10, Character: 5},
					End:   Position{Line: 10, Character: 15},
				},
				Severity: DiagnosticSeverityError,
				Code:     "unused",
				Source:   "gopls",
				Message:  "variable declared but not used",
				Tags:     []DiagnosticTag{DiagnosticTagUnnecessary},
			},
		},
		{
			name:     "empty diagnostic",
			input:    map[string]any{},
			expected: Diagnostic{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDiagnosticFromMap(tt.input)
			if result.Severity != tt.expected.Severity {
				t.Errorf("parseDiagnosticFromMap() severity = %v, want %v", result.Severity, tt.expected.Severity)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("parseDiagnosticFromMap() message = %v, want %v", result.Message, tt.expected.Message)
			}
			if len(result.Tags) != len(tt.expected.Tags) {
				t.Errorf("parseDiagnosticFromMap() tags count = %v, want %v", len(result.Tags), len(tt.expected.Tags))
			}
		})
	}
}

func TestParseDiagnosticsFromResponse(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]any
		expected      []Diagnostic
		expectedError bool
	}{
		{
			name: "valid diagnostics response (direct array)",
			input: map[string]any{
				"result": []any{
					map[string]any{
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
						"severity": float64(1),
						"message":  "variable declared but not used",
					},
				},
			},
			expected: []Diagnostic{
				{
					Range: Range{
						Start: Position{Line: 10, Character: 5},
						End:   Position{Line: 10, Character: 15},
					},
					Severity: DiagnosticSeverityError,
					Message:  "variable declared but not used",
				},
			},
		},
		{
			name: "valid diagnostics response (items object)",
			input: map[string]any{
				"result": map[string]any{
					"items": []any{
						map[string]any{
							"range": map[string]any{
								"start": map[string]any{
									"line":      float64(5),
									"character": float64(0),
								},
								"end": map[string]any{
									"line":      float64(5),
									"character": float64(10),
								},
							},
							"severity": float64(2),
							"message":  "unused import",
						},
					},
				},
			},
			expected: []Diagnostic{
				{
					Range: Range{
						Start: Position{Line: 5, Character: 0},
						End:   Position{Line: 5, Character: 10},
					},
					Severity: DiagnosticSeverityWarning,
					Message:  "unused import",
				},
			},
		},
		{
			name: "empty diagnostics response",
			input: map[string]any{
				"result": []any{},
			},
			expected: []Diagnostic{},
		},
		{
			name: "missing result",
			input: map[string]any{
				"error": "some error",
			},
			expectedError: true,
		},
		{
			name: "invalid result type",
			input: map[string]any{
				"result": "invalid",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDiagnosticsFromResponse(tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("parseDiagnosticsFromResponse() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("parseDiagnosticsFromResponse() error = %v", err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("parseDiagnosticsFromResponse() count = %v, want %v", len(result), len(tt.expected))
			}
			for i, diag := range result {
				if i < len(tt.expected) && diag.Message != tt.expected[i].Message {
					t.Errorf("parseDiagnosticsFromResponse() diagnostic[%d].message = %v, want %v",
						i, diag.Message, tt.expected[i].Message)
				}
			}
		})
	}
}

func TestParseCompletionItemFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected CompletionItem
	}{
		{
			name: "valid completion item",
			input: map[string]any{
				"label":            "TestFunction",
				"kind":             float64(3),
				"detail":           "func TestFunction()",
				"documentation":    "Test function documentation",
				"deprecated":       false,
				"preselect":        true,
				"sortText":         "TestFunction",
				"filterText":       "TestFunction",
				"insertText":       "TestFunction()",
				"insertTextFormat": float64(1),
				"tags":             []any{},
				"commitCharacters": []any{"(", ")"},
			},
			expected: CompletionItem{
				Label:            "TestFunction",
				Kind:             CompletionItemKindFunction,
				Detail:           "func TestFunction()",
				Documentation:    "Test function documentation",
				Deprecated:       false,
				Preselect:        true,
				SortText:         "TestFunction",
				FilterText:       "TestFunction",
				InsertText:       "TestFunction()",
				InsertTextFormat: 1,
				Tags:             []int{},
				CommitCharacters: []string{"(", ")"},
			},
		},
		{
			name:     "empty completion item",
			input:    map[string]any{},
			expected: CompletionItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCompletionItemFromMap(tt.input)
			if result.Label != tt.expected.Label {
				t.Errorf("parseCompletionItemFromMap() label = %v, want %v", result.Label, tt.expected.Label)
			}
			if result.Kind != tt.expected.Kind {
				t.Errorf("parseCompletionItemFromMap() kind = %v, want %v", result.Kind, tt.expected.Kind)
			}
			if result.Preselect != tt.expected.Preselect {
				t.Errorf("parseCompletionItemFromMap() preselect = %v, want %v", result.Preselect, tt.expected.Preselect)
			}
			if len(result.CommitCharacters) != len(tt.expected.CommitCharacters) {
				t.Errorf("parseCompletionItemFromMap() commitCharacters count = %v, want %v",
					len(result.CommitCharacters), len(tt.expected.CommitCharacters))
			}
		})
	}
}

func TestParseCompletionsFromResponse(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]any
		expected      []CompletionItem
		expectedError bool
	}{
		{
			name: "valid completions response (direct array)",
			input: map[string]any{
				"result": []any{
					map[string]any{
						"label": "TestFunction",
						"kind":  float64(3),
					},
				},
			},
			expected: []CompletionItem{
				{
					Label: "TestFunction",
					Kind:  CompletionItemKindFunction,
				},
			},
		},
		{
			name: "valid completions response (items object)",
			input: map[string]any{
				"result": map[string]any{
					"items": []any{
						map[string]any{
							"label": "TestVariable",
							"kind":  float64(6),
						},
					},
				},
			},
			expected: []CompletionItem{
				{
					Label: "TestVariable",
					Kind:  CompletionItemKindVariable,
				},
			},
		},
		{
			name: "empty completions response",
			input: map[string]any{
				"result": []any{},
			},
			expected: []CompletionItem{},
		},
		{
			name: "missing result",
			input: map[string]any{
				"error": "some error",
			},
			expectedError: true,
		},
		{
			name: "invalid result type",
			input: map[string]any{
				"result": "invalid",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCompletionsFromResponse(tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("parseCompletionsFromResponse() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("parseCompletionsFromResponse() error = %v", err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("parseCompletionsFromResponse() count = %v, want %v", len(result), len(tt.expected))
			}
			for i, item := range result {
				if i < len(tt.expected) && item.Label != tt.expected[i].Label {
					t.Errorf("parseCompletionsFromResponse() item[%d].label = %v, want %v", i, item.Label, tt.expected[i].Label)
				}
			}
		})
	}
}

func TestParseCallHierarchyItemFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected CallHierarchyItem
	}{
		{
			name: "valid call hierarchy item",
			input: map[string]any{
				"name":   "TestFunction",
				"kind":   float64(12), // Function
				"detail": "func TestFunction()",
				"uri":    "file:///test.go",
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
			expected: CallHierarchyItem{
				Name:   "TestFunction",
				Kind:   SymbolKindFunction,
				Detail: "func TestFunction()",
				URI:    "file:///test.go",
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
		{
			name:  "empty call hierarchy item",
			input: map[string]any{},
			expected: CallHierarchyItem{
				Name:           "",
				Kind:           0,
				Detail:         "",
				URI:            "",
				Range:          Range{},
				SelectionRange: Range{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCallHierarchyItemFromMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseCallHierarchyItemFromMap() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseCallHierarchyItemsFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []CallHierarchyItem
		wantErr  bool
	}{
		{
			name: "valid call hierarchy items response",
			input: map[string]any{
				"result": []any{
					map[string]any{
						"name": "TestFunction",
						"kind": float64(12), // Function
						"uri":  "file:///test.go",
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
			expected: []CallHierarchyItem{
				{
					Name: "TestFunction",
					Kind: SymbolKindFunction,
					URI:  "file:///test.go",
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
			result, err := parseCallHierarchyItemsFromResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCallHierarchyItemsFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseCallHierarchyItemsFromResponse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseIncomingCallFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected CallHierarchyIncomingCall
	}{
		{
			name: "valid incoming call",
			input: map[string]any{
				"from": map[string]any{
					"name": "CallerFunction",
					"kind": float64(12), // Function
					"uri":  "file:///caller.go",
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
					"selectionRange": map[string]any{
						"start": map[string]any{
							"line":      float64(5),
							"character": float64(5),
						},
						"end": map[string]any{
							"line":      float64(5),
							"character": float64(19),
						},
					},
				},
				"fromRanges": []any{
					map[string]any{
						"start": map[string]any{
							"line":      float64(7),
							"character": float64(4),
						},
						"end": map[string]any{
							"line":      float64(7),
							"character": float64(16),
						},
					},
				},
			},
			expected: CallHierarchyIncomingCall{
				From: CallHierarchyItem{
					Name: "CallerFunction",
					Kind: SymbolKindFunction,
					URI:  "file:///caller.go",
					Range: Range{
						Start: Position{Line: 5, Character: 0},
						End:   Position{Line: 10, Character: 1},
					},
					SelectionRange: Range{
						Start: Position{Line: 5, Character: 5},
						End:   Position{Line: 5, Character: 19},
					},
				},
				FromRanges: []Range{
					{
						Start: Position{Line: 7, Character: 4},
						End:   Position{Line: 7, Character: 16},
					},
				},
			},
		},
		{
			name:  "empty incoming call",
			input: map[string]any{},
			expected: CallHierarchyIncomingCall{
				From:       CallHierarchyItem{},
				FromRanges: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIncomingCallFromMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseIncomingCallFromMap() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseSignatureHelpFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected *SignatureHelp
		wantErr  bool
	}{
		{
			name: "valid signature help response",
			input: map[string]any{
				"result": map[string]any{
					"signatures": []any{
						map[string]any{
							"label":         "TestFunction(param1 string, param2 int) error",
							"documentation": "TestFunction performs a test operation",
							"parameters": []any{
								map[string]any{
									"label":         "param1 string",
									"documentation": "First parameter",
								},
								map[string]any{
									"label":         "param2 int",
									"documentation": "Second parameter",
								},
							},
						},
					},
					"activeSignature": float64(0),
					"activeParameter": float64(1),
				},
			},
			expected: &SignatureHelp{
				Signatures: []SignatureInformation{
					{
						Label:         "TestFunction(param1 string, param2 int) error",
						Documentation: "TestFunction performs a test operation",
						Parameters: []ParameterInformation{
							{
								Label:         "param1 string",
								Documentation: "First parameter",
							},
							{
								Label:         "param2 int",
								Documentation: "Second parameter",
							},
						},
					},
				},
				ActiveSignature: 0,
				ActiveParameter: 1,
			},
			wantErr: false,
		},
		{
			name: "empty signature help response",
			input: map[string]any{
				"result": map[string]any{},
			},
			expected: &SignatureHelp{
				Signatures:      nil,
				ActiveSignature: 0,
				ActiveParameter: 0,
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
			result, err := parseSignatureHelpFromResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSignatureHelpFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseSignatureHelpFromResponse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseTypeHierarchyItemFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected TypeHierarchyItem
	}{
		{
			name: "valid type hierarchy item",
			input: map[string]any{
				"name":   "TestInterface",
				"kind":   float64(11), // Interface
				"detail": "interface TestInterface",
				"uri":    "file:///test.go",
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
				"selectionRange": map[string]any{
					"start": map[string]any{
						"line":      float64(5),
						"character": float64(10),
					},
					"end": map[string]any{
						"line":      float64(5),
						"character": float64(23),
					},
				},
			},
			expected: TypeHierarchyItem{
				Name:   "TestInterface",
				Kind:   SymbolKindInterface,
				Detail: "interface TestInterface",
				URI:    "file:///test.go",
				Range: Range{
					Start: Position{Line: 5, Character: 0},
					End:   Position{Line: 10, Character: 1},
				},
				SelectionRange: Range{
					Start: Position{Line: 5, Character: 10},
					End:   Position{Line: 5, Character: 23},
				},
			},
		},
		{
			name:  "empty type hierarchy item",
			input: map[string]any{},
			expected: TypeHierarchyItem{
				Name:           "",
				Kind:           0,
				Detail:         "",
				URI:            "",
				Range:          Range{},
				SelectionRange: Range{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTypeHierarchyItemFromMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseTypeHierarchyItemFromMap() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}
