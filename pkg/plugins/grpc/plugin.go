package grpc

import (
	"context"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/mwantia/forge-sdk/pkg/plugins"
	channelgrpc "github.com/mwantia/forge-sdk/pkg/plugins/grpc/channel"
	channelproto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/channel/proto"
	drivergrpc "github.com/mwantia/forge-sdk/pkg/plugins/grpc/driver"
	driverproto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/driver/proto"
	providergrpc "github.com/mwantia/forge-sdk/pkg/plugins/grpc/provider"
	providerproto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/provider/proto"
	resourcegrpc "github.com/mwantia/forge-sdk/pkg/plugins/grpc/resource"
	resourceproto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/resource/proto"
	sandboxgrpc "github.com/mwantia/forge-sdk/pkg/plugins/grpc/sandbox"
	sandboxproto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/sandbox/proto"
	toolsgrpc "github.com/mwantia/forge-sdk/pkg/plugins/grpc/tools"
	toolsproto "github.com/mwantia/forge-sdk/pkg/plugins/grpc/tools/proto"
	"google.golang.org/grpc"
)

// Handshake is the plugin handshake configuration.
// Plugins and hosts must use the same values to communicate.
var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  2,
	MagicCookieKey:   "FORGE_PLUGIN",
	MagicCookieValue: "forge",
}

// Plugins is the map of supported plugin types for use with go-plugin.
var Plugins = map[string]goplugin.Plugin{
	"driver": &DriverPlugin{},
}

// DriverPlugin is the hashicorp/go-plugin wrapper for the Driver interface.
type DriverPlugin struct {
	goplugin.Plugin
	Impl plugins.Driver
}

func (p *DriverPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	driverproto.RegisterDriverServiceServer(s, drivergrpc.NewServer(p.Impl, broker))
	providerproto.RegisterProviderServiceServer(s, providergrpc.NewServer(p.Impl))
	resourceproto.RegisterResourceServiceServer(s, resourcegrpc.NewServer(p.Impl))
	channelproto.RegisterChannelServiceServer(s, channelgrpc.NewServer(p.Impl))
	toolsproto.RegisterToolsServiceServer(s, toolsgrpc.NewServer(p.Impl))
	sandboxproto.RegisterSandboxServiceServer(s, sandboxgrpc.NewServer(p.Impl))
	return nil
}

func (p *DriverPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
	return drivergrpc.NewClient(c, broker), nil
}
