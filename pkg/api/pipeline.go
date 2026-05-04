package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
// DispatchOptions configures branch-aware dispatching. Zero values pick
// the defaults: HEAD branch, no fork, server-managed chunking.
type DispatchOptions struct {
	NoStore  bool
	Raw      bool
	Ref      string // dispatch on a non-HEAD ref
	ForkFrom string // hash or prefix; auto-creates a branch off that message's parent
}

// SendMessage dispatches a user message and returns the active ref name and a
// channel of streamed pipeline events. The ref comes from the X-Forge-Ref
// response header — useful for renaming auto-created fork-* refs via
// RenameBranch after the stream completes.
func (c *Client) SendMessage(ctx context.Context, sessionID, content string, opts DispatchOptions) (ref string, ch <-chan WireEvent, err error) {
	body := map[string]any{
		"session_id": sessionID,
		"content":    content,
		"no_store":   opts.NoStore,
	}
	path := pipelinePath + "/dispatch"
	q := []string{}
	if opts.Raw {
		q = append(q, "raw=true")
	}
	if opts.Ref != "" {
		q = append(q, "ref="+opts.Ref)
	}
	if opts.ForkFrom != "" {
		q = append(q, "fork_from="+opts.ForkFrom)
	}
	if len(q) > 0 {
		path += "?" + joinQuery(q)
	}
	resp, rerr := c.postRaw(path, body)
	if rerr != nil {
		return "", nil, rerr
	}

	ref = resp.Header.Get("X-Forge-Ref")
	events := make(chan WireEvent, 32)
	go func() {
		defer close(events)
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
			events <- ev
			if ev.Type == "done" || ev.Type == "error" {
				return
			}
		}
	}()

	return ref, events, nil
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

func joinQuery(parts []string) string {
	return strings.Join(parts, "&")
}

// SystemSnapshot is the current system message for a session (root of HEAD chain).
type SystemSnapshot struct {
	Hash    string `json:"hash"`
	Content string `json:"content"`
	Message string `json:"message,omitempty"`
}

// SystemRegenResult is the response from system edit or regen operations.
// Branch is non-empty when the operation forked the existing chain; empty when
// it wrote the system to an empty HEAD (fresh session).
type SystemRegenResult struct {
	Hash   string `json:"hash"`
	Branch string `json:"branch,omitempty"`
}

// GetSystemSnapshot returns the system message for a session (root of HEAD chain).
// If no messages exist yet, Hash and Content will be empty.
func (c *Client) GetSystemSnapshot(ctx context.Context, sessionID string) (*SystemSnapshot, error) {
	var snap SystemSnapshot
	if err := c.get(sessionsPath+"/"+sessionID+"/system", &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// EditSystemSnapshot replaces the system message with template-rendered content.
// Returns the new hash and the fork branch name (empty on a fresh session).
func (c *Client) EditSystemSnapshot(ctx context.Context, sessionID, content string) (string, string, error) {
	body := map[string]any{"content": content}
	var out SystemRegenResult
	if err := c.patch(sessionsPath+"/"+sessionID+"/system", body, &out); err != nil {
		return "", "", err
	}
	return out.Hash, out.Branch, nil
}

// RegenSystemSnapshot re-assembles the system prompt from current plugin state
// and stores it as a new root MessageObj. system is an optional session-layer
// template rendered and appended to the assembled prompt. toolsVerbosity and
// plugins are optional overrides (empty = use session defaults).
// Returns the new hash and the fork branch name (empty on a fresh session).
func (c *Client) RegenSystemSnapshot(ctx context.Context, sessionID, system, toolsVerbosity string, plugins []string) (string, string, error) {
	body := map[string]any{}
	if system != "" {
		body["system"] = system
	}
	if toolsVerbosity != "" {
		body["tools_verbosity"] = toolsVerbosity
	}
	if len(plugins) > 0 {
		body["plugins"] = plugins
	}
	var out SystemRegenResult
	if err := c.post(sessionsPath+"/"+sessionID+"/system/regen", body, &out); err != nil {
		return "", "", err
	}
	return out.Hash, out.Branch, nil
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
