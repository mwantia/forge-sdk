package plugins

import (
	"context"

	"github.com/mwantia/forge-sdk/pkg/errors"
)

// MemoryPlugin acts as resource-level memory management (semantic store + retrieve).
// Session and message lifecycle belongs to SessionsPlugin.
type MemoryPlugin interface {
	BasePlugin

	StoreResource(ctx context.Context, namespace, content string, metadata map[string]any) (*MemoryResource, error)
	RetrieveResource(ctx context.Context, namespace, query string, limit int, filter map[string]any) ([]*MemoryResource, error)
}

type MemoryResource struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Score    float64        `json:"score"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type StoreRequest struct {
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Namespace string         `json:"namespace,omitempty"`
}

// UnimplementedMemoryPlugin can be embedded to satisfy MemoryPlugin with default
// implementations that return errors.ErrPluginCapabilityNotSupported.
type UnimplementedMemoryPlugin struct{}

func (UnimplementedMemoryPlugin) GetLifecycle() Lifecycle { return nil }

func (UnimplementedMemoryPlugin) StoreResource(_ context.Context, _, _ string, _ map[string]any) (*MemoryResource, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedMemoryPlugin) RetrieveResource(_ context.Context, _, _ string, _ int, _ map[string]any) ([]*MemoryResource, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

var _ MemoryPlugin = (*UnimplementedMemoryPlugin)(nil)
