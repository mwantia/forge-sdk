package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
)

const (
	pipelinePath = "/v1/pipeline"
	sessionsPath = "/v1/sessions"
)

// ListSessions returns sessions, optionally filtered by parent session ID.
func (c *Client) ListSessions(ctx context.Context, parent string, offset, limit int) ([]*SessionMetadata, error) {
	path := fmt.Sprintf("%s?offset=%d&limit=%d", sessionsPath, offset, limit)
	if parent != "" {
		path += "&parent=" + parent
	}
	var resp struct {
		Sessions []*SessionMetadata `json:"sessions"`
	}
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Sessions, nil
}

// CreateSession creates a new session and returns its metadata.
func (c *Client) CreateSession(ctx context.Context, req CreateSessionRequest) (*SessionMetadata, error) {
	var meta SessionMetadata
	if err := c.post(sessionsPath, req, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetSession returns metadata for an existing session by ID or name.
func (c *Client) GetSession(ctx context.Context, id string) (*SessionMetadata, error) {
	var meta SessionMetadata
	if err := c.get(sessionsPath+"/"+id, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// DeleteSession deletes a session and all its messages.
func (c *Client) DeleteSession(ctx context.Context, id string) error {
	return c.delete(sessionsPath + "/" + id)
}

// ListMessages returns messages for a session in chronological order.
func (c *Client) ListMessages(ctx context.Context, sessionID string, offset, limit int) ([]*Message, error) {
	path := fmt.Sprintf("%s/%s/messages?offset=%d&limit=%d", sessionsPath, sessionID, offset, limit)
	var resp struct {
		Messages []*Message `json:"messages"`
	}
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

// GetMessage returns a single message by session ID and message ID.
func (c *Client) GetMessage(ctx context.Context, sessionID, msgID string) (*Message, error) {
	var msg Message
	if err := c.get(sessionsPath+"/"+sessionID+"/messages/"+msgID, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// SendMessage dispatches a user message to the pipeline and returns a channel
// of pipeline events streamed as NDJSON. The channel is closed after the
// DoneEvent or ErrorEvent. The caller must drain or cancel the channel;
// cancelling ctx stops the stream. When noStore is true, the generated
// messages are not persisted to storage. When raw is true, the server's
// output chunking and pacing policy is bypassed: deltas are forwarded as
// token-boundary chunks with no pacing. Use this for programmatic consumers
// that want raw throughput.
func (c *Client) SendMessage(ctx context.Context, sessionID, content string, noStore, raw bool) (<-chan WireEvent, error) {
	body := map[string]any{
		"session_id": sessionID,
		"content":    content,
		"no_store":   noStore,
	}
	path := pipelinePath + "/dispatch"
	if raw {
		path += "?raw=true"
	}
	resp, err := c.postRaw(path, body)
	if err != nil {
		return nil, err
	}

	ch := make(chan WireEvent, 32)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line := scanner.Text()
			if line == "" {
				continue
			}
			var ev WireEvent
			if err := json.Unmarshal([]byte(line), &ev); err != nil {
				continue
			}
			ch <- ev
			if ev.Type == "done" || ev.Type == "error" {
				return
			}
		}
	}()

	return ch, nil
}

// PreviewUsage is a heuristic size summary for a single fragment. EstTokens
// is computed as runes/4 — accuracy is roughly ±20% across common LLM
// tokenizers. Use for relative comparison, not billing.
type PreviewUsage struct {
	Bytes     int `json:"bytes"`
	Runes     int `json:"runes"`
	EstTokens int `json:"est_tokens"`
}

// PreviewMessage is one entry in a PreviewResponse — the role/content pair
// that would be sent as a chat message (excluding the assembled system block,
// which is exposed separately on the response).
type PreviewMessage struct {
	Role    string       `json:"role"`
	Content string       `json:"content,omitempty"`
	Usage   PreviewUsage `json:"usage"`
}

// PreviewResponse mirrors what the pipeline /preview endpoint returns: the
// assembled system prompt, the rendered message history (plus the optional
// new user content), and the number of tools that would be advertised.
type PreviewResponse struct {
	SessionID   string           `json:"session_id"`
	System      string           `json:"system"`
	SystemUsage PreviewUsage     `json:"system_usage"`
	Messages    []PreviewMessage `json:"messages"`
	Total       PreviewUsage     `json:"total"`
	ToolCount   int              `json:"tool_count"`
	EstAccuracy string           `json:"est_accuracy"`
}

// PreviewPipeline asks the agent what it would send to the LLM for the given
// session and (optional) new user content. Useful for inspecting prompt
// composition without running a real turn — nothing is persisted and no LLM
// call is made.
func (c *Client) PreviewPipeline(ctx context.Context, sessionID, content string) (*PreviewResponse, error) {
	body := map[string]any{
		"session_id": sessionID,
		"content":    content,
	}
	var resp PreviewResponse
	if err := c.post(pipelinePath+"/preview", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CompactMessages removes intermediate tool-call messages from the session.
func (c *Client) CompactMessages(ctx context.Context, sessionID string) (*CompactResult, error) {
	var result CompactResult
	body := map[string]any{"strip_tools": true}
	if err := c.patch(sessionsPath+"/"+sessionID+"/messages/compact", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
