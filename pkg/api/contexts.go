package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
)

const contextsPath = "/v1/contexts"

// PromptContextMessage is one resolved entry returned by GET
// /v1/contexts/:hash/materialized.
type PromptContextMessage struct {
	Hash      string         `json:"hash"`
	Role      string         `json:"role"`
	Content   string         `json:"content,omitempty"`
	ToolCalls []any          `json:"tool_calls,omitempty"`
}

// MaterializedContext mirrors the pipeline /v1/contexts/:hash/materialized
// payload: a fully resolved chat slice plus the original prompt-context
// header fields.
type MaterializedContext struct {
	Hash            string                 `json:"hash"`
	Provider        string                 `json:"provider"`
	Model           string                 `json:"model"`
	ToolCatalogHash string                 `json:"tool_catalog_hash,omitempty"`
	Options         map[string]any         `json:"options,omitempty"`
	Messages        []PromptContextMessage `json:"messages"`
}

// GetContext returns the raw PromptContext blob.
func (c *Client) GetContext(ctx context.Context, hash string) (map[string]any, error) {
	var out map[string]any
	if err := c.get(contextsPath+"/"+hash, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MaterializeContext returns the resolved prompt-context messages.
func (c *Client) MaterializeContext(ctx context.Context, hash string) (*MaterializedContext, error) {
	var out MaterializedContext
	if err := c.get(contextsPath+"/"+hash+"/materialized", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReplayContext re-dispatches a stored PromptContext and streams the
// response back as a channel of WireEvents. The original session is
// untouched. modelOverride is optional ("" = use the recorded model).
func (c *Client) ReplayContext(ctx context.Context, hash, modelOverride string) (<-chan WireEvent, error) {
	body := map[string]any{}
	if modelOverride != "" {
		body["model"] = modelOverride
	}
	resp, err := c.postRaw(fmt.Sprintf("%s/%s/replay", contextsPath, hash), body)
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
