package api

import (
	"encoding/json"
	"fmt"
)

// ListMetafields returns metafields for a given owner resource GID.
func (c *Client) ListMetafields(ownerID string, first int) (*MetafieldConnection, error) {
	const gql = `
		query ListMetafields($ownerId: ID!, $first: Int!) {
			metafields(owner: $ownerId, first: $first) {
				edges {
					cursor
					node {
						id namespace key value type createdAt updatedAt
					}
				}
				pageInfo { hasNextPage endCursor }
			}
		}`
	resp, err := c.Do(gql, map[string]any{
		"ownerId": ownerID,
		"first":   first,
	})
	if err != nil {
		return nil, err
	}
	var data struct {
		Metafields MetafieldConnection `json:"metafields"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing metafields: %w", err)
	}
	return &data.Metafields, nil
}

// GetMetafield returns a single metafield by ID.
func (c *Client) GetMetafield(id string) (*Metafield, error) {
	const gql = `
		query GetMetafield($id: ID!) {
			metafield(id: $id) {
				id namespace key value type createdAt updatedAt
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("Metafield", id)})
	if err != nil {
		return nil, err
	}
	var data struct {
		Metafield *Metafield `json:"metafield"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing metafield: %w", err)
	}
	if data.Metafield == nil {
		return nil, fmt.Errorf("metafield %s not found", id)
	}
	return data.Metafield, nil
}

// SetMetafield creates or updates a metafield using metafieldsSet.
func (c *Client) SetMetafield(ownerID, namespace, key, value, metafieldType string) (*Metafield, error) {
	const gql = `
		mutation metafieldsSet($metafields: [MetafieldsSetInput!]!) {
			metafieldsSet(metafields: $metafields) {
				metafields { id namespace key value type updatedAt }
				userErrors { field message }
			}
		}`
	metafields := []map[string]any{
		{
			"ownerId":   ownerID,
			"namespace": namespace,
			"key":       key,
			"value":     value,
			"type":      metafieldType,
		},
	}
	resp, err := c.Do(gql, map[string]any{"metafields": metafields})
	if err != nil {
		return nil, err
	}
	var data struct {
		MetafieldsSet struct {
			Metafields []Metafield `json:"metafields"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"metafieldsSet"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(data.MetafieldsSet.UserErrors); err != nil {
		return nil, err
	}
	if len(data.MetafieldsSet.Metafields) == 0 {
		return nil, fmt.Errorf("no metafield returned")
	}
	m := data.MetafieldsSet.Metafields[0]
	return &m, nil
}

// DeleteMetafield deletes a metafield by ID.
func (c *Client) DeleteMetafield(id string) error {
	const gql = `
		mutation metafieldDelete($input: MetafieldDeleteInput!) {
			metafieldDelete(input: $input) {
				deletedId
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"input": map[string]any{"id": ToGID("Metafield", id)}})
	if err != nil {
		return err
	}
	var data struct {
		MetafieldDelete struct {
			DeletedId  string      `json:"deletedId"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"metafieldDelete"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.MetafieldDelete.UserErrors)
}
