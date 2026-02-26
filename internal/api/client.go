package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const apiVersion = "2026-01"

// Client is the Shopify Admin GraphQL API client.
type Client struct {
	shop        string
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a new Shopify Admin API client.
// shop can be "mystore" or "mystore.myshopify.com".
func NewClient(shop, accessToken string) *Client {
	return &Client{
		shop:        shop,
		accessToken: accessToken,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) endpoint() string {
	shop := c.shop
	if !strings.Contains(shop, ".") {
		shop = shop + ".myshopify.com"
	}
	return fmt.Sprintf("https://%s/admin/api/%s/graphql.json", shop, apiVersion)
}

// Do executes a GraphQL query or mutation.
func (c *Client) Do(query string, variables map[string]any) (*GraphQLResponse, error) {
	payload := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &ShopifyError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return nil, &ShopifyError{
			StatusCode: 200,
			Message:    strings.Join(msgs, "; "),
		}
	}

	return &gqlResp, nil
}

// ToGID converts a plain numeric ID to a Shopify GID.
// If id already starts with "gid://", it is returned as-is.
func ToGID(resourceType, id string) string {
	if strings.HasPrefix(id, "gid://") {
		return id
	}
	return "gid://shopify/" + resourceType + "/" + id
}

// ShortID extracts the numeric portion from a Shopify GID.
// "gid://shopify/Product/1234567890" -> "1234567890"
func ShortID(gid string) string {
	idx := strings.LastIndex(gid, "/")
	if idx >= 0 && idx < len(gid)-1 {
		return gid[idx+1:]
	}
	return gid
}

// userErrorsToError converts a slice of UserError into a single error.
func userErrorsToError(errs []UserError) error {
	if len(errs) == 0 {
		return nil
	}
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Message
	}
	return fmt.Errorf("%s", strings.Join(msgs, "; "))
}
