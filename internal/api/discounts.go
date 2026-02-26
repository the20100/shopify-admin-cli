package api

import (
	"encoding/json"
	"fmt"
)

// ListDiscounts returns a paginated list of discount nodes (all discount types).
func (c *Client) ListDiscounts(first int, after, query string) (*DiscountNodeConnection, error) {
	const gql = `
		query ListDiscounts($first: Int!, $after: String, $query: String) {
			discountNodes(first: $first, after: $after, query: $query) {
				edges {
					cursor
					node {
						id
						discount {
							__typename
							... on DiscountCodeBasic {
								title status startsAt endsAt asyncUsageCount
							}
							... on DiscountCodeBxgy {
								title status startsAt endsAt asyncUsageCount
							}
							... on DiscountCodeFreeShipping {
								title status startsAt endsAt asyncUsageCount
							}
							... on DiscountAutomaticBasic {
								title status startsAt endsAt asyncUsageCount
							}
							... on DiscountAutomaticBxgy {
								title status startsAt endsAt asyncUsageCount
							}
							... on DiscountAutomaticFreeShipping {
								title status startsAt endsAt
							}
							... on DiscountAutomaticApp {
								title status startsAt endsAt
							}
							... on DiscountCodeApp {
								title status startsAt endsAt asyncUsageCount
							}
						}
					}
				}
				pageInfo { hasNextPage endCursor }
			}
		}`
	vars := map[string]any{"first": first}
	if after != "" {
		vars["after"] = after
	}
	if query != "" {
		vars["query"] = query
	}
	resp, err := c.Do(gql, vars)
	if err != nil {
		return nil, err
	}
	var data struct {
		DiscountNodes DiscountNodeConnection `json:"discountNodes"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing discounts: %w", err)
	}
	return &data.DiscountNodes, nil
}

// DeactivateDiscount deactivates a discount code.
func (c *Client) DeactivateDiscountCode(id string) error {
	const gql = `
		mutation discountCodeDeactivate($id: ID!) {
			discountCodeDeactivate(id: $id) {
				codeDiscountNode {
					id
					discount {
						... on DiscountCodeBasic { status }
					}
				}
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("DiscountCodeNode", id)})
	if err != nil {
		return err
	}
	var data struct {
		DiscountCodeDeactivate struct {
			UserErrors []UserError `json:"userErrors"`
		} `json:"discountCodeDeactivate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.DiscountCodeDeactivate.UserErrors)
}

// DeactivateAutomaticDiscount deactivates an automatic discount.
func (c *Client) DeactivateAutomaticDiscount(id string) error {
	const gql = `
		mutation discountAutomaticDeactivate($id: ID!) {
			discountAutomaticDeactivate(id: $id) {
				automaticDiscountNode {
					id
				}
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("DiscountAutomaticNode", id)})
	if err != nil {
		return err
	}
	var data struct {
		DiscountAutomaticDeactivate struct {
			UserErrors []UserError `json:"userErrors"`
		} `json:"discountAutomaticDeactivate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.DiscountAutomaticDeactivate.UserErrors)
}
