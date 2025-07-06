package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	transportHTTP  = "http"
	transportStdio = "stdio"
)

func main() {
	// Initialize logger
	logger := initLogger()
	slog.SetDefault(logger)

	// Parse command line flags
	workspacesFlag := flag.String("workspaces", "", "Comma-separated list of workspace paths (required)")
	transportType := flag.String("transport", "http", "Transport type: http or stdio")
	flag.Parse()

	// Validate that workspaces are provided
	if *workspacesFlag == "" {
		logger.Error("workspaces flag is required")
		os.Exit(1)
	}

	// Parse workspaces from comma-separated string
	workspacePaths := strings.Split(*workspacesFlag, ",")

	// Trim whitespace from each workspace path
	for i, path := range workspacePaths {
		workspacePaths[i] = strings.TrimSpace(path)
	}

	// Validate that we have at least one workspace
	if len(workspacePaths) == 0 {
		logger.Error("at least one workspace path is required")
		os.Exit(1)
	}

	// Validate that all workspace paths are non-empty
	for _, path := range workspacePaths {
		if path == "" {
			logger.Error("workspace path cannot be empty")
			os.Exit(1)
		}
	}

	// Validate transport type
	if *transportType != transportHTTP && *transportType != transportStdio {
		logger.Error("transport must be either 'http' or 'stdio'", "provided", *transportType)
		os.Exit(1)
	}

	// Create workspace manager
	workspaceManager := NewWorkspaceManager(workspacePaths, logger)

	// Start all workspaces
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := workspaceManager.Start(ctx); err != nil {
		logger.Error("failed to start workspaces", "error", err)
		return
	}
	defer func() { _ = workspaceManager.Stop() }()

	// Create and setup MCP server
	server := setupMCPServer(workspaceManager)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down server")
		cancel()
		_ = workspaceManager.Stop()
		os.Exit(0)
	}()

	logger.Info("starting gopls-mcp server",
		"workspaces", workspacePaths,
		"transport", *transportType)

	// Start server based on transport type
	if *transportType == transportStdio {
		logger.Info("using stdio transport")
		if err := server.Run(ctx, mcp.NewStdioTransport()); err != nil {
			logger.Error("stdio server failed", "error", err)
		}
	} else {
		// HTTP transport
		handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
			return server
		}, nil)

		logger.Info("HTTP server available", "url", "http://localhost:8080")

		httpServer := &http.Server{
			Addr:         ":8080",
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		if err := httpServer.ListenAndServe(); err != nil {
			logger.Error("HTTP server failed to start", "error", err)
		}
	}
}

// setupMCPServer creates and configures the MCP server with gopls tools.
func setupMCPServer(workspaceManager *WorkspaceManager) *mcp.Server {
	// Create MCP server
	server := mcp.NewServer("gopls-mcp", "v0.1.0", nil)

	// Add gopls tools (these will be updated to handle multiple workspaces)
	server.AddTools(
		workspaceManager.CreateGoToDefinitionTool(),
		workspaceManager.CreateFindReferencesTool(),
		workspaceManager.CreateGetHoverTool(),
		workspaceManager.CreateListWorkspacesTool(),
		workspaceManager.CreateGetDocumentSymbolsTool(),
		workspaceManager.CreateSearchWorkspaceSymbolsTool(),
		workspaceManager.CreateGoToTypeDefinitionTool(),
		workspaceManager.CreateGetDiagnosticsTool(),
		workspaceManager.CreateFindImplementationsTool(),
		workspaceManager.CreateGetCompletionsTool(),
		workspaceManager.CreateGetCallHierarchyTool(),
		workspaceManager.CreateGetSignatureHelpTool(),
		workspaceManager.CreateGetTypeHierarchyTool(),
	)

	return server
}
