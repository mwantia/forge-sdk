package api

import "time"

// SystemPromptEntry is a named system prompt segment.
// Multiple entries are joined in order into a single system message.
type SystemPromptEntry struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// SessionMetadata holds the persistent metadata for a pipeline session.
type SessionMetadata struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Title         string               `json:"title,omitempty"`
	Description   string               `json:"description,omitempty"`
	Parent        string               `json:"parent,omitempty"`
	Model         string               `json:"model"`
	SystemPrompts []SystemPromptEntry  `json:"system_prompts,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// Message is a single stored message in a session.
type Message struct {
	ID        string            `json:"id"`
	Role      string            `json:"role"`
	Content   string            `json:"content"`
	ToolCalls []MessageToolCall `json:"tool_calls,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// MessageToolCall records a tool invocation within a message.
type MessageToolCall struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Result    string         `json:"result,omitempty"`
	IsError   bool           `json:"is_error,omitempty"`
}

// CreateSessionRequest is the POST body for creating a new session.
type CreateSessionRequest struct {
	Name              string               `json:"name,omitempty"`
	Model             string               `json:"model"`
	SystemPrompts     []SystemPromptEntry  `json:"system_prompts,omitempty"`
	Memory            string               `json:"memory,omitempty"`
	Tools             []string             `json:"tools,omitempty"`
	MaxToolIterations int                  `json:"max_tool_iterations,omitempty"`
}

// CompactResult is the response from PATCH …/messages/compact.
type CompactResult struct {
	Before  int `json:"before"`
	After   int `json:"after"`
	Deleted int `json:"deleted"`
}
