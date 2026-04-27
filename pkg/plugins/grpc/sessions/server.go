package sessions

import (
	"context"
	"fmt"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/sessions/proto"
)

// Server implements SessionsServiceServer, bridging gRPC to the SessionsPlugin interface.
type Server struct {
	proto.UnimplementedSessionsServiceServer
	impl plugins.Driver
}

func NewServer(impl plugins.Driver) *Server {
	return &Server{impl: impl}
}

func (s *Server) plugin(ctx context.Context) (plugins.SessionsPlugin, error) {
	p, err := s.impl.GetSessionsPlugin(ctx)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("sessions plugin not available")
	}
	return p, nil
}

func (s *Server) CreateSession(ctx context.Context, req *proto.CreateSessionRequest) (*proto.SessionProto, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	sess, err := p.CreateSession(ctx, req.Author, req.Metadata.AsMap())
	if err != nil {
		return nil, err
	}
	return sessionToProto(sess)
}

func (s *Server) GetSession(ctx context.Context, req *proto.GetSessionRequest) (*proto.SessionProto, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	sess, err := p.GetSession(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return sessionToProto(sess)
}

func (s *Server) ListSessions(ctx context.Context, req *proto.ListSessionsRequest) (*proto.ListSessionsResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	list, err := p.ListSessions(ctx, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, err
	}
	resp := &proto.ListSessionsResponse{}
	for _, sess := range list {
		ps, err := sessionToProto(sess)
		if err != nil {
			return nil, err
		}
		resp.Sessions = append(resp.Sessions, ps)
	}
	return resp, nil
}

func (s *Server) DeleteSession(ctx context.Context, req *proto.DeleteSessionRequest) (*proto.BoolResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := p.DeleteSession(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return &proto.BoolResponse{Ok: ok}, nil
}

func (s *Server) CommitSession(ctx context.Context, req *proto.CommitSessionRequest) (*proto.BoolResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := p.CommitSession(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return &proto.BoolResponse{Ok: ok}, nil
}

func (s *Server) AddMessage(ctx context.Context, req *proto.AddMessageRequest) (*proto.BoolResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := p.AddMessage(ctx, req.SessionId, messageFromProto(req.Message))
	if err != nil {
		return nil, err
	}
	return &proto.BoolResponse{Ok: ok}, nil
}

func (s *Server) GetMessage(ctx context.Context, req *proto.GetMessageRequest) (*proto.MessageProto, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := p.GetMessage(ctx, req.SessionId, req.MessageId)
	if err != nil {
		return nil, err
	}
	return messageToProto(msg)
}

func (s *Server) ListMessages(ctx context.Context, req *proto.ListMessagesRequest) (*proto.ListMessagesResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	list, err := p.ListMessages(ctx, req.SessionId, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, err
	}
	resp := &proto.ListMessagesResponse{}
	for _, m := range list {
		pm, err := messageToProto(m)
		if err != nil {
			return nil, err
		}
		resp.Messages = append(resp.Messages, pm)
	}
	return resp, nil
}

func (s *Server) CountMessages(ctx context.Context, req *proto.CountMessagesRequest) (*proto.CountResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	n, err := p.CountMessages(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return &proto.CountResponse{Count: int32(n)}, nil
}

func (s *Server) CompactMessages(ctx context.Context, req *proto.CompactMessagesRequest) (*proto.CompactResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	n, err := p.CompactMessages(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return &proto.CompactResponse{Removed: int32(n)}, nil
}
