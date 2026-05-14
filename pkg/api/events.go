package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

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

type EventState string

const (
	EventStateRunning EventState = "running"
	EventStatePaused  EventState = "paused"
	EventStateFailed  EventState = "failed"
)

// EventBranch describes a single dispatched event allocation (one branch per fire).
type EventBranch struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	FiredAt time.Time `json:"fired_at"`
}

// EventStatus describes a configured event and its live queue state.
type EventStatus struct {
	ID          string     `json:"id"`
	Description string     `json:"description,omitempty"`
	Session     string     `json:"session"`
	State       EventState `json:"state"`
	Options     *EventOpts `json:"options,omitempty"`
	Queue       *EventQueue `json:"queue,omitempty"`
	LastBranch  string     `json:"last_branch,omitempty"`
}

// EventOpts mirrors the options block from the HCL config.
type EventOpts struct {
	Timespan string `json:"timespan,omitempty"`
	MaxQueue int    `json:"max_queue,omitempty"`
}

// EventQueue is the live queue state returned by the server.
type EventQueue struct {
	Size            int        `json:"size"`
	WindowExpiresAt *time.Time `json:"window_expires_at,omitempty"`
}

// FireResponse is the JSON body returned by the fire endpoint.
type FireResponse struct {
	EventID         string     `json:"event_id"`
	Status          string     `json:"status"`
	FiredAt         time.Time  `json:"fired_at"`
	Branch          string     `json:"branch,omitempty"`
	QueueSize       int        `json:"queue_size,omitempty"`
	QueueCapacity   int        `json:"queue_capacity,omitempty"`
	Evicted         bool       `json:"evicted,omitempty"`
	WindowExpiresAt *time.Time `json:"window_expires_at,omitempty"`
}

// ListEventBranches returns all branches created by the named event in the given
// session, sorted newest-first. Branch names follow the convention
// "event/<eventID>-<RFC3339>" set by the service dispatch path.
func (c *Client) ListEventBranches(ctx context.Context, sessionID, eventID string) ([]EventBranch, error) {
	prefix := "event/" + eventID + "-"
	refs, err := c.ListBranchesWithPrefix(ctx, sessionID, prefix)
	if err != nil {
		return nil, err
	}
	branches := make([]EventBranch, 0, len(refs))
	for name, hash := range refs {
		ts := strings.TrimPrefix(name, prefix)
		firedAt, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}
		branches = append(branches, EventBranch{Name: name, Hash: hash, FiredAt: firedAt})
	}
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].FiredAt.After(branches[j].FiredAt)
	})
	return branches, nil
}

// ListEvents returns all configured events and their live queue state.
func (c *Client) ListEvents(_ context.Context) ([]*EventStatus, error) {
	var out []*EventStatus
	return out, c.get("/v1/events", &out)
}

// GetEvent returns a single event by ID.
func (c *Client) GetEvent(_ context.Context, id string) (*EventStatus, error) {
	var out EventStatus
	return &out, c.get(fmt.Sprintf("/v1/events/%s", id), &out)
}

// PauseEvent pauses the named event. Fires while paused return 503.
func (c *Client) PauseEvent(_ context.Context, id string) (*EventStatus, error) {
	req, err := http.NewRequest(http.MethodPost, c.addr+fmt.Sprintf("/v1/events/%s/pause", id), nil)
	if err != nil {
		return nil, err
	}
	var out EventStatus
	return &out, c.do(req, &out)
}

// ResumeEvent resumes a paused event.
func (c *Client) ResumeEvent(_ context.Context, id string) (*EventStatus, error) {
	req, err := http.NewRequest(http.MethodPost, c.addr+fmt.Sprintf("/v1/events/%s/resume", id), nil)
	if err != nil {
		return nil, err
	}
	var out EventStatus
	return &out, c.do(req, &out)
}

// FireEvent fires the named event with an optional JSON payload and branch
// override. payload may be nil (bare fire with no body).
func (c *Client) FireEvent(_ context.Context, id string, payload any, ref string) (*FireResponse, error) {
	var body []byte
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = b
	}

	path := fmt.Sprintf("/v1/events/%s/fire", id)
	if ref != "" {
		path += "?ref=" + ref
	}

	req, err := http.NewRequest(http.MethodPost, c.addr+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var resp FireResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
