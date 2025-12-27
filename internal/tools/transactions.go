package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/mark3labs/mcp-go/mcp"
)

// NewListTransactionsTool creates the list_transactions tool
func NewListTransactionsTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "list_transactions",
		Description: "List transactions in a budget. Can filter by date (since_date) or type (uncategorized/unapproved). Returns up to most recent transactions.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"since_date": map[string]interface{}{
					"type":        "string",
					"description": "Only return transactions on or after this date (YYYY-MM-DD format). Optional.",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Filter by type: 'uncategorized' or 'unapproved'. Optional.",
					"enum":        []string{"uncategorized", "unapproved"},
				},
				"account_id": map[string]interface{}{
					"type":        "string",
					"description": "Only return transactions for this specific account ID. Optional.",
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

		// Build query
		query := &ynab.TransactionQuery{}
		if sinceDate, ok := args["since_date"].(string); ok && sinceDate != "" {
			query.SinceDate = sinceDate
		}
		if txType, ok := args["type"].(string); ok && txType != "" {
			query.Type = txType
		}

		var transactions []ynab.Transaction
		var err error

		// Check if account_id is specified
		if accountID, ok := args["account_id"].(string); ok && accountID != "" {
			transactions, err = client.ListAccountTransactions(budgetID, accountID, query)
		} else {
			transactions, err = client.ListTransactions(budgetID, query)
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch transactions: %v", err)), nil
		}

		if len(transactions) == 0 {
			return mcp.NewToolResultText("No transactions found."), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Found %d transaction(s):\n\n", len(transactions)))

		// Limit display to most recent 50 transactions
		displayCount := len(transactions)
		if displayCount > 50 {
			displayCount = 50
		}

		totalAmount := int64(0)
		for i := 0; i < displayCount; i++ {
			tx := transactions[i]
			if tx.Deleted {
				continue
			}

			totalAmount += tx.Amount

			// Format cleared status
			clearedSymbol := "⚪" // uncleared
			if tx.Cleared == "cleared" {
				clearedSymbol = "✓"
			} else if tx.Cleared == "reconciled" {
				clearedSymbol = "🔒"
			}

			// Format approval status
			approvalSymbol := ""
			if !tx.Approved {
				approvalSymbol = " [UNAPPROVED]"
			}

			result.WriteString(fmt.Sprintf("%d. %s %s - %s%s\n",
				i+1,
				tx.Date,
				clearedSymbol,
				tx.PayeeName,
				approvalSymbol))
			result.WriteString(fmt.Sprintf("   ID: %s\n", tx.ID))
			result.WriteString(fmt.Sprintf("   Amount: %s\n", ynab.FormatCurrency(tx.Amount)))
			result.WriteString(fmt.Sprintf("   Account: %s\n", tx.AccountName))
			if tx.CategoryName != "" {
				result.WriteString(fmt.Sprintf("   Category: %s\n", tx.CategoryName))
			}
			if tx.Memo != "" {
				result.WriteString(fmt.Sprintf("   Memo: %s\n", tx.Memo))
			}
			result.WriteString("\n")
		}

		if len(transactions) > displayCount {
			result.WriteString(fmt.Sprintf("... and %d more transactions (showing most recent %d)\n\n",
				len(transactions)-displayCount, displayCount))
		}

		result.WriteString(fmt.Sprintf("Total Amount (displayed): %s\n", ynab.FormatCurrency(totalAmount)))

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewGetTransactionTool creates the get_transaction tool
func NewGetTransactionTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "get_transaction_details",
		Description: "Get detailed information about a specific transaction including all fields and any subtransactions (splits).",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"transaction_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the transaction",
				},
			},
			Required: []string{"budget_id", "transaction_id"},
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

		transactionID, ok := args["transaction_id"].(string)
		if !ok || transactionID == "" {
			return mcp.NewToolResultError("transaction_id is required"), nil
		}

		tx, err := client.GetTransaction(budgetID, transactionID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch transaction: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString("Transaction Details\n\n")
		result.WriteString(fmt.Sprintf("Date: %s\n", tx.Date))
		result.WriteString(fmt.Sprintf("Payee: %s\n", tx.PayeeName))
		result.WriteString(fmt.Sprintf("Amount: %s\n", ynab.FormatCurrency(tx.Amount)))
		result.WriteString(fmt.Sprintf("Account: %s\n", tx.AccountName))
		if tx.CategoryName != "" {
			result.WriteString(fmt.Sprintf("Category: %s\n", tx.CategoryName))
		}
		if tx.Memo != "" {
			result.WriteString(fmt.Sprintf("Memo: %s\n", tx.Memo))
		}

		result.WriteString("\nStatus:\n")
		result.WriteString(fmt.Sprintf("  Cleared: %s\n", tx.Cleared))
		result.WriteString(fmt.Sprintf("  Approved: %t\n", tx.Approved))
		if tx.FlagColor != "" {
			result.WriteString(fmt.Sprintf("  Flag: %s\n", tx.FlagColor))
		}

		if len(tx.Subtransactions) > 0 {
			result.WriteString(fmt.Sprintf("\nSplit into %d subtransactions:\n", len(tx.Subtransactions)))
			for i, sub := range tx.Subtransactions {
				result.WriteString(fmt.Sprintf("  %d. %s - %s: %s\n",
					i+1, sub.CategoryName, sub.PayeeName, ynab.FormatCurrency(sub.Amount)))
				if sub.Memo != "" {
					result.WriteString(fmt.Sprintf("     Memo: %s\n", sub.Memo))
				}
			}
		}

		result.WriteString(fmt.Sprintf("\nID: %s\n", tx.ID))

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewCreateTransactionTool creates the create_transaction tool
func NewCreateTransactionTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "create_transaction",
		Description: "Create a new transaction in a budget. Requires account_id, date, and amount. Optionally specify payee, category, and memo.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"account_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the account for this transaction",
				},
				"date": map[string]interface{}{
					"type":        "string",
					"description": "Transaction date in YYYY-MM-DD format (e.g., 2024-01-15)",
				},
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "Transaction amount in currency units (e.g., -45.67 for an expense, 100.00 for income)",
				},
				"payee_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the payee. Optional.",
				},
				"category_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the category for this transaction. Optional.",
				},
				"memo": map[string]interface{}{
					"type":        "string",
					"description": "Memo/note for this transaction. Optional.",
				},
				"cleared": map[string]interface{}{
					"type":        "string",
					"description": "Cleared status: 'cleared', 'uncleared', or 'reconciled'. Default is 'uncleared'.",
					"enum":        []string{"cleared", "uncleared", "reconciled"},
				},
			},
			Required: []string{"budget_id", "account_id", "date", "amount"},
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

		date, ok := args["date"].(string)
		if !ok || date == "" {
			date = time.Now().Format("2006-01-02")
		}

		amount, ok := args["amount"].(float64)
		if !ok {
			return mcp.NewToolResultError("amount is required and must be a number"), nil
		}

		// Create transaction request
		req := &ynab.CreateTransactionRequest{}
		req.Transaction.AccountID = accountID
		req.Transaction.Date = date
		req.Transaction.Amount = ynab.FloatToMilliunits(amount)

		if payeeName, ok := args["payee_name"].(string); ok && payeeName != "" {
			req.Transaction.PayeeName = payeeName
		}

		if categoryID, ok := args["category_id"].(string); ok && categoryID != "" {
			req.Transaction.CategoryID = categoryID
		}

		if memo, ok := args["memo"].(string); ok && memo != "" {
			req.Transaction.Memo = memo
		}

		if cleared, ok := args["cleared"].(string); ok && cleared != "" {
			req.Transaction.Cleared = cleared
		} else {
			req.Transaction.Cleared = "uncleared"
		}

		req.Transaction.Approved = true

		tx, err := client.CreateTransaction(budgetID, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create transaction: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString("Transaction created successfully!\n\n")
		result.WriteString(fmt.Sprintf("ID: %s\n", tx.ID))
		result.WriteString(fmt.Sprintf("Date: %s\n", tx.Date))
		result.WriteString(fmt.Sprintf("Payee: %s\n", tx.PayeeName))
		result.WriteString(fmt.Sprintf("Amount: %s\n", ynab.FormatCurrency(tx.Amount)))
		result.WriteString(fmt.Sprintf("Account: %s\n", tx.AccountName))
		if tx.CategoryName != "" {
			result.WriteString(fmt.Sprintf("Category: %s\n", tx.CategoryName))
		}
		if tx.Memo != "" {
			result.WriteString(fmt.Sprintf("Memo: %s\n", tx.Memo))
		}

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}

