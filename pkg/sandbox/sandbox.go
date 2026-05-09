package sandbox

import (
	"context"
	"time"
)

// Sandbox provides scoped file and execution access rooted at a single directory.
type Sandbox interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte) error

	MkdirAll(ctx context.Context, path string) error
	ReadDir(ctx context.Context, path string) ([]FileInfo, error)

	Remove(ctx context.Context, path string) error

	RegisterProfile(profile ExecProfile) error
	UnregisterProfile(name string) (bool, error)
	ExecProfile(ctx context.Context, req ExecRequest) (ExecResult, error)

	Root() string
	Close() error
}

// FileInfo describes an entry within the sandbox.
type FileInfo struct {
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}
