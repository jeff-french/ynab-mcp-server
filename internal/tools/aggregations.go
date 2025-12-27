package tools

import (
	"fmt"
	"time"

	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
)

// Helper functions for aggregation tools

// isTransfer checks if a transaction is a transfer (should be excluded from spending analysis)
func isTransfer(tx ynab.Transaction) bool {
	return tx.TransferAccountID != ""
}

// parseDate validates and parses a date string in YYYY-MM-DD format
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date string is empty")
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return t, nil
}

// validateDateRange checks that dates are valid and range is reasonable
func validateDateRange(sinceDate, untilDate string) error {
	since, err := parseDate(sinceDate)
	if err != nil {
		return fmt.Errorf("since_date: %w", err)
	}

	until, err := parseDate(untilDate)
	if err != nil {
		return fmt.Errorf("until_date: %w", err)
	}

	if until.Before(since) {
		return fmt.Errorf("until_date must be after since_date")
	}

	// Check if range is too large (> 2 years)
	duration := until.Sub(since)
	if duration > 730*24*time.Hour { // ~2 years
		return fmt.Errorf("date range too large (max 2 years), consider using a smaller range for better performance")
	}

	return nil
}

// getMonthString returns YYYY-MM format for a time
func getMonthString(t time.Time) string {
	return t.Format("2006-01")
}

// parseMonth validates and parses a month string in YYYY-MM format
func parseMonth(monthStr string) (time.Time, error) {
	if monthStr == "" {
		return time.Time{}, fmt.Errorf("month string is empty")
	}
	t, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month format, expected YYYY-MM: %w", err)
	}
	return t, nil
}

// getCurrentMonth returns current month in YYYY-MM format
func getCurrentMonth() string {
	return time.Now().Format("2006-01")
}

// getLastNMonths returns a list of month strings for the last N months including current
func getLastNMonths(n int) []string {
	if n < 1 {
		n = 1
	}
	if n > 24 {
		n = 24
	}

	months := make([]string, n)
	now := time.Now()

	for i := n - 1; i >= 0; i-- {
		monthDate := now.AddDate(0, -i, 0)
		months[n-1-i] = getMonthString(monthDate)
	}

	return months
}

// categorySummary holds aggregated data for a category
type categorySummary struct {
	CategoryID        string  `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	CategoryGroupName string  `json:"category_group_name"`
	TotalOutflow      float64 `json:"total_outflow"`
	TotalInflow       float64 `json:"total_inflow"`
	Net               float64 `json:"net"`
	TransactionCount  int     `json:"transaction_count"`
}

// monthSummary holds aggregated data for a month
type monthSummary struct {
	Month            string  `json:"month"`
	TotalOutflow     float64 `json:"total_outflow"`
	TotalInflow      float64 `json:"total_inflow"`
	Net              float64 `json:"net"`
	TransactionCount int     `json:"transaction_count"`
}

// payeeSummary holds aggregated data for a payee
type payeeSummary struct {
	PayeeID          string  `json:"payee_id"`
	PayeeName        string  `json:"payee_name"`
	TotalOutflow     float64 `json:"total_outflow"`
	TotalInflow      float64 `json:"total_inflow"`
	Net              float64 `json:"net"`
	TransactionCount int     `json:"transaction_count"`
}

// accountBalance holds account balance information
type accountBalance struct {
	AccountID        string  `json:"account_id"`
	AccountName      string  `json:"account_name"`
	AccountType      string  `json:"account_type"`
	OnBudget         bool    `json:"on_budget"`
	Closed           bool    `json:"closed"`
	ClearedBalance   float64 `json:"cleared_balance"`
	UnclearedBalance float64 `json:"uncleared_balance"`
	CurrentBalance   float64 `json:"current_balance"`
}

// aggregateByCategory groups transactions by category and sums amounts
func aggregateByCategory(transactions []ynab.Transaction) map[string]*categorySummary {
	summaries := make(map[string]*categorySummary)

	for _, tx := range transactions {
		// Skip transfers
		if isTransfer(tx) {
			continue
		}

		// Skip deleted transactions
		if tx.Deleted {
			continue
		}

		// Get or create category summary
		categoryID := tx.CategoryID
		if categoryID == "" {
			categoryID = "uncategorized"
		}

		summary, exists := summaries[categoryID]
		if !exists {
			summary = &categorySummary{
				CategoryID:        categoryID,
				CategoryName:      tx.CategoryName,
				CategoryGroupName: "", // Will be filled in if we have category data
			}
			if categoryID == "uncategorized" {
				summary.CategoryName = "Uncategorized"
			}
			summaries[categoryID] = summary
		}

		// Aggregate amounts
		amount := ynab.MilliunitsToFloat(tx.Amount)
		if amount < 0 {
			summary.TotalOutflow += -amount // Store as positive
		} else {
			summary.TotalInflow += amount
		}
		summary.Net += amount
		summary.TransactionCount++
	}

	return summaries
}

// aggregateByMonth groups transactions by month and sums amounts
func aggregateByMonth(transactions []ynab.Transaction, months []string) map[string]*monthSummary {
	summaries := make(map[string]*monthSummary)

	// Initialize all months with zero values
	for _, month := range months {
		summaries[month] = &monthSummary{
			Month: month,
		}
	}

	// Aggregate transactions
	for _, tx := range transactions {
		// Skip transfers
		if isTransfer(tx) {
			continue
		}

		// Skip deleted transactions
		if tx.Deleted {
			continue
		}

		// Get month from transaction date
		txDate, err := parseDate(tx.Date)
		if err != nil {
			continue // Skip invalid dates
		}
		month := getMonthString(txDate)

		// Only include if in our month list
		summary, exists := summaries[month]
		if !exists {
			continue
		}

		// Aggregate amounts
		amount := ynab.MilliunitsToFloat(tx.Amount)
		if amount < 0 {
			summary.TotalOutflow += -amount // Store as positive
		} else {
			summary.TotalInflow += amount
		}
		summary.Net += amount
		summary.TransactionCount++
	}

	return summaries
}

// aggregateByPayee groups transactions by payee and sums amounts
func aggregateByPayee(transactions []ynab.Transaction) map[string]*payeeSummary {
	summaries := make(map[string]*payeeSummary)

	for _, tx := range transactions {
		// Skip transfers
		if isTransfer(tx) {
			continue
		}

		// Skip deleted transactions
		if tx.Deleted {
			continue
		}

		// Get or create payee summary
		payeeID := tx.PayeeID
		if payeeID == "" {
			payeeID = "no-payee"
		}

		summary, exists := summaries[payeeID]
		if !exists {
			summary = &payeeSummary{
				PayeeID:   payeeID,
				PayeeName: tx.PayeeName,
			}
			if payeeID == "no-payee" {
				summary.PayeeName = "No Payee"
			}
			summaries[payeeID] = summary
		}

		// Aggregate amounts
		amount := ynab.MilliunitsToFloat(tx.Amount)
		if amount < 0 {
			summary.TotalOutflow += -amount // Store as positive
		} else {
			summary.TotalInflow += amount
		}
		summary.Net += amount
		summary.TransactionCount++
	}

	return summaries
}
