package ynab

import "fmt"

// ListAccounts returns all accounts for a budget
func (c *Client) ListAccounts(budgetID string) ([]Account, error) {
	var resp AccountsResponse
	path := fmt.Sprintf("/budgets/%s/accounts", budgetID)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Data.Accounts, nil
}

// GetAccount returns a single account
func (c *Client) GetAccount(budgetID, accountID string) (*Account, error) {
	accounts, err := c.ListAccounts(budgetID)
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.ID == accountID {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account not found: %s", accountID)
}
