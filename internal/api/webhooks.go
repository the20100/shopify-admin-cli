package api

import (
	"encoding/json"
	"fmt"
)

// ListWebhooks returns a paginated list of webhook subscriptions.
func (c *Client) ListWebhooks(first int, after string) (*WebhookSubscriptionConnection, error) {
	const gql = `
		query ListWebhooks($first: Int!, $after: String) {
			webhookSubscriptions(first: $first, after: $after) {
				edges {
					cursor
					node {
						id topic format createdAt
						endpoint {
							__typename
							... on WebhookHttpEndpoint { callbackUrl }
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
	resp, err := c.Do(gql, vars)
	if err != nil {
		return nil, err
	}
	// Parse with inline endpoint handling
	var rawData struct {
		WebhookSubscriptions struct {
			Edges []struct {
				Cursor string `json:"cursor"`
				Node   struct {
					ID        string `json:"id"`
					Topic     string `json:"topic"`
					Format    string `json:"format"`
					CreatedAt string `json:"createdAt"`
					Endpoint  struct {
						CallbackURL string `json:"callbackUrl"`
					} `json:"endpoint"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"webhookSubscriptions"`
	}
	if err := json.Unmarshal(resp.Data, &rawData); err != nil {
		return nil, fmt.Errorf("parsing webhooks: %w", err)
	}
	conn := &WebhookSubscriptionConnection{
		PageInfo: rawData.WebhookSubscriptions.PageInfo,
	}
	for _, e := range rawData.WebhookSubscriptions.Edges {
		conn.Edges = append(conn.Edges, WebhookSubscriptionEdge{
			Cursor: e.Cursor,
			Node: WebhookSubscription{
				ID:        e.Node.ID,
				Topic:     e.Node.Topic,
				Format:    e.Node.Format,
				CreatedAt: e.Node.CreatedAt,
				Endpoint:  WebhookEndpoint{CallbackURL: e.Node.Endpoint.CallbackURL},
			},
		})
	}
	return conn, nil
}

// CreateWebhook creates a new HTTP webhook subscription.
func (c *Client) CreateWebhook(topic, callbackURL, format string) (*WebhookSubscription, error) {
	const gql = `
		mutation webhookSubscriptionCreate($topic: WebhookSubscriptionTopic!, $webhookSubscription: WebhookSubscriptionInput!) {
			webhookSubscriptionCreate(topic: $topic, webhookSubscription: $webhookSubscription) {
				webhookSubscription {
					id topic format createdAt
					endpoint {
						... on WebhookHttpEndpoint { callbackUrl }
					}
				}
				userErrors { field message }
			}
		}`
	if format == "" {
		format = "JSON"
	}
	resp, err := c.Do(gql, map[string]any{
		"topic": topic,
		"webhookSubscription": map[string]any{
			"callbackUrl": callbackURL,
			"format":      format,
		},
	})
	if err != nil {
		return nil, err
	}
	var rawData struct {
		WebhookSubscriptionCreate struct {
			WebhookSubscription *struct {
				ID        string `json:"id"`
				Topic     string `json:"topic"`
				Format    string `json:"format"`
				CreatedAt string `json:"createdAt"`
				Endpoint  struct {
					CallbackURL string `json:"callbackUrl"`
				} `json:"endpoint"`
			} `json:"webhookSubscription"`
			UserErrors []UserError `json:"userErrors"`
		} `json:"webhookSubscriptionCreate"`
	}
	if err := json.Unmarshal(resp.Data, &rawData); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if err := userErrorsToError(rawData.WebhookSubscriptionCreate.UserErrors); err != nil {
		return nil, err
	}
	if rawData.WebhookSubscriptionCreate.WebhookSubscription == nil {
		return nil, fmt.Errorf("no webhook returned")
	}
	wh := rawData.WebhookSubscriptionCreate.WebhookSubscription
	return &WebhookSubscription{
		ID:        wh.ID,
		Topic:     wh.Topic,
		Format:    wh.Format,
		CreatedAt: wh.CreatedAt,
		Endpoint:  WebhookEndpoint{CallbackURL: wh.Endpoint.CallbackURL},
	}, nil
}

// DeleteWebhook deletes a webhook subscription.
func (c *Client) DeleteWebhook(id string) error {
	const gql = `
		mutation webhookSubscriptionDelete($id: ID!) {
			webhookSubscriptionDelete(id: $id) {
				deletedWebhookSubscriptionId
				userErrors { field message }
			}
		}`
	resp, err := c.Do(gql, map[string]any{"id": ToGID("WebhookSubscription", id)})
	if err != nil {
		return err
	}
	var data struct {
		WebhookSubscriptionDelete struct {
			DeletedWebhookSubscriptionId string      `json:"deletedWebhookSubscriptionId"`
			UserErrors                   []UserError `json:"userErrors"`
		} `json:"webhookSubscriptionDelete"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return userErrorsToError(data.WebhookSubscriptionDelete.UserErrors)
}
