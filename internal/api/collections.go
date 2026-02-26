package api

import (
	"encoding/json"
	"fmt"
)

// ListCollections returns a paginated list of collections.
func (c *Client) ListCollections(first int, after, query string) (*CollectionConnection, error) {
	const gql = `
		query ListCollections($first: Int!, $after: String, $query: String) {
			collections(first: $first, after: $after, query: $query) {
				edges {
					cursor
					node {
						id title handle updatedAt
						productsCount { count }
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
		Collections CollectionConnection `json:"collections"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing collections: %w", err)
	}
	return &data.Collections, nil
}

// GetCollection returns a single collection by ID.
func (c *Client) GetCollection(id string) (*Collection, error) {
	const gql = `
		query GetCollection($id: ID!) {
			collection(id: $id) {
				id title handle description updatedAt
				productsCount { count }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Collection", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Collection *Collection `json:"collection"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing collection: %w", err)
	}
	if data.Collection == nil {
		return nil, fmt.Errorf("collection %s not found", id)
	}
	return data.Collection, nil
}

// CreateCollection creates a new custom collection.
func (c *Client) CreateCollection(title, description string) (*Collection, error) {
	const gql = `
		mutation collectionCreate($input: CollectionInput!) {
			collectionCreate(input: $input) {
				collection { id title handle updatedAt }
				userErrors { field message }
			}
		}`
	input := map[string]any{"title": title}
	if description != "" {
		input["descriptionHtml"] = description
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		CollectionCreate struct {
			Collection *Collection `json:"collection"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"collectionCreate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.CollectionCreate.UserErrors); err != nil {
		return nil, err
	}
	return data.CollectionCreate.Collection, nil
}

// UpdateCollection updates an existing collection.
func (c *Client) UpdateCollection(id, title, description string) (*Collection, error) {
	const gql = `
		mutation collectionUpdate($input: CollectionInput!) {
			collectionUpdate(input: $input) {
				collection { id title handle updatedAt }
				userErrors { field message }
			}
		}`
	input := map[string]any{"id": ToGID("Collection", id)}
	if title != "" {
		input["title"] = title
	}
	if description != "" {
		input["descriptionHtml"] = description
	}
	resp, err := c.Do(gql, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		CollectionUpdate struct {
			Collection *Collection `json:"collection"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"collectionUpdate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.CollectionUpdate.UserErrors); err != nil {
		return nil, err
	}
	return data.CollectionUpdate.Collection, nil
}

// DeleteCollection deletes a collection.
func (c *Client) DeleteCollection(id string) error {
	const gql = `
		mutation collectionDelete($input: CollectionDeleteInput!) {
			collectionDelete(input: $input) {
				deletedCollectionId
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"input": map[string]any{"id": ToGID("Collection", id)}})
	if err != nil {
		return err
	}
	var data struct {
		CollectionDelete struct {
			DeletedCollectionId string      `json:"deletedCollectionId"`
			UserErrors          []UserError `json:"userErrors"`
		} `json:"collectionDelete"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.CollectionDelete.UserErrors)
}
