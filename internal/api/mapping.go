package api

import (
	"fmt"
	"strings"

	"github.com/tanq16/expenseowl/internal/storage"
)

// MappingEngine applies subcategory mapping rules to transaction names
type MappingEngine struct {
	rules []storage.SubCategoryMappingRule
}

// NewMappingEngine creates a new mapping engine with the provided rules
func NewMappingEngine(rules []storage.SubCategoryMappingRule) (*MappingEngine, error) {
	// Validate rules
	for i, rule := range rules {
		if err := validateMappingRule(rule); err != nil {
			return nil, fmt.Errorf("invalid rule at index %d: %w", i, err)
		}
	}
	
	return &MappingEngine{
		rules: rules,
	}, nil
}

// ApplyMapping applies mapping rules to a transaction name and category
// Returns the matched subcategory or empty string if no match
// First matching rule wins (rule precedence)
func (m *MappingEngine) ApplyMapping(transactionName string, category string) string {
	for _, rule := range m.rules {
		// Skip rules that don't match the category
		if rule.Category != category {
			continue
		}
		
		// Apply pattern matching based on match type
		if m.matchesPattern(transactionName, rule) {
			return rule.SubCategory
		}
	}
	
	return ""
}

// ApplyMappingWithCategory applies mapping rules to a transaction name
// Returns both the matched category and subcategory, or empty strings if no match
// This is used when CSV doesn't have a category column
// First matching rule wins (rule precedence)
func (m *MappingEngine) ApplyMappingWithCategory(transactionName string) (string, string) {
	for _, rule := range m.rules {
		// Apply pattern matching based on match type
		if m.matchesPattern(transactionName, rule) {
			return rule.Category, rule.SubCategory
		}
	}
	
	return "", ""
}

// matchesPattern checks if a transaction name matches a rule's pattern
func (m *MappingEngine) matchesPattern(transactionName string, rule storage.SubCategoryMappingRule) bool {
	switch rule.MatchType {
	case "exact":
		return rule.Pattern == transactionName
	case "contains":
		return strings.Contains(
			strings.ToLower(transactionName),
			strings.ToLower(rule.Pattern),
		)
	default:
		return false
	}
}

// validateMappingRule validates a single mapping rule
func validateMappingRule(rule storage.SubCategoryMappingRule) error {
	if rule.Pattern == "" {
		return fmt.Errorf("mapping rule missing required field: pattern")
	}
	if rule.MatchType == "" {
		return fmt.Errorf("mapping rule missing required field: matchType")
	}
	if rule.Category == "" {
		return fmt.Errorf("mapping rule missing required field: category")
	}
	if rule.SubCategory == "" {
		return fmt.Errorf("mapping rule missing required field: subCategory")
	}
	
	// Validate match type
	if rule.MatchType != "exact" && rule.MatchType != "contains" {
		return fmt.Errorf("invalid match type '%s', must be 'exact' or 'contains'", rule.MatchType)
	}
	
	return nil
}
