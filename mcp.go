package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCP tool parameter types

// GoToDefinitionParams represents parameters for go to definition requests.
type GoToDefinitionParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// FindReferencesParams represents parameters for find references requests.
type FindReferencesParams struct {
	Workspace          string `json:"workspace"`
	Path               string `json:"path"`
	Line               int    `json:"line"`
	Character          int    `json:"character"`
	IncludeDeclaration bool   `json:"includeDeclaration"`
}

// GetHoverParams represents parameters for get hover info requests.
type GetHoverParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// GetDiagnosticsParams represents parameters for get diagnostics requests.
type GetDiagnosticsParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
}

// GetDocumentSymbolsParams represents parameters for get document symbols requests.
type GetDocumentSymbolsParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
}

// GetWorkspaceSymbolsParams represents parameters for get workspace symbols requests.
type GetWorkspaceSymbolsParams struct {
	Workspace string `json:"workspace"`
	Query     string `json:"query"`
}

// GetSignatureHelpParams represents parameters for get signature help requests.
type GetSignatureHelpParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// GetCompletionsParams represents parameters for get completions requests.
type GetCompletionsParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// GetTypeDefinitionParams represents parameters for get type definition requests.
type GetTypeDefinitionParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// FindImplementationsParams represents parameters for find implementations requests.
type FindImplementationsParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// FormatDocumentParams represents parameters for format document requests.
type FormatDocumentParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
}

// OrganizeImportsParams represents parameters for organize imports requests.
type OrganizeImportsParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
}

// GetInlayHintsParams represents parameters for get inlay hints requests.
type GetInlayHintsParams struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
	StartLine int    `json:"startLine"`
	StartChar int    `json:"startChar"`
	EndLine   int    `json:"endLine"`
	EndChar   int    `json:"endChar"`
}

// ListWorkspacesParams represents parameters for list workspaces requests.
type ListWorkspacesParams struct {
	// No parameters needed
}

// MCP tool result types

// LocationResult represents a location result.
type LocationResult struct {
	URI          string `json:"uri"`
	Line         int    `json:"line"`
	Character    int    `json:"character"`
	EndLine      int    `json:"endLine"`
	EndCharacter int    `json:"endCharacter"`
}

// GoToDefinitionResult represents the result of a go to definition request.
type GoToDefinitionResult struct {
	Locations []LocationResult `json:"locations"`
}

// FindReferencesResult represents the result of a find references request.
type FindReferencesResult struct {
	Locations []LocationResult `json:"locations"`
}

// GetHoverResult represents the result of a get hover request.
type GetHoverResult struct {
	Contents []string        `json:"contents"`
	HasRange bool            `json:"hasRange"`
	Range    *LocationResult `json:"range,omitempty"`
}

// DiagnosticResult represents a diagnostic result.
type DiagnosticResult struct {
	Range    LocationResult `json:"range"`
	Severity int            `json:"severity"`
	Code     string         `json:"code,omitempty"`
	Source   string         `json:"source,omitempty"`
	Message  string         `json:"message"`
}

// GetDiagnosticsResult represents the result of a get diagnostics request.
type GetDiagnosticsResult struct {
	Diagnostics []DiagnosticResult `json:"diagnostics"`
}

// DocumentSymbolResult represents a document symbol result.
type DocumentSymbolResult struct {
	Name           string                 `json:"name"`
	Detail         string                 `json:"detail,omitempty"`
	Kind           int                    `json:"kind"`
	Deprecated     bool                   `json:"deprecated,omitempty"`
	Range          LocationResult         `json:"range"`
	SelectionRange LocationResult         `json:"selectionRange"`
	Children       []DocumentSymbolResult `json:"children,omitempty"`
}

// GetDocumentSymbolsResult represents the result of a get document symbols request.
type GetDocumentSymbolsResult struct {
	Symbols []DocumentSymbolResult `json:"symbols"`
}

