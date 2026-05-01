package plugins

import (
	"context"
	"time"

	"github.com/mwantia/forge-sdk/pkg/errors"
)

// ResourcePlugin owns namespaced resource storage and retrieval — the
// surface formerly known as MemoryPlugin. Verbs are deliberately narrow
// (Store/Recall/Forget); broader querying (ListAll, Fetch) is deferred
// until a caller needs it.
type ResourcePlugin interface {
	BasePlugin

	Store(ctx context.Context, namespace, content string, metadata map[string]any) (*Resource, error)
	Recall(ctx context.Context, namespace, query string, limit int, filter map[string]any) ([]*Resource, error)
	Forget(ctx context.Context, namespace, id string) error
}

type Resource struct {
	ID        string         `json:"id"`
	Namespace string         `json:"namespace,omitempty"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Score     float64        `json:"score,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
}

// UnimplementedResourcePlugin can be embedded to satisfy ResourcePlugin
// with default implementations that return ErrPluginCapabilityNotSupported.
type UnimplementedResourcePlugin struct{}

func (UnimplementedResourcePlugin) GetLifecycle() Lifecycle { return nil }

func (UnimplementedResourcePlugin) Store(_ context.Context, _, _ string, _ map[string]any) (*Resource, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedResourcePlugin) Recall(_ context.Context, _, _ string, _ int, _ map[string]any) ([]*Resource, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedResourcePlugin) Forget(_ context.Context, _, _ string) error {
	return errors.ErrPluginCapabilityNotSupported
}

var _ ResourcePlugin = (*UnimplementedResourcePlugin)(nil)
