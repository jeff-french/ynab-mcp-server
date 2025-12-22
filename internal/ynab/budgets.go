package ynab

import "fmt"

// ListBudgets returns all budgets
func (c *Client) ListBudgets() ([]Budget, error) {
	var resp BudgetSummaryResponse
	if err := c.get("/budgets", &resp); err != nil {
		return nil, err
	}
	return resp.Data.Budgets, nil
}

// GetBudget returns a single budget with all related entities
func (c *Client) GetBudget(budgetID string) (*Budget, error) {
	var resp BudgetDetailResponse
	path := fmt.Sprintf("/budgets/%s", budgetID)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp.Data.Budget, nil
}

// GetBudgetSettings returns budget settings (summary without all entities)
func (c *Client) GetBudgetSettings(budgetID string) (*Budget, error) {
	var resp BudgetDetailResponse
	path := fmt.Sprintf("/budgets/%s/settings", budgetID)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp.Data.Budget, nil
}
