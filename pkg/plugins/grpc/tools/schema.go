package tools

import (
	"encoding/json"

	"github.com/mwantia/forge-sdk/pkg/plugins"
	proto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/tools/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// ToolParametersToProto converts plugins.ToolParameters to *proto.ToolParametersProto.
func ToolParametersToProto(p plugins.ToolParameters) *proto.ToolParametersProto {
	if len(p.Properties) == 0 && len(p.Required) == 0 {
		return nil
	}
	out := &proto.ToolParametersProto{
		Required: p.Required,
	}
	for name, prop := range p.Properties {
		if out.Properties == nil {
			out.Properties = make(map[string]*proto.ToolPropertyProto)
		}
		out.Properties[name] = &proto.ToolPropertyProto{
			Type:        prop.Type,
			Description: prop.Description,
			Enum:        prop.Enum,
			Format:      prop.Format,
		}
	}
	return out
}

// ProtoToToolParameters converts *proto.ToolParametersProto to plugins.ToolParameters.
func ProtoToToolParameters(p *proto.ToolParametersProto) plugins.ToolParameters {
	if p == nil {
		return plugins.ToolParameters{}
	}
	out := plugins.ToolParameters{
		Type:     "object",
		Required: p.Required,
	}
	for name, prop := range p.Properties {
		if out.Properties == nil {
			out.Properties = make(map[string]plugins.ToolProperties)
		}
		out.Properties[name] = plugins.ToolProperties{
			Type:        prop.Type,
			Description: prop.Description,
			Enum:        prop.Enum,
			Format:      prop.Format,
		}
	}
	return out
}

// toValue converts any value to *structpb.Value via a JSON round-trip.
func toValue(v any) (*structpb.Value, error) {
	if v == nil {
		return structpb.NewNullValue(), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var normalized any
	if err := json.Unmarshal(b, &normalized); err != nil {
		return nil, err
	}
	return structpb.NewValue(normalized)
}

// toStruct converts any JSON-serializable value to *structpb.Struct via a JSON round-trip.
func toStruct(v any) (*structpb.Struct, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var normalized map[string]any
	if err := json.Unmarshal(b, &normalized); err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return structpb.NewStruct(normalized)
}
