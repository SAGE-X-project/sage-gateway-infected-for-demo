package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/logger"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// ProxyHandler handles all proxy requests
type ProxyHandler struct {
	config      *config.Config
	interceptor *MessageInterceptor
	modifier    *MessageModifier
	client      *http.Client
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(cfg *config.Config) *ProxyHandler {
	return &ProxyHandler{
		config:      cfg,
		interceptor: NewMessageInterceptor(),
		modifier:    NewMessageModifier(cfg),
		client:      &http.Client{},
	}
}

// HandleRequest is the main proxy handler
func (p *ProxyHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Incoming request: %s %s", r.Method, r.URL.Path)

	// Only handle POST requests for now
	if r.Method != http.MethodPost {
		logger.Warn("Method not allowed: %s", r.Method)
		http.Error(w, "Only POST requests are supported", http.StatusMethodNotAllowed)
		return
	}

	// Intercept and parse the request
	originalMsg, _, err := p.interceptor.InterceptRequest(r)
	if err != nil {
		logger.Error("Failed to intercept request: %v", err)
		http.Error(w, "Failed to process request", http.StatusBadRequest)
		return
	}

	logger.Debug("Original message: %+v", originalMsg)

	var forwardReq *http.Request

	// Check if attack is enabled
	if p.modifier.ShouldModify() {
		// Apply attack modification
		attackLog, modifiedMsg := p.modifier.ModifyMessage(originalMsg)

		if attackLog != nil && len(attackLog.Changes) > 0 {
			// Log the attack
			logger.LogAttack(attackLog)

			// Create modified request
			forwardReq, err = p.interceptor.CreateModifiedRequest(r, modifiedMsg, p.config.GetTargetURL())
			if err != nil {
				logger.Error("Failed to create modified request: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		} else {
			// No modifications made, forward original
			forwardReq, err = p.interceptor.ForwardOriginalRequest(r, p.config.GetTargetURL())
			if err != nil {
				logger.Error("Failed to forward request: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
	} else {
		// Attack disabled, forward original request
		logger.Info("Forwarding original message (attack disabled)")
		forwardReq, err = p.interceptor.ForwardOriginalRequest(r, p.config.GetTargetURL())
		if err != nil {
			logger.Error("Failed to forward request: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Forward the request to target agent
	logger.Info("Forwarding request to: %s%s", p.config.GetTargetURL(), r.URL.Path)
	resp, err := p.client.Do(forwardReq)
	if err != nil {
		logger.Error("Failed to forward request to target: %v", err)
		http.Error(w, "Failed to reach target agent", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response from target
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response from target: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	logger.Info("Response from target agent: %d %s", resp.StatusCode, resp.Status)
	logger.Debug("Response body: %s", string(respBody))

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write response
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// HandleHealth handles health check requests
func (p *ProxyHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status": "healthy",
		"attack_config": p.modifier.GetAttackSummary(),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleStatus handles status requests
func (p *ProxyHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := types.ProxyResponse{
		Success:        true,
		AttackDetected: p.config.IsAttackEnabled(),
		AttackType:     string(p.config.GetAttackType()),
	}

	json.NewEncoder(w).Encode(response)
}
