package sessions

import (
	"context"
	"time"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/sessions/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client implements plugins.SessionsPlugin over gRPC.
type Client struct {
	plugins.UnimplementedSessionsPlugin
	client proto.SessionsServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{client: proto.NewSessionsServiceClient(conn)}
}

func (c *Client) CreateSession(ctx context.Context, author string, metadata map[string]any) (*plugins.PluginSession, error) {
	meta, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.CreateSession(ctx, &proto.CreateSessionRequest{
		Author:   author,
		Metadata: meta,
	})
	if err != nil {
		return nil, err
	}
	return sessionFromProto(resp), nil
}

func (c *Client) GetSession(ctx context.Context, sessionID string) (*plugins.PluginSession, error) {
	resp, err := c.client.GetSession(ctx, &proto.GetSessionRequest{SessionId: sessionID})
	if err != nil {
		return nil, err
	}
	return sessionFromProto(resp), nil
}

func (c *Client) ListSessions(ctx context.Context, offset, limit int) ([]*plugins.PluginSession, error) {
	resp, err := c.client.ListSessions(ctx, &proto.ListSessionsRequest{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*plugins.PluginSession, 0, len(resp.Sessions))
	for _, s := range resp.Sessions {
		out = append(out, sessionFromProto(s))
	}
	return out, nil
}

func (c *Client) DeleteSession(ctx context.Context, sessionID string) (bool, error) {
	resp, err := c.client.DeleteSession(ctx, &proto.DeleteSessionRequest{SessionId: sessionID})
	if err != nil {
		return false, err
	}
	return resp.Ok, nil
}

func (c *Client) CommitSession(ctx context.Context, sessionID string) (bool, error) {
	resp, err := c.client.CommitSession(ctx, &proto.CommitSessionRequest{SessionId: sessionID})
	if err != nil {
		return false, err
	}
	return resp.Ok, nil
}

func (c *Client) AddMessage(ctx context.Context, sessionID string, msg *plugins.PluginMessage) (bool, error) {
	pm, err := messageToProto(msg)
	if err != nil {
		return false, err
	}
	resp, err := c.client.AddMessage(ctx, &proto.AddMessageRequest{
		SessionId: sessionID,
		Message:   pm,
	})
	if err != nil {
		return false, err
	}
	return resp.Ok, nil
}

func (c *Client) GetMessage(ctx context.Context, sessionID, messageID string) (*plugins.PluginMessage, error) {
	resp, err := c.client.GetMessage(ctx, &proto.GetMessageRequest{
		SessionId: sessionID,
		MessageId: messageID,
	})
	if err != nil {
		return nil, err
	}
	return messageFromProto(resp), nil
}

func (c *Client) ListMessages(ctx context.Context, sessionID string, offset, limit int) ([]*plugins.PluginMessage, error) {
	resp, err := c.client.ListMessages(ctx, &proto.ListMessagesRequest{
		SessionId: sessionID,
		Offset:    int32(offset),
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*plugins.PluginMessage, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		out = append(out, messageFromProto(m))
	}
	return out, nil
}

func (c *Client) CountMessages(ctx context.Context, sessionID string) (int, error) {
	resp, err := c.client.CountMessages(ctx, &proto.CountMessagesRequest{SessionId: sessionID})
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

func (c *Client) CompactMessages(ctx context.Context, sessionID string) (int, error) {
	resp, err := c.client.CompactMessages(ctx, &proto.CompactMessagesRequest{SessionId: sessionID})
	if err != nil {
		return 0, err
	}
	return int(resp.Removed), nil
}

func sessionFromProto(p *proto.SessionProto) *plugins.PluginSession {
	if p == nil {
		return nil
	}
	return &plugins.PluginSession{
		ID:           p.Id,
		Author:       p.Author,
		Committed:    p.Committed,
		Archived:     p.Archived,
		MessageCount: int(p.MessageCount),
		Metadata:     p.Metadata.AsMap(),
	}
}

func sessionToProto(s *plugins.PluginSession) (*proto.SessionProto, error) {
	if s == nil {
		return nil, nil
	}
	meta, err := structpb.NewStruct(s.Metadata)
	if err != nil {
		return nil, err
	}
	return &proto.SessionProto{
		Id:           s.ID,
		Author:       s.Author,
		Committed:    s.Committed,
		Archived:     s.Archived,
		MessageCount: int32(s.MessageCount),
		Metadata:     meta,
	}, nil
}

func messageFromProto(p *proto.MessageProto) *plugins.PluginMessage {
	if p == nil {
		return nil
	}
	m := &plugins.PluginMessage{
		ID:        p.Id,
		SessionID: p.SessionId,
		Role:      p.Role,
		Content:   p.Content,
	}
	if p.CreatedAtUnixNano != 0 {
		m.CreatedAt = time.Unix(0, p.CreatedAtUnixNano)
	}
	for _, tc := range p.ToolCalls {
		m.ToolCalls = append(m.ToolCalls, plugins.PluginToolCall{
			ID:        tc.Id,
			Name:      tc.Name,
			Arguments: tc.Arguments.AsMap(),
			Result:    tc.Result,
			IsError:   tc.IsError,
		})
	}
	return m
}

func messageToProto(m *plugins.PluginMessage) (*proto.MessageProto, error) {
	if m == nil {
		return nil, nil
	}
	pm := &proto.MessageProto{
		Id:                 m.ID,
		SessionId:          m.SessionID,
		Role:               m.Role,
		Content:            m.Content,
		CreatedAtUnixNano:  m.CreatedAt.UnixNano(),
	}
	for _, tc := range m.ToolCalls {
		args, err := structpb.NewStruct(tc.Arguments)
		if err != nil {
			return nil, err
		}
		pm.ToolCalls = append(pm.ToolCalls, &proto.ToolCallProto{
			Id:        tc.ID,
			Name:      tc.Name,
			Arguments: args,
			Result:    tc.Result,
			IsError:   tc.IsError,
		})
	}
	return pm, nil
}
