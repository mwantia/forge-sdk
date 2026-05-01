package resource

import (
	"context"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/resource/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client implements plugins.ResourcePlugin over gRPC.
type Client struct {
	plugins.UnimplementedResourcePlugin
	client proto.ResourceServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{client: proto.NewResourceServiceClient(conn)}
}

func (c *Client) Store(ctx context.Context, namespace, content string, metadata map[string]any) (*plugins.Resource, error) {
	meta, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Store(ctx, &proto.StoreRequest{
		Namespace: namespace,
		Content:   content,
		Metadata:  meta,
	})
	if err != nil {
		return nil, err
	}
	return resourceFromProto(resp.Resource), nil
}

func (c *Client) Recall(ctx context.Context, namespace, query string, limit int, filter map[string]any) ([]*plugins.Resource, error) {
	filterStruct, err := structpb.NewStruct(filter)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Recall(ctx, &proto.RecallRequest{
		Namespace: namespace,
		Query:     query,
		Limit:     int32(limit),
		Filter:    filterStruct,
	})
	if err != nil {
		return nil, err
	}
	results := make([]*plugins.Resource, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, resourceFromProto(r))
	}
	return results, nil
}

func (c *Client) Forget(ctx context.Context, namespace, id string) error {
	_, err := c.client.Forget(ctx, &proto.ForgetRequest{Namespace: namespace, Id: id})
	return err
}

func resourceFromProto(p *proto.Resource) *plugins.Resource {
	if p == nil {
		return nil
	}
	r := &plugins.Resource{
		ID:        p.Id,
		Namespace: p.Namespace,
		Content:   p.Content,
		Score:     p.Score,
	}
	if p.Metadata != nil {
		r.Metadata = p.Metadata.AsMap()
	}
	if p.CreatedAt != nil {
		r.CreatedAt = p.CreatedAt.AsTime()
	}
	return r
}
