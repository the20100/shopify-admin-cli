package api

import (
	"encoding/json"
	"fmt"
)

// RunShopifyQL executes a ShopifyQL analytics query.
func (c *Client) RunShopifyQL(query string) (*ShopifyQLResult, error) {
	const gql = `
		query shopifyqlQuery($query: String!) {
			shopifyqlQuery(query: $query) {
				parseErrors { code message }
				tableData {
					columnDefinitions { name dataType displayName }
					rows
				}
			}
		}`
	resp, err := c.Do(gql, map[string]any{"query": query})
	if err != nil {
		return nil, err
	}
	var data struct {
		ShopifyqlQuery ShopifyQLResult `json:"shopifyqlQuery"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing analytics result: %w", err)
	}
	return &data.ShopifyqlQuery, nil
}