// WorkspaceSymbolResult represents a workspace symbol result.
type WorkspaceSymbolResult struct {
	Name          string         `json:"name"`
	Kind          int            `json:"kind"`
	Deprecated    bool           `json:"deprecated,omitempty"`
	Location      LocationResult `json:"location"`
	ContainerName string         `json:"containerName,omitempty"`
}

// GetWorkspaceSymbolsResult represents the result of a get workspace symbols request.
type GetWorkspaceSymbolsResult struct {
	Symbols []WorkspaceSymbolResult `json:"symbols"`
}

// ParameterInformationResult represents parameter information.
type ParameterInformationResult struct {
	Label         string `json:"label"`
	Documentation string `json:"documentation,omitempty"`
}

// SignatureInformationResult represents signature information.
type SignatureInformationResult struct {
	Label         string                       `json:"label"`
	Documentation string                       `json:"documentation,omitempty"`
	Parameters    []ParameterInformationResult `json:"parameters,omitempty"`
}

// GetSignatureHelpResult represents the result of a get signature help request.
type GetSignatureHelpResult struct {
	Signatures      []SignatureInformationResult `json:"signatures"`
	ActiveSignature int                          `json:"activeSignature,omitempty"`
	ActiveParameter int                          `json:"activeParameter,omitempty"`
}

// CompletionItemResult represents a completion item.
type CompletionItemResult struct {
	Label            string `json:"label"`
	Kind             int    `json:"kind,omitempty"`
	Detail           string `json:"detail,omitempty"`
	Documentation    string `json:"documentation,omitempty"`
	InsertText       string `json:"insertText,omitempty"`
	InsertTextFormat int    `json:"insertTextFormat,omitempty"`
	SortText         string `json:"sortText,omitempty"`
	FilterText       string `json:"filterText,omitempty"`
}

// GetCompletionsResult represents the result of a get completions request.
type GetCompletionsResult struct {
	Items        []CompletionItemResult `json:"items"`
	IsIncomplete bool                   `json:"isIncomplete"`
}

// GetTypeDefinitionResult represents the result of a get type definition request.
type GetTypeDefinitionResult struct {
	Locations []LocationResult `json:"locations"`
}

// FindImplementationsResult represents the result of a find implementations request.
type FindImplementationsResult struct {
	Locations []LocationResult `json:"locations"`
}

// TextEditResult represents a text edit result.
type TextEditResult struct {
	Range   LocationResult `json:"range"`
	NewText string         `json:"newText"`
}

// FormatDocumentResult represents the result of a format document request.
type FormatDocumentResult struct {
	Edits []TextEditResult `json:"edits"`
}

// OrganizeImportsResult represents the result of an organize imports request.
type OrganizeImportsResult struct {
	Edits []TextEditResult `json:"edits"`
}

// InlayHintResult represents an inlay hint result.
type InlayHintResult struct {
	Position LocationResult `json:"position"`
	Label    string         `json:"label"`
	Kind     int            `json:"kind,omitempty"`
	Tooltip  string         `json:"tooltip,omitempty"`
}

// GetInlayHintsResult represents the result of a get inlay hints request.
type GetInlayHintsResult struct {
	Hints []InlayHintResult `json:"hints"`
}

