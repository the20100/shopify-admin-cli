package api

import (
	"encoding/json"
	"fmt"
)

// ListOrders returns a paginated list of orders.
func (c *Client) ListOrders(first int, after, query string) (*OrderConnection, error) {
	const gql = `
		query ListOrders($first: Int!, $after: String, $query: String) {
			orders(first: $first, after: $after, query: $query) {
				edges {
					cursor
					node {
						id name email financialStatus displayFulfillmentStatus
						totalPriceSet { shopMoney { amount currencyCode } }
						createdAt
						customer { firstName lastName }
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
		Orders OrderConnection `json:"orders"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing orders: %w", err)
	}
	return &data.Orders, nil
}

// GetOrder returns a single order by ID with full details.
func (c *Client) GetOrder(id string) (*Order, error) {
	const gql = `
		query GetOrder($id: ID!) {
			order(id: $id) {
				id name email phone
				financialStatus displayFulfillmentStatus
				totalPriceSet { shopMoney { amount currencyCode } }
				subtotalPriceSet { shopMoney { amount currencyCode } }
				totalTaxSet { shopMoney { amount currencyCode } }
				createdAt processedAt note tags
				customer { id firstName lastName email }
				shippingAddress {
					firstName lastName address1 address2
					city province zip country phone
				}
				lineItems(first: 50) {
					edges {
						node {
							id title quantity sku
							originalUnitPriceSet { shopMoney { amount currencyCode } }
						}
					}
				}
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Order", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Order *Order `json:"order"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing order: %w", err)
	}
	if data.Order == nil {
		return nil, fmt.Errorf("order %s not found", id)
	}
	return data.Order, nil
}

// CloseOrder marks an order as closed.
func (c *Client) CloseOrder(id string) (*Order, error) {
	const gql = `
		mutation orderClose($input: OrderCloseInput!) {
			orderClose(input: $input) {
				order { id name financialStatus displayFulfillmentStatus }
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"input": map[string]any{"id": ToGID("Order", id)}})
	if err != nil {
		return nil, err
	}
	var data struct {
		OrderClose struct {
			Order      *Order      `json:"order"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"orderClose"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.OrderClose.UserErrors); err != nil {
		return nil, err
	}
	return data.OrderClose.Order, nil
}

// CancelOrder cancels an order.
func (c *Client) CancelOrder(id, reason string, refund, restock bool) error {
	const gql = `
		mutation orderCancel($orderId: ID!, $reason: OrderCancelReason!, $refund: Boolean!, $restock: Boolean!, $notifyCustomer: Boolean!) {
			orderCancel(orderId: $orderId, reason: $reason, refund: $refund, restock: $restock, notifyCustomer: $notifyCustomer) {
				job { id done }
				orderCancelUserErrors { field message }
			}
		}`
	if reason == "" {
		reason = "OTHER"
	}
	resp, err := c.Do(gql, map[string]any{
		"orderId":        ToGID("Order", id),
		"reason":         reason,
		"refund":         refund,
		"restock":        restock,
		"notifyCustomer": false,
	})
	if err != nil {
		return err
	}
	var data struct {
		OrderCancel struct {
			OrderCancelUserErrors []UserError `json:"orderCancelUserErrors"`
		} `json:"orderCancel"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.OrderCancel.OrderCancelUserErrors)
}

// MarkOrderAsPaid marks an order as paid.
func (c *Client) MarkOrderAsPaid(id string) (*Order, error) {
	const gql = `
		mutation orderMarkAsPaid($input: OrderMarkAsPaidInput!) {
			orderMarkAsPaid(input: $input) {
				order { id name financialStatus }
				errors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"input": map[string]any{"id": ToGID("Order", id)}})
	if err != nil {
		return nil, err
	}
	var data struct {
		OrderMarkAsPaid struct {
			Order  *Order      `json:"order"`
			Errors []UserError `json:"errors"`
		} `json:"orderMarkAsPaid"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.OrderMarkAsPaid.Errors); err != nil {
		return nil, err
	}
	return data.OrderMarkAsPaid.Order, nil
}
