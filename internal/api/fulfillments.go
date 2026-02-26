package api

import (
	"encoding/json"
	"fmt"
)

// ListFulfillmentOrders returns fulfillment orders for a given order ID.
func (c *Client) ListFulfillmentOrders(orderID string) (*FulfillmentOrderConnection, error) {
	const gql = `
		query ListFulfillmentOrders($id: ID!) {
			order(id: $id) {
				fulfillmentOrders(first: 50) {
					edges {
						cursor
						node {
							id status requestStatus fulfillAt
							assignedLocation { name }
							destination {
								firstName lastName address1 city country zip
							}
							lineItems(first: 50) {
								edges {
									node {
										id remainingQuantity totalQuantity
										lineItem { id title sku }
									}
								}
							}
						}
					}
					pageInfo { hasNextPage endCursor }
				}
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Order", orderID)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Order struct {
			FulfillmentOrders FulfillmentOrderConnection `json:"fulfillmentOrders"`
		} `json:"order"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing fulfillment orders: %w", err)
	}
	return &data.Order.FulfillmentOrders, nil
}

// CreateFulfillment creates a fulfillment for a fulfillment order.
func (c *Client) CreateFulfillment(fulfillmentOrderID, trackingCompany, trackingNumber, trackingURL string) (*Fulfillment, error) {
	const gql = `
		mutation fulfillmentCreateV2($fulfillment: FulfillmentV2Input!) {
			fulfillmentCreateV2(fulfillment: $fulfillment) {
				fulfillment {
					id status createdAt
					trackingInfo { company number url }
				}
				userErrors { field message }
			}
		}`
	lineItemsByFulfillmentOrder := []map[string]any{
		{"fulfillmentOrderId": ToGID("FulfillmentOrder", fulfillmentOrderID)},
	}
	fulfillment := map[string]any{
		"lineItemsByFulfillmentOrder": lineItemsByFulfillmentOrder,
	}
	if trackingCompany != "" || trackingNumber != "" {
		trackingInfo := map[string]any{}
		if trackingCompany != "" {
			trackingInfo["company"] = trackingCompany
		}
		if trackingNumber != "" {
			trackingInfo["number"] = trackingNumber
		}
		if trackingURL != "" {
			trackingInfo["url"] = trackingURL
		}
		fulfillment["trackingInfo"] = trackingInfo
	}
	resp, err := c.Do(gql, map[string]any{"fulfillment": fulfillment})
	if err != nil {
		return nil, err
	}
	var rawData struct {
		FulfillmentCreateV2 struct {
			Fulfillment *struct {
				ID          string `json:"id"`
				Status      string `json:"status"`
				CreatedAt   string `json:"createdAt"`
				TrackingInfo []struct {
					Company string `json:"company"`
					Number  string `json:"number"`
					URL     string `json:"url"`
				} `json:"trackingInfo"`
			} `json:"fulfillment"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"fulfillmentCreateV2"`
	}
	if err := json.Unmarshal(resp.Data, &rawData); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(rawData.FulfillmentCreateV2.UserErrors); err != nil {
		return nil, err
	}
	if rawData.FulfillmentCreateV2.Fulfillment == nil {
		return nil, fmt.Errorf("no fulfillment returned")
	}
	f := rawData.FulfillmentCreateV2.Fulfillment
	result := &Fulfillment{
		ID:        f.ID,
		Status:    f.Status,
		CreatedAt: f.CreatedAt,
	}
	if len(f.TrackingInfo) > 0 {
		result.TrackingCompany = f.TrackingInfo[0].Company
		for _, t := range f.TrackingInfo {
			result.TrackingNumbers = append(result.TrackingNumbers, t.Number)
			result.TrackingURLs = append(result.TrackingURLs, t.URL)
		}
	}
	return result, nil
}
