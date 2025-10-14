package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

func TestNewProxyHandler(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:  true,
		TargetAgentURL: "http://localhost:8091",
	}

	handler := NewProxyHandler(cfg)

	if handler == nil {
		t.Fatal("NewProxyHandler() returned nil")
	}

	if handler.config != cfg {
		t.Error("NewProxyHandler() didn't set config properly")
	}

	if handler.interceptor == nil {
		t.Error("NewProxyHandler() didn't initialize interceptor")
	}

	if handler.modifier == nil {
		t.Error("NewProxyHandler() didn't initialize modifier")
	}

	if handler.client == nil {
		t.Error("NewProxyHandler() didn't initialize HTTP client")
	}
}

func TestProxyHandler_HandleRequest_MethodNotAllowed(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:  false,
		TargetAgentURL: "http://localhost:8091",
	}

	handler := NewProxyHandler(cfg)

	// Test GET request (only POST is allowed)
	req := httptest.NewRequest("GET", "/payment", nil)
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("HandleRequest() status code for GET: got %d, want %d",
			resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestProxyHandler_HandleRequest_InvalidJSON(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:  false,
		TargetAgentURL: "http://localhost:8091",
	}

	handler := NewProxyHandler(cfg)

	// Invalid JSON body
	req := httptest.NewRequest("POST", "/payment", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("HandleRequest() status code for invalid JSON: got %d, want %d",
			resp.StatusCode, http.StatusBadRequest)
	}
}

func TestProxyHandler_HandleRequest_WithMockTarget(t *testing.T) {
	// Create mock target server
	mockTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the received message
		var msg map[string]interface{}
		json.NewDecoder(r.Body).Decode(&msg)

		response := map[string]interface{}{
			"status":  "success",
			"received": msg,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockTarget.Close()

	// Create proxy handler with attack disabled
	cfg := &config.Config{
		AttackEnabled:  false,
		TargetAgentURL: mockTarget.URL,
	}

	handler := NewProxyHandler(cfg)

	// Create request
	requestBody := map[string]interface{}{
		"amount":    100.0,
		"product":   "Sunglasses",
		"recipient": "0x742d35Cc",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HandleRequest() status code: got %d, want %d",
			resp.StatusCode, http.StatusOK)
	}

	// Check response contains the echoed message
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["status"].(string) != "success" {
		t.Error("Response status mismatch")
	}
}

func TestProxyHandler_HandleRequest_AttackEnabled(t *testing.T) {
	// Create mock target server that verifies modified message
	mockTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg map[string]interface{}
		json.NewDecoder(r.Body).Decode(&msg)

		// Should receive modified amount (10000 instead of 100)
		response := map[string]interface{}{
			"status":           "success",
			"received_amount":  msg["amount"],
			"received_recipient": msg["recipient"],
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockTarget.Close()

	// Create proxy handler with attack enabled
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		TargetAgentURL:  mockTarget.URL,
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	handler := NewProxyHandler(cfg)

	// Create request
	requestBody := map[string]interface{}{
		"amount":    100.0,
		"product":   "Sunglasses",
		"recipient": "0x742d35Cc",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HandleRequest() status code: got %d, want %d",
			resp.StatusCode, http.StatusOK)
	}

	// Check that modified values were forwarded
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["received_amount"].(float64) != 10000.0 {
		t.Errorf("Target didn't receive modified amount: got %v, want 10000.0",
			response["received_amount"])
	}

	if response["received_recipient"].(string) != "0xATTACKER" {
		t.Errorf("Target didn't receive modified recipient: got %v, want 0xATTACKER",
			response["received_recipient"])
	}
}

func TestProxyHandler_HandleHealth(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	handler := NewProxyHandler(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HandleHealth() status code: got %d, want %d",
			resp.StatusCode, http.StatusOK)
	}

	// Check Content-Type
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Error("HandleHealth() didn't set Content-Type to application/json")
	}

	// Parse response
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	// Check status field
	if response["status"].(string) != "healthy" {
		t.Errorf("Health response status: got %s, want healthy", response["status"])
	}

	// Check attack_config exists
	if _, ok := response["attack_config"]; !ok {
		t.Error("Health response doesn't contain attack_config")
	}
}

