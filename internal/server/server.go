package server

import (
	"github.com/jeff-french/ynab-mcp-server/internal/tools"
	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/server"
)

// NewMCPServer creates and configures the MCP server with all YNAB tools
func NewMCPServer(ynabClient *ynab.Client) (*server.MCPServer, error) {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"ynab-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register all tools with their handlers
	allTools := tools.GetAllTools(ynabClient)
	for _, toolDef := range allTools {
		mcpServer.AddTool(toolDef.Tool, toolDef.Handler)
	}

	return mcpServer, nil
}
