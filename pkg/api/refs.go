package api

import (
	"context"
	"fmt"
)

// ListBranches returns all refs for a session as a name -> resolved-hash map.
// HEAD is included with its resolved hash (symrefs are dereferenced server-side).
func (c *Client) ListBranches(ctx context.Context, sessionID string) (map[string]string, error) {
	refs, _, err := c.ListBranchesWithSymrefs(ctx, sessionID)
	return refs, err
}

// ListBranchesWithSymrefs returns both the name->hash map and a symrefs map
// (e.g. {"HEAD": "main"}) that shows which refs are symbolic pointers.
func (c *Client) ListBranchesWithSymrefs(ctx context.Context, sessionID string) (refs, symrefs map[string]string, err error) {
	var out struct {
		Refs    map[string]string `json:"refs"`
		Symrefs map[string]string `json:"symrefs"`
	}
	if err := c.get(sessionsPath+"/"+sessionID+"/branch", &out); err != nil {
		return nil, nil, err
	}
	if out.Symrefs == nil {
		out.Symrefs = map[string]string{}
	}
	return out.Refs, out.Symrefs, nil
}

// CreateBranch points name at hash. Conflicts return an error.
func (c *Client) CreateBranch(ctx context.Context, sessionID, name, hash string) error {
	body := map[string]any{"name": name, "hash": hash}
	return c.post(sessionsPath+"/"+sessionID+"/branch", body, nil)
}

// MoveBranch advances a ref. When expectedHash is non-empty the move is
// CAS-conditional and may fail with a 409.
func (c *Client) MoveBranch(ctx context.Context, sessionID, name, hash, expectedHash string) error {
	body := map[string]any{"hash": hash}
	if expectedHash != "" {
		body["expected_hash"] = expectedHash
	}
	return c.patch(fmt.Sprintf("%s/%s/branch/%s", sessionsPath, sessionID, name), body, nil)
}

// RenameBranch atomically renames a ref from oldName to newName.
// Returns an error if newName already exists.
func (c *Client) RenameBranch(ctx context.Context, sessionID, oldName, newName string) error {
	body := map[string]any{"name": newName}
	return c.patch(fmt.Sprintf("%s/%s/branch/%s", sessionsPath, sessionID, oldName), body, nil)
}

// DeleteBranch removes a ref. Missing refs are not an error.
func (c *Client) DeleteBranch(ctx context.Context, sessionID, name string) error {
	return c.delete(fmt.Sprintf("%s/%s/branch/%s", sessionsPath, sessionID, name))
}

// CheckoutBranch sets HEAD to point symbolically at targetBranch, so
// subsequent dispatches advance targetBranch. Returns an error if the
// branch does not exist or targetBranch is "HEAD".
func (c *Client) CheckoutBranch(ctx context.Context, sessionID, targetBranch string) error {
	body := map[string]any{"checkout": targetBranch}
	return c.patch(fmt.Sprintf("%s/%s/branch/%s", sessionsPath, sessionID, "HEAD"), body, nil)
}
