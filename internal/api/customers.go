package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ListCustomers returns a paginated list of customers.
func (c *Client) ListCustomers(first int, after, query string) (*CustomerConnection, error) {
	const gql = `
		query ListCustomers($first: Int!, $after: String, $query: String) {
			customers(first: $first, after: $after, query: $query) {
				edges {
					cursor
					node {
						id firstName lastName email phone state
						numberOfOrders amountSpent { amount currencyCode }
						createdAt
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
		Customers CustomerConnection `json:"customers"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing customers: %w", err)
	}
	return &data.Customers, nil
}

// GetCustomer returns a single customer by ID.
func (c *Client) GetCustomer(id string) (*Customer, error) {
	const gql = `
		query GetCustomer($id: ID!) {
			customer(id: $id) {
				id firstName lastName email phone state tags
				numberOfOrders amountSpent { amount currencyCode }
				createdAt updatedAt
				defaultAddress {
					address1 address2 city province zip country
				}
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Customer", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Customer *Customer `json:"customer"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing customer: %w", err)
	}
	if data.Customer == nil {
		return nil, fmt.Errorf("customer %s not found", id)
	}
	return data.Customer, nil
}

// CreateCustomer creates a new customer.
func (c *Client) CreateCustomer(firstName, lastName, email, phone string, tags []string) (*Customer, error) {
	const gql = `
		mutation customerCreate($input: CustomerInput!) {
			customerCreate(input: $input) {
				customer { id firstName lastName email phone state createdAt }
				userErrors { field message }
			}
		}`
	input := map[string]any{}
	if firstName != "" {
		input["firstName"] = firstName
	}
	if lastName != "" {
		input["lastName"] = lastName
	}
	if email != "" {
		input["email"] = email
	}
	if phone != "" {
		input["phone"] = phone
	}
	if len(tags) > 0 {
		input["tags"] = strings.Join(tags, ",")
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		CustomerCreate struct {
			Customer   *Customer   `json:"customer"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"customerCreate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.CustomerCreate.UserErrors); err != nil {
		return nil, err
	}
	return data.CustomerCreate.Customer, nil
}

// UpdateCustomer updates an existing customer.
func (c *Client) UpdateCustomer(id, firstName, lastName, email, phone string, tags []string) (*Customer, error) {
	const gql = `
		mutation customerUpdate($input: CustomerInput!) {
			customerUpdate(input: $input) {
				customer { id firstName lastName email phone state updatedAt }
				userErrors { field message }
			}
		}`
	input := map[string]any{"id": ToGID("Customer", id)}
	if firstName != "" {
		input["firstName"] = firstName
	}
	if lastName != "" {
		input["lastName"] = lastName
	}
	if email != "" {
		input["email"] = email
	}
	if phone != "" {
		input["phone"] = phone
	}
	if len(tags) > 0 {
		input["tags"] = strings.Join(tags, ",")
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		CustomerUpdate struct {
			Customer   *Customer   `json:"customer"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"customerUpdate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.CustomerUpdate.UserErrors); err != nil {
		return nil, err
	}
	return data.CustomerUpdate.Customer, nil
}

// DeleteCustomer deletes a customer.
func (c *Client) DeleteCustomer(id string) error {
	const gql = `
		mutation customerDelete($input: CustomerDeleteInput!) {
			customerDelete(input: $input) {
				deletedCustomerId
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"input": map[string]any{"id": ToGID("Customer", id)}})
	if err != nil {
		return err
	}
	var data struct {
		CustomerDelete struct {
			DeletedCustomerId string      `json:"deletedCustomerId"`
			UserErrors        []UserError `json:"userErrors"`
		} `json:"customerDelete"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.CustomerDelete.UserErrors)
}
