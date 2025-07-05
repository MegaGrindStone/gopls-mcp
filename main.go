package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Get workspace path from environment or use current directory
	workspacePath := os.Getenv("WORKSPACE_PATH")
	if workspacePath == "" {
		var err error
		workspacePath, err = os.Getwd()
		if err != nil {
			log.Fatal("Failed to get current working directory:", err)
		}
	}

	// Create gopls manager
	goplsManager := NewManager(workspacePath)

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
		mcp.NewServerTool[any, string](
			"ping",
			"Simple ping tool to test server connectivity",
			handlePing,
		),
		goplsManager.CreateGoToDefinitionTool(),
		goplsManager.CreateFindReferencesTool(),
		goplsManager.CreateGetHoverTool(),
	)

	// Create SSE handler
	handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return server
	})

	// Set up HTTP server
	http.HandleFunc("/sse", handler.ServeHTTP)

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
	log.Printf("Workspace path: %s", workspacePath)
	log.Printf("SSE endpoint available at: http://localhost:8080/sse")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("Server failed to start: %v", err)
	}
}

func handlePing(_ context.Context, _ *mcp.ServerSession, _ *mcp.CallToolParamsFor[any]) (*mcp.CallToolResultFor[string], error) {
	return &mcp.CallToolResultFor[string]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "pong",
			},
		},
	}, nil
}
