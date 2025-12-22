package ynab

import "fmt"

// ListCategories returns all category groups and categories for a budget
func (c *Client) ListCategories(budgetID string) ([]CategoryGroup, error) {
	var resp CategoriesResponse
	path := fmt.Sprintf("/budgets/%s/categories", budgetID)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Data.CategoryGroups, nil
}

// GetCategory returns a single category by ID
func (c *Client) GetCategory(budgetID, categoryID string) (*Category, error) {
	categoryGroups, err := c.ListCategories(budgetID)
	if err != nil {
		return nil, err
	}

	// Search for category across all groups
	for _, group := range categoryGroups {
		for _, category := range group.Categories {
			if category.ID == categoryID {
				return &category, nil
			}
		}
	}

	return nil, fmt.Errorf("category not found: %s", categoryID)
}

// GetCategoryByMonth returns category details for a specific month
func (c *Client) GetCategoryByMonth(budgetID, month, categoryID string) (*Category, error) {
	var resp struct {
		Data struct {
			Category Category `json:"category"`
		} `json:"data"`
	}
	path := fmt.Sprintf("/budgets/%s/months/%s/categories/%s", budgetID, month, categoryID)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp.Data.Category, nil
}
