package plugins

import (
	"context"
	"time"

	"github.com/mwantia/forge-sdk/pkg/errors"
)

// SessionsPlugin owns session + message lifecycle for plugins that manage
// their own storage (e.g. OpenViking). Sessions and messages form one
// feature: a plugin either supports both or neither.
type SessionsPlugin interface {
	BasePlugin

	CreateSession(ctx context.Context, author string, metadata map[string]any) (*PluginSession, error)
	GetSession(ctx context.Context, sessionID string) (*PluginSession, error)
	ListSessions(ctx context.Context, offset, limit int) ([]*PluginSession, error)
	DeleteSession(ctx context.Context, sessionID string) (bool, error)
	CommitSession(ctx context.Context, sessionID string) (bool, error)

	AddMessage(ctx context.Context, sessionID string, msg *PluginMessage) (bool, error)
	GetMessage(ctx context.Context, sessionID, messageID string) (*PluginMessage, error)
	ListMessages(ctx context.Context, sessionID string, offset, limit int) ([]*PluginMessage, error)
	CountMessages(ctx context.Context, sessionID string) (int, error)
	CompactMessages(ctx context.Context, sessionID string) (int, error)
}

type PluginSession struct {
	ID           string         `json:"id"`
	Author       string         `json:"author"`
	Committed    bool           `json:"committed,omitempty"`
	Archived     bool           `json:"archived,omitempty"`
	MessageCount int            `json:"message_count"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type PluginMessage struct {
	ID        string             `json:"id"`
	SessionID string             `json:"session_id"`
	Role      string             `json:"role"`
	Content   string             `json:"content"`
	ToolCalls []PluginToolCall   `json:"tool_calls,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
}

type PluginToolCall struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Result    string         `json:"result,omitempty"`
	IsError   bool           `json:"is_error,omitempty"`
}

// UnimplementedSessionsPlugin can be embedded to satisfy SessionsPlugin with
// default implementations that return errors.ErrPluginCapabilityNotSupported.
type UnimplementedSessionsPlugin struct{}

func (UnimplementedSessionsPlugin) GetLifecycle() Lifecycle { return nil }

func (UnimplementedSessionsPlugin) CreateSession(_ context.Context, _ string, _ map[string]any) (*PluginSession, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) GetSession(_ context.Context, _ string) (*PluginSession, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) ListSessions(_ context.Context, _, _ int) ([]*PluginSession, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) DeleteSession(_ context.Context, _ string) (bool, error) {
	return false, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) CommitSession(_ context.Context, _ string) (bool, error) {
	return false, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) AddMessage(_ context.Context, _ string, _ *PluginMessage) (bool, error) {
	return false, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) GetMessage(_ context.Context, _, _ string) (*PluginMessage, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) ListMessages(_ context.Context, _ string, _, _ int) ([]*PluginMessage, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) CountMessages(_ context.Context, _ string) (int, error) {
	return 0, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedSessionsPlugin) CompactMessages(_ context.Context, _ string) (int, error) {
	return 0, errors.ErrPluginCapabilityNotSupported
}

var _ SessionsPlugin = (*UnimplementedSessionsPlugin)(nil)
