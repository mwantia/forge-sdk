package resource

import (
	"context"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/resource/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client implements plugins.ResourcePlugin over gRPC.
type Client struct {
	plugins.UnimplementedResourcePlugin
	client proto.ResourceServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{client: proto.NewResourceServiceClient(conn)}
}

func (c *Client) Store(ctx context.Context, path, content string, tags []string, metadata map[string]any) (*plugins.Resource, error) {
	meta, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Store(ctx, &proto.StoreRequest{
		Path:     path,
		Content:  content,
		Tags:     tags,
		Metadata: meta,
	})
	if err != nil {
		return nil, err
	}
	return resourceFromProto(resp.Resource), nil
}

func (c *Client) Recall(ctx context.Context, q plugins.RecallQuery) ([]*plugins.Resource, error) {
	pq := &proto.RecallQuery{
		Path:  q.Path,
		Query: q.Query,
		Tags:  q.Tags,
		Limit: int32(q.Limit),
	}
	for _, f := range q.Filter {
		val, _ := structpb.NewValue(f.Value)
		pq.Filter = append(pq.Filter, &proto.FilterPredicate{
			Key:   f.Key,
			Op:    string(f.Op),
			Value: val,
		})
	}
	if !q.CreatedAfter.IsZero() {
		pq.CreatedAfter = timestamppb.New(q.CreatedAfter)
	}
	if !q.CreatedBefore.IsZero() {
		pq.CreatedBefore = timestamppb.New(q.CreatedBefore)
	}

	resp, err := c.client.Recall(ctx, &proto.RecallRequest{Query: pq})
	if err != nil {
		return nil, err
	}
	results := make([]*plugins.Resource, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, resourceFromProto(r))
	}
	return results, nil
}

func (c *Client) Forget(ctx context.Context, path, id string) error {
	_, err := c.client.Forget(ctx, &proto.ForgetRequest{Path: path, Id: id})
	return err
}

func resourceFromProto(p *proto.Resource) *plugins.Resource {
	if p == nil {
		return nil
	}
	r := &plugins.Resource{
		ID:      p.Id,
		Path:    p.Path,
		Content: p.Content,
		Tags:    p.Tags,
		Score:   p.Score,
	}
	if p.Metadata != nil {
		r.Metadata = p.Metadata.AsMap()
	}
	if p.CreatedAt != nil {
		r.CreatedAt = p.CreatedAt.AsTime()
	}
	return r
}
