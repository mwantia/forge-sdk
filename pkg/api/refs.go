package api

import (
	"context"
	"fmt"
)

// ListRefs returns all refs for a session as a name -> hash map.
func (c *Client) ListRefs(ctx context.Context, sessionID string) (map[string]string, error) {
	var out struct {
		Refs map[string]string `json:"refs"`
	}
	if err := c.get(sessionsPath+"/"+sessionID+"/refs", &out); err != nil {
		return nil, err
	}
	return out.Refs, nil
}

// CreateRef points name at hash. Conflicts return an error.
func (c *Client) CreateRef(ctx context.Context, sessionID, name, hash string) error {
	body := map[string]any{"name": name, "hash": hash}
	return c.post(sessionsPath+"/"+sessionID+"/refs", body, nil)
}

// MoveRef advances a ref. When expectedHash is non-empty the move is
// CAS-conditional and may fail with a 409.
func (c *Client) MoveRef(ctx context.Context, sessionID, name, hash, expectedHash string) error {
	body := map[string]any{"hash": hash}
	if expectedHash != "" {
		body["expected_hash"] = expectedHash
	}
	return c.patch(fmt.Sprintf("%s/%s/refs/%s", sessionsPath, sessionID, name), body, nil)
}

// DeleteRef removes a ref. Missing refs are not an error.
func (c *Client) DeleteRef(ctx context.Context, sessionID, name string) error {
	return c.delete(fmt.Sprintf("%s/%s/refs/%s", sessionsPath, sessionID, name))
}
