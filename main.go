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
	workspaceFlag := flag.String("workspace", "", "Comma-separated list of Go workspace directories (required)")
	transportType := flag.String("transport", "http", "Transport type: http or stdio")
	flag.Parse()

	// Validate that workspace path is provided
	if *workspaceFlag == "" {
		logger.Error("workspace flag is required")
		os.Exit(1)
	}

	// Parse and validate workspace paths
	workspacePaths := parseAndValidateWorkspaces(*workspaceFlag, logger)

	// Validate transport type
	if *transportType != transportHTTP && *transportType != transportStdio {
		logger.Error("transport must be either 'http' or 'stdio'", "provided", *transportType)
		os.Exit(1)
	}

	// Create gopls clients for each workspace
	goplsClients := make(map[string]*goplsClient)
	for _, workspacePath := range workspacePaths {
		goplsClients[workspacePath] = newClient(workspacePath, logger)
	}

	// Start all gopls clients
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for workspacePath, client := range goplsClients {
		if err := client.start(ctx); err != nil {
			logger.Error("failed to start gopls", "workspace", workspacePath, "error", err)
			return
		}
	}
	defer func() {
		for _, client := range goplsClients {
			_ = client.stop()
		}
	}()

	// Create and setup MCP server
	server := setupMCPServer(goplsClients)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down server")
		cancel()
		for _, client := range goplsClients {
			_ = client.stop()
		}
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

// parseAndValidateWorkspaces parses comma-separated workspace paths and validates they exist.
func parseAndValidateWorkspaces(workspaceFlag string, logger *slog.Logger) []string {
	// Parse workspace paths
	workspacePaths := strings.Split(workspaceFlag, ",")
	for i, path := range workspacePaths {
		workspacePaths[i] = strings.TrimSpace(path)
	}

	// Validate all workspace paths exist
	for _, workspacePath := range workspacePaths {
		if workspacePath == "" {
			logger.Error("empty workspace path found")
			os.Exit(1)
		}
		info, err := os.Stat(workspacePath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Error("workspace path does not exist", "path", workspacePath)
			} else {
				logger.Error("failed to access workspace path", "path", workspacePath, "error", err)
			}
			os.Exit(1)
		}
		if !info.IsDir() {
			logger.Error("workspace path is not a directory", "path", workspacePath)
			os.Exit(1)
		}
	}

	return workspacePaths
}
