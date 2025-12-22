package ynab

import "fmt"

// APIErrorResponse represents an error response from the YNAB API
type APIErrorResponse struct {
	Error struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Detail string `json:"detail"`
	} `json:"error"`
}

// Budget represents a YNAB budget
type Budget struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	LastModifiedOn       string          `json:"last_modified_on"`
	FirstMonth           string          `json:"first_month"`
	LastMonth            string          `json:"last_month"`
	DateFormat           *DateFormat     `json:"date_format"`
	CurrencyFormat       *CurrencyFormat `json:"currency_format"`
	Accounts             []Account       `json:"accounts,omitempty"`
	Categories           []Category      `json:"categories,omitempty"`
	CategoryGroups       []CategoryGroup `json:"category_groups,omitempty"`
	Payees               []Payee         `json:"payees,omitempty"`
	Months               []Month         `json:"months,omitempty"`
	Transactions         []Transaction   `json:"transactions,omitempty"`
	ScheduledTransactions []ScheduledTransaction `json:"scheduled_transactions,omitempty"`
}

// DateFormat represents budget date format settings
type DateFormat struct {
	Format string `json:"format"`
}

// CurrencyFormat represents budget currency format settings
type CurrencyFormat struct {
	ISOCode          string `json:"iso_code"`
	ExampleFormat    string `json:"example_format"`
	DecimalDigits    int    `json:"decimal_digits"`
	DecimalSeparator string `json:"decimal_separator"`
	SymbolFirst      bool   `json:"symbol_first"`
	GroupSeparator   string `json:"group_separator"`
	CurrencySymbol   string `json:"currency_symbol"`
	DisplaySymbol    bool   `json:"display_symbol"`
}

// Account represents a YNAB account
type Account struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"` // checking, savings, creditCard, etc.
	OnBudget            bool   `json:"on_budget"`
	Closed              bool   `json:"closed"`
	Note                string `json:"note"`
	Balance             int64  `json:"balance"` // in milliunits
	ClearedBalance      int64  `json:"cleared_balance"`
	UnclearedBalance    int64  `json:"uncleared_balance"`
	TransferPayeeID     string `json:"transfer_payee_id"`
	DirectImportLinked  bool   `json:"direct_import_linked"`
	DirectImportInError bool   `json:"direct_import_in_error"`
	Deleted             bool   `json:"deleted"`
}

// Transaction represents a YNAB transaction
type Transaction struct {
	ID                  string              `json:"id"`
	Date                string              `json:"date"`
	Amount              int64               `json:"amount"` // in milliunits
	Memo                string              `json:"memo"`
	Cleared             string              `json:"cleared"` // cleared, uncleared, reconciled
	Approved            bool                `json:"approved"`
	FlagColor           string              `json:"flag_color"`
	FlagName            string              `json:"flag_name"`
	AccountID           string              `json:"account_id"`
	AccountName         string              `json:"account_name"`
	PayeeID             string              `json:"payee_id"`
	PayeeName           string              `json:"payee_name"`
	CategoryID          string              `json:"category_id"`
	CategoryName        string              `json:"category_name"`
	TransferAccountID   string              `json:"transfer_account_id"`
	TransferTransactionID string            `json:"transfer_transaction_id"`
	MatchedTransactionID string             `json:"matched_transaction_id"`
	ImportID            string              `json:"import_id"`
	ImportPayeeName     string              `json:"import_payee_name"`
	ImportPayeeNameOriginal string          `json:"import_payee_name_original"`
	DebtTransactionType string              `json:"debt_transaction_type"`
	Deleted             bool                `json:"deleted"`
	Subtransactions     []SubTransaction    `json:"subtransactions,omitempty"`
}

// SubTransaction represents a split transaction
type SubTransaction struct {
	ID                 string `json:"id"`
	TransactionID      string `json:"transaction_id"`
	Amount             int64  `json:"amount"` // in milliunits
	Memo               string `json:"memo"`
	PayeeID            string `json:"payee_id"`
	PayeeName          string `json:"payee_name"`
	CategoryID         string `json:"category_id"`
	CategoryName       string `json:"category_name"`
	TransferAccountID  string `json:"transfer_account_id"`
	TransferTransactionID string `json:"transfer_transaction_id"`
	Deleted            bool   `json:"deleted"`
}

