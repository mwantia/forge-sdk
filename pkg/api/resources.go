package api

import (
	"context"
	"net/url"

	sdkplugins "github.com/mwantia/forge-sdk/pkg/plugins"
)

type resourceListResponse struct {
	Resources []*sdkplugins.Resource `json:"resources"`
}

func (c *Client) ListResources(ctx context.Context, namespace string) ([]*sdkplugins.Resource, error) {
	var resp resourceListResponse
	if err := c.get("/v1/resources/"+url.PathEscape(namespace), &resp); err != nil {
		return nil, err
	}
	return resp.Resources, nil
}

func (c *Client) GetResource(ctx context.Context, namespace, id string) (*sdkplugins.Resource, error) {
	var res sdkplugins.Resource
	if err := c.get("/v1/resources/"+url.PathEscape(namespace)+"/"+url.PathEscape(id), &res); err != nil {
		return nil, err
	}
	return &res, nil
}
