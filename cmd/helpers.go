package cmd

import (
	"strings"

	"github.com/the20100/shopify-admin-cli/internal/api"
)

// shortID extracts the numeric portion from a Shopify GID for display.
func shortID(gid string) string {
	return api.ShortID(gid)
}

// splitTags splits a comma-separated tag string into a slice.
func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// formatMoney formats a MoneyV2 value as "amount currencyCode" or "-".
func formatMoney(amount, currency string) string {
	if amount == "" {
		return "-"
	}
	return amount + " " + currency
}
