package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/mcp"
)

// NewListPayeesTool creates the list_payees tool
func NewListPayeesTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "list_payees",
		Description: "List all payees in a budget. Payees are the people or entities you pay money to or receive money from.",
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

		payees, err := client.ListPayees(budgetID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch payees: %v", err)), nil
		}

		if len(payees) == 0 {
			return mcp.NewToolResultText("No payees found."), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Found %d payee(s):\n\n", len(payees)))

		transferPayees := []ynab.Payee{}
		regularPayees := []ynab.Payee{}

		// Separate transfer payees from regular payees
		for _, payee := range payees {
			if payee.Deleted {
				continue
			}

			if payee.TransferAccountID != "" {
				transferPayees = append(transferPayees, payee)
			} else {
				regularPayees = append(regularPayees, payee)
			}
		}

		// Display regular payees
		if len(regularPayees) > 0 {
			result.WriteString("Payees:\n")
			for i, payee := range regularPayees {
				result.WriteString(fmt.Sprintf("%d. %s\n", i+1, payee.Name))
				result.WriteString(fmt.Sprintf("   ID: %s\n", payee.ID))
			}
			result.WriteString("\n")
		}

		// Display transfer payees
		if len(transferPayees) > 0 {
			result.WriteString("Transfer Payees (Account Transfers):\n")
			for i, payee := range transferPayees {
				result.WriteString(fmt.Sprintf("%d. %s\n", i+1, payee.Name))
				result.WriteString(fmt.Sprintf("   ID: %s\n", payee.ID))
				result.WriteString(fmt.Sprintf("   Transfer Account: %s\n", payee.TransferAccountID))
			}
		}

		result.WriteString(fmt.Sprintf("\nTotal: %d regular payees, %d transfer payees\n",
			len(regularPayees), len(transferPayees)))

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}
