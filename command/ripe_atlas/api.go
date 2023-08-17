package ripe_atlas

type CreditsResponse struct {
	CurrentBalance            int    `json:"current_balance"`
	CreditChecked             bool   `json:"credit_checked"`
	MaxDailyCredits           int    `json:"max_daily_credits"`
	EstimatedDailyIncome      int    `json:"estimated_daily_income"`
	EstimatedDailyExpenditure int    `json:"estimated_daily_expenditure"`
	EstimatedDailyBalance     int    `json:"estimated_daily_balance"`
	CalculationTime           string `json:"calculation_time"`
	EstimatedRunoutSeconds    any    `json:"estimated_runout_seconds"`
	PastDayMeasurementResults int    `json:"past_day_measurement_results"`
	PastDayCreditsSpent       int    `json:"past_day_credits_spent"`
	LastDateDebited           string `json:"last_date_debited"`
	LastDateCredited          string `json:"last_date_credited"`
	IncomeItems               string `json:"income_items"`
	ExpenseItems              string `json:"expense_items"`
	Transactions              string `json:"transactions"`
}

