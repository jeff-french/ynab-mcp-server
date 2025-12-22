package ynab

import "fmt"

// ListPayees returns all payees for a budget
func (c *Client) ListPayees(budgetID string) ([]Payee, error) {
	var resp PayeesResponse
	path := fmt.Sprintf("/budgets/%s/payees", budgetID)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Data.Payees, nil
}

// GetPayee returns a single payee by ID
func (c *Client) GetPayee(budgetID, payeeID string) (*Payee, error) {
	payees, err := c.ListPayees(budgetID)
	if err != nil {
		return nil, err
	}

	for _, payee := range payees {
		if payee.ID == payeeID {
			return &payee, nil
		}
	}

	return nil, fmt.Errorf("payee not found: %s", payeeID)
}
