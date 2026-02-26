package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ListProducts returns a paginated list of products.
func (c *Client) ListProducts(first int, after, query string) (*ProductConnection, error) {
	const gql = `
		query ListProducts($first: Int!, $after: String, $query: String) {
			products(first: $first, after: $after, query: $query) {
				edges {
					cursor
					node {
						id title status totalInventory
						vendor productType createdAt updatedAt
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
		Products ProductConnection `json:"products"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing products: %w", err)
	}
	return &data.Products, nil
}

// GetProduct returns a single product by ID.
func (c *Client) GetProduct(id string) (*Product, error) {
	const gql = `
		query GetProduct($id: ID!) {
			product(id: $id) {
				id title status handle description totalInventory
				vendor productType tags createdAt updatedAt
				variants(first: 100) {
					edges {
						node {
							id title price compareAtPrice sku
							inventoryQuantity barcode weight weightUnit
						}
					}
				}
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Product", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Product *Product `json:"product"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing product: %w", err)
	}
	if data.Product == nil {
		return nil, fmt.Errorf("product %s not found", id)
	}
	return data.Product, nil
}

// CreateProduct creates a new product.
func (c *Client) CreateProduct(title, vendor, productType, status, descriptionHTML string, tags []string) (*Product, error) {
	const gql = `
		mutation productCreate($input: ProductInput!) {
			productCreate(input: $input) {
				product { id title status handle vendor productType createdAt }
				userErrors { field message }
			}
		}`
	input := map[string]any{"title": title}
	if vendor != "" {
		input["vendor"] = vendor
	}
	if productType != "" {
		input["productType"] = productType
	}
	if status != "" {
		input["status"] = strings.ToUpper(status)
	}
	if descriptionHTML != "" {
		input["descriptionHtml"] = descriptionHTML
	}
	if len(tags) > 0 {
		input["tags"] = tags
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		ProductCreate struct {
			Product    *Product    `json:"product"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"productCreate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.ProductCreate.UserErrors); err != nil {
		return nil, err
	}
	return data.ProductCreate.Product, nil
}

// UpdateProduct updates an existing product.
func (c *Client) UpdateProduct(id, title, vendor, productType, status, descriptionHTML string, tags []string) (*Product, error) {
	const gql = `
		mutation productUpdate($input: ProductInput!) {
			productUpdate(input: $input) {
				product { id title status handle vendor productType updatedAt }
				userErrors { field message }
			}
		}`
	input := map[string]any{"id": ToGID("Product", id)}
	if title != "" {
		input["title"] = title
	}
	if vendor != "" {
		input["vendor"] = vendor
	}
	if productType != "" {
		input["productType"] = productType
	}
	if status != "" {
		input["status"] = strings.ToUpper(status)
	}
	if descriptionHTML != "" {
		input["descriptionHtml"] = descriptionHTML
	}
	if len(tags) > 0 {
		input["tags"] = tags
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		ProductUpdate struct {
			Product    *Product    `json:"product"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"productUpdate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.ProductUpdate.UserErrors); err != nil {
		return nil, err
	}
	return data.ProductUpdate.Product, nil
}

// DeleteProduct deletes a product.
func (c *Client) DeleteProduct(id string) error {
	const gql = `
		mutation productDelete($input: ProductDeleteInput!) {
			productDelete(input: $input) {
				deletedProductId
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"input": map[string]any{"id": ToGID("Product", id)}})
	if err != nil {
		return err
	}
	var data struct {
		ProductDelete struct {
			DeletedProductId string      `json:"deletedProductId"`
			UserErrors       []UserError `json:"userErrors"`
		} `json:"productDelete"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.ProductDelete.UserErrors)
}

// GetVariant returns a single product variant.
func (c *Client) GetVariant(id string) (*ProductVariant, error) {
	const gql = `
		query GetVariant($id: ID!) {
			productVariant(id: $id) {
				id title price compareAtPrice sku
				inventoryQuantity barcode weight weightUnit
				createdAt updatedAt
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("ProductVariant", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		ProductVariant *ProductVariant `json:"productVariant"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing variant: %w", err)
	}
	if data.ProductVariant == nil {
		return nil, fmt.Errorf("variant %s not found", id)
	}
	return data.ProductVariant, nil
}

// UpdateVariant updates a product variant.
func (c *Client) UpdateVariant(id, price, sku, barcode string) (*ProductVariant, error) {
	const gql = `
		mutation productVariantUpdate($input: ProductVariantInput!) {
			productVariantUpdate(input: $input) {
				productVariant {
					id title price sku inventoryQuantity updatedAt
				}
				userErrors { field message }
			}
		}`
	input := map[string]any{"id": ToGID("ProductVariant", id)}
	if price != "" {
		input["price"] = price
	}
	if sku != "" {
		input["sku"] = sku
	}
	if barcode != "" {
		input["barcode"] = barcode
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		ProductVariantUpdate struct {
			ProductVariant *ProductVariant `json:"productVariant"`
			UserErrors     []UserError     `json:"userErrors"`
		} `json:"productVariantUpdate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.ProductVariantUpdate.UserErrors); err != nil {
		return nil, err
	}
	return data.ProductVariantUpdate.ProductVariant, nil
}