// WorkspaceInfo represents information about a workspace.
type WorkspaceInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// ListWorkspacesResult represents the result of a list workspaces request.
type ListWorkspacesResult struct {
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

// Line number conversion functions for MCP layer (1-based) to LSP layer (0-based)

// convertLineToLSP converts a 1-based line number from MCP to 0-based for LSP.
func convertLineToLSP(line int) int {
	if line <= 0 {
		return 0
	}
	return line - 1
}

// convertLineFromLSP converts a 0-based line number from LSP to 1-based for MCP.
func convertLineFromLSP(line int) int {
	return line + 1
}

// mcpTools wraps multiple goplsClients to provide MCP tool functionality.
type mcpTools struct {
	clients map[string]*goplsClient
}

// newMCPTools creates a new MCP tools instance wrapping the given goplsClients.
func newMCPTools(clients map[string]*goplsClient) mcpTools {
	return mcpTools{
		clients: clients,
	}
}

// getClient returns the goplsClient for the specified workspace.
func (m mcpTools) getClient(workspace string) (*goplsClient, error) {
	client, exists := m.clients[workspace]
	if !exists {
		return nil, fmt.Errorf("workspace not found: %s", workspace)
	}
	if !client.isRunning() {
		return nil, fmt.Errorf("gopls is not running for workspace: %s", workspace)
	}
	return client, nil
}

// convertLocationsToResults converts Location structs to LocationResult structs.
func (m mcpTools) convertLocationsToResults(locations []Location) []LocationResult {
	results := make([]LocationResult, len(locations))
	for i, loc := range locations {
		results[i] = LocationResult{
			URI:          loc.URI,
			Line:         convertLineFromLSP(loc.Range.Start.Line),
			Character:    loc.Range.Start.Character,
			EndLine:      convertLineFromLSP(loc.Range.End.Line),
			EndCharacter: loc.Range.End.Character,
		}
	}
	return results
}

// convertLocationToResult converts a Location struct to LocationResult struct.
func (m mcpTools) convertLocationToResult(location Location) LocationResult {
	return LocationResult{
		URI:          location.URI,
		Line:         convertLineFromLSP(location.Range.Start.Line),
		Character:    location.Range.Start.Character,
		EndLine:      convertLineFromLSP(location.Range.End.Line),
		EndCharacter: location.Range.End.Character,
	}
}

// convertDocumentSymbolToResult converts a DocumentSymbol struct to DocumentSymbolResult struct.
func (m mcpTools) convertDocumentSymbolToResult(symbol DocumentSymbol) DocumentSymbolResult {
	children := make([]DocumentSymbolResult, len(symbol.Children))
	for i, child := range symbol.Children {
		children[i] = m.convertDocumentSymbolToResult(child)
	}

	return DocumentSymbolResult{
		Name:       symbol.Name,
		Detail:     symbol.Detail,
		Kind:       symbol.Kind,
		Deprecated: symbol.Deprecated,
		Range: LocationResult{
			URI:          "",
			Line:         convertLineFromLSP(symbol.Range.Start.Line),
			Character:    symbol.Range.Start.Character,
			EndLine:      convertLineFromLSP(symbol.Range.End.Line),
			EndCharacter: symbol.Range.End.Character,
		},
		SelectionRange: LocationResult{
			URI:          "",
			Line:         convertLineFromLSP(symbol.SelectionRange.Start.Line),
			Character:    symbol.SelectionRange.Start.Character,
			EndLine:      convertLineFromLSP(symbol.SelectionRange.End.Line),
			EndCharacter: symbol.SelectionRange.End.Character,
		},
		Children: children,
	}
}

// convertTextEditsToResults converts TextEdit structs to TextEditResult structs.
func (m mcpTools) convertTextEditsToResults(textEdits []TextEdit) []TextEditResult {
	results := make([]TextEditResult, len(textEdits))
	for i, edit := range textEdits {
		results[i] = TextEditResult{
			Range: LocationResult{
				URI:          "",
				Line:         convertLineFromLSP(edit.Range.Start.Line),
				Character:    edit.Range.Start.Character,
				EndLine:      convertLineFromLSP(edit.Range.End.Line),
				EndCharacter: edit.Range.End.Character,
			},
			NewText: edit.NewText,
		}
	}
	return results
}

// convertInlayHintsToResults converts InlayHint structs to InlayHintResult structs.
func (m mcpTools) convertInlayHintsToResults(inlayHints []InlayHint) []InlayHintResult {
	results := make([]InlayHintResult, len(inlayHints))
	for i, hint := range inlayHints {
		results[i] = InlayHintResult{
			Position: LocationResult{
				URI:          "",
				Line:         convertLineFromLSP(hint.Position.Line),
				Character:    hint.Position.Character,
				EndLine:      convertLineFromLSP(hint.Position.Line),
				EndCharacter: hint.Position.Character,
			},
			Label:   hint.Label,
			Kind:    hint.Kind,
			Tooltip: hint.Tooltip,
		}
	}
	return results
}

// MCP tool handlers

// HandleListWorkspaces handles list workspaces requests.
func (m mcpTools) HandleListWorkspaces(
	_ context.Context,
	_ *mcp.ServerSession,
	_ *mcp.CallToolParamsFor[ListWorkspacesParams],
) (*mcp.CallToolResultFor[ListWorkspacesResult], error) {
	workspaces := make([]WorkspaceInfo, 0, len(m.clients))
	for workspacePath := range m.clients {
		workspaces = append(workspaces, WorkspaceInfo{
			Path: workspacePath,
			Name: filepath.Base(workspacePath),
		})
	}

	result := ListWorkspacesResult{
		Workspaces: workspaces,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[ListWorkspacesResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGoToDefinition handles go to definition requests.
//
//nolint:dupl // Similar pattern across location-based handlers is acceptable
func (m mcpTools) HandleGoToDefinition(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GoToDefinitionParams],
) (*mcp.CallToolResultFor[GoToDefinitionResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	locations, err := client.goToDefinition(
		params.Arguments.Path,
		convertLineToLSP(params.Arguments.Line),
		params.Arguments.Character,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	result := GoToDefinitionResult{
		Locations: m.convertLocationsToResults(locations),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GoToDefinitionResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleFindReferences handles find references requests.
func (m mcpTools) HandleFindReferences(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[FindReferencesParams],
) (*mcp.CallToolResultFor[FindReferencesResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	locations, err := client.findReferences(
		params.Arguments.Path,
		convertLineToLSP(params.Arguments.Line),
		params.Arguments.Character,
		params.Arguments.IncludeDeclaration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	result := FindReferencesResult{
		Locations: m.convertLocationsToResults(locations),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[FindReferencesResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetHover handles get hover info requests.
func (m mcpTools) HandleGetHover(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetHoverParams],
) (*mcp.CallToolResultFor[GetHoverResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	hover, err := client.getHover(
		params.Arguments.Path,
		convertLineToLSP(params.Arguments.Line),
		params.Arguments.Character,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get hover info: %w", err)
	}

	result := GetHoverResult{
		Contents: hover.Contents,
		HasRange: hover.Range != nil,
	}

	if hover.Range != nil {
		result.Range = &LocationResult{
			URI:          params.Arguments.Path,
			Line:         convertLineFromLSP(hover.Range.Start.Line),
			Character:    hover.Range.Start.Character,
			EndLine:      convertLineFromLSP(hover.Range.End.Line),
			EndCharacter: hover.Range.End.Character,
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetHoverResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetDiagnostics handles get diagnostics requests.
func (m mcpTools) HandleGetDiagnostics(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetDiagnosticsParams],
) (*mcp.CallToolResultFor[GetDiagnosticsResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	diagnostics, err := client.getDiagnostics(params.Arguments.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get diagnostics: %w", err)
	}

	// Convert diagnostics to results
	diagResults := make([]DiagnosticResult, len(diagnostics))
	for i, diag := range diagnostics {
		diagResults[i] = DiagnosticResult{
			Range: LocationResult{
				URI:          params.Arguments.Path,
				Line:         convertLineFromLSP(diag.Range.Start.Line),
				Character:    diag.Range.Start.Character,
				EndLine:      convertLineFromLSP(diag.Range.End.Line),
				EndCharacter: diag.Range.End.Character,
			},
			Severity: int(diag.Severity),
			Code:     diag.Code,
			Source:   diag.Source,
			Message:  diag.Message,
		}
	}

	result := GetDiagnosticsResult{
		Diagnostics: diagResults,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetDiagnosticsResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetDocumentSymbols handles get document symbols requests.
func (m mcpTools) HandleGetDocumentSymbols(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetDocumentSymbolsParams],
) (*mcp.CallToolResultFor[GetDocumentSymbolsResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	symbols, err := client.getDocumentSymbols(params.Arguments.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get document symbols: %w", err)
	}

	// Convert symbols to results
	symbolResults := make([]DocumentSymbolResult, len(symbols))
	for i, sym := range symbols {
		symbolResults[i] = m.convertDocumentSymbolToResult(sym)
	}

	result := GetDocumentSymbolsResult{
		Symbols: symbolResults,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetDocumentSymbolsResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetWorkspaceSymbols handles get workspace symbols requests.
func (m mcpTools) HandleGetWorkspaceSymbols(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetWorkspaceSymbolsParams],
) (*mcp.CallToolResultFor[GetWorkspaceSymbolsResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	symbols, err := client.getWorkspaceSymbols(params.Arguments.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace symbols: %w", err)
	}

	// Convert symbols to results
	symbolResults := make([]WorkspaceSymbolResult, len(symbols))
	for i, sym := range symbols {
		symbolResults[i] = WorkspaceSymbolResult{
			Name:          sym.Name,
			Kind:          sym.Kind,
			Deprecated:    sym.Deprecated,
			Location:      m.convertLocationToResult(sym.Location),
			ContainerName: sym.ContainerName,
		}
	}

	result := GetWorkspaceSymbolsResult{
		Symbols: symbolResults,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetWorkspaceSymbolsResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetSignatureHelp handles get signature help requests.
func (m mcpTools) HandleGetSignatureHelp(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetSignatureHelpParams],
) (*mcp.CallToolResultFor[GetSignatureHelpResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	signatureHelp, err := client.getSignatureHelp(
		params.Arguments.Path, convertLineToLSP(params.Arguments.Line), params.Arguments.Character)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature help: %w", err)
	}

	// Convert signature help to result
	sigResults := make([]SignatureInformationResult, len(signatureHelp.Signatures))
	for i, sig := range signatureHelp.Signatures {
		paramResults := make([]ParameterInformationResult, len(sig.Parameters))
		for j, param := range sig.Parameters {
			paramResults[j] = ParameterInformationResult(param)
		}
		sigResults[i] = SignatureInformationResult{
			Label:         sig.Label,
			Documentation: sig.Documentation,
			Parameters:    paramResults,
		}
	}

	result := GetSignatureHelpResult{
		Signatures:      sigResults,
		ActiveSignature: signatureHelp.ActiveSignature,
		ActiveParameter: signatureHelp.ActiveParameter,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetSignatureHelpResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetCompletions handles get completions requests.
func (m mcpTools) HandleGetCompletions(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetCompletionsParams],
) (*mcp.CallToolResultFor[GetCompletionsResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	completions, err := client.getCompletions(
		params.Arguments.Path,
		convertLineToLSP(params.Arguments.Line),
		params.Arguments.Character,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get completions: %w", err)
	}

	// Convert completion items to results
	itemResults := make([]CompletionItemResult, len(completions.Items))
	for i, item := range completions.Items {
		itemResults[i] = CompletionItemResult(item)
	}

	result := GetCompletionsResult{
		Items:        itemResults,
		IsIncomplete: completions.IsIncomplete,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetCompletionsResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetTypeDefinition handles get type definition requests.
//
//nolint:dupl // Similar pattern across location-based handlers is acceptable
func (m mcpTools) HandleGetTypeDefinition(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetTypeDefinitionParams],
) (*mcp.CallToolResultFor[GetTypeDefinitionResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	locations, err := client.getTypeDefinition(
		params.Arguments.Path,
		convertLineToLSP(params.Arguments.Line),
		params.Arguments.Character,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get type definition: %w", err)
	}

	result := GetTypeDefinitionResult{
		Locations: m.convertLocationsToResults(locations),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetTypeDefinitionResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleFindImplementations handles find implementations requests.
//
//nolint:dupl // Similar pattern across location-based handlers is acceptable
func (m mcpTools) HandleFindImplementations(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[FindImplementationsParams],
) (*mcp.CallToolResultFor[FindImplementationsResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	locations, err := client.findImplementations(
		params.Arguments.Path, convertLineToLSP(params.Arguments.Line), params.Arguments.Character)
	if err != nil {
		return nil, fmt.Errorf("failed to find implementations: %w", err)
	}

	result := FindImplementationsResult{
		Locations: m.convertLocationsToResults(locations),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[FindImplementationsResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleFormatDocument handles format document requests.
func (m mcpTools) HandleFormatDocument(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[FormatDocumentParams],
) (*mcp.CallToolResultFor[FormatDocumentResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	textEdits, err := client.formatDocument(params.Arguments.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to format document: %w", err)
	}

	result := FormatDocumentResult{
		Edits: m.convertTextEditsToResults(textEdits),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[FormatDocumentResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleOrganizeImports handles organize imports requests.
func (m mcpTools) HandleOrganizeImports(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[OrganizeImportsParams],
) (*mcp.CallToolResultFor[OrganizeImportsResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	textEdits, err := client.organizeImports(params.Arguments.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to organize imports: %w", err)
	}

	result := OrganizeImportsResult{
		Edits: m.convertTextEditsToResults(textEdits),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[OrganizeImportsResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// HandleGetInlayHints handles get inlay hints requests.
func (m mcpTools) HandleGetInlayHints(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetInlayHintsParams],
) (*mcp.CallToolResultFor[GetInlayHintsResult], error) {
	client, err := m.getClient(params.Arguments.Workspace)
	if err != nil {
		return nil, err
	}

	inlayHints, err := client.getInlayHints(
		params.Arguments.Path,
		convertLineToLSP(params.Arguments.StartLine),
		params.Arguments.StartChar,
		convertLineToLSP(params.Arguments.EndLine),
		params.Arguments.EndChar,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get inlay hints: %w", err)
	}

	result := GetInlayHintsResult{
		Hints: m.convertInlayHintsToResults(inlayHints),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[GetInlayHintsResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil
}

// MCP tool creation methods

// CreateListWorkspacesTool creates the list workspaces MCP tool.
func (m mcpTools) CreateListWorkspacesTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"list_workspaces",
		"List all available Go workspaces configured in the server",
		m.HandleListWorkspaces,
		mcp.Input(),
	)
}

// CreateGoToDefinitionTool creates the go to definition MCP tool.
func (m mcpTools) CreateGoToDefinitionTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"go_to_definition",
		"Navigate to the definition of a symbol at the specified position in a Go file",
		m.HandleGoToDefinition,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (1-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateFindReferencesTool creates the find references MCP tool.
func (m mcpTools) CreateFindReferencesTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"find_references",
		"Find all references to a symbol at the specified position in a Go file",
		m.HandleFindReferences,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (1-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
			mcp.Property("includeDeclaration", mcp.Description("Include declaration in results"), mcp.Required(false)),
		),
	)
}

// CreateGetHoverTool creates the get hover info MCP tool.
func (m mcpTools) CreateGetHoverTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_hover_info",
		"Get hover information (documentation, type info) for a symbol at the specified position",
		m.HandleGetHover,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (1-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateGetDiagnosticsTool creates the get diagnostics MCP tool.
func (m mcpTools) CreateGetDiagnosticsTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_diagnostics",
		"Get compilation errors, warnings, and other diagnostics for a Go file",
		m.HandleGetDiagnostics,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
		),
	)
}

// CreateGetDocumentSymbolsTool creates the get document symbols MCP tool.
func (m mcpTools) CreateGetDocumentSymbolsTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_document_symbols",
		"Get outline of symbols (functions, types, etc.) defined in a Go file",
		m.HandleGetDocumentSymbols,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
		),
	)
}

// CreateGetWorkspaceSymbolsTool creates the get workspace symbols MCP tool.
func (m mcpTools) CreateGetWorkspaceSymbolsTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_workspace_symbols",
		"Search for symbols across the entire Go workspace/project",
		m.HandleGetWorkspaceSymbols,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("query",
				mcp.Description("Search query for symbol names (supports fuzzy matching)"),
				mcp.Required(true)),
		),
	)
}

// CreateGetSignatureHelpTool creates the get signature help MCP tool.
func (m mcpTools) CreateGetSignatureHelpTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_signature_help",
		"Get function signature help (parameter information) at the specified position",
		m.HandleGetSignatureHelp,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (1-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateGetCompletionsTool creates the get completions MCP tool.
func (m mcpTools) CreateGetCompletionsTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_completions",
		"Get code completion suggestions at the specified position",
		m.HandleGetCompletions,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (1-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateGetTypeDefinitionTool creates the get type definition MCP tool.
func (m mcpTools) CreateGetTypeDefinitionTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_type_definition",
		"Navigate to the type definition of a symbol at the specified position",
		m.HandleGetTypeDefinition,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (1-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateFindImplementationsTool creates the find implementations MCP tool.
func (m mcpTools) CreateFindImplementationsTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"find_implementations",
		"Find all implementations of an interface or method at the specified position",
		m.HandleFindImplementations,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (1-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateFormatDocumentTool creates the format document MCP tool.
func (m mcpTools) CreateFormatDocumentTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"format_document",
		"Format a Go source file according to gofmt standards",
		m.HandleFormatDocument,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
		),
	)
}

// CreateOrganizeImportsTool creates the organize imports MCP tool.
func (m mcpTools) CreateOrganizeImportsTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"organize_imports",
		"Organize and clean up import statements in a Go file",
		m.HandleOrganizeImports,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
		),
	)
}

// CreateGetInlayHintsTool creates the get inlay hints MCP tool.
func (m mcpTools) CreateGetInlayHintsTool() *mcp.ServerTool {
	return mcp.NewServerTool(
		"get_inlay_hints",
		"Get inlay hints (implicit parameter names, type information) for a range in a Go file",
		m.HandleGetInlayHints,
		mcp.Input(
			mcp.Property("workspace", mcp.Description("Workspace path to use for this request"), mcp.Required(true)),
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("startLine", mcp.Description("Start line number (1-based)"), mcp.Required(true)),
			mcp.Property("startChar", mcp.Description("Start character position (0-based)"), mcp.Required(true)),
			mcp.Property("endLine", mcp.Description("End line number (1-based)"), mcp.Required(true)),
			mcp.Property("endChar", mcp.Description("End character position (0-based)"), mcp.Required(true)),
		),
	)
}

// setupMCPServer creates and configures the MCP server with gopls tools.
func setupMCPServer(clients map[string]*goplsClient) *mcp.Server {
	// Create MCP server
	server := mcp.NewServer("gopls-mcp", "v0.3.0", nil)

	// Create MCP tools wrapper
	tools := newMCPTools(clients)

	// Add gopls tools
	server.AddTools(
		// Workspace management tools
		tools.CreateListWorkspacesTool(),

		// Core navigation tools
		tools.CreateGoToDefinitionTool(),
		tools.CreateFindReferencesTool(),
		tools.CreateGetHoverTool(),

		// Diagnostic and analysis tools
		tools.CreateGetDiagnosticsTool(),
		tools.CreateGetDocumentSymbolsTool(),
		tools.CreateGetWorkspaceSymbolsTool(),

		// Code assistance tools
		tools.CreateGetSignatureHelpTool(),
		tools.CreateGetCompletionsTool(),

		// Advanced navigation tools
		tools.CreateGetTypeDefinitionTool(),
		tools.CreateFindImplementationsTool(),

		// Code maintenance tools
		tools.CreateFormatDocumentTool(),
		tools.CreateOrganizeImportsTool(),
		tools.CreateGetInlayHintsTool(),
	)

	return server
}
