package api

import (
	"context"
	"strings"
	"time"

	sdkplugins "github.com/mwantia/forge-sdk/pkg/plugins"
)

// ResourceRevision is one entry in a resource's parent-chain history.
type ResourceRevision struct {
	Hash      string         `json:"hash"`
	CreatedAt time.Time      `json:"created_at"`
	Tags      []string       `json:"tags,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	IndexedAt *time.Time     `json:"indexed_at,omitempty"`
	IndexedBy string         `json:"indexed_by,omitempty"`
}

type resourceListResponse struct {
	Resources []*sdkplugins.Resource `json:"resources"`
}

type resourceHistoryResponse struct {
	History []*ResourceRevision `json:"history"`
}

// resourcePath normalizes a path for use in a URL: strips leading slash so it
// can be safely appended after "/v1/resources/".
func resourcePath(path string) string {
	return strings.TrimPrefix(path, "/")
}

func (c *Client) ListResources(ctx context.Context, path string) ([]*sdkplugins.Resource, error) {
	var resp resourceListResponse
	if err := c.get("/v1/resources/"+resourcePath(path), &resp); err != nil {
		return nil, err
	}
	return resp.Resources, nil
}

func (c *Client) GetResource(ctx context.Context, path, name string) (*sdkplugins.Resource, error) {
	var res sdkplugins.Resource
	if err := c.get("/v1/resources/"+resourcePath(path)+"?id="+name, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type StoreResourceRequest struct {
	Name     string         `json:"name,omitempty"`
	Content  string         `json:"content"`
	Tags     []string       `json:"tags,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (c *Client) StoreResource(ctx context.Context, path string, req StoreResourceRequest) (*sdkplugins.Resource, error) {
	var res sdkplugins.Resource
	if err := c.put("/v1/resources/"+resourcePath(path), req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type RecallResourcesRequest struct {
	Query         string                    `json:"query,omitempty"`
	Tags          []string                  `json:"tags,omitempty"`
	Filter        []sdkplugins.FilterPredicate `json:"filter,omitempty"`
	CreatedAfter  *time.Time                `json:"created_after,omitempty"`
	CreatedBefore *time.Time                `json:"created_before,omitempty"`
	Limit         int                       `json:"limit,omitempty"`
}

func (c *Client) RecallResources(ctx context.Context, path string, req RecallResourcesRequest) ([]*sdkplugins.Resource, error) {
	var resp resourceListResponse
	if err := c.post("/v1/resources/"+resourcePath(path), req, &resp); err != nil {
		return nil, err
	}
	return resp.Resources, nil
}

func (c *Client) ForgetResource(ctx context.Context, path, name string) error {
	return c.delete("/v1/resources/" + resourcePath(path) + "?id=" + name)
}

func (c *Client) ListResourceHistory(ctx context.Context, path, name string) ([]*ResourceRevision, error) {
	var resp resourceHistoryResponse
	if err := c.get("/v1/resources/"+resourcePath(path)+"/history?name="+name, &resp); err != nil {
		return nil, err
	}
	return resp.History, nil
}
