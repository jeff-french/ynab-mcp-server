package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/mcp"
)

// NewListAccountsTool creates the list_accounts tool
func NewListAccountsTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "list_accounts",
		Description: "List all accounts in a budget. Shows account name, type, balance, and status (open/closed, on/off budget).",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
			},
			Required: []string{"budget_id"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		budgetID, ok := args["budget_id"].(string)
		if !ok || budgetID == "" {
			return mcp.NewToolResultError("budget_id is required"), nil
		}

		accounts, err := client.ListAccounts(budgetID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch accounts: %v", err)), nil
		}

		if len(accounts) == 0 {
			return mcp.NewToolResultText("No accounts found."), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Found %d account(s):\n\n", len(accounts)))

		onBudgetTotal := int64(0)
		offBudgetTotal := int64(0)

		for i, account := range accounts {
			if account.Deleted {
				continue
			}

			status := []string{}
			if account.Closed {
				status = append(status, "CLOSED")
			}
			if !account.OnBudget {
				status = append(status, "OFF BUDGET")
				offBudgetTotal += account.Balance
			} else {
				onBudgetTotal += account.Balance
			}

			statusStr := ""
			if len(status) > 0 {
				statusStr = fmt.Sprintf(" [%s]", strings.Join(status, ", "))
			}

			result.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, account.Name, statusStr))
			result.WriteString(fmt.Sprintf("   ID: %s\n", account.ID))
			result.WriteString(fmt.Sprintf("   Type: %s\n", account.Type))
			result.WriteString(fmt.Sprintf("   Balance: %s\n", ynab.FormatCurrency(account.Balance)))
			result.WriteString(fmt.Sprintf("   Cleared: %s\n", ynab.FormatCurrency(account.ClearedBalance)))
			result.WriteString(fmt.Sprintf("   Uncleared: %s\n", ynab.FormatCurrency(account.UnclearedBalance)))
			if account.Note != "" {
				result.WriteString(fmt.Sprintf("   Note: %s\n", account.Note))
			}
			result.WriteString("\n")
		}

		result.WriteString(fmt.Sprintf("On Budget Total: %s\n", ynab.FormatCurrency(onBudgetTotal)))
		result.WriteString(fmt.Sprintf("Off Budget Total: %s\n", ynab.FormatCurrency(offBudgetTotal)))
		result.WriteString(fmt.Sprintf("Net Worth: %s\n", ynab.FormatCurrency(onBudgetTotal+offBudgetTotal)))

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetAccountTool creates the get_account_details tool
func NewGetAccountTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_account_details",
		Description: "Get detailed information about a specific account including balance breakdown and account settings.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"account_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the account",
				},
			},
			Required: []string{"budget_id", "account_id"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		budgetID, ok := args["budget_id"].(string)
		if !ok || budgetID == "" {
			return mcp.NewToolResultError("budget_id is required"), nil
		}

		accountID, ok := args["account_id"].(string)
		if !ok || accountID == "" {
			return mcp.NewToolResultError("account_id is required"), nil
		}

		account, err := client.GetAccount(budgetID, accountID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch account: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Account: %s\n", account.Name))
		result.WriteString(fmt.Sprintf("ID: %s\n", account.ID))
		result.WriteString(fmt.Sprintf("Type: %s\n\n", account.Type))

		result.WriteString("Balances:\n")
		result.WriteString(fmt.Sprintf("  Total: %s\n", ynab.FormatCurrency(account.Balance)))
		result.WriteString(fmt.Sprintf("  Cleared: %s\n", ynab.FormatCurrency(account.ClearedBalance)))
		result.WriteString(fmt.Sprintf("  Uncleared: %s\n\n", ynab.FormatCurrency(account.UnclearedBalance)))

		result.WriteString("Status:\n")
		result.WriteString(fmt.Sprintf("  On Budget: %t\n", account.OnBudget))
		result.WriteString(fmt.Sprintf("  Closed: %t\n", account.Closed))
		result.WriteString(fmt.Sprintf("  Direct Import Linked: %t\n", account.DirectImportLinked))
		if account.DirectImportInError {
			result.WriteString("  ⚠️  Direct Import Error: true\n")
		}

		if account.Note != "" {
			result.WriteString(fmt.Sprintf("\nNote: %s\n", account.Note))
		}

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}
