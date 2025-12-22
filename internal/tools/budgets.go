package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/mcp"
)

// NewListBudgetsTool creates the list_budgets tool
func NewListBudgetsTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "list_budgets",
		Description: "List all YNAB budgets accessible with the current token. Returns budget ID, name, and last modified date for each budget.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		budgets, err := client.ListBudgets()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch budgets: %v", err)), nil
		}

		if len(budgets) == 0 {
			return mcp.NewToolResultText("No budgets found."), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Found %d budget(s):\n\n", len(budgets)))

		for i, budget := range budgets {
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, budget.Name))
			result.WriteString(fmt.Sprintf("   ID: %s\n", budget.ID))
			result.WriteString(fmt.Sprintf("   Last Modified: %s\n", budget.LastModifiedOn))
			if budget.CurrencyFormat != nil {
				result.WriteString(fmt.Sprintf("   Currency: %s\n", budget.CurrencyFormat.ISOCode))
			}
			result.WriteString("\n")
		}

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetBudgetTool creates the get_budget_details tool
func NewGetBudgetTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_budget_details",
		Description: "Get detailed information about a specific budget including accounts, categories, and payees. Requires a budget ID from list_budgets.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget to retrieve",
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

		budget, err := client.GetBudget(budgetID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch budget: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Budget: %s\n", budget.Name))
		result.WriteString(fmt.Sprintf("ID: %s\n", budget.ID))
		result.WriteString(fmt.Sprintf("Last Modified: %s\n", budget.LastModifiedOn))
		result.WriteString(fmt.Sprintf("First Month: %s\n", budget.FirstMonth))
		result.WriteString(fmt.Sprintf("Last Month: %s\n\n", budget.LastMonth))

		if budget.CurrencyFormat != nil {
			result.WriteString(fmt.Sprintf("Currency: %s (%s)\n\n",
				budget.CurrencyFormat.ISOCode,
				budget.CurrencyFormat.CurrencySymbol))
		}

		// Accounts summary
		if len(budget.Accounts) > 0 {
			result.WriteString(fmt.Sprintf("Accounts (%d):\n", len(budget.Accounts)))
			onBudgetBalance := int64(0)
			offBudgetBalance := int64(0)

			for _, account := range budget.Accounts {
				if account.Deleted || account.Closed {
					continue
				}
				if account.OnBudget {
					onBudgetBalance += account.Balance
				} else {
					offBudgetBalance += account.Balance
				}
				status := ""
				if account.Closed {
					status = " [CLOSED]"
				}
				result.WriteString(fmt.Sprintf("  - %s: %s%s\n",
					account.Name,
					ynab.FormatCurrency(account.Balance),
					status))
			}
			result.WriteString(fmt.Sprintf("\nOn Budget Total: %s\n", ynab.FormatCurrency(onBudgetBalance)))
			result.WriteString(fmt.Sprintf("Off Budget Total: %s\n\n", ynab.FormatCurrency(offBudgetBalance)))
		}

		// Category groups summary
		if len(budget.CategoryGroups) > 0 {
			result.WriteString(fmt.Sprintf("Category Groups (%d):\n", len(budget.CategoryGroups)))
			for _, group := range budget.CategoryGroups {
				if group.Deleted || group.Hidden {
					continue
				}
				result.WriteString(fmt.Sprintf("  - %s (%d categories)\n", group.Name, len(group.Categories)))
			}
			result.WriteString("\n")
		}

		// Payees count
		if len(budget.Payees) > 0 {
			activePayees := 0
			for _, payee := range budget.Payees {
				if !payee.Deleted {
					activePayees++
				}
			}
			result.WriteString(fmt.Sprintf("Total Payees: %d\n", activePayees))
		}

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}
