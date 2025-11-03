package handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/logger"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries       int
	BackoffBase      int // milliseconds
	HTTPTimeout      int // seconds
}

// RetryableHTTPClient wraps http.Client with retry logic
type RetryableHTTPClient struct {
	client      *http.Client
	retryConfig *RetryConfig
}

// NewRetryableHTTPClient creates a new retryable HTTP client
func NewRetryableHTTPClient(retryConfig *RetryConfig) *RetryableHTTPClient {
	return &RetryableHTTPClient{
		client: &http.Client{
			Timeout: time.Duration(retryConfig.HTTPTimeout) * time.Second,
		},
		retryConfig: retryConfig,
	}
}

// Do executes an HTTP request with retry logic and exponential backoff
func (r *RetryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	maxRetries := r.retryConfig.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Clone request for retry (body can only be read once)
		reqClone := req.Clone(req.Context())

		resp, err = r.client.Do(reqClone)

		// Success - return immediately
		if err == nil && resp.StatusCode < 500 {
			if attempt > 0 {
				logger.Info("✅ Request succeeded after %d retries", attempt)
			}
			return resp, nil
		}

		// Last attempt - return error
		if attempt == maxRetries {
			if err != nil {
				logger.Error("❌ Request failed after %d retries: %v", maxRetries, err)
			} else {
				logger.Error("❌ Request failed after %d retries: HTTP %d", maxRetries, resp.StatusCode)
			}
			return resp, err
		}

		// Calculate backoff with exponential backoff
		backoffTime := r.calculateBackoff(attempt)

		// Log retry attempt
		if err != nil {
			logger.Warn("⚠️  Request failed (attempt %d/%d): %v - retrying in %dms...",
				attempt+1, maxRetries+1, err, backoffTime)
		} else {
			logger.Warn("⚠️  Request failed (attempt %d/%d): HTTP %d - retrying in %dms...",
				attempt+1, maxRetries+1, resp.StatusCode, backoffTime)
			// Close failed response body
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}

		// Wait before retry
		time.Sleep(time.Duration(backoffTime) * time.Millisecond)
	}

	return resp, err
}

// calculateBackoff calculates exponential backoff time in milliseconds
// Formula: backoffBase * (2 ^ attempt) with jitter
func (r *RetryableHTTPClient) calculateBackoff(attempt int) int {
	base := r.retryConfig.BackoffBase
	if base <= 0 {
		base = 100
	}

	// Exponential backoff: base * 2^attempt
	backoff := float64(base) * math.Pow(2, float64(attempt))

	// Cap at 10 seconds
	if backoff > 10000 {
		backoff = 10000
	}

	// Add 10% jitter to avoid thundering herd
	jitter := backoff * 0.1 * (float64(time.Now().UnixNano()%100) / 100.0)

	return int(backoff + jitter)
}

// isRetryable determines if an HTTP status code is retryable
func isRetryable(statusCode int) bool {
	// Retry on 5xx server errors and 429 Too Many Requests
	return statusCode >= 500 || statusCode == 429
}

// GetHTTPTimeout returns the configured HTTP timeout
func (r *RetryableHTTPClient) GetHTTPTimeout() time.Duration {
	return r.client.Timeout
}

// GetRetryConfig returns the retry configuration
func (r *RetryableHTTPClient) GetRetryConfig() *RetryConfig {
	return r.retryConfig
}

// FormatRetryStats formats retry statistics for logging
func FormatRetryStats(retryConfig *RetryConfig) string {
	if retryConfig.MaxRetries == 0 {
		return "Retries: disabled"
	}
	return fmt.Sprintf("Retries: %d max, %dms base backoff, %ds timeout",
		retryConfig.MaxRetries, retryConfig.BackoffBase, retryConfig.HTTPTimeout)
}
