package api

import (
	"encoding/json"
	"fmt"
)

// ListMetaobjectDefinitions returns all metaobject definitions.
func (c *Client) ListMetaobjectDefinitions(first int) (*MetaobjectDefinitionConnection, error) {
	const gql = `
		query ListMetaobjectDefinitions($first: Int!) {
			metaobjectDefinitions(first: $first) {
				edges {
					cursor
					node { id name type description }
				}
				pageInfo { hasNextPage endCursor }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"first": first})
	if err != nil {
		return nil, err
	}
	var data struct {
		MetaobjectDefinitions MetaobjectDefinitionConnection `json:"metaobjectDefinitions"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing metaobject definitions: %w", err)
	}
	return &data.MetaobjectDefinitions, nil
}

// ListMetaobjects returns a paginated list of metaobjects of a given type.
func (c *Client) ListMetaobjects(metaobjectType string, first int, after string) (*MetaobjectConnection, error) {
	const gql = `
		query ListMetaobjects($type: String!, $first: Int!, $after: String) {
			metaobjects(type: $type, first: $first, after: $after) {
				edges {
					cursor
					node {
						id handle type updatedAt
						fields { key value type }
					}
				}
				pageInfo { hasNextPage endCursor }
			}
		}`
	vars := map[string]any{"type": metaobjectType, "first": first}
	if after != "" {
		vars["after"] = after
	}
	resp, err := c.Do(gql, vars)
	if err != nil {
		return nil, err
	}
	var data struct {
		Metaobjects MetaobjectConnection `json:"metaobjects"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing metaobjects: %w", err)
	}
	return &data.Metaobjects, nil
}

// GetMetaobject returns a single metaobject by ID.
func (c *Client) GetMetaobject(id string) (*Metaobject, error) {
	const gql = `
		query GetMetaobject($id: ID!) {
			metaobject(id: $id) {
				id handle type updatedAt
				fields { key value type }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Metaobject", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Metaobject *Metaobject `json:"metaobject"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing metaobject: %w", err)
	}
	if data.Metaobject == nil {
		return nil, fmt.Errorf("metaobject %s not found", id)
	}
	return data.Metaobject, nil
}

// CreateMetaobject creates a new metaobject.
func (c *Client) CreateMetaobject(metaobjectType, handle string, fields []MetaobjectField) (*Metaobject, error) {
	const gql = `
		mutation metaobjectCreate($metaobject: MetaobjectCreateInput!) {
			metaobjectCreate(metaobject: $metaobject) {
				metaobject { id handle type updatedAt fields { key value } }
				userErrors { field message }
			}
		}`
	fieldInputs := make([]map[string]any, len(fields))
	for i, f := range fields {
		fieldInputs[i] = map[string]any{"key": f.Key, "value": f.Value}
	}
	input := map[string]any{
		"type":   metaobjectType,
		"fields": fieldInputs,
	}
	if handle != "" {
		input["handle"] = handle
	}
	resp, err := c.Do(gql, map[string]any{"metaobject": input})
	if err != nil {
		return nil, err
	}
	var data struct {
		MetaobjectCreate struct {
			Metaobject *Metaobject `json:"metaobject"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"metaobjectCreate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.MetaobjectCreate.UserErrors); err != nil {
		return nil, err
	}
	return data.MetaobjectCreate.Metaobject, nil
}

// UpdateMetaobject updates an existing metaobject.
func (c *Client) UpdateMetaobject(id string, fields []MetaobjectField) (*Metaobject, error) {
	const gql = `
		mutation metaobjectUpdate($id: ID!, $metaobject: MetaobjectUpdateInput!) {
			metaobjectUpdate(id: $id, metaobject: $metaobject) {
				metaobject { id handle type updatedAt fields { key value } }
				userErrors { field message }
			}
		}`
	fieldInputs := make([]map[string]any, len(fields))
	for i, f := range fields {
		fieldInputs[i] = map[string]any{"key": f.Key, "value": f.Value}
	}
	resp, err := c.Do(gql, map[string]any{
		"id":         ToGID("Metaobject", id),
		"metaobject": map[string]any{"fields": fieldInputs},
	})
	if err != nil {
		return nil, err
	}
	var data struct {
		MetaobjectUpdate struct {
			Metaobject *Metaobject `json:"metaobject"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"metaobjectUpdate"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.MetaobjectUpdate.UserErrors); err != nil {
		return nil, err
	}
	return data.MetaobjectUpdate.Metaobject, nil
}

// DeleteMetaobject deletes a metaobject.
func (c *Client) DeleteMetaobject(id string) error {
	const gql = `
		mutation metaobjectDelete($id: ID!) {
			metaobjectDelete(id: $id) {
				deletedId
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Metaobject", id)})
	if err != nil {
		return err
	}
	var data struct {
		MetaobjectDelete struct {
			DeletedId  string      `json:"deletedId"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"metaobjectDelete"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.MetaobjectDelete.UserErrors)
}
