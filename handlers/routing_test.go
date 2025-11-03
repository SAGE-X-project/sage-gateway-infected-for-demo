package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// TestDynamicRoutingByToField tests that gateway routes messages based on "to" field
func TestDynamicRoutingByToField(t *testing.T) {
	// Setup: Mock backend agents
	paymentAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "payment_received"}`))
	}))
	defer paymentAgent.Close()

	medicalAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "medical_received"}`))
	}))
	defer medicalAgent.Close()

	// Set AGENT_URLS environment variable
	agentURLs := map[string]string{
		"payment": paymentAgent.URL,
		"medical": medicalAgent.URL,
	}
	agentURLsJSON, _ := json.Marshal(agentURLs)
	os.Setenv("AGENT_URLS", string(agentURLsJSON))
	defer os.Unsetenv("AGENT_URLS")

	// Disable attack for clean routing test
	os.Setenv("ATTACK_ENABLED", "false")
	defer os.Unsetenv("ATTACK_ENABLED")

	// Create gateway proxy handler
	cfg := config.LoadConfig()
	proxyHandler := NewProxyHandler(cfg)

	tests := []struct {
		name           string
		targetAgent    string
		expectedCalled string
	}{
		{
			name:           "Route to payment agent",
			targetAgent:    "payment",
			expectedCalled: paymentAgent.URL,
		},
		{
			name:           "Route to medical agent",
			targetAgent:    "medical",
			expectedCalled: medicalAgent.URL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create AgentMessage with "to" field
			agentMsg := types.AgentMessage{
				ID:        "test-msg-1",
				From:      "test-client",
				To:        tt.targetAgent,
				Content:   "Test message",
				Timestamp: time.Now(),
				Type:      "request",
			}

			msgBytes, err := json.Marshal(agentMsg)
			if err != nil {
				t.Fatalf("Failed to marshal agent message: %v", err)
			}

			// Send request to gateway
			req := httptest.NewRequest("POST", "/", bytes.NewBuffer(msgBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			proxyHandler.HandleRequest(w, req)

			// Verify response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// Verify correct agent was called (check response)
			respBody := w.Body.String()
			if tt.targetAgent == "payment" && respBody != `{"status": "payment_received"}` {
				t.Errorf("Expected payment agent response, got: %s", respBody)
			}
			if tt.targetAgent == "medical" && respBody != `{"status": "medical_received"}` {
				t.Errorf("Expected medical agent response, got: %s", respBody)
			}
		})
	}
}

// TestAgentURLsLoading tests loading agent URLs from environment variable
func TestAgentURLsLoading(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedURLs   map[string]string
		shouldHaveURLs bool
	}{
		{
			name:     "Valid JSON agent URLs",
			envValue: `{"root":"http://localhost:18080","payment":"http://localhost:19083","medical":"http://localhost:19082"}`,
			expectedURLs: map[string]string{
				"root":    "http://localhost:18080",
				"payment": "http://localhost:19083",
				"medical": "http://localhost:19082",
			},
			shouldHaveURLs: true,
		},
		{
			name:     "Empty environment variable (use defaults)",
			envValue: "",
			expectedURLs: map[string]string{
				"root":     "http://localhost:18080",
				"payment":  "http://localhost:19083",
				"medical":  "http://localhost:19082",
				"planning": "http://localhost:19081",
			},
			shouldHaveURLs: true,
		},
		{
			name:     "Invalid JSON (fallback to defaults)",
			envValue: `{invalid json}`,
			expectedURLs: map[string]string{
				"root":     "http://localhost:18080",
				"payment":  "http://localhost:19083",
				"medical":  "http://localhost:19082",
				"planning": "http://localhost:19081",
			},
			shouldHaveURLs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("AGENT_URLS", tt.envValue)
				defer os.Unsetenv("AGENT_URLS")
			}

			// Load config
			cfg := config.LoadConfig()

			// Verify agent URLs
			if tt.shouldHaveURLs {
				for agentName, expectedURL := range tt.expectedURLs {
					actualURL := cfg.GetAgentURL(agentName)
					if actualURL != expectedURL {
						t.Errorf("Agent %s: expected URL %s, got %s", agentName, expectedURL, actualURL)
					}
				}
			}
		})
	}
}

// TestFallbackToLegacyTargetURL tests that gateway falls back to TARGET_AGENT_URL if "to" field is not found
func TestFallbackToLegacyTargetURL(t *testing.T) {
	// Setup: Mock backend agent
	targetAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "received"}`))
	}))
	defer targetAgent.Close()

	// Set legacy TARGET_AGENT_URL
	os.Setenv("TARGET_AGENT_URL", targetAgent.URL)
	defer os.Unsetenv("TARGET_AGENT_URL")

	// Disable attack
	os.Setenv("ATTACK_ENABLED", "false")
	defer os.Unsetenv("ATTACK_ENABLED")

	// Create gateway proxy handler
	cfg := config.LoadConfig()
	proxyHandler := NewProxyHandler(cfg)

	// Send message WITHOUT "to" field (legacy format)
	legacyMsg := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x123",
	}
	msgBytes, _ := json.Marshal(legacyMsg)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(msgBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	proxyHandler.HandleRequest(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	respBody := w.Body.String()
	if respBody != `{"status": "received"}` {
		t.Errorf("Expected target agent response, got: %s", respBody)
	}
}

// TestUnknownAgentFallback tests handling of unknown agent names in "to" field
func TestUnknownAgentFallback(t *testing.T) {
	// Setup: Mock backend agent
	defaultAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "default_received"}`))
	}))
	defer defaultAgent.Close()

	// Set TARGET_AGENT_URL as fallback
	os.Setenv("TARGET_AGENT_URL", defaultAgent.URL)
	defer os.Unsetenv("TARGET_AGENT_URL")

	// Disable attack
	os.Setenv("ATTACK_ENABLED", "false")
	defer os.Unsetenv("ATTACK_ENABLED")

	// Create gateway proxy handler
	cfg := config.LoadConfig()
	proxyHandler := NewProxyHandler(cfg)

	// Send message with unknown agent in "to" field
	agentMsg := types.AgentMessage{
		ID:        "test-msg-1",
		From:      "test-client",
		To:        "unknown-agent",
		Content:   "Test message",
		Timestamp: time.Now(),
		Type:      "request",
	}
	msgBytes, _ := json.Marshal(agentMsg)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(msgBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	proxyHandler.HandleRequest(w, req)

	// Verify falls back to default agent
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	respBody := w.Body.String()
	if respBody != `{"status": "default_received"}` {
		t.Errorf("Expected default agent response, got: %s", respBody)
	}
}
