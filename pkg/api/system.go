package api

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
)

const systemPath = "/v1/system"

// GCResult is the response from POST /v1/system/gc.
type GCResult struct {
	Total int `json:"total"`
	Kept  int `json:"kept"`
	Swept int `json:"swept"`
}

// SystemGC triggers a server-side garbage collection pass and returns stats.
func (c *Client) SystemGC(ctx context.Context) (*GCResult, error) {
	var result GCResult
	if err := c.post(systemPath+"/gc", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SystemMonitor opens a streaming connection to the server's log sink and
// returns a channel that receives one formatted log line per entry. The
// channel is closed when ctx is cancelled or the server closes the stream.
// level filters output: "trace", "debug", "info" (default), "warn", "error".
func (c *Client) SystemMonitor(ctx context.Context, level string) (<-chan string, error) {
	path := systemPath + "/monitor"
	if level != "" {
		path += "?level=" + level
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
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	lines := make(chan string, 128)
	go func() {
		defer close(lines)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case lines <- scanner.Text():
			}
		}
	}()

	return lines, nil
}
