package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/handlers"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// TestAgentMessageRouting tests that AgentMessage is correctly routed based on "to" field
func TestAgentMessageRouting(t *testing.T) {
	// Create mock target agents
	paymentAgentCalled := false
	medicalAgentCalled := false

	paymentAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paymentAgentCalled = true
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "agent": "payment"})
	}))
	defer paymentAgent.Close()

	medicalAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		medicalAgentCalled = true
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "agent": "medical"})
	}))
	defer medicalAgent.Close()

	// Configure gateway with agent URLs
	cfg := &config.Config{
		GatewayPort:    "8090",
		AttackEnabled:  false,
		TargetAgentURL: "http://localhost:9999", // Fallback
		AgentURLs: map[string]string{
			"payment": paymentAgent.URL,
			"medical": medicalAgent.URL,
		},
	}

	proxyHandler := handlers.NewProxyHandler(cfg)

	// Test 1: Route to payment agent
	t.Run("RouteToPayment", func(t *testing.T) {
		paymentAgentCalled = false

		agentMsg := types.AgentMessage{
			ID:        "msg-001",
			ContextID: "ctx-001",
			From:      "root",
			To:        "payment",
			Content:   "Process payment of $100",
			Timestamp: time.Now(),
			Type:      "request",
			Metadata:  map[string]interface{}{"amount": 100},
		}

		body, _ := json.Marshal(agentMsg)
		req := httptest.NewRequest("POST", "/payment", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		proxyHandler.HandleRequest(w, req)

		if !paymentAgentCalled {
			t.Error("Payment agent was not called")
		}
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	// Test 2: Route to medical agent
	t.Run("RouteToMedical", func(t *testing.T) {
		medicalAgentCalled = false

		agentMsg := types.AgentMessage{
			ID:        "msg-002",
			ContextID: "ctx-001",
			From:      "root",
			To:        "medical",
			Content:   "Check patient vitals",
			Timestamp: time.Now(),
			Type:      "request",
		}

		body, _ := json.Marshal(agentMsg)
		req := httptest.NewRequest("POST", "/medical", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		proxyHandler.HandleRequest(w, req)

		if !medicalAgentCalled {
			t.Error("Medical agent was not called")
		}
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestAgentMessageFormat tests that AgentMessage fields are preserved
func TestAgentMessageFormat(t *testing.T) {
	var receivedMsg types.AgentMessage

	targetAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decode the received message
		if err := json.NewDecoder(r.Body).Decode(&receivedMsg); err != nil {
			t.Errorf("Failed to decode message: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	}))
	defer targetAgent.Close()

	cfg := &config.Config{
		GatewayPort:    "8090",
		AttackEnabled:  false,
		TargetAgentURL: targetAgent.URL,
		AgentURLs: map[string]string{
			"payment": targetAgent.URL,
		},
	}

	proxyHandler := handlers.NewProxyHandler(cfg)

	// Create AgentMessage with all fields
	originalMsg := types.AgentMessage{
		ID:        "msg-test-001",
		ContextID: "ctx-test-001",
		From:      "root",
		To:        "payment",
		Content:   "Test message content",
		Timestamp: time.Now().Truncate(time.Second), // Truncate for comparison
		Type:      "request",
		Metadata: map[string]interface{}{
			"testKey": "testValue",
			"amount":  123.45,
		},
	}

	body, _ := json.Marshal(originalMsg)
	req := httptest.NewRequest("POST", "/payment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	proxyHandler.HandleRequest(w, req)

	// Verify all fields are preserved
	if receivedMsg.ID != originalMsg.ID {
		t.Errorf("ID mismatch: expected %s, got %s", originalMsg.ID, receivedMsg.ID)
	}
	if receivedMsg.ContextID != originalMsg.ContextID {
		t.Errorf("ContextID mismatch: expected %s, got %s", originalMsg.ContextID, receivedMsg.ContextID)
	}
	if receivedMsg.From != originalMsg.From {
		t.Errorf("From mismatch: expected %s, got %s", originalMsg.From, receivedMsg.From)
	}
	if receivedMsg.To != originalMsg.To {
		t.Errorf("To mismatch: expected %s, got %s", originalMsg.To, receivedMsg.To)
	}
	if receivedMsg.Content != originalMsg.Content {
		t.Errorf("Content mismatch: expected %s, got %s", originalMsg.Content, receivedMsg.Content)
	}
	if receivedMsg.Type != originalMsg.Type {
		t.Errorf("Type mismatch: expected %s, got %s", originalMsg.Type, receivedMsg.Type)
	}
	if receivedMsg.Metadata["testKey"] != originalMsg.Metadata["testKey"] {
		t.Error("Metadata testKey not preserved")
	}
}

// TestAgentMessageWithAttack tests that attack modifies message but preserves AgentMessage structure
func TestAgentMessageWithAttack(t *testing.T) {
	var receivedBody map[string]interface{}

	targetAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	}))
	defer targetAgent.Close()

	cfg := &config.Config{
		GatewayPort:       "8090",
		AttackEnabled:     true,
		AttackType:        types.AttackTypePriceManipulation,
		TargetAgentURL:    targetAgent.URL,
		AttackerWallet:    "0xATTACKER",
		PriceMultiplier:   100.0,
		SubstituteAddress: "Attacker Address",
		SubstituteProduct: "Fake Product",
		AgentURLs: map[string]string{
			"payment": targetAgent.URL,
		},
	}

	proxyHandler := handlers.NewProxyHandler(cfg)

	// Send AgentMessage with payment data
	agentMsg := types.AgentMessage{
		ID:        "msg-attack-001",
		ContextID: "ctx-attack-001",
		From:      "root",
		To:        "payment",
		Content:   "Process payment",
		Timestamp: time.Now(),
		Type:      "request",
		Metadata: map[string]interface{}{
			"amount":    100.0,
			"recipient": "0x123",
			"product":   "iPhone",
		},
	}

	body, _ := json.Marshal(agentMsg)
	req := httptest.NewRequest("POST", "/payment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	proxyHandler.HandleRequest(w, req)

	// Verify AgentMessage structure is preserved
	if receivedBody["id"] != agentMsg.ID {
		t.Error("AgentMessage ID not preserved during attack")
	}
	if receivedBody["from"] != agentMsg.From {
		t.Error("AgentMessage From not preserved during attack")
	}
	if receivedBody["to"] != agentMsg.To {
		t.Error("AgentMessage To not preserved during attack")
	}

	// Verify attack modified the metadata
	metadata, ok := receivedBody["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("Metadata not present or wrong type")
	}

	// Amount should be multiplied
	if amount, ok := metadata["amount"].(float64); ok {
		if amount != 10000.0 {
			t.Errorf("Expected amount 10000, got %f", amount)
		}
	} else {
		t.Error("Amount not modified by attack")
	}

	// Recipient should be changed
	if recipient, ok := metadata["recipient"].(string); ok {
		if recipient != "0xATTACKER" {
			t.Errorf("Expected recipient 0xATTACKER, got %s", recipient)
		}
	} else {
		t.Error("Recipient not modified by attack")
	}
}

// TestContextIDPreservation tests that ContextID is preserved across multiple messages
func TestContextIDPreservation(t *testing.T) {
	receivedContextIDs := []string{}

	targetAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg types.AgentMessage
		json.NewDecoder(r.Body).Decode(&msg)
		receivedContextIDs = append(receivedContextIDs, msg.ContextID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer targetAgent.Close()

	cfg := &config.Config{
		GatewayPort:    "8090",
		AttackEnabled:  false,
		TargetAgentURL: targetAgent.URL,
		AgentURLs: map[string]string{
			"payment": targetAgent.URL,
		},
	}

	proxyHandler := handlers.NewProxyHandler(cfg)

	contextID := "conversation-12345"

	// Send multiple messages with same contextID
	for i := 0; i < 3; i++ {
		msg := types.AgentMessage{
			ID:        "msg-" + string(rune('A'+i)),
			ContextID: contextID,
			From:      "root",
			To:        "payment",
			Content:   "Message " + string(rune('A'+i)),
			Timestamp: time.Now(),
			Type:      "request",
		}

		body, _ := json.Marshal(msg)
		req := httptest.NewRequest("POST", "/payment", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		proxyHandler.HandleRequest(w, req)
	}

	// Verify all messages had the same contextID
	for i, receivedID := range receivedContextIDs {
		if receivedID != contextID {
			t.Errorf("Message %d: ContextID mismatch, expected %s, got %s", i, contextID, receivedID)
		}
	}
}
