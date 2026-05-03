package plugins

import (
	"context"
	"time"

	"github.com/mwantia/forge-sdk/pkg/errors"
)

type FilterOp string

const (
	FilterOpEq       FilterOp = "eq"
	FilterOpPrefix   FilterOp = "prefix"
	FilterOpContains FilterOp = "contains"
	FilterOpGte      FilterOp = "gte"
	FilterOpLte      FilterOp = "lte"
)

type FilterPredicate struct {
	Key   string   `json:"key"`
	Op    FilterOp `json:"op"`
	Value any      `json:"value"`
}

// RecallQuery is the structured query passed to Recall. All fields are
// optional except Path; zero values mean "no constraint".
type RecallQuery struct {
	// Path is an exact path or a glob pattern (*, **, ? supported).
	// Examples: "/sessions/abc123", "/sessions/**", "/archives/**/consul*".
	Path string `json:"path"`

	// Query is a content text search string. Empty = no content filter.
	Query string `json:"query,omitempty"`

	// Tags filters resources by tag (AND: resource must carry all listed tags).
	Tags []string `json:"tags,omitempty"`

	// Filter is a list of typed metadata predicates (AND).
	Filter []FilterPredicate `json:"filter,omitempty"`

	// CreatedAfter and CreatedBefore are inclusive time bounds.
	// Zero value means unbounded.
	CreatedAfter  time.Time `json:"created_after,omitempty"`
	CreatedBefore time.Time `json:"created_before,omitempty"`

	// Limit caps the result count. 0 means backend default (typically 5).
	Limit int `json:"limit,omitempty"`
}

// ResourcePlugin owns path-based resource storage and retrieval.
type ResourcePlugin interface {
	BasePlugin

	Store(ctx context.Context, path, content string, tags []string, metadata map[string]any) (*Resource, error)
	Recall(ctx context.Context, q RecallQuery) ([]*Resource, error)
	Forget(ctx context.Context, path, id string) error
}

type Resource struct {
	ID        string         `json:"id"`
	Path      string         `json:"path,omitempty"`
	Content   string         `json:"content"`
	Tags      []string       `json:"tags,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Score     float64        `json:"score,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
}

// UnimplementedResourcePlugin can be embedded to satisfy ResourcePlugin
// with default implementations that return ErrPluginCapabilityNotSupported.
type UnimplementedResourcePlugin struct{}

func (UnimplementedResourcePlugin) GetLifecycle() Lifecycle { return nil }

func (UnimplementedResourcePlugin) Store(_ context.Context, _, _ string, _ []string, _ map[string]any) (*Resource, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedResourcePlugin) Recall(_ context.Context, _ RecallQuery) ([]*Resource, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedResourcePlugin) Forget(_ context.Context, _, _ string) error {
	return errors.ErrPluginCapabilityNotSupported
}

var _ ResourcePlugin = (*UnimplementedResourcePlugin)(nil)