// Category represents a budget category
type Category struct {
	ID                      string `json:"id"`
	CategoryGroupID         string `json:"category_group_id"`
	CategoryGroupName       string `json:"category_group_name"`
	Name                    string `json:"name"`
	Hidden                  bool   `json:"hidden"`
	OriginalCategoryGroupID string `json:"original_category_group_id"`
	Note                    string `json:"note"`
	Budgeted                int64  `json:"budgeted"` // in milliunits
	Activity                int64  `json:"activity"`
	Balance                 int64  `json:"balance"`
	GoalType                string `json:"goal_type"`
	GoalDay                 int    `json:"goal_day"`
	GoalCadence             int    `json:"goal_cadence"`
	GoalCadenceFrequency    int    `json:"goal_cadence_frequency"`
	GoalCreationMonth       string `json:"goal_creation_month"`
	GoalTarget              int64  `json:"goal_target"`
	GoalTargetMonth         string `json:"goal_target_month"`
	GoalPercentageComplete  int    `json:"goal_percentage_complete"`
	GoalMonthsToBudget      int    `json:"goal_months_to_budget"`
	GoalUnderFunded         int64  `json:"goal_under_funded"`
	GoalOverallFunded       int64  `json:"goal_overall_funded"`
	GoalOverallLeft         int64  `json:"goal_overall_left"`
	Deleted                 bool   `json:"deleted"`
}

// CategoryGroup represents a group of categories
type CategoryGroup struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Hidden  bool       `json:"hidden"`
	Deleted bool       `json:"deleted"`
	Categories []Category `json:"categories,omitempty"`
}

// Payee represents a payee
type Payee struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	TransferAccountID string `json:"transfer_account_id"`
	Deleted           bool   `json:"deleted"`
}

// Month represents a budget month
type Month struct {
	Month      string     `json:"month"`
	Note       string     `json:"note"`
	Income     int64      `json:"income"` // in milliunits
	Budgeted   int64      `json:"budgeted"`
	Activity   int64      `json:"activity"`
	ToBeBudgeted int64    `json:"to_be_budgeted"`
	AgeOfMoney int        `json:"age_of_money"`
	Deleted    bool       `json:"deleted"`
	Categories []Category `json:"categories,omitempty"`
}

// ScheduledTransaction represents a scheduled transaction
type ScheduledTransaction struct {
	ID                string `json:"id"`
	DateFirst         string `json:"date_first"`
	DateNext          string `json:"date_next"`
	Frequency         string `json:"frequency"`
	Amount            int64  `json:"amount"` // in milliunits
	Memo              string `json:"memo"`
	FlagColor         string `json:"flag_color"`
	FlagName          string `json:"flag_name"`
	AccountID         string `json:"account_id"`
	AccountName       string `json:"account_name"`
	PayeeID           string `json:"payee_id"`
	PayeeName         string `json:"payee_name"`
	CategoryID        string `json:"category_id"`
	CategoryName      string `json:"category_name"`
	TransferAccountID string `json:"transfer_account_id"`
	Deleted           bool   `json:"deleted"`
}

// API response wrappers

// BudgetSummaryResponse wraps budget list response
type BudgetSummaryResponse struct {
	Data struct {
		Budgets       []Budget `json:"budgets"`
		DefaultBudget *Budget  `json:"default_budget"`
	} `json:"data"`
}

// BudgetDetailResponse wraps single budget response
type BudgetDetailResponse struct {
	Data struct {
		Budget          Budget  `json:"budget"`
		ServerKnowledge int64   `json:"server_knowledge"`
	} `json:"data"`
}

// AccountsResponse wraps accounts list response
type AccountsResponse struct {
	Data struct {
		Accounts        []Account `json:"accounts"`
		ServerKnowledge int64     `json:"server_knowledge"`
	} `json:"data"`
}

// TransactionsResponse wraps transactions list response
type TransactionsResponse struct {
	Data struct {
		Transactions    []Transaction `json:"transactions"`
		ServerKnowledge int64         `json:"server_knowledge"`
	} `json:"data"`
}

// TransactionResponse wraps single transaction response
type TransactionResponse struct {
	Data struct {
		Transaction     Transaction `json:"transaction"`
		ServerKnowledge int64       `json:"server_knowledge"`
	} `json:"data"`
}

// CategoriesResponse wraps categories response
type CategoriesResponse struct {
	Data struct {
		CategoryGroups  []CategoryGroup `json:"category_groups"`
		ServerKnowledge int64           `json:"server_knowledge"`
	} `json:"data"`
}

// PayeesResponse wraps payees list response
type PayeesResponse struct {
	Data struct {
		Payees          []Payee `json:"payees"`
		ServerKnowledge int64   `json:"server_knowledge"`
	} `json:"data"`
}

// Helper functions

// MilliunitsToFloat converts YNAB milliunits (1/1000 of currency unit) to float
func MilliunitsToFloat(milliunits int64) float64 {
	return float64(milliunits) / 1000.0
}

// FloatToMilliunits converts float to YNAB milliunits
func FloatToMilliunits(amount float64) int64 {
	return int64(amount * 1000)
}

// FormatCurrency formats milliunits as currency string
func FormatCurrency(milliunits int64) string {
	return fmt.Sprintf("$%.2f", MilliunitsToFloat(milliunits))
}
