package main

const (
	fileScheme = "file"
)

// LSP types for gopls communication

// Position represents a position in a document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
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

// DiagnosticSeverity represents the severity level of a diagnostic.
type DiagnosticSeverity int

const (
	// DiagnosticSeverityError represents an error-level diagnostic.
	DiagnosticSeverityError DiagnosticSeverity = 1
	// DiagnosticSeverityWarning represents a warning-level diagnostic.
	DiagnosticSeverityWarning DiagnosticSeverity = 2
	// DiagnosticSeverityInformation represents an information-level diagnostic.
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	// DiagnosticSeverityHint represents a hint-level diagnostic.
	DiagnosticSeverityHint DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic message.
type Diagnostic struct {
	Range       Range                          `json:"range"`
	Severity    DiagnosticSeverity             `json:"severity"`
	Code        string                         `json:"code,omitempty"`
	Source      string                         `json:"source,omitempty"`
	Message     string                         `json:"message"`
	RelatedInfo []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

// DiagnosticRelatedInformation represents related information for a diagnostic.
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// DocumentSymbol represents a symbol in a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Deprecated     bool             `json:"deprecated,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation represents symbol information for workspace symbols.
type SymbolInformation struct {
	Name          string   `json:"name"`
	Kind          int      `json:"kind"`
	Deprecated    bool     `json:"deprecated,omitempty"`
	Location      Location `json:"location"`
	ContainerName string   `json:"containerName,omitempty"`
}

// SignatureHelp represents signature help information.
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature,omitempty"`
	ActiveParameter int                    `json:"activeParameter,omitempty"`
}

// SignatureInformation represents information about a function signature.
type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation string                 `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

// ParameterInformation represents information about a function parameter.
type ParameterInformation struct {
	Label         string `json:"label"`
	Documentation string `json:"documentation,omitempty"`
}

// CompletionItem represents a completion item.
type CompletionItem struct {
	Label            string `json:"label"`
	Kind             int    `json:"kind,omitempty"`
	Detail           string `json:"detail,omitempty"`
	Documentation    string `json:"documentation,omitempty"`
	InsertText       string `json:"insertText,omitempty"`
	InsertTextFormat int    `json:"insertTextFormat,omitempty"`
	SortText         string `json:"sortText,omitempty"`
	FilterText       string `json:"filterText,omitempty"`
}

// CompletionList represents a list of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// InlayHint represents an inlay hint.
type InlayHint struct {
	Position Position `json:"position"`
	Label    string   `json:"label"`
	Kind     int      `json:"kind,omitempty"`
	Tooltip  string   `json:"tooltip,omitempty"`
}

// TextEdit represents a text edit.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// WorkspaceEdit represents a workspace edit.
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes,omitempty"`
}
