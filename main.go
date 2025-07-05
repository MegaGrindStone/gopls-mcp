package main

import (
	"context"
	"flag"
	"log"
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
	// Parse command line flags
	workspacePath := flag.String("workspace", "", "Path to the Go workspace directory (required)")
	transportType := flag.String("transport", "http", "Transport type: http or stdio")
	flag.Parse()

	// Validate that workspace path is provided
	if *workspacePath == "" {
		log.Fatal("Error: -workspace flag is required")
	}

	// Validate transport type
	if *transportType != transportHTTP && *transportType != transportStdio {
		log.Fatal("Error: -transport must be either 'http' or 'stdio'")
	}

	// Create gopls manager
	goplsManager := NewManager(*workspacePath)

	// Start gopls
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := goplsManager.Start(ctx); err != nil {
		log.Printf("Failed to start gopls: %v", err)
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
		log.Println("Shutting down server...")
		cancel()
		_ = goplsManager.Stop()
		os.Exit(0)
	}()

	log.Printf("Starting gopls-mcp server")
	log.Printf("Workspace path: %s", *workspacePath)
	log.Printf("Transport: %s", *transportType)

	// Start server based on transport type
	if *transportType == transportStdio {
		log.Println("Using stdio transport")
		if err := server.Run(ctx, mcp.NewStdioTransport()); err != nil {
			log.Printf("Stdio server failed: %v", err)
		}
	} else {
		// HTTP transport
		handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
			return server
		}, nil)

		log.Printf("HTTP server available at: http://localhost:8080")

		httpServer := &http.Server{
			Addr:         ":8080",
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		if err := httpServer.ListenAndServe(); err != nil {
			log.Printf("HTTP server failed to start: %v", err)
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
