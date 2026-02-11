package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tanq16/expenseowl/internal/storage"
)

// mockStorage is a simple mock implementation for testing
type mockStorage struct {
	expenses  []storage.Expense
	startDate int
}

func (m *mockStorage) GetAllExpenses() ([]storage.Expense, error) {
	return m.expenses, nil
}

func (m *mockStorage) GetStartDate() (int, error) {
	return m.startDate, nil
}

func (m *mockStorage) GetConfig() (*storage.Config, error) {
	return &storage.Config{}, nil
}

func (m *mockStorage) GetExpense(string) (storage.Expense, error) {
	return storage.Expense{}, nil
}

func (m *mockStorage) GetRecurringExpense(string) (storage.RecurringExpense, error) {
	return storage.RecurringExpense{}, nil
}

func (m *mockStorage) FindDuplicateExpense(string, string, float64, time.Time) (bool, error) {
	return false, nil
}

func (m *mockStorage) AddMultipleExpenses([]storage.Expense) error {
	return nil
}

func (m *mockStorage) GetCategories() ([]string, error) {
	return []string{}, nil
}

func (m *mockStorage) GetCurrency() (string, error) {
	return "usd", nil
}

func (m *mockStorage) UpdateCategories([]string) error {
	return nil
}

func (m *mockStorage) UpdateCurrency(string) error {
	return nil
}

func (m *mockStorage) UpdateStartDate(int) error {
	return nil
}

func (m *mockStorage) AddExpense(storage.Expense) error {
	return nil
}

func (m *mockStorage) UpdateExpense(string, storage.Expense) error {
	return nil
}

func (m *mockStorage) RemoveExpense(string) error {
	return nil
}

func (m *mockStorage) RemoveMultipleExpenses([]string) error {
	return nil
}

func (m *mockStorage) AddRecurringExpense(storage.RecurringExpense) error {
	return nil
}

func (m *mockStorage) GetRecurringExpenses() ([]storage.RecurringExpense, error) {
	return []storage.RecurringExpense{}, nil
}

func (m *mockStorage) UpdateRecurringExpense(string, storage.RecurringExpense, bool) error {
	return nil
}

func (m *mockStorage) RemoveRecurringExpense(string, bool) error {
	return nil
}

func (m *mockStorage) GetSubCategories(string) ([]string, error) {
	return []string{}, nil
}

func (m *mockStorage) AddSubCategory(string, string) error {
	return nil
}

func (m *mockStorage) RemoveSubCategory(string, string) error {
	return nil
}

func (m *mockStorage) RenameSubCategory(string, string, string) error {
	return nil
}

func (m *mockStorage) GetSubCategoryMappings() ([]storage.SubCategoryMappingRule, error) {
	return []storage.SubCategoryMappingRule{}, nil
}

func (m *mockStorage) UpdateSubCategoryMappings([]storage.SubCategoryMappingRule) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

// TestGetMonthlyExpenses_DefaultMonths tests the default behavior (12 months)
func TestGetMonthlyExpenses_DefaultMonths(t *testing.T) {
	// Create mock storage with sample expenses
	now := time.Now()
	mock := &mockStorage{
		expenses: []storage.Expense{
			{
				ID:       "1",
				Date:     now.AddDate(0, -1, 0),
				Amount:   -100.0,
				Category: "Food",
				Name:     "Test expense",
			},
			{
				ID:       "2",
				Date:     now.AddDate(0, -2, 0),
				Amount:   -200.0,
				Category: "Transport",
				Name:     "Test expense 2",
			},
		},
		startDate: 1,
	}

	handler := NewHandler(mock)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/expenses/monthly", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetMonthlyExpenses(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var monthlyData []MonthlyData
	if err := json.NewDecoder(w.Body).Decode(&monthlyData); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should return 12 months by default
	if len(monthlyData) != 12 {
		t.Errorf("Expected 12 months, got %d", len(monthlyData))
	}
}

// TestGetMonthlyExpenses_CustomMonths tests custom month parameter
func TestGetMonthlyExpenses_CustomMonths(t *testing.T) {
	mock := &mockStorage{
		expenses:  []storage.Expense{},
		startDate: 1,
	}

	handler := NewHandler(mock)

	// Create request with months=6
	req := httptest.NewRequest(http.MethodGet, "/api/expenses/monthly?months=6", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetMonthlyExpenses(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var monthlyData []MonthlyData
	if err := json.NewDecoder(w.Body).Decode(&monthlyData); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should return 6 months
	if len(monthlyData) != 6 {
		t.Errorf("Expected 6 months, got %d", len(monthlyData))
	}
}

// TestGetMonthlyExpenses_InvalidMethod tests that non-GET methods are rejected
func TestGetMonthlyExpenses_InvalidMethod(t *testing.T) {
	mock := &mockStorage{
		expenses:  []storage.Expense{},
		startDate: 1,
	}

	handler := NewHandler(mock)

	// Create POST request
	req := httptest.NewRequest(http.MethodPost, "/api/expenses/monthly", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetMonthlyExpenses(w, req)

	// Check response
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestFilterExpensesByCategories tests category filtering
func TestFilterExpensesByCategories(t *testing.T) {
	expenses := []storage.Expense{
		{ID: "1", Category: "Food", Amount: -100},
		{ID: "2", Category: "Transport", Amount: -200},
		{ID: "3", Category: "Food", Amount: -150},
		{ID: "4", Category: "Entertainment", Amount: -50},
	}

	// Test single category filter
	filtered := filterExpensesByCategories(expenses, []string{"Food"})
	if len(filtered) != 2 {
		t.Errorf("Expected 2 Food expenses, got %d", len(filtered))
	}

	// Test multiple category filter
	filtered = filterExpensesByCategories(expenses, []string{"Food", "Transport"})
	if len(filtered) != 3 {
		t.Errorf("Expected 3 expenses (Food + Transport), got %d", len(filtered))
	}

	// Test empty filter (should return all)
	filtered = filterExpensesByCategories(expenses, []string{})
	if len(filtered) != 4 {
		t.Errorf("Expected all 4 expenses, got %d", len(filtered))
	}

	// Test non-existent category
	filtered = filterExpensesByCategories(expenses, []string{"NonExistent"})
	if len(filtered) != 0 {
		t.Errorf("Expected 0 expenses for non-existent category, got %d", len(filtered))
	}
}