// NewUpdateTransactionTool creates the update_transaction tool
func NewUpdateTransactionTool(client *ynab.Client) ToolDefinition {
	tool := mcp.Tool{
		Name:        "update_transaction",
		Description: "Update an existing transaction. Specify the fields you want to change. All fields are optional except budget_id and transaction_id.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"budget_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the budget",
				},
				"transaction_id": map[string]interface{}{
					"type":        "string",
					"description": "The ID of the transaction to update",
				},
				"date": map[string]interface{}{
					"type":        "string",
					"description": "New date in YYYY-MM-DD format. Optional.",
				},
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "New amount in currency units. Optional.",
				},
				"payee_name": map[string]interface{}{
					"type":        "string",
					"description": "New payee name. Optional.",
				},
				"category_id": map[string]interface{}{
					"type":        "string",
					"description": "New category ID. Optional.",
				},
				"memo": map[string]interface{}{
					"type":        "string",
					"description": "New memo. Optional.",
				},
				"cleared": map[string]interface{}{
					"type":        "string",
					"description": "New cleared status: 'cleared', 'uncleared', or 'reconciled'. Optional.",
					"enum":        []string{"cleared", "uncleared", "reconciled"},
				},
			},
			Required: []string{"budget_id", "transaction_id"},
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

		transactionID, ok := args["transaction_id"].(string)
		if !ok || transactionID == "" {
			return mcp.NewToolResultError("transaction_id is required"), nil
		}

		// Build update request with only specified fields
		req := &ynab.UpdateTransactionRequest{}

		if date, ok := args["date"].(string); ok && date != "" {
			req.Transaction.Date = date
		}

		if amount, ok := args["amount"].(float64); ok {
			milliunits := ynab.FloatToMilliunits(amount)
			req.Transaction.Amount = milliunits
		}

		if payeeName, ok := args["payee_name"].(string); ok && payeeName != "" {
			req.Transaction.PayeeName = payeeName
		}

		if categoryID, ok := args["category_id"].(string); ok && categoryID != "" {
			req.Transaction.CategoryID = categoryID
		}

		if memo, ok := args["memo"].(string); ok && memo != "" {
			req.Transaction.Memo = memo
		}

		if cleared, ok := args["cleared"].(string); ok && cleared != "" {
			req.Transaction.Cleared = cleared
		}

		tx, err := client.UpdateTransaction(budgetID, transactionID, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update transaction: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString("Transaction updated successfully!\n\n")
		result.WriteString(fmt.Sprintf("ID: %s\n", tx.ID))
		result.WriteString(fmt.Sprintf("Date: %s\n", tx.Date))
		result.WriteString(fmt.Sprintf("Payee: %s\n", tx.PayeeName))
		result.WriteString(fmt.Sprintf("Amount: %s\n", ynab.FormatCurrency(tx.Amount)))
		result.WriteString(fmt.Sprintf("Account: %s\n", tx.AccountName))
		if tx.CategoryName != "" {
			result.WriteString(fmt.Sprintf("Category: %s\n", tx.CategoryName))
		}
		if tx.Memo != "" {
			result.WriteString(fmt.Sprintf("Memo: %s\n", tx.Memo))
		}
		result.WriteString(fmt.Sprintf("Cleared: %s\n", tx.Cleared))

		return mcp.NewToolResultText(result.String()), nil
	}

	return ToolDefinition{Tool: tool, Handler: handler}
}
