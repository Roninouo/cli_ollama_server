package ollamaapi

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/url"
	"syscall"
	"time"
)

// RetryConfig configures the retry behavior for API requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (0 = no retries).
	MaxRetries int
	// InitialBackoff is the initial wait time between retries.
	InitialBackoff time.Duration
	// MaxBackoff is the maximum wait time between retries.
	MaxBackoff time.Duration
	// BackoffMultiplier is the factor by which backoff increases each retry.
	BackoffMultiplier float64
	// Jitter adds randomness to backoff to prevent thundering herd.
	Jitter bool
}

// DefaultRetryConfig provides sensible defaults for retry behavior.
var DefaultRetryConfig = RetryConfig{
	MaxRetries:        3,
	InitialBackoff:    100 * time.Millisecond,
	MaxBackoff:        5 * time.Second,
	BackoffMultiplier: 2.0,
	Jitter:            true,
}

// NoRetry disables retry behavior.
var NoRetry = RetryConfig{MaxRetries: 0}

// WithRetry sets custom retry configuration.
func WithRetry(cfg RetryConfig) ClientOption {
	return func(c *clientConfig) { c.retry = cfg }
}

// WithDefaultRetry enables default retry behavior.
func WithDefaultRetry() ClientOption {
	return func(c *clientConfig) { c.retry = DefaultRetryConfig }
}

// IsRetryableError returns true if the error is a transient error worth retrying.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation - don't retry these.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for API errors with specific status codes.
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// Retry on server errors (5xx) but not client errors (4xx).
		switch apiErr.StatusCode {
		case 500, 502, 503, 504:
			return true
		default:
			return false
		}
	}

	// Check for network-related errors.
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Retry on timeouts.
		if netErr.Timeout() {
			return true
		}
	}

	// Check for URL errors (DNS failures, connection refused, etc.).
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return true
		}
		// Unwrap and check for specific syscall errors.
		if urlErr.Err != nil {
			if isRetryableSyscallError(urlErr.Err) {
				return true
			}
		}
	}

	// Check for connection reset/refused errors.
	if isRetryableSyscallError(err) {
		return true
	}

	return false
}

// isRetryableSyscallError checks for common retryable syscall errors.
func isRetryableSyscallError(err error) bool {
	if err == nil {
		return false
	}

	// Connection reset or refused are often transient.
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	// Check by error message for cross-platform compatibility.
	errStr := err.Error()
	transientMessages := []string{
		"connection reset",
		"connection refused",
		"broken pipe",
		"no such host",
		"i/o timeout",
	}
	for _, msg := range transientMessages {
		if containsIgnoreCase(errStr, msg) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldSlice(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalFoldSlice compares two strings case-insensitively.
func equalFoldSlice(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// calculateBackoff computes the next backoff duration with optional jitter.
func calculateBackoff(attempt int, cfg RetryConfig) time.Duration {
	if attempt <= 0 {
		return cfg.InitialBackoff
	}

	backoff := float64(cfg.InitialBackoff)
	for i := 0; i < attempt; i++ {
		backoff *= cfg.BackoffMultiplier
	}

	if backoff > float64(cfg.MaxBackoff) {
		backoff = float64(cfg.MaxBackoff)
	}

	if cfg.Jitter {
		// Add ±25% jitter.
		jitterRange := backoff * 0.25
		backoff = backoff - jitterRange + (rand.Float64() * 2 * jitterRange)
	}

	return time.Duration(backoff)
}

// sleepWithContext sleeps for the given duration, returning early if context is cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
