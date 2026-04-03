package channel

import (
	"context"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/channel/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client implements plugins.ChannelPlugin over gRPC.
type Client struct {
	client proto.ChannelServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{client: proto.NewChannelServiceClient(conn)}
}

func (c *Client) GetLifecycle() plugins.Lifecycle { return nil }

func (c *Client) Send(ctx context.Context, channel, content string, metadata map[string]any) (string, error) {
	meta, err := structpb.NewStruct(metadata)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Send(ctx, &proto.SendRequest{
		ChannelId: channel,
		Content:   content,
		Metadata:  meta,
	})
	if err != nil {
		return "", err
	}
	return resp.MessageId, nil
}

func (c *Client) Receive(ctx context.Context) (<-chan plugins.ChannelMessage, error) {
	stream, err := c.client.Receive(ctx, &proto.ReceiveRequest{})
	if err != nil {
		return nil, err
	}

	ch := make(chan plugins.ChannelMessage, 1)
	go func() {
		defer close(ch)
		for {
			evt, err := stream.Recv()
			if err != nil {
				return
			}
			ch <- plugins.ChannelMessage{
				ID:       evt.Id,
				Channel:  evt.ChannelId,
				Author:   evt.AuthorId,
				Content:  evt.Content,
				Metadata: evt.Metadata.AsMap(),
			}
		}
	}()
	return ch, nil
}
