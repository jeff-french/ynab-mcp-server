package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/mcp"
)

// NewGetSpendingByCategoryTool creates the get_spending_by_category aggregation tool
func NewGetSpendingByCategoryTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_spending_by_category",
		Description: "Get total spending per category for a date range. Returns aggregated data without fetching every transaction individually. Useful for understanding spending patterns across categories.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"since_date": map[string]interface{}{
					"type":        "string",
					"description": "Start date in YYYY-MM-DD format (e.g., 2024-01-01)",
				},
				"until_date": map[string]interface{}{
					"type":        "string",
					"description": "End date in YYYY-MM-DD format (e.g., 2024-12-31)",
				},
				"account_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional: filter to specific account ID",
				},
			},
			Required: []string{"budget_id", "since_date", "until_date"},
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

		sinceDate, ok := args["since_date"].(string)
		if !ok || sinceDate == "" {
			return mcp.NewToolResultError("since_date is required (YYYY-MM-DD format)"), nil
		}

		untilDate, ok := args["until_date"].(string)
		if !ok || untilDate == "" {
			return mcp.NewToolResultError("until_date is required (YYYY-MM-DD format)"), nil
		}

		// Validate date range
		if err := validateDateRange(sinceDate, untilDate); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Fetch transactions for date range
		query := &ynab.TransactionQuery{
			SinceDate: sinceDate,
		}

		var transactions []ynab.Transaction
		var err error

		if accountID, ok := args["account_id"].(string); ok && accountID != "" {
			transactions, err = client.ListAccountTransactions(budgetID, accountID, query)
		} else {
			transactions, err = client.ListTransactions(budgetID, query)
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch transactions: %v", err)), nil
		}

		// Filter transactions to until_date (YNAB's since_date doesn't have until)
		untilTime, _ := parseDate(untilDate)
		filteredTxs := make([]ynab.Transaction, 0)
		for _, tx := range transactions {
			txDate, err := parseDate(tx.Date)
			if err != nil {
				continue
			}
			if !txDate.After(untilTime) {
				filteredTxs = append(filteredTxs, tx)
			}
		}

		// Aggregate by category
		summaries := aggregateByCategory(filteredTxs)

		// Convert to sorted slice
		categories := make([]categorySummary, 0, len(summaries))
		totalOutflow := 0.0
		totalInflow := 0.0

		for _, summary := range summaries {
			categories = append(categories, *summary)
			totalOutflow += summary.TotalOutflow
			totalInflow += summary.TotalInflow
		}

		// Sort by total outflow descending
		sort.Slice(categories, func(i, j int) bool {
			return categories[i].TotalOutflow > categories[j].TotalOutflow
		})

		// Build result
		result := map[string]interface{}{
			"categories":    categories,
			"total_outflow": totalOutflow,
			"total_inflow":  totalInflow,
			"date_range": map[string]string{
				"since": sinceDate,
				"until": untilDate,
			},
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetSpendingByMonthTool creates the get_spending_by_month aggregation tool
func NewGetSpendingByMonthTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_spending_by_month",
		Description: "Get monthly spending totals for trend analysis. Returns aggregated spending data for the last N months.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"category_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional: specific category ID to analyze. Omit for all categories.",
				},
				"num_months": map[string]interface{}{
					"type":        "number",
					"description": "Number of months including current (1-24)",
					"minimum":     1,
					"maximum":     24,
				},
				"account_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional: filter to specific account ID",
				},
			},
			Required: []string{"budget_id", "num_months"},
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

		numMonthsFloat, ok := args["num_months"].(float64)
		if !ok {
			return mcp.NewToolResultError("num_months is required (1-24)"), nil
		}
		numMonths := int(numMonthsFloat)
		if numMonths < 1 || numMonths > 24 {
			return mcp.NewToolResultError("num_months must be between 1 and 24"), nil
		}

		// Get category name if filtering by category
		categoryName := "All Categories"
		categoryID := ""
		if catID, ok := args["category_id"].(string); ok && catID != "" {
			categoryID = catID
			// Fetch category details to get name
			category, err := client.GetCategory(budgetID, categoryID)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch category: %v", err)), nil
			}
			categoryName = category.Name
		}

		// Calculate date range for last N months
		months := getLastNMonths(numMonths)
		if len(months) == 0 {
			return mcp.NewToolResultError("Failed to calculate month range"), nil
		}

		// Get since date from oldest month
		sinceDate := months[0] + "-01"

		// Fetch transactions
		query := &ynab.TransactionQuery{
			SinceDate: sinceDate,
		}

		var transactions []ynab.Transaction
		var err error

		if accountID, ok := args["account_id"].(string); ok && accountID != "" {
			transactions, err = client.ListAccountTransactions(budgetID, accountID, query)
		} else {
			transactions, err = client.ListTransactions(budgetID, query)
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch transactions: %v", err)), nil
		}

		// Filter by category if specified
		if categoryID != "" {
			filtered := make([]ynab.Transaction, 0)
			for _, tx := range transactions {
				if tx.CategoryID == categoryID {
					filtered = append(filtered, tx)
				}
			}
			transactions = filtered
		}

		// Aggregate by month
		summaries := aggregateByMonth(transactions, months)

		// Convert to sorted slice (chronological order)
		monthData := make([]monthSummary, len(months))
		totalOutflow := 0.0
		totalInflow := 0.0

		for i, month := range months {
			summary := summaries[month]
			monthData[i] = *summary
			totalOutflow += summary.TotalOutflow
			totalInflow += summary.TotalInflow
		}

		// Calculate averages
		avgOutflow := 0.0
		avgInflow := 0.0
		if numMonths > 0 {
			avgOutflow = totalOutflow / float64(numMonths)
			avgInflow = totalInflow / float64(numMonths)
		}

		// Build result
		result := map[string]interface{}{
			"months":                  monthData,
			"category_name":           categoryName,
			"average_monthly_outflow": avgOutflow,
			"average_monthly_inflow":  avgInflow,
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetBudgetSummaryTool creates the get_budget_summary aggregation tool
func NewGetBudgetSummaryTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_budget_summary",
		Description: "Get current budget state showing budgeted vs actual for all categories. Returns structured budget data for a specific month.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"month": map[string]interface{}{
					"type":        "string",
					"description": "Optional: month in YYYY-MM format. Defaults to current month.",
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

		// Get month (default to current)
		month := getCurrentMonth()
		if monthArg, ok := args["month"].(string); ok && monthArg != "" {
			if _, err := parseMonth(monthArg); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid month format: %v", err)), nil
			}
			month = monthArg
		}

		// Fetch budget with category data
		budget, err := client.GetBudget(budgetID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch budget: %v", err)), nil
		}

		// Build category groups structure
		categoryGroups := make([]map[string]interface{}, 0)

		for _, group := range budget.CategoryGroups {
			if group.Deleted || group.Hidden {
				continue
			}

			categories := make([]map[string]interface{}, 0)
			for _, cat := range group.Categories {
				if cat.Deleted || cat.Hidden {
					continue
				}

				category := map[string]interface{}{
					"category_id":   cat.ID,
					"category_name": cat.Name,
					"budgeted":      ynab.MilliunitsToFloat(cat.Budgeted),
					"activity":      ynab.MilliunitsToFloat(cat.Activity),
					"available":     ynab.MilliunitsToFloat(cat.Balance),
					"goal_target":   nil,
					"goal_type":     nil,
				}

				if cat.GoalTarget > 0 {
					category["goal_target"] = ynab.MilliunitsToFloat(cat.GoalTarget)
				}
				if cat.GoalType != "" {
					category["goal_type"] = cat.GoalType
				}

				categories = append(categories, category)
			}

			if len(categories) > 0 {
				categoryGroup := map[string]interface{}{
					"category_group_id":   group.ID,
					"category_group_name": group.Name,
					"categories":          categories,
				}
				categoryGroups = append(categoryGroups, categoryGroup)
			}
		}

		// Build result
		result := map[string]interface{}{
			"month":           month,
			"category_groups": categoryGroups,
			"age_of_money":    nil, // YNAB doesn't provide this in budget endpoint
			"to_be_budgeted":  nil, // Would need month-specific endpoint
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetPayeeSummaryTool creates the get_payee_summary aggregation tool
func NewGetPayeeSummaryTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_payee_summary",
		Description: "See where money is going by payee. Returns top payees by spending for a date range.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"since_date": map[string]interface{}{
					"type":        "string",
					"description": "Start date in YYYY-MM-DD format",
				},
				"until_date": map[string]interface{}{
					"type":        "string",
					"description": "End date in YYYY-MM-DD format",
				},
				"top_n": map[string]interface{}{
					"type":        "number",
					"description": "Optional: return top N payees (default 20)",
					"default":     20,
				},
			},
			Required: []string{"budget_id", "since_date", "until_date"},
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

		sinceDate, ok := args["since_date"].(string)
		if !ok || sinceDate == "" {
			return mcp.NewToolResultError("since_date is required (YYYY-MM-DD format)"), nil
		}

		untilDate, ok := args["until_date"].(string)
		if !ok || untilDate == "" {
			return mcp.NewToolResultError("until_date is required (YYYY-MM-DD format)"), nil
		}

		// Validate date range
		if err := validateDateRange(sinceDate, untilDate); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		topN := 20
		if topNFloat, ok := args["top_n"].(float64); ok {
			topN = int(topNFloat)
			if topN < 1 {
				topN = 20
			}
		}

		// Fetch transactions
		query := &ynab.TransactionQuery{
			SinceDate: sinceDate,
		}

		transactions, err := client.ListTransactions(budgetID, query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch transactions: %v", err)), nil
		}

		// Filter to until_date
		untilTime, _ := parseDate(untilDate)
		filteredTxs := make([]ynab.Transaction, 0)
		for _, tx := range transactions {
			txDate, err := parseDate(tx.Date)
			if err != nil {
				continue
			}
			if !txDate.After(untilTime) {
				filteredTxs = append(filteredTxs, tx)
			}
		}

		// Aggregate by payee
		summaries := aggregateByPayee(filteredTxs)

		// Convert to sorted slice
		payees := make([]payeeSummary, 0, len(summaries))
		for _, summary := range summaries {
			payees = append(payees, *summary)
		}

		// Sort by total outflow descending
		sort.Slice(payees, func(i, j int) bool {
			return payees[i].TotalOutflow > payees[j].TotalOutflow
		})

		// Limit to top N
		if len(payees) > topN {
			payees = payees[:topN]
		}

		// Build result
		result := map[string]interface{}{
			"payees": payees,
			"date_range": map[string]string{
				"since": sinceDate,
				"until": untilDate,
			},
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetAccountBalancesTool creates the get_account_balances aggregation tool
func NewGetAccountBalancesTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_account_balances",
		Description: "Quick snapshot of all account balances. Returns current balances for all accounts with totals.",
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

		// Fetch accounts
		accounts, err := client.ListAccounts(budgetID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch accounts: %v", err)), nil
		}

		// Build account balances
		accountBalances := make([]accountBalance, 0)
		totalOnBudget := 0.0
		totalOffBudget := 0.0

		for _, account := range accounts {
			if account.Deleted {
				continue
			}

			balance := accountBalance{
				AccountID:        account.ID,
				AccountName:      account.Name,
				AccountType:      account.Type,
				OnBudget:         account.OnBudget,
				Closed:           account.Closed,
				ClearedBalance:   ynab.MilliunitsToFloat(account.ClearedBalance),
				UnclearedBalance: ynab.MilliunitsToFloat(account.UnclearedBalance),
				CurrentBalance:   ynab.MilliunitsToFloat(account.Balance),
			}

			accountBalances = append(accountBalances, balance)

			// Accumulate totals (exclude closed accounts)
			if !account.Closed {
				if account.OnBudget {
					totalOnBudget += balance.CurrentBalance
				} else {
					totalOffBudget += balance.CurrentBalance
				}
			}
		}

		// Build result
		result := map[string]interface{}{
			"accounts":          accountBalances,
			"total_on_budget":   totalOnBudget,
			"total_off_budget":  totalOffBudget,
			"net_worth":         totalOnBudget + totalOffBudget,
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}
