package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRetryableHTTPClient(t *testing.T) {
	retryConfig := &RetryConfig{
		MaxRetries:  3,
		BackoffBase: 100,
		HTTPTimeout: 30,
	}

	client := NewRetryableHTTPClient(retryConfig)

	if client == nil {
		t.Fatal("NewRetryableHTTPClient() returned nil")
	}

	if client.retryConfig.MaxRetries != 3 {
		t.Errorf("MaxRetries: got %d, want 3", client.retryConfig.MaxRetries)
	}

	if client.GetHTTPTimeout() != 30*time.Second {
		t.Errorf("HTTPTimeout: got %v, want 30s", client.GetHTTPTimeout())
	}
}

func TestRetryableHTTPClient_Do_Success(t *testing.T) {
	// Create mock server that responds successfully
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	retryConfig := &RetryConfig{
		MaxRetries:  3,
		BackoffBase: 10,
		HTTPTimeout: 5,
	}

	client := NewRetryableHTTPClient(retryConfig)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "success" {
		t.Errorf("Body: got %s, want success", string(body))
	}
}

func TestRetryableHTTPClient_Do_RetryOn500(t *testing.T) {
	attempts := 0

	// Create mock server that fails twice, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	retryConfig := &RetryConfig{
		MaxRetries:  3,
		BackoffBase: 10,
		HTTPTimeout: 5,
	}

	client := NewRetryableHTTPClient(retryConfig)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if attempts != 3 {
		t.Errorf("Attempts: got %d, want 3", attempts)
	}
}

func TestRetryableHTTPClient_Do_ExhaustsRetries(t *testing.T) {
	attempts := 0

	// Create mock server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("service unavailable"))
	}))
	defer server.Close()

	retryConfig := &RetryConfig{
		MaxRetries:  2,
		BackoffBase: 10,
		HTTPTimeout: 5,
	}

	client := NewRetryableHTTPClient(retryConfig)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode: got %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}

	// Should try initial + 2 retries = 3 total attempts
	expectedAttempts := 3
	if attempts != expectedAttempts {
		t.Errorf("Attempts: got %d, want %d", attempts, expectedAttempts)
	}
}

func TestRetryableHTTPClient_Do_NoRetryOn4xx(t *testing.T) {
	attempts := 0

	// Create mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	retryConfig := &RetryConfig{
		MaxRetries:  3,
		BackoffBase: 10,
		HTTPTimeout: 5,
	}

	client := NewRetryableHTTPClient(retryConfig)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	// Should not retry on 4xx errors
	if attempts != 1 {
		t.Errorf("Attempts: got %d, want 1 (no retry)", attempts)
	}
}

func TestRetryableHTTPClient_Do_WithRequestBody(t *testing.T) {
	attempts := 0

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		body, _ := io.ReadAll(r.Body)

		// Verify body is preserved across retries
		if string(body) != "test body" {
			t.Errorf("Request body: got %s, want 'test body'", string(body))
		}

		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	retryConfig := &RetryConfig{
		MaxRetries:  3,
		BackoffBase: 10,
		HTTPTimeout: 5,
	}

	client := NewRetryableHTTPClient(retryConfig)

	req, _ := http.NewRequest("POST", server.URL, strings.NewReader("test body"))
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if attempts != 2 {
		t.Errorf("Attempts: got %d, want 2", attempts)
	}
}

func TestCalculateBackoff(t *testing.T) {
	retryConfig := &RetryConfig{
		MaxRetries:  3,
		BackoffBase: 100,
		HTTPTimeout: 30,
	}

	client := NewRetryableHTTPClient(retryConfig)

	tests := []struct {
		attempt int
		min     int
		max     int
	}{
		{0, 90, 110},   // 100 * 2^0 = 100 (+/- jitter)
		{1, 180, 220},  // 100 * 2^1 = 200 (+/- jitter)
		{2, 360, 440},  // 100 * 2^2 = 400 (+/- jitter)
		{3, 720, 880},  // 100 * 2^3 = 800 (+/- jitter)
		{10, 9000, 11000}, // Should cap at 10000
	}

	for _, tt := range tests {
		backoff := client.calculateBackoff(tt.attempt)
		if backoff < tt.min || backoff > tt.max {
			t.Errorf("calculateBackoff(%d): got %d, want between %d and %d",
				tt.attempt, backoff, tt.min, tt.max)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, false},
		{201, false},
		{400, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tt := range tests {
		result := isRetryable(tt.statusCode)
		if result != tt.expected {
			t.Errorf("isRetryable(%d): got %v, want %v",
				tt.statusCode, result, tt.expected)
		}
	}
}

func TestFormatRetryStats(t *testing.T) {
	tests := []struct {
		name     string
		config   *RetryConfig
		expected string
	}{
		{
			"With retries",
			&RetryConfig{MaxRetries: 3, BackoffBase: 100, HTTPTimeout: 30},
			"Retries: 3 max, 100ms base backoff, 30s timeout",
		},
		{
			"No retries",
			&RetryConfig{MaxRetries: 0, BackoffBase: 100, HTTPTimeout: 30},
			"Retries: disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRetryStats(tt.config)
			if result != tt.expected {
				t.Errorf("FormatRetryStats(): got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRetryableHTTPClient_GetRetryConfig(t *testing.T) {
	retryConfig := &RetryConfig{
		MaxRetries:  5,
		BackoffBase: 200,
		HTTPTimeout: 60,
	}

	client := NewRetryableHTTPClient(retryConfig)
	config := client.GetRetryConfig()

	if config.MaxRetries != 5 {
		t.Errorf("MaxRetries: got %d, want 5", config.MaxRetries)
	}

	if config.BackoffBase != 200 {
		t.Errorf("BackoffBase: got %d, want 200", config.BackoffBase)
	}

	if config.HTTPTimeout != 60 {
		t.Errorf("HTTPTimeout: got %d, want 60", config.HTTPTimeout)
	}
}
