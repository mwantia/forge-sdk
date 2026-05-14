package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const dagPath = "/v1/system/dag"

// DagObject is returned by DagCat.
type DagObject struct {
	Type string          `json:"type"` // from X-Forge-Object-Type header
	Raw  json.RawMessage // decoded body
}

// DagLogEntry is one line from DagLog.
type DagLogEntry struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"short_hash"`
	Role      string    `json:"role"`
	Preview   string    `json:"preview"`
	CreatedAt time.Time `json:"created_at"`
}

// DagObjectEntry is one line from DagObjects when list=true.
type DagObjectEntry struct {
	Hash  string `json:"hash"`
	Shard string `json:"shard"`
}

// DagVerifyResult is the response from DagVerify.
type DagVerifyResult struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

// DagCat fetches the raw canonical JSON for a DAG object by hash or prefix.
// The returned DagObject carries the detected type and raw bytes.
func (c *Client) DagCat(ctx context.Context, hash string, pretty bool) (*DagObject, error) {
	path := dagPath + "/objects/" + hash
	if pretty {
		path += "?pretty=true"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+path, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &DagObject{
		Type: resp.Header.Get("X-Forge-Object-Type"),
		Raw:  json.RawMessage(raw),
	}, nil
}

// DagType returns the type string for a DAG object.
func (c *Client) DagType(ctx context.Context, hash string) (string, error) {
	var result struct {
		Hash string `json:"hash"`
		Type string `json:"type"`
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+dagPath+"/objects/"+hash+"/type", nil)
	if err != nil {
		return "", err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", parseErrorResponse(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Type, nil
}

// DagLog walks a session ref and streams log entries over the returned channel.
// The channel is closed when the stream ends or ctx is cancelled.
func (c *Client) DagLog(ctx context.Context, sessionID, ref string) (<-chan DagLogEntry, error) {
	path := fmt.Sprintf("%s/sessions/%s/log", dagPath, sessionID)
	if ref != "" {
		path += "?ref=" + ref
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+path, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, parseErrorResponse(resp)
	}

	ch := make(chan DagLogEntry, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var entry DagLogEntry
			if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case ch <- entry:
			}
		}
	}()
	return ch, nil
}

// DagDiff fetches the unified diff between two DAG objects and returns it as plain text.
func (c *Client) DagDiff(ctx context.Context, hashA, hashB string) (string, error) {
	path := fmt.Sprintf("%s/diff?a=%s&b=%s", dagPath, hashA, hashB)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+path, nil)
	if err != nil {
		return "", err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", parseErrorResponse(resp)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// DagVerify checks integrity of reachable objects for a session ref.
// Returns a result even on 422 (verify found errors) so the caller can
// inspect the error list; only network/auth failures return a non-nil error.
func (c *Client) DagVerify(ctx context.Context, sessionID, ref string, all bool) (*DagVerifyResult, error) {
	body := map[string]any{"session_id": sessionID, "ref": ref, "all": all}
	var result DagVerifyResult
	_ = c.post(dagPath+"/verify", body, &result)
	return &result, nil
}

// DagObjectsCount returns the number of objects in the store, optionally filtered by shard prefix.
func (c *Client) DagObjectsCount(ctx context.Context, prefix string) (int, error) {
	path := dagPath + "/objects"
	if prefix != "" {
		path += "?prefix=" + prefix
	}
	var result struct {
		Count int `json:"count"`
	}
	if err := c.get(path, &result); err != nil {
		return 0, err
	}
	return result.Count, nil
}

// DagObjectsList streams object entries over the returned channel.
func (c *Client) DagObjectsList(ctx context.Context, prefix string) (<-chan DagObjectEntry, error) {
	path := dagPath + "/objects?list=true"
	if prefix != "" {
		path += "&prefix=" + prefix
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+path, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, parseErrorResponse(resp)
	}

	ch := make(chan DagObjectEntry, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var entry DagObjectEntry
			if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case ch <- entry:
			}
		}
	}()
	return ch, nil
}

// DagGC runs (or dry-runs) a GC pass and returns stats.
func (c *Client) DagGC(ctx context.Context, dryRun bool) (*GCResult, error) {
	path := dagPath + "/gc"
	if dryRun {
		path += "?dry_run=true"
	}
	var result GCResult
	if err := c.post(path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