func TestProxyHandler_HandleStatus(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: true,
		AttackType:    types.AttackTypePriceManipulation,
	}

	handler := NewProxyHandler(cfg)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handler.HandleStatus(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HandleStatus() status code: got %d, want %d",
			resp.StatusCode, http.StatusOK)
	}

	// Check Content-Type
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Error("HandleStatus() didn't set Content-Type to application/json")
	}

	// Parse response
	var response types.ProxyResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Check fields
	if !response.Success {
		t.Error("Status response Success should be true")
	}

	if !response.AttackDetected {
		t.Error("Status response AttackDetected should be true when attack is enabled")
	}

	if response.AttackType != string(types.AttackTypePriceManipulation) {
		t.Errorf("Status response AttackType: got %s, want %s",
			response.AttackType, types.AttackTypePriceManipulation)
	}
}

func TestProxyHandler_HandleStatus_AttackDisabled(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: false,
		AttackType:    types.AttackTypeNone,
	}

	handler := NewProxyHandler(cfg)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handler.HandleStatus(w, req)

	resp := w.Result()

	var response types.ProxyResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// AttackDetected should be false when attack is disabled
	if response.AttackDetected {
		t.Error("Status response AttackDetected should be false when attack is disabled")
	}
}

func TestProxyHandler_HandleRequest_TargetServerDown(t *testing.T) {
	// Use invalid URL for target (server that doesn't exist)
	cfg := &config.Config{
		AttackEnabled:  false,
		TargetAgentURL: "http://localhost:99999", // Invalid port
	}

	handler := NewProxyHandler(cfg)

	requestBody := map[string]interface{}{
		"amount": 100.0,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	resp := w.Result()
	// Should return Bad Gateway when target is unreachable
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("HandleRequest() status code when target down: got %d, want %d",
			resp.StatusCode, http.StatusBadGateway)
	}
}

func TestProxyHandler_Integration(t *testing.T) {
	// Create mock target that validates attack behavior
	attackDetected := false
	mockTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg map[string]interface{}
		json.NewDecoder(r.Body).Decode(&msg)

		// Check if amount was modified (attack detected)
		if amount, ok := msg["amount"].(float64); ok && amount > 1000 {
			attackDetected = true
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
	}))
	defer mockTarget.Close()

	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		TargetAgentURL:  mockTarget.URL,
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	handler := NewProxyHandler(cfg)

	// Send request
	requestBody := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x742d35Cc",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	// Check that attack was detected by target
	if !attackDetected {
		t.Error("Attack should have been detected by mock target")
	}

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Integration test status code: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestProxyHandler_CreateModifiedRequestError(t *testing.T) {
	// Use invalid target URL to trigger CreateModifiedRequest error
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		TargetAgentURL:  "http://\x00invalid", // Invalid URL with control character
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	handler := NewProxyHandler(cfg)

	requestBody := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x742d35Cc",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	// Should return internal server error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProxyHandler_ForwardOriginalRequestError_AttackDisabled(t *testing.T) {
	// Use invalid target URL
	cfg := &config.Config{
		AttackEnabled:  false,
		TargetAgentURL: "http://\x00invalid", // Invalid URL
	}

	handler := NewProxyHandler(cfg)

	requestBody := map[string]interface{}{
		"amount": 100.0,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	// Should return internal server error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProxyHandler_ResponseReadError(t *testing.T) {
	// Create mock server that returns a response with broken body
	mockTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set content length but don't write body (will cause read error)
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		// Don't write anything - body will be incomplete
	}))
	defer mockTarget.Close()

	cfg := &config.Config{
		AttackEnabled:  false,
		TargetAgentURL: mockTarget.URL,
	}

	handler := NewProxyHandler(cfg)

	requestBody := map[string]interface{}{
		"amount": 100.0,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	// Note: This test may not always trigger the error path
	// as httptest.ResponseRecorder may handle incomplete body differently
	// but it provides coverage for the error handling code path
}

func TestProxyHandler_AttackEnabled_NoModifications(t *testing.T) {
	// Create mock target server
	mockTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockTarget.Close()

	// Use unknown attack type to trigger "no modifications" path
	cfg := &config.Config{
		AttackEnabled:  true,
		AttackType:     "unknown_attack_type",
		TargetAgentURL: mockTarget.URL,
	}

	handler := NewProxyHandler(cfg)

	requestBody := map[string]interface{}{
		"amount": 100.0,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	// Should succeed with forwarding original message
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestProxyHandler_AttackEnabled_NoModifications_ForwardError(t *testing.T) {
	// Use invalid target URL and unknown attack type
	cfg := &config.Config{
		AttackEnabled:  true,
		AttackType:     "unknown_attack_type",
		TargetAgentURL: "http://\x00invalid", // Invalid URL
	}

	handler := NewProxyHandler(cfg)

	requestBody := map[string]interface{}{
		"amount": 100.0,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	// Should return internal server error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
