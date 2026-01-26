package api

import (
	"log"
	"net/http"
	"time"
)

// TRMNLResponse represents the data structure for TRMNL polling
type TRMNLResponse struct {
	Month          string              `json:"month"`           // e.g., "January 2026"
	TotalIncome    float64             `json:"total_income"`    // positive amounts
	TotalExpenses  float64             `json:"total_expenses"`  // negative amounts (absolute value)
	Balance        float64             `json:"balance"`         // income - expenses
	Currency       string              `json:"currency"`        // e.g., "usd"
	TopCategories  []CategorySummary   `json:"top_categories"`  // top 5 expense categories
	AllCategories  []CategorySummary   `json:"all_categories"`  // all categories with amounts and percentages
	MonthlyTrend   []MonthlyData       `json:"monthly_trend"`   // last 12 months trend
	LastUpdated    string              `json:"last_updated"`    // ISO timestamp
}

type CategorySummary struct {
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`     // absolute value
	Percentage float64 `json:"percentage"` // percentage of total expenses
}

type MonthlyData struct {
	Month         string  `json:"month"`          // e.g., "2026-01" or "Jan 2026"
	TotalIncome   float64 `json:"total_income"`   // income for the month
	TotalExpenses float64 `json:"total_expenses"` // expenses for the month
	Balance       float64 `json:"balance"`        // net balance
}

// GetTRMNLData returns current month's expense summary for TRMNL polling
func (h *Handler) GetTRMNLData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get all expenses
	expenses, err := h.storage.GetAllExpenses()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve expenses"})
		log.Printf("API ERROR: Failed to retrieve expenses for TRMNL: %v\n", err)
		return
	}

	// Get currency
	currency, err := h.storage.GetCurrency()
	if err != nil {
		currency = "usd" // default fallback
	}

	// Get start date
	startDate, err := h.storage.GetStartDate()
	if err != nil {
		startDate = 1 // default fallback
	}

	// Calculate current month range based on start date
	now := time.Now()
	var monthStart, monthEnd time.Time

	if now.Day() >= startDate {
		// Current period: startDate of this month to startDate-1 of next month
		monthStart = time.Date(now.Year(), now.Month(), startDate, 0, 0, 0, 0, time.UTC)
		monthEnd = time.Date(now.Year(), now.Month()+1, startDate, 0, 0, 0, 0, time.UTC)
	} else {
		// Current period: startDate of last month to startDate-1 of this month
		monthStart = time.Date(now.Year(), now.Month()-1, startDate, 0, 0, 0, 0, time.UTC)
		monthEnd = time.Date(now.Year(), now.Month(), startDate, 0, 0, 0, 0, time.UTC)
	}

	// Calculate totals and category breakdown for current month
	var totalIncome, totalExpenses float64
	categoryTotals := make(map[string]float64)

	for _, expense := range expenses {
		// Check if expense is in current period
		if expense.Date.Before(monthStart) || expense.Date.After(monthEnd) {
			continue
		}

		if expense.Amount >= 0 {
			// Income
			totalIncome += expense.Amount
		} else {
			// Expense
			absAmount := -expense.Amount
			totalExpenses += absAmount
			categoryTotals[expense.Category] += absAmount
		}
	}

	// Get top 5 categories by spending
	topCategories := getTopCategories(categoryTotals, totalExpenses, 5)

	// Get all categories sorted by amount
	allCategories := getTopCategories(categoryTotals, totalExpenses, len(categoryTotals))

	// Calculate last 12 months trend
	monthlyTrend := calculateMonthlyTrend(expenses, startDate, 12)

	response := TRMNLResponse{
		Month:         monthStart.Format("January 2006"),
		TotalIncome:   totalIncome,
		TotalExpenses: totalExpenses,
		Balance:       totalIncome - totalExpenses,
		Currency:      currency,
		TopCategories: topCategories,
		AllCategories: allCategories,
		MonthlyTrend:  monthlyTrend,
		LastUpdated:   time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, response)
	log.Println("HTTP: Served TRMNL data")
}

// getTopCategories returns top N categories sorted by amount with percentages
func getTopCategories(categoryTotals map[string]float64, totalExpenses float64, limit int) []CategorySummary {
	// Convert map to slice
	categories := make([]CategorySummary, 0, len(categoryTotals))
	for name, amount := range categoryTotals {
		percentage := 0.0
		if totalExpenses > 0 {
			percentage = (amount / totalExpenses) * 100
		}
		categories = append(categories, CategorySummary{
			Name:       name,
			Amount:     amount,
			Percentage: percentage,
		})
	}

	// Simple bubble sort (good enough for small datasets)
	for i := 0; i < len(categories); i++ {
		for j := i + 1; j < len(categories); j++ {
			if categories[j].Amount > categories[i].Amount {
				categories[i], categories[j] = categories[j], categories[i]
			}
		}
	}

	// Return top N
	if len(categories) > limit {
		return categories[:limit]
	}
	return categories
}

// calculateMonthlyTrend calculates income, expenses, and balance for the last N months
func calculateMonthlyTrend(expenses []Expense, startDate int, months int) []MonthlyData {
	now := time.Now()
	trend := make([]MonthlyData, 0, months)

	// Calculate for each of the last N months
	for i := months - 1; i >= 0; i-- {
		var monthStart, monthEnd time.Time

		// Calculate the target month (going back i months from now)
		targetDate := now.AddDate(0, -i, 0)

		if targetDate.Day() >= startDate {
			// Period: startDate of target month to startDate-1 of next month
			monthStart = time.Date(targetDate.Year(), targetDate.Month(), startDate, 0, 0, 0, 0, time.UTC)
			monthEnd = time.Date(targetDate.Year(), targetDate.Month()+1, startDate, 0, 0, 0, 0, time.UTC)
		} else {
			// Period: startDate of previous month to startDate-1 of target month
			monthStart = time.Date(targetDate.Year(), targetDate.Month()-1, startDate, 0, 0, 0, 0, time.UTC)
			monthEnd = time.Date(targetDate.Year(), targetDate.Month(), startDate, 0, 0, 0, 0, time.UTC)
		}

		// Calculate totals for this month
		var income, expenseTotal float64
		for _, expense := range expenses {
			if expense.Date.Before(monthStart) || expense.Date.After(monthEnd) {
				continue
			}

			if expense.Amount >= 0 {
				income += expense.Amount
			} else {
				expenseTotal += -expense.Amount
			}
		}

		trend = append(trend, MonthlyData{
			Month:         monthStart.Format("Jan 2006"),
			TotalIncome:   income,
			TotalExpenses: expenseTotal,
			Balance:       income - expenseTotal,
		})
	}

	return trend
}
