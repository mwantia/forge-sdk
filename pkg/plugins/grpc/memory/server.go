package memory

import (
	"context"
	"fmt"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/memory/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Server implements MemoryServiceServer, bridging gRPC to the MemoryPlugin interface.
type Server struct {
	proto.UnimplementedMemoryServiceServer
	impl plugins.Driver
}

func NewServer(impl plugins.Driver) *Server {
	return &Server{impl: impl}
}

func (s *Server) Store(ctx context.Context, req *proto.StoreRequest) (*proto.StoreResponse, error) {
	plugin, err := s.impl.GetMemoryPlugin(ctx)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, fmt.Errorf("memory plugin not available")
	}

	resource, err := plugin.StoreResource(ctx, req.Namespace, req.Content, req.Metadata.AsMap())
	if err != nil {
		return nil, err
	}
	return &proto.StoreResponse{Id: resource.ID}, nil
}

func (s *Server) Retrieve(ctx context.Context, req *proto.RetrieveRequest) (*proto.RetrieveResponse, error) {
	plugin, err := s.impl.GetMemoryPlugin(ctx)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, fmt.Errorf("memory plugin not available")
	}

	resources, err := plugin.RetrieveResource(ctx, req.Namespace, req.Query, int(req.Limit), req.Filter.AsMap())
	if err != nil {
		return nil, err
	}

	protoResp := &proto.RetrieveResponse{}
	for _, r := range resources {
		meta, err := structpb.NewStruct(r.Metadata)
		if err != nil {
			return nil, err
		}
		protoResp.Results = append(protoResp.Results, &proto.MemoryResultProto{
			Id:       r.ID,
			Content:  r.Content,
			Score:    r.Score,
			Metadata: meta,
		})
	}
	return protoResp, nil
}
