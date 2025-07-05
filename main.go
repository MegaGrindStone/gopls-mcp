package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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
	workspacePath := flag.String("workspace", "", "Path to the Go workspace directory (required)")
	transportType := flag.String("transport", "http", "Transport type: http or stdio")
	flag.Parse()

	// Validate that workspace path is provided
	if *workspacePath == "" {
		logger.Error("workspace flag is required")
		os.Exit(1)
	}

	// Validate transport type
	if *transportType != transportHTTP && *transportType != transportStdio {
		logger.Error("transport must be either 'http' or 'stdio'", "provided", *transportType)
		os.Exit(1)
	}

	// Create gopls manager
	goplsManager := NewManager(*workspacePath, logger)

	// Start gopls
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := goplsManager.Start(ctx); err != nil {
		logger.Error("failed to start gopls", "error", err)
		return
	}
	defer func() { _ = goplsManager.Stop() }()

	// Create and setup MCP server
	server := setupMCPServer(goplsManager)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down server")
		cancel()
		_ = goplsManager.Stop()
		os.Exit(0)
	}()

	logger.Info("starting gopls-mcp server",
		"workspace", *workspacePath,
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
func setupMCPServer(goplsManager *Manager) *mcp.Server {
	// Create MCP server
	server := mcp.NewServer("gopls-mcp", "v0.1.0", nil)

	// Add gopls tools
	server.AddTools(
		goplsManager.CreateGoToDefinitionTool(),
		goplsManager.CreateFindReferencesTool(),
		goplsManager.CreateGetHoverTool(),
	)

	return server
}
