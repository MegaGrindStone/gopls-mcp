package tools

import (
	"context"
	"fmt"

	"github.com/MegaGrindStone/gopls-mcp/internal/gopls"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Manager manages MCP tools that interact with gopls
type Manager struct {
	goplsManager *gopls.Manager
}

// NewManager creates a new tools manager
func NewManager(goplsManager *gopls.Manager) Manager {
	return Manager{
		goplsManager: goplsManager,
	}
}

// GoToDefinitionParams represents parameters for go to definition requests
type GoToDefinitionParams struct {
	URI       string `json:"uri"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// FindReferencesParams represents parameters for find references requests
type FindReferencesParams struct {
	URI                string `json:"uri"`
	Line               int    `json:"line"`
	Character          int    `json:"character"`
	IncludeDeclaration bool   `json:"includeDeclaration"`
}

// GetHoverParams represents parameters for get hover info requests
type GetHoverParams struct {
	URI       string `json:"uri"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// LocationResult represents a location result
type LocationResult struct {
	URI   string `json:"uri"`
	Line  int    `json:"line"`
	Character int `json:"character"`
	EndLine int `json:"endLine"`
	EndCharacter int `json:"endCharacter"`
}

// GoToDefinitionResult represents the result of a go to definition request
type GoToDefinitionResult struct {
	Locations []LocationResult `json:"locations"`
}

// FindReferencesResult represents the result of a find references request
type FindReferencesResult struct {
	Locations []LocationResult `json:"locations"`
}

// GetHoverResult represents the result of a get hover request
type GetHoverResult struct {
	Contents []string `json:"contents"`
	HasRange bool     `json:"hasRange"`
	Range    *LocationResult `json:"range,omitempty"`
}

// HandleGoToDefinition handles go to definition requests
func (m Manager) HandleGoToDefinition(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GoToDefinitionParams]) (*mcp.CallToolResultFor[GoToDefinitionResult], error) {
	if !m.goplsManager.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}
	
	locations, err := m.goplsManager.GoToDefinition(ctx, params.Arguments.URI, params.Arguments.Line, params.Arguments.Character)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}
	
	result := GoToDefinitionResult{
		Locations: make([]LocationResult, len(locations)),
	}
	
	for i, loc := range locations {
		result.Locations[i] = LocationResult{
			URI:          loc.URI,
			Line:         loc.Range.Start.Line,
			Character:    loc.Range.Start.Character,
			EndLine:      loc.Range.End.Line,
			EndCharacter: loc.Range.End.Character,
		}
	}
	
	return &mcp.CallToolResultFor[GoToDefinitionResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d definition(s)", len(result.Locations)),
			},
		},
	}, nil
}

// HandleFindReferences handles find references requests
func (m Manager) HandleFindReferences(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[FindReferencesParams]) (*mcp.CallToolResultFor[FindReferencesResult], error) {
	if !m.goplsManager.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}
	
	locations, err := m.goplsManager.FindReferences(ctx, params.Arguments.URI, params.Arguments.Line, params.Arguments.Character, params.Arguments.IncludeDeclaration)
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}
	
	result := FindReferencesResult{
		Locations: make([]LocationResult, len(locations)),
	}
	
	for i, loc := range locations {
		result.Locations[i] = LocationResult{
			URI:          loc.URI,
			Line:         loc.Range.Start.Line,
			Character:    loc.Range.Start.Character,
			EndLine:      loc.Range.End.Line,
			EndCharacter: loc.Range.End.Character,
		}
	}
	
	return &mcp.CallToolResultFor[FindReferencesResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d reference(s)", len(result.Locations)),
			},
		},
	}, nil
}

// HandleGetHover handles get hover info requests
func (m Manager) HandleGetHover(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GetHoverParams]) (*mcp.CallToolResultFor[GetHoverResult], error) {
	if !m.goplsManager.IsRunning() {
		return nil, fmt.Errorf("gopls is not running")
	}
	
	hover, err := m.goplsManager.GetHover(ctx, params.Arguments.URI, params.Arguments.Line, params.Arguments.Character)
	if err != nil {
		return nil, fmt.Errorf("failed to get hover info: %w", err)
	}
	
	result := GetHoverResult{
		Contents: hover.Contents,
		HasRange: hover.Range != nil,
	}
	
	if hover.Range != nil {
		result.Range = &LocationResult{
			URI:          params.Arguments.URI,
			Line:         hover.Range.Start.Line,
			Character:    hover.Range.Start.Character,
			EndLine:      hover.Range.End.Line,
			EndCharacter: hover.Range.End.Character,
		}
	}
	
	contentText := "Hover information:\n"
	for _, content := range hover.Contents {
		contentText += content + "\n"
	}
	
	return &mcp.CallToolResultFor[GetHoverResult]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: contentText,
			},
		},
	}, nil
}

// CreateGoToDefinitionTool creates the go to definition MCP tool
func (m Manager) CreateGoToDefinitionTool() *mcp.ServerTool {
	return mcp.NewServerTool[GoToDefinitionParams, GoToDefinitionResult](
		"go_to_definition",
		"Navigate to the definition of a symbol at the specified position in a Go file",
		m.HandleGoToDefinition,
		mcp.Input(
			mcp.Property("uri", mcp.Description("File URI (e.g., file:///path/to/file.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

// CreateFindReferencesTool creates the find references MCP tool
func (m Manager) CreateFindReferencesTool() *mcp.ServerTool {
	return mcp.NewServerTool[FindReferencesParams, FindReferencesResult](
		"find_references",
		"Find all references to a symbol at the specified position in a Go file",
		m.HandleFindReferences,
		mcp.Input(
			mcp.Property("uri", mcp.Description("File URI (e.g., file:///path/to/file.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
			mcp.Property("includeDeclaration", mcp.Description("Include declaration in results"), mcp.Required(false)),
		),
	)
}

// CreateGetHoverTool creates the get hover info MCP tool
func (m Manager) CreateGetHoverTool() *mcp.ServerTool {
	return mcp.NewServerTool[GetHoverParams, GetHoverResult](
		"get_hover_info",
		"Get hover information (documentation, type info) for a symbol at the specified position",
		m.HandleGetHover,
		mcp.Input(
			mcp.Property("uri", mcp.Description("File URI (e.g., file:///path/to/file.go)"), mcp.Required(true)),
			mcp.Property("line", mcp.Description("Line number (0-based)"), mcp.Required(true)),
			mcp.Property("character", mcp.Description("Character position (0-based)"), mcp.Required(true)),
		),
	)
}

