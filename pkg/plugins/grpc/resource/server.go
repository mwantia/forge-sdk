package resource

import (
	"context"
	"fmt"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/resource/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements ResourceServiceServer, bridging gRPC to the
// ResourcePlugin interface.
type Server struct {
	proto.UnimplementedResourceServiceServer
	impl plugins.Driver
}

func NewServer(impl plugins.Driver) *Server {
	return &Server{impl: impl}
}

func (s *Server) plugin(ctx context.Context) (plugins.ResourcePlugin, error) {
	p, err := s.impl.GetResourcePlugin(ctx)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("resource plugin not available")
	}
	return p, nil
}

func (s *Server) Store(ctx context.Context, req *proto.StoreRequest) (*proto.StoreResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	resource, err := p.Store(ctx, req.Namespace, req.Content, req.Metadata.AsMap())
	if err != nil {
		return nil, err
	}
	rp, err := resourceToProto(resource)
	if err != nil {
		return nil, err
	}
	return &proto.StoreResponse{Resource: rp}, nil
}

func (s *Server) Recall(ctx context.Context, req *proto.RecallRequest) (*proto.RecallResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	resources, err := p.Recall(ctx, req.Namespace, req.Query, int(req.Limit), req.Filter.AsMap())
	if err != nil {
		return nil, err
	}
	resp := &proto.RecallResponse{}
	for _, r := range resources {
		rp, err := resourceToProto(r)
		if err != nil {
			return nil, err
		}
		resp.Results = append(resp.Results, rp)
	}
	return resp, nil
}

func (s *Server) Forget(ctx context.Context, req *proto.ForgetRequest) (*proto.ForgetResponse, error) {
	p, err := s.plugin(ctx)
	if err != nil {
		return nil, err
	}
	if err := p.Forget(ctx, req.Namespace, req.Id); err != nil {
		return nil, err
	}
	return &proto.ForgetResponse{}, nil
}

func resourceToProto(r *plugins.Resource) (*proto.Resource, error) {
	if r == nil {
		return nil, nil
	}
	out := &proto.Resource{
		Id:        r.ID,
		Namespace: r.Namespace,
		Content:   r.Content,
		Score:     r.Score,
	}
	if len(r.Metadata) > 0 {
		meta, err := structpb.NewStruct(r.Metadata)
		if err != nil {
			return nil, err
		}
		out.Metadata = meta
	}
	if !r.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(r.CreatedAt)
	}
	return out, nil
}
