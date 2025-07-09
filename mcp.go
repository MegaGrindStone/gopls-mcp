package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCP tool parameter types

// GoToDefinitionParams represents parameters for go to definition requests.
type GoToDefinitionParams struct {
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// FindReferencesParams represents parameters for find references requests.
type FindReferencesParams struct {
	Path               string `json:"path"`
	Line               int    `json:"line"`
	Character          int    `json:"character"`
	IncludeDeclaration bool   `json:"includeDeclaration"`
}

// GetHoverParams represents parameters for get hover info requests.
type GetHoverParams struct {
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
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

// mcpTools wraps a goplsClient to provide MCP tool functionality.
type mcpTools struct {
	client *goplsClient
}

// newMCPTools creates a new MCP tools instance wrapping the given goplsClient.
func newMCPTools(client *goplsClient) mcpTools {
	return mcpTools{
		client: client,
	}
}

// convertLocationsToResults converts Location structs to LocationResult structs.
func (m mcpTools) convertLocationsToResults(locations []Location) []LocationResult {
	results := make([]LocationResult, len(locations))
	for i, loc := range locations {
		results[i] = LocationResult{
			URI:          loc.URI,
			Line:         loc.Range.Start.Line,
			Character:    loc.Range.Start.Character,
			EndLine:      loc.Range.End.Line,
			EndCharacter: loc.Range.End.Character,
		}
	}
	return results
}

// MCP tool handlers

// HandleGoToDefinition handles go to definition requests.
func (m mcpTools) HandleGoToDefinition(
	_ context.Context,
	_ *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GoToDefinitionParams],
) (*mcp.CallToolResultFor[GoToDefinitionResult], error) {
	if !m.client.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	locations, err := m.client.goToDefinition(params.Arguments.Path, params.Arguments.Line, params.Arguments.Character)
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
	if !m.client.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	locations, err := m.client.findReferences(
		params.Arguments.Path,
		params.Arguments.Line,
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
	if !m.client.isRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}

	hover, err := m.client.getHover(params.Arguments.Path, params.Arguments.Line, params.Arguments.Character)
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
			Line:         hover.Range.Start.Line,
			Character:    hover.Range.Start.Character,
			EndLine:      hover.Range.End.Line,
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

// MCP tool creation methods

// CreateGoToDefinitionTool creates the go to definition MCP tool.
func (m mcpTools) CreateGoToDefinitionTool() *mcp.ServerTool {
	return mcp.NewServerTool[GoToDefinitionParams, GoToDefinitionResult](
		"go_to_definition",
		"Navigate to the definition of a symbol at the specified position in a Go file",
		m.HandleGoToDefinition,
		mcp.Input(
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateFindReferencesTool creates the find references MCP tool.
func (m mcpTools) CreateFindReferencesTool() *mcp.ServerTool {
	return mcp.NewServerTool[FindReferencesParams, FindReferencesResult](
		"find_references",
		"Find all references to a symbol at the specified position in a Go file",
		m.HandleFindReferences,
		mcp.Input(
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
			mcp.Property("includeDeclaration", mcp.Description("Include declaration in results"), mcp.Required(false)),
		),
	)
}

// CreateGetHoverTool creates the get hover info MCP tool.
func (m mcpTools) CreateGetHoverTool() *mcp.ServerTool {
	return mcp.NewServerTool[GetHoverParams, GetHoverResult](
		"get_hover_info",
		"Get hover information (documentation, type info) for a symbol at the specified position",
		m.HandleGetHover,
		mcp.Input(
			mcp.Property("path", mcp.Description("Relative path to Go file (e.g., main.go, pkg/client.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// setupMCPServer creates and configures the MCP server with gopls tools.
func setupMCPServer(client *goplsClient) *mcp.Server {
	// Create MCP server
	server := mcp.NewServer("gopls-mcp", "v0.1.0", nil)

	// Create MCP tools wrapper
	tools := newMCPTools(client)

	// Add gopls tools
	server.AddTools(
		tools.CreateGoToDefinitionTool(),
		tools.CreateFindReferencesTool(),
		tools.CreateGetHoverTool(),
	)

	return server
}
