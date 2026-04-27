package retry

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// Config controls retry behaviour.
type Config struct {
	// MaxAttempts is the total number of attempts (including the first). Must be >= 1.
	MaxAttempts int

	// BaseDelay is the initial backoff duration. Doubles on each attempt.
	BaseDelay time.Duration

	// MaxDelay caps the per-attempt backoff before jitter is applied.
	MaxDelay time.Duration
}

// DefaultConfig is a sensible default for provider HTTP calls.
var DefaultConfig = Config{
	MaxAttempts: 3,
	BaseDelay:   500 * time.Millisecond,
	MaxDelay:    10 * time.Second,
}

// RetryableError wraps an error with HTTP status metadata so Do() can decide
// whether to retry.
type RetryableError struct {
	StatusCode int
	RetryAfter time.Duration // parsed from Retry-After header, 0 if absent
	Err        error
}

func (e *RetryableError) Error() string { return e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

// HTTPError constructs a RetryableError from an HTTP response.
// The caller is responsible for reading and closing the body before calling this.
func HTTPError(resp *http.Response, err error) *RetryableError {
	re := &RetryableError{StatusCode: resp.StatusCode, Err: err}
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, parseErr := strconv.Atoi(ra); parseErr == nil {
			re.RetryAfter = time.Duration(secs) * time.Second
		}
	}
	return re
}

// IsRetryable reports whether err should trigger a retry.
// Only *RetryableError values with a 5xx status or 429 are retried;
// everything else (4xx, non-HTTP errors) fails fast.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	re, ok := err.(*RetryableError)
	if !ok {
		// Plain error (e.g. connection refused) — retry.
		return true
	}
	return re.StatusCode == http.StatusTooManyRequests || re.StatusCode >= 500
}

// Do calls fn up to cfg.MaxAttempts times, backing off between attempts.
// It stops early if ctx is cancelled or fn returns a non-retryable error.
func Do(ctx context.Context, cfg Config, fn func(ctx context.Context) error) error {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}

	var lastErr error
	for attempt := range cfg.MaxAttempts {
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}
		if !IsRetryable(lastErr) {
			return lastErr
		}
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		delay := backoffDelay(cfg, attempt)

		// Honour Retry-After when present.
		if re, ok := lastErr.(*RetryableError); ok && re.RetryAfter > 0 {
			delay = re.RetryAfter
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastErr
}

// backoffDelay returns the wait duration for the given attempt index (0-based)
// using exponential backoff with ±25% jitter.
func backoffDelay(cfg Config, attempt int) time.Duration {
	base := cfg.BaseDelay * (1 << attempt)
	base = min(base, cfg.MaxDelay)
	// Add ±25% jitter.
	jitter := time.Duration(float64(base) * 0.25 * (rand.Float64()*2 - 1))
	return base + jitter
}
