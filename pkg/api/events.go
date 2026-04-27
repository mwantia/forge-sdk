package api

import (
	"encoding/json"
	"fmt"

	"github.com/mwantia/forge-sdk/pkg/plugins"
)

// WireEvent is the JSON envelope sent over the NDJSON stream.
type WireEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Concrete event payloads — one per Type value.

// ChunkBoundary classifies a ChunkEvent's cut point.
type ChunkBoundary string

const (
	ChunkBoundaryToken    ChunkBoundary = "token"
	ChunkBoundarySentence ChunkBoundary = "sentence"
	ChunkBoundaryBlock    ChunkBoundary = "block"
	ChunkBoundaryFinal    ChunkBoundary = "final"
)

// ChunkEvent carries a text chunk emitted at a boundary chosen by the server's
// pipeline.output policy. Concatenating all non-thinking Text fields in order
// reproduces the full assistant response.
type ChunkEvent struct {
	Text     string        `json:"text"`
	Thinking string        `json:"thinking,omitempty"`
	Boundary ChunkBoundary `json:"boundary"`
}

type ToolCallEvent struct {
	CallID string         `json:"call_id,omitempty"`
	Name   string         `json:"name"`
	Args   map[string]any `json:"args,omitempty"`
}

type ToolResultEvent struct {
	CallID  string `json:"call_id,omitempty"`
	Name    string `json:"name"`
	Result  any    `json:"result"`
	IsError bool   `json:"is_error,omitempty"`
}

type ErrorEvent struct {
	Message string `json:"message"`
}

type DoneEvent struct {
	Usage    *plugins.TokenUsage `json:"usage,omitempty"`
	Metadata map[string]any      `json:"metadata,omitempty"`
}

// ParseWireEvent deserialises the Data field of a WireEvent into the
// appropriate concrete type. Returns nil for unknown event types.
func ParseWireEvent(w WireEvent) (any, error) {
	switch w.Type {
	case "chunk":
		var ev ChunkEvent
		return ev, json.Unmarshal(w.Data, &ev)
	case "tool_call":
		var ev ToolCallEvent
		return ev, json.Unmarshal(w.Data, &ev)
	case "tool_result":
		var ev ToolResultEvent
		return ev, json.Unmarshal(w.Data, &ev)
	case "error":
		var ev ErrorEvent
		return ev, json.Unmarshal(w.Data, &ev)
	case "done":
		var ev DoneEvent
		return ev, json.Unmarshal(w.Data, &ev)
	default:
		return nil, fmt.Errorf("unknown wire event type %q", w.Type)
	}
}
