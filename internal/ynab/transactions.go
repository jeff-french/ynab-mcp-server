package ynab

import (
	"fmt"
	"net/url"
)

// TransactionQuery holds parameters for querying transactions
type TransactionQuery struct {
	SinceDate string // YYYY-MM-DD format
	Type      string // uncategorized, unapproved
}

// ListTransactions returns all transactions for a budget
func (c *Client) ListTransactions(budgetID string, query *TransactionQuery) ([]Transaction, error) {
	path := fmt.Sprintf("/budgets/%s/transactions", budgetID)

	// Add query parameters if provided
	if query != nil {
		params := url.Values{}
		if query.SinceDate != "" {
			params.Add("since_date", query.SinceDate)
		}
		if query.Type != "" {
			params.Add("type", query.Type)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp TransactionsResponse
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Data.Transactions, nil
}

// GetTransaction returns a single transaction
func (c *Client) GetTransaction(budgetID, transactionID string) (*Transaction, error) {
	var resp TransactionResponse
	path := fmt.Sprintf("/budgets/%s/transactions/%s", budgetID, transactionID)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp.Data.Transaction, nil
}

// CreateTransactionRequest represents a request to create a transaction
type CreateTransactionRequest struct {
	Transaction struct {
		AccountID  string `json:"account_id"`
		Date       string `json:"date"` // YYYY-MM-DD
		Amount     int64  `json:"amount"` // in milliunits
		PayeeID    string `json:"payee_id,omitempty"`
		PayeeName  string `json:"payee_name,omitempty"`
		CategoryID string `json:"category_id,omitempty"`
		Memo       string `json:"memo,omitempty"`
		Cleared    string `json:"cleared,omitempty"` // cleared, uncleared, reconciled
		Approved   bool   `json:"approved,omitempty"`
		FlagColor  string `json:"flag_color,omitempty"` // red, orange, yellow, green, blue, purple
	} `json:"transaction"`
}

// CreateTransaction creates a new transaction
func (c *Client) CreateTransaction(budgetID string, req *CreateTransactionRequest) (*Transaction, error) {
	var resp TransactionResponse
	path := fmt.Sprintf("/budgets/%s/transactions", budgetID)
	if err := c.post(path, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data.Transaction, nil
}

// UpdateTransactionRequest represents a request to update a transaction
type UpdateTransactionRequest struct {
	Transaction struct {
		AccountID  string `json:"account_id,omitempty"`
		Date       string `json:"date,omitempty"`
		Amount     int64  `json:"amount,omitempty"`
		PayeeID    string `json:"payee_id,omitempty"`
		PayeeName  string `json:"payee_name,omitempty"`
		CategoryID string `json:"category_id,omitempty"`
		Memo       string `json:"memo,omitempty"`
		Cleared    string `json:"cleared,omitempty"`
		Approved   *bool  `json:"approved,omitempty"`
		FlagColor  string `json:"flag_color,omitempty"`
	} `json:"transaction"`
}

// UpdateTransaction updates an existing transaction
func (c *Client) UpdateTransaction(budgetID, transactionID string, req *UpdateTransactionRequest) (*Transaction, error) {
	var resp TransactionResponse
	path := fmt.Sprintf("/budgets/%s/transactions/%s", budgetID, transactionID)
	if err := c.put(path, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data.Transaction, nil
}

// ListAccountTransactions returns all transactions for a specific account
func (c *Client) ListAccountTransactions(budgetID, accountID string, query *TransactionQuery) ([]Transaction, error) {
	path := fmt.Sprintf("/budgets/%s/accounts/%s/transactions", budgetID, accountID)

	// Add query parameters if provided
	if query != nil {
		params := url.Values{}
		if query.SinceDate != "" {
			params.Add("since_date", query.SinceDate)
		}
		if query.Type != "" {
			params.Add("type", query.Type)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var resp TransactionsResponse
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Data.Transactions, nil
}
