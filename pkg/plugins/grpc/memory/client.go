package memory

import (
	"context"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/memory/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client implements plugins.MemoryPlugin over gRPC.
type Client struct {
	plugins.UnimplementedMemoryPlugin
	client proto.MemoryServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{client: proto.NewMemoryServiceClient(conn)}
}

func (c *Client) StoreResource(ctx context.Context, sessionID, content string, metadata map[string]any) (*plugins.MemoryResource, error) {
	meta, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Store(ctx, &proto.StoreRequest{
		Content:   content,
		Namespace: sessionID,
		Metadata:  meta,
	})
	if err != nil {
		return nil, err
	}
	return &plugins.MemoryResource{ID: resp.Id, Content: content}, nil
}

func (c *Client) RetrieveResource(ctx context.Context, sessionID, query string, limit int, filter map[string]any) ([]*plugins.MemoryResource, error) {
	filterStruct, err := structpb.NewStruct(filter)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Retrieve(ctx, &proto.RetrieveRequest{
		Query:     query,
		Limit:     int32(limit),
		Namespace: sessionID,
		Filter:    filterStruct,
	})
	if err != nil {
		return nil, err
	}

	var results []*plugins.MemoryResource
	for _, r := range resp.Results {
		results = append(results, &plugins.MemoryResource{
			ID:       r.Id,
			Content:  r.Content,
			Score:    r.Score,
			Metadata: r.Metadata.AsMap(),
		})
	}
	return results, nil
}
