package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "ynab-mcp-server",
	Short: "YNAB MCP Server with stdio and HTTP transport support",
	Long: `A Model Context Protocol (MCP) server for YNAB (You Need A Budget).

Supports both stdio transport (for local use with Claude Desktop) and HTTP transport
(for remote deployment). Provides MCP tools for accessing YNAB budgets, accounts,
transactions, categories, and payees.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
