package api

import (
	"encoding/json"
	"fmt"
)

// GetShop returns store information.
func (c *Client) GetShop() (*Shop, error) {
	const query = `{
		shop {
			id
			name
			email
			myshopifyDomain
			primaryDomain { url host }
			currencyCode
			countryCode
			timezoneAbbreviation
			createdAt
			plan { displayName shopifyPlus }
		}
	}`
	resp, err := c.Do(query, nil)
	if err != nil {
		return nil, err
	}
	var data struct {
		Shop Shop `json:"shop"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing shop: %w", err)
	}
	return &data.Shop, nil
}
