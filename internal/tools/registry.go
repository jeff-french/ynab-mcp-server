package tools

import (
	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolDefinition represents a tool and its handler
type ToolDefinition struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

// GetAllTools returns all available YNAB MCP tools
func GetAllTools(client *ynab.Client) []ToolDefinition {
	return []ToolDefinition{
		// Budget tools
		NewListBudgetsTool(client),
		NewGetBudgetTool(client),

		// Account tools
		NewListAccountsTool(client),
		NewGetAccountTool(client),

		// Transaction tools
		NewListTransactionsTool(client),
		NewGetTransactionTool(client),
		NewCreateTransactionTool(client),
		NewUpdateTransactionTool(client),

		// Category tools
		NewListCategoriesTool(client),
		NewGetCategoryTool(client),

		// Payee tools
		NewListPayeesTool(client),
	}
}
