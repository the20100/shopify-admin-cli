package api

import (
	"encoding/json"
	"fmt"
)

// ListMarkets returns a list of Shopify Markets.
func (c *Client) ListMarkets(first int) (*MarketConnection, error) {
	const gql = `
		query ListMarkets($first: Int!) {
			markets(first: $first) {
				edges {
					cursor
					node {
						id name handle enabled primary
						regions(first: 20) {
							edges { node { name } }
						}
					}
				}
				pageInfo { hasNextPage endCursor }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"first": first})
	if err != nil {
		return nil, err
	}
	var data struct {
		Markets MarketConnection `json:"markets"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing markets: %w", err)
	}
	return &data.Markets, nil
}

// GetMarket returns a single market by ID.
func (c *Client) GetMarket(id string) (*Market, error) {
	const gql = `
		query GetMarket($id: ID!) {
			market(id: $id) {
				id name handle enabled primary
				regions(first: 50) {
					edges { node { name } }
				}
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Market", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Market *Market `json:"market"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing market: %w", err)
	}
	if data.Market == nil {
		return nil, fmt.Errorf("market %s not found", id)
	}
	return data.Market, nil
}
