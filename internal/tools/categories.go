package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/mcp"
)

// NewListCategoriesTool creates the list_categories tool
func NewListCategoriesTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "list_categories",
		Description: "List all category groups and their categories in a budget. Shows budgeted amounts, activity, and balances for each category.",
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

		categoryGroups, err := client.ListCategories(budgetID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch categories: %v", err)), nil
		}

		if len(categoryGroups) == 0 {
			return mcp.NewToolResultText("No category groups found."), nil
		}

		var result strings.Builder
		result.WriteString("Category Groups and Categories:\n\n")

		totalBudgeted := int64(0)
		totalActivity := int64(0)
		totalBalance := int64(0)

		for _, group := range categoryGroups {
			if group.Deleted || group.Hidden {
				continue
			}

			result.WriteString(fmt.Sprintf("📁 %s\n", group.Name))

			activeCategories := 0
			for _, category := range group.Categories {
				if category.Deleted || category.Hidden {
					continue
				}
				activeCategories++

				totalBudgeted += category.Budgeted
				totalActivity += category.Activity
				totalBalance += category.Balance

				// Determine if category is overspent
				overspent := ""
				if category.Balance < 0 {
					overspent = " ⚠️ OVERSPENT"
				}

				result.WriteString(fmt.Sprintf("  - %s%s\n", category.Name, overspent))
				result.WriteString(fmt.Sprintf("    ID: %s\n", category.ID))
				result.WriteString(fmt.Sprintf("    Budgeted: %s | Activity: %s | Available: %s\n",
					ynab.FormatCurrency(category.Budgeted),
					ynab.FormatCurrency(category.Activity),
					ynab.FormatCurrency(category.Balance)))

				// Show goal information if present
				if category.GoalType != "" {
					result.WriteString(fmt.Sprintf("    Goal: %s", category.GoalType))
					if category.GoalTarget > 0 {
						result.WriteString(fmt.Sprintf(" - Target: %s", ynab.FormatCurrency(category.GoalTarget)))
					}
					if category.GoalPercentageComplete > 0 {
						result.WriteString(fmt.Sprintf(" (%d%% complete)", category.GoalPercentageComplete))
					}
					result.WriteString("\n")
				}
			}

			if activeCategories == 0 {
				result.WriteString("  (no active categories)\n")
			}
			result.WriteString("\n")
		}

		result.WriteString("Summary:\n")
		result.WriteString(fmt.Sprintf("  Total Budgeted: %s\n", ynab.FormatCurrency(totalBudgeted)))
		result.WriteString(fmt.Sprintf("  Total Activity: %s\n", ynab.FormatCurrency(totalActivity)))
		result.WriteString(fmt.Sprintf("  Total Available: %s\n", ynab.FormatCurrency(totalBalance)))

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetCategoryTool creates the get_category_details tool
func NewGetCategoryTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_category_details",
		Description: "Get detailed information about a specific category including budget, activity, balance, and goal information.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"category_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the category",
				},
			},
			Required: []string{"budget_id", "category_id"},
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

		categoryID, ok := args["category_id"].(string)
		if !ok || categoryID == "" {
			return mcp.NewToolResultError("category_id is required"), nil
		}

		category, err := client.GetCategory(budgetID, categoryID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch category: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Category: %s\n", category.Name))
		result.WriteString(fmt.Sprintf("ID: %s\n", category.ID))
		result.WriteString(fmt.Sprintf("Group: %s\n\n", category.CategoryGroupName))

		result.WriteString("Budget Information:\n")
		result.WriteString(fmt.Sprintf("  Budgeted: %s\n", ynab.FormatCurrency(category.Budgeted)))
		result.WriteString(fmt.Sprintf("  Activity: %s\n", ynab.FormatCurrency(category.Activity)))
		result.WriteString(fmt.Sprintf("  Available: %s\n\n", ynab.FormatCurrency(category.Balance)))

		if category.Balance < 0 {
			result.WriteString("⚠️  This category is overspent!\n\n")
		}

		// Goal information
		if category.GoalType != "" {
			result.WriteString("Goal Information:\n")
			result.WriteString(fmt.Sprintf("  Type: %s\n", category.GoalType))

			if category.GoalTarget > 0 {
				result.WriteString(fmt.Sprintf("  Target: %s\n", ynab.FormatCurrency(category.GoalTarget)))
			}

			if category.GoalTargetMonth != "" {
				result.WriteString(fmt.Sprintf("  Target Month: %s\n", category.GoalTargetMonth))
			}

			if category.GoalPercentageComplete > 0 {
				result.WriteString(fmt.Sprintf("  Progress: %d%% complete\n", category.GoalPercentageComplete))
			}

			if category.GoalUnderFunded > 0 {
				result.WriteString(fmt.Sprintf("  Under Funded: %s\n", ynab.FormatCurrency(category.GoalUnderFunded)))
			}

			result.WriteString("\n")
		}

		if category.Note != "" {
			result.WriteString(fmt.Sprintf("Note: %s\n", category.Note))
		}

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}
