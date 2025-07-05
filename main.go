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

func main() {
	// Parse command line flags
	workspacePath := flag.String("workspace", "", "Path to the Go workspace directory (required)")
	flag.Parse()

	// Validate that workspace path is provided
	if *workspacePath == "" {
		log.Fatal("Error: -workspace flag is required")
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

	// Create MCP server
	server := mcp.NewServer("gopls-mcp", "v0.1.0", nil)

	// Add gopls tools
	server.AddTools(
		goplsManager.CreateGoToDefinitionTool(),
		goplsManager.CreateFindReferencesTool(),
		goplsManager.CreateGetHoverTool(),
	)

	// Create SSE handler
	handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return server
	})

	// Set up HTTP server with mux
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", handler.ServeHTTP)

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

	log.Printf("Starting gopls-mcp server on :8080")
	log.Printf("Workspace path: %s", *workspacePath)
	log.Printf("SSE endpoint available at: http://localhost:8080/sse")

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		log.Printf("Server failed to start: %v", err)
	}
}
