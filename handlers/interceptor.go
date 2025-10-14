package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/logger"
)

// MessageInterceptor intercepts and parses HTTP messages
type MessageInterceptor struct{}

// NewMessageInterceptor creates a new message interceptor
func NewMessageInterceptor() *MessageInterceptor {
	return &MessageInterceptor{}
}

// InterceptRequest reads and parses the incoming request
func (i *MessageInterceptor) InterceptRequest(r *http.Request) (map[string]interface{}, []byte, error) {
	// Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body: %v", err)
		return nil, nil, err
	}

	// Close the original body
	r.Body.Close()

	// Restore the body for further processing
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse JSON
	var message map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &message); err != nil {
		logger.Error("Failed to parse JSON: %v", err)
		return nil, bodyBytes, err
	}

	logger.Debug("Intercepted request body: %s", string(bodyBytes))
	return message, bodyBytes, nil
}

// CreateModifiedRequest creates a new HTTP request with modified message
func (i *MessageInterceptor) CreateModifiedRequest(originalReq *http.Request, modifiedMsg map[string]interface{}, targetURL string) (*http.Request, error) {
	// Marshal modified message to JSON
	modifiedBody, err := json.Marshal(modifiedMsg)
	if err != nil {
		logger.Error("Failed to marshal modified message: %v", err)
		return nil, err
	}

	// Create new request with modified body
	newReq, err := http.NewRequest(originalReq.Method, targetURL+originalReq.URL.Path, bytes.NewBuffer(modifiedBody))
	if err != nil {
		logger.Error("Failed to create new request: %v", err)
		return nil, err
	}

	// Copy headers from original request
	for key, values := range originalReq.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	// Update Content-Length
	newReq.Header.Set("Content-Length", string(rune(len(modifiedBody))))
	newReq.ContentLength = int64(len(modifiedBody))

	logger.Debug("Created modified request to: %s", targetURL+originalReq.URL.Path)
	return newReq, nil
}

// ForwardOriginalRequest forwards the original request without modification
func (i *MessageInterceptor) ForwardOriginalRequest(originalReq *http.Request, targetURL string) (*http.Request, error) {
	// Read original body
	bodyBytes, err := io.ReadAll(originalReq.Body)
	if err != nil {
		return nil, err
	}
	originalReq.Body.Close()

	// Create new request with same body
	newReq, err := http.NewRequest(originalReq.Method, targetURL+originalReq.URL.Path, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Copy all headers
	for key, values := range originalReq.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	logger.Debug("Forwarding original request to: %s", targetURL+originalReq.URL.Path)
	return newReq, nil
}
