package server

import (
	"github.com/mark3labs/mcp-go/server"
)

// ServeStdio starts the MCP server in stdio mode (reads from stdin, writes to stdout)
// This mode is used for local execution with Claude Desktop
func ServeStdio(mcpServer *server.MCPServer) error {
	return server.ServeStdio(mcpServer)
}
