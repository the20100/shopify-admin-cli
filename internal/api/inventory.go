package api

import (
	"encoding/json"
	"fmt"
)

// ListLocations returns a list of inventory locations.
func (c *Client) ListLocations(first int) (*LocationConnection, error) {
	const gql = `
		query ListLocations($first: Int!) {
			locations(first: $first) {
				edges {
					cursor
					node {
						id name isActive
						address { address1 city country }
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
		Locations LocationConnection `json:"locations"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing locations: %w", err)
	}
	return &data.Locations, nil
}

// ListInventoryItems returns a paginated list of inventory items.
func (c *Client) ListInventoryItems(first int, after, query string) (*InventoryItemConnection, error) {
	const gql = `
		query ListInventoryItems($first: Int!, $after: String, $query: String) {
			inventoryItems(first: $first, after: $after, query: $query) {
				edges {
					cursor
					node {
						id sku tracked requiresShipping
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
		InventoryItems InventoryItemConnection `json:"inventoryItems"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing inventory items: %w", err)
	}
	return &data.InventoryItems, nil
}

// ListInventoryLevels returns inventory levels for a specific location.
func (c *Client) ListInventoryLevels(locationID string, first int) (*InventoryLevelConnection, error) {
	const gql = `
		query ListInventoryLevels($id: ID!, $first: Int!) {
			location(id: $id) {
				inventoryLevels(first: $first) {
					edges {
						cursor
						node {
							id
							quantities(names: ["available", "on_hand"]) {
								name quantity
							}
							item { id sku }
						}
					}
					pageInfo { hasNextPage endCursor }
				}
			}
		}`
	resp, err := c.Do(gql, map[string]any{
		"id":    ToGID("Location", locationID),
		"first": first,
	})
	if err != nil {
		return nil, err
	}
	var data struct {
		Location struct {
			InventoryLevels InventoryLevelConnection `json:"inventoryLevels"`
		} `json:"location"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing inventory levels: %w", err)
	}
	return &data.Location.InventoryLevels, nil
}

// AdjustInventory adjusts the inventory quantity for an item at a location.
func (c *Client) AdjustInventory(inventoryItemID, locationID string, delta int, reason string) error {
	const gql = `
		mutation inventoryAdjustQuantities($input: InventoryAdjustQuantitiesInput!) {
			inventoryAdjustQuantities(input: $input) {
				inventoryAdjustmentGroup {
					reason
					changes {
						name delta quantityAfterChange
						item { id sku }
						location { name }
					}
				}
				userErrors { field message }
			}
		}`
	if reason == "" {
		reason = "correction"
	}
	input := map[string]any{
		"reason": reason,
		"name":   "available",
		"changes": []map[string]any{
			{
				"inventoryItemId": ToGID("InventoryItem", inventoryItemID),
				"locationId":      ToGID("Location", locationID),
				"delta":           delta,
			},
		},
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return err
	}
	var data struct {
		InventoryAdjustQuantities struct {
			UserErrors []UserError `json:"userErrors"`
		} `json:"inventoryAdjustQuantities"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.InventoryAdjustQuantities.UserErrors)
}
