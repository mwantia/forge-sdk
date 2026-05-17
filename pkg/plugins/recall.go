package plugins

import "context"

// RecallPlugin owns the semantic index layer for stored resources.
// It is intentionally separate from ResourcePlugin: storage is always
// the built-in DAG; search is pluggable (or built-in HNSW).
type RecallPlugin interface {
	BasePlugin

	// Index is called after every successful resource store.
	// The plugin computes and stores an embedding or other index entry.
	// Failure here is non-fatal: the content is safe in the DAG; the
	// resource will simply be absent from semantic recall until re-indexed.
	Index(ctx context.Context, req IndexRequest) error

	// Recall returns content hashes ranked by relevance.
	// The agent resolves hashes to full content via the DAG object store.
	// The plugin never stores or returns content directly.
	Recall(ctx context.Context, q RecallQuery) ([]RankedHash, error)

	// Forget removes the index entry for hash.
	// Called on ForgetResource. Idempotent — missing entries are not errors.
	Forget(ctx context.Context, hash string) error
}

// IndexRequest carries all context needed for the plugin to build an index entry.
type IndexRequest struct {
	Hash        string         `json:"hash"`
	ContentType string         `json:"content_type"`
	Content     string         `json:"content"`
	Namespace   string         `json:"namespace"`
	Name        string         `json:"name"`
	Tags        []string       `json:"tags,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// RankedHash is a content hash with its relevance score.
type RankedHash struct {
	Hash  string  `json:"hash"`
	Score float64 `json:"score"`
}

// UnimplementedRecallPlugin satisfies RecallPlugin with no-op stubs.
type UnimplementedRecallPlugin struct{}

func (UnimplementedRecallPlugin) GetLifecycle() Lifecycle { return nil }

func (UnimplementedRecallPlugin) Index(_ context.Context, _ IndexRequest) error { return nil }

func (UnimplementedRecallPlugin) Recall(_ context.Context, _ RecallQuery) ([]RankedHash, error) {
	return nil, nil
}

func (UnimplementedRecallPlugin) Forget(_ context.Context, _ string) error { return nil }

var _ RecallPlugin = (*UnimplementedRecallPlugin)(nil)
