package main

import (
	"context"
	"fmt"
	"time"
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

// SymbolKind represents the kind of a symbol.
type SymbolKind int

// SymbolKind constants represent the kinds of symbols as defined in the LSP specification.
const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// DocumentSymbol represents a symbol within a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Tags           []int            `json:"tags,omitempty"`
	Deprecated     bool             `json:"deprecated,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation represents information about a symbol.
type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Tags          []int      `json:"tags,omitempty"`
	Deprecated    bool       `json:"deprecated,omitempty"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

// WorkspaceSymbolParams represents parameters for workspace symbol requests.
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

// DiagnosticSeverity constants as defined in the LSP specification.
const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// DiagnosticTag represents a tag for a diagnostic.
type DiagnosticTag int

// DiagnosticTag constants as defined in the LSP specification.
const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// Diagnostic represents a diagnostic, such as a compile error or warning.
type Diagnostic struct {
	Range              Range              `json:"range"`
	Severity           DiagnosticSeverity `json:"severity"`
	Code               string             `json:"code,omitempty"`
	Source             string             `json:"source,omitempty"`
	Message            string             `json:"message"`
	Tags               []DiagnosticTag    `json:"tags,omitempty"`
	RelatedInformation []any              `json:"relatedInformation,omitempty"`
}

// CompletionItemKind represents the kind of a completion item.
type CompletionItemKind int

// CompletionItemKind constants as defined in the LSP specification.
const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// CompletionItem represents a completion item.
type CompletionItem struct {
	Label               string             `json:"label"`
	Kind                CompletionItemKind `json:"kind"`
	Tags                []int              `json:"tags,omitempty"`
	Detail              string             `json:"detail,omitempty"`
	Documentation       string             `json:"documentation,omitempty"`
	Deprecated          bool               `json:"deprecated,omitempty"`
	Preselect           bool               `json:"preselect,omitempty"`
	SortText            string             `json:"sortText,omitempty"`
	FilterText          string             `json:"filterText,omitempty"`
	InsertText          string             `json:"insertText,omitempty"`
	InsertTextFormat    int                `json:"insertTextFormat,omitempty"`
	TextEdit            any                `json:"textEdit,omitempty"`
	AdditionalTextEdits []any              `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string           `json:"commitCharacters,omitempty"`
	Command             any                `json:"command,omitempty"`
	Data                any                `json:"data,omitempty"`
}

// GoToDefinition sends a textDocument/definition request to gopls.
func (m *Manager) GoToDefinition(_ context.Context, uri string, line, character int) ([]Location, error) {
	m.logger.Debug("GoToDefinition called", "uri", uri, "line", line, "character", character)

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
	m.logger.Debug("FindReferences called",
		"uri", uri,
		"line", line,
		"character", character,
		"includeDeclaration", includeDeclaration)

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
	m.logger.Debug("GetHover called", "uri", uri, "line", line, "character", character)

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

// GetDocumentSymbols sends a textDocument/documentSymbol request to gopls.
func (m *Manager) GetDocumentSymbols(_ context.Context, uri string) ([]DocumentSymbol, error) {
	m.logger.Debug("GetDocumentSymbols called", "uri", uri)

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
		"method":  "textDocument/documentSymbol",
		"params": map[string]any{
			"textDocument": TextDocumentIdentifier{URI: uri},
		},
	}

	response, err := m.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get document symbols: %w", err)
	}

	// Extract document symbols from response
	return parseDocumentSymbolsFromResponse(response)
}

// SearchWorkspaceSymbols sends a workspace/symbol request to gopls.
func (m *Manager) SearchWorkspaceSymbols(_ context.Context, query string) ([]SymbolInformation, error) {
	m.logger.Debug("SearchWorkspaceSymbols called", "query", query)

	if !m.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	// Wait for workspace to be ready before making requests
	if err := m.waitForWorkspaceReady(30 * time.Second); err != nil {
		return nil, fmt.Errorf("workspace not ready: %w", err)
	}

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      m.nextRequestID(),
		"method":  "workspace/symbol",
		"params": WorkspaceSymbolParams{
			Query: query,
		},
	}

	response, err := m.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to search workspace symbols: %w", err)
	}

	// Extract workspace symbols from response
	return parseWorkspaceSymbolsFromResponse(response)
}

// GoToTypeDefinition sends a textDocument/typeDefinition request to gopls.
func (m *Manager) GoToTypeDefinition(_ context.Context, uri string, line, character int) ([]Location, error) {
	m.logger.Debug("GoToTypeDefinition called", "uri", uri, "line", line, "character", character)

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
		"method":  "textDocument/typeDefinition",
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
		return nil, fmt.Errorf("failed to get type definition: %w", err)
	}

	// Extract locations from response (reuse existing parser)
	return parseLocationsFromResponse(response)
}

// GetDiagnostics sends a textDocument/diagnostic request to gopls.
func (m *Manager) GetDiagnostics(_ context.Context, uri string) ([]Diagnostic, error) {
	m.logger.Debug("GetDiagnostics called", "uri", uri)

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
		"method":  "textDocument/diagnostic",
		"params": map[string]any{
			"textDocument": TextDocumentIdentifier{URI: uri},
		},
	}

	response, err := m.sendRequestAndWait(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get diagnostics: %w", err)
	}

	// Extract diagnostics from response
	return parseDiagnosticsFromResponse(response)
}

// FindImplementations sends a textDocument/implementation request to gopls.
func (m *Manager) FindImplementations(_ context.Context, uri string, line, character int) ([]Location, error) {
	m.logger.Debug("FindImplementations called", "uri", uri, "line", line, "character", character)

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
		"method":  "textDocument/implementation",
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
		return nil, fmt.Errorf("failed to find implementations: %w", err)
	}

	// Extract locations from response (reuse existing parser)
	return parseLocationsFromResponse(response)
}

// GetCompletions sends a textDocument/completion request to gopls.
func (m *Manager) GetCompletions(_ context.Context, uri string, line, character int) ([]CompletionItem, error) {
	m.logger.Debug("GetCompletions called", "uri", uri, "line", line, "character", character)

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
		"method":  "textDocument/completion",
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
		return nil, fmt.Errorf("failed to get completions: %w", err)
	}

	// Extract completions from response
	return parseCompletionsFromResponse(response)
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

// parseDocumentSymbolFromMap parses a single document symbol from a map.
func parseDocumentSymbolFromMap(symbolMap map[string]any) DocumentSymbol {
	var symbol DocumentSymbol

	if name, nameOk := symbolMap["name"].(string); nameOk {
		symbol.Name = name
	}
	if detail, detailOk := symbolMap["detail"].(string); detailOk {
		symbol.Detail = detail
	}
	if kind, kindOk := symbolMap["kind"].(float64); kindOk {
		symbol.Kind = SymbolKind(int(kind))
	}
	if deprecated, deprecatedOk := symbolMap["deprecated"].(bool); deprecatedOk {
		symbol.Deprecated = deprecated
	}
	if rangeMap, rangeMapOk := symbolMap["range"].(map[string]any); rangeMapOk {
		symbol.Range = parseRange(rangeMap)
	}
	if selectionRangeMap, selectionRangeMapOk := symbolMap["selectionRange"].(map[string]any); selectionRangeMapOk {
		symbol.SelectionRange = parseRange(selectionRangeMap)
	}
	if children, childrenOk := symbolMap["children"].([]any); childrenOk {
		for _, child := range children {
			if childMap, childMapOk := child.(map[string]any); childMapOk {
				symbol.Children = append(symbol.Children, parseDocumentSymbolFromMap(childMap))
			}
		}
	}

	return symbol
}

// parseDocumentSymbolsFromResponse extracts document symbols from LSP response.
func parseDocumentSymbolsFromResponse(response map[string]any) ([]DocumentSymbol, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	symbols, symbolsOk := result.([]any)
	if !symbolsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	var documentSymbols []DocumentSymbol
	for _, symbol := range symbols {
		if symbolMap, symbolMapOk := symbol.(map[string]any); symbolMapOk {
			documentSymbol := parseDocumentSymbolFromMap(symbolMap)
			documentSymbols = append(documentSymbols, documentSymbol)
		}
	}
	return documentSymbols, nil
}

// parseSymbolInformationFromMap parses a single symbol information from a map.
func parseSymbolInformationFromMap(symbolMap map[string]any) SymbolInformation {
	var symbol SymbolInformation

	if name, nameOk := symbolMap["name"].(string); nameOk {
		symbol.Name = name
	}
	if kind, kindOk := symbolMap["kind"].(float64); kindOk {
		symbol.Kind = SymbolKind(int(kind))
	}
	if deprecated, deprecatedOk := symbolMap["deprecated"].(bool); deprecatedOk {
		symbol.Deprecated = deprecated
	}
	if location, locationOk := symbolMap["location"].(map[string]any); locationOk {
		symbol.Location = parseLocationFromMap(location)
	}
	if containerName, containerNameOk := symbolMap["containerName"].(string); containerNameOk {
		symbol.ContainerName = containerName
	}

	return symbol
}

// parseWorkspaceSymbolsFromResponse extracts workspace symbols from LSP response.
func parseWorkspaceSymbolsFromResponse(response map[string]any) ([]SymbolInformation, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	symbols, symbolsOk := result.([]any)
	if !symbolsOk {
		return nil, fmt.Errorf("invalid response format")
	}

	var workspaceSymbols []SymbolInformation
	for _, symbol := range symbols {
		if symbolMap, symbolMapOk := symbol.(map[string]any); symbolMapOk {
			symbolInfo := parseSymbolInformationFromMap(symbolMap)
			workspaceSymbols = append(workspaceSymbols, symbolInfo)
		}
	}
	return workspaceSymbols, nil
}

// parseDiagnosticFromMap parses a single diagnostic from a map.
func parseDiagnosticFromMap(diagMap map[string]any) Diagnostic {
	var diagnostic Diagnostic

	if rangeMap, rangeMapOk := diagMap["range"].(map[string]any); rangeMapOk {
		diagnostic.Range = parseRange(rangeMap)
	}
	if severity, severityOk := diagMap["severity"].(float64); severityOk {
		diagnostic.Severity = DiagnosticSeverity(int(severity))
	}
	if code, codeOk := diagMap["code"].(string); codeOk {
		diagnostic.Code = code
	}
	if source, sourceOk := diagMap["source"].(string); sourceOk {
		diagnostic.Source = source
	}
	if message, messageOk := diagMap["message"].(string); messageOk {
		diagnostic.Message = message
	}
	if tags, tagsOk := diagMap["tags"].([]any); tagsOk {
		for _, tag := range tags {
			if tagFloat, tagFloatOk := tag.(float64); tagFloatOk {
				diagnostic.Tags = append(diagnostic.Tags, DiagnosticTag(int(tagFloat)))
			}
		}
	}
	if relatedInfo, relatedInfoOk := diagMap["relatedInformation"]; relatedInfoOk {
		diagnostic.RelatedInformation = []any{relatedInfo}
	}

	return diagnostic
}

// parseDiagnosticsFromResponse extracts diagnostics from LSP response.
func parseDiagnosticsFromResponse(response map[string]any) ([]Diagnostic, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	// Handle both direct array and object with items array
	var diagnosticsArray []any
	if diagArray, diagArrayOk := result.([]any); diagArrayOk {
		diagnosticsArray = diagArray
	} else if resultMap, resultMapOk := result.(map[string]any); resultMapOk {
		if items, itemsOk := resultMap["items"].([]any); itemsOk {
			diagnosticsArray = items
		} else {
			return nil, fmt.Errorf("invalid diagnostics response format")
		}
	} else {
		return nil, fmt.Errorf("invalid diagnostics response format")
	}

	var diagnostics []Diagnostic
	for _, diag := range diagnosticsArray {
		if diagMap, diagMapOk := diag.(map[string]any); diagMapOk {
			diagnostic := parseDiagnosticFromMap(diagMap)
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	return diagnostics, nil
}

// parseCompletionItemFromMap parses a single completion item from a map.
func parseCompletionItemFromMap(itemMap map[string]any) CompletionItem {
	var item CompletionItem

	if label, labelOk := itemMap["label"].(string); labelOk {
		item.Label = label
	}
	if kind, kindOk := itemMap["kind"].(float64); kindOk {
		item.Kind = CompletionItemKind(int(kind))
	}
	if detail, detailOk := itemMap["detail"].(string); detailOk {
		item.Detail = detail
	}
	if documentation, docOk := itemMap["documentation"].(string); docOk {
		item.Documentation = documentation
	}
	if deprecated, deprecatedOk := itemMap["deprecated"].(bool); deprecatedOk {
		item.Deprecated = deprecated
	}
	if preselect, preselectOk := itemMap["preselect"].(bool); preselectOk {
		item.Preselect = preselect
	}
	if sortText, sortTextOk := itemMap["sortText"].(string); sortTextOk {
		item.SortText = sortText
	}
	if filterText, filterTextOk := itemMap["filterText"].(string); filterTextOk {
		item.FilterText = filterText
	}
	if insertText, insertTextOk := itemMap["insertText"].(string); insertTextOk {
		item.InsertText = insertText
	}
	if insertTextFormat, insertTextFormatOk := itemMap["insertTextFormat"].(float64); insertTextFormatOk {
		item.InsertTextFormat = int(insertTextFormat)
	}
	if tags, tagsOk := itemMap["tags"].([]any); tagsOk {
		for _, tag := range tags {
			if tagFloat, tagFloatOk := tag.(float64); tagFloatOk {
				item.Tags = append(item.Tags, int(tagFloat))
			}
		}
	}
	if commitCharacters, commitCharactersOk := itemMap["commitCharacters"].([]any); commitCharactersOk {
		for _, char := range commitCharacters {
			if charStr, charStrOk := char.(string); charStrOk {
				item.CommitCharacters = append(item.CommitCharacters, charStr)
			}
		}
	}
	if textEdit, textEditOk := itemMap["textEdit"]; textEditOk {
		item.TextEdit = textEdit
	}
	if additionalTextEdits, additionalTextEditsOk := itemMap["additionalTextEdits"]; additionalTextEditsOk {
		if editsArray, editsArrayOk := additionalTextEdits.([]any); editsArrayOk {
			item.AdditionalTextEdits = editsArray
		}
	}
	if command, commandOk := itemMap["command"]; commandOk {
		item.Command = command
	}
	if data, dataOk := itemMap["data"]; dataOk {
		item.Data = data
	}

	return item
}

// parseCompletionsFromResponse extracts completion items from LSP response.
func parseCompletionsFromResponse(response map[string]any) ([]CompletionItem, error) {
	result, resultOk := response["result"]
	if !resultOk {
		return nil, fmt.Errorf("invalid response format")
	}

	// Handle both direct array and object with items array
	var itemsArray []any
	if items, itemsOk := result.([]any); itemsOk {
		itemsArray = items
	} else if resultMap, resultMapOk := result.(map[string]any); resultMapOk {
		if itemsData, itemsDataOk := resultMap["items"].([]any); itemsDataOk {
			itemsArray = itemsData
		} else {
			return nil, fmt.Errorf("invalid completions response format")
		}
	} else {
		return nil, fmt.Errorf("invalid completions response format")
	}

	var completions []CompletionItem
	for _, item := range itemsArray {
		if itemMap, itemMapOk := item.(map[string]any); itemMapOk {
			completion := parseCompletionItemFromMap(itemMap)
			completions = append(completions, completion)
		}
	}
	return completions, nil
}
