package handlers

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"
)

func TestNewMessageInterceptor(t *testing.T) {
	interceptor := NewMessageInterceptor()

	if interceptor == nil {
		t.Fatal("NewMessageInterceptor() returned nil")
	}
}

func TestInterceptRequest_ValidJSON(t *testing.T) {
	interceptor := NewMessageInterceptor()

	jsonBody := `{"amount": 100, "product": "Sunglasses", "recipient": "0x742d35Cc"}`
	req := httptest.NewRequest("POST", "/payment", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	message, bodyBytes, err := interceptor.InterceptRequest(req)

	if err != nil {
		t.Fatalf("InterceptRequest() error: %v", err)
	}

	if message == nil {
		t.Fatal("InterceptRequest() returned nil message")
	}

	if bodyBytes == nil {
		t.Fatal("InterceptRequest() returned nil bodyBytes")
	}

	// Check parsed message
	if message["amount"].(float64) != 100.0 {
		t.Errorf("Parsed amount: got %v, want 100.0", message["amount"])
	}

	if message["product"].(string) != "Sunglasses" {
		t.Errorf("Parsed product: got %v, want Sunglasses", message["product"])
	}

	// Check that body bytes match
	if string(bodyBytes) != jsonBody {
		t.Errorf("Body bytes mismatch: got %s, want %s", string(bodyBytes), jsonBody)
	}

	// Check that request body can be read again (restored)
	restoredBody, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read restored body: %v", err)
	}
	if string(restoredBody) != jsonBody {
		t.Error("Request body was not properly restored")
	}
}

func TestInterceptRequest_InvalidJSON(t *testing.T) {
	interceptor := NewMessageInterceptor()

	invalidJSON := `{"amount": invalid}`
	req := httptest.NewRequest("POST", "/payment", bytes.NewBufferString(invalidJSON))

	message, bodyBytes, err := interceptor.InterceptRequest(req)

	// Should return error for invalid JSON
	if err == nil {
		t.Error("InterceptRequest() should return error for invalid JSON")
	}

	// Message should be nil
	if message != nil {
		t.Error("InterceptRequest() should return nil message for invalid JSON")
	}

	// Body bytes should still be returned
	if bodyBytes == nil {
		t.Error("InterceptRequest() should return bodyBytes even for invalid JSON")
	}
}

func TestInterceptRequest_EmptyBody(t *testing.T) {
	interceptor := NewMessageInterceptor()

	req := httptest.NewRequest("POST", "/payment", bytes.NewBufferString(""))

	message, bodyBytes, err := interceptor.InterceptRequest(req)

	// Should return error for empty body
	if err == nil {
		t.Error("InterceptRequest() should return error for empty body")
	}

	// Check that empty body bytes are returned
	if len(bodyBytes) != 0 {
		t.Errorf("InterceptRequest() bodyBytes length: got %d, want 0", len(bodyBytes))
	}

	// Message should be nil
	if message != nil {
		t.Error("InterceptRequest() should return nil message for empty body")
	}
}

func TestCreateModifiedRequest(t *testing.T) {
	interceptor := NewMessageInterceptor()

	// Original request
	originalReq := httptest.NewRequest("POST", "/payment", bytes.NewBufferString("original"))
	originalReq.Header.Set("Content-Type", "application/json")
	originalReq.Header.Set("X-Custom-Header", "custom-value")

	// Modified message
	modifiedMsg := map[string]interface{}{
		"amount":    10000.0,
		"recipient": "0xATTACKER",
	}

	targetURL := "http://localhost:8091"

	newReq, err := interceptor.CreateModifiedRequest(originalReq, modifiedMsg, targetURL)

	if err != nil {
		t.Fatalf("CreateModifiedRequest() error: %v", err)
	}

	if newReq == nil {
		t.Fatal("CreateModifiedRequest() returned nil request")
	}

	// Check URL
	expectedURL := targetURL + "/payment"
	if newReq.URL.String() != expectedURL {
		t.Errorf("New request URL: got %s, want %s", newReq.URL.String(), expectedURL)
	}

	// Check method
	if newReq.Method != "POST" {
		t.Errorf("New request method: got %s, want POST", newReq.Method)
	}

	// Check headers are copied
	if newReq.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header was not copied")
	}

	if newReq.Header.Get("X-Custom-Header") != "custom-value" {
		t.Error("Custom header was not copied")
	}

	// Check body contains modified message
	body, err := io.ReadAll(newReq.Body)
	if err != nil {
		t.Fatalf("Failed to read new request body: %v", err)
	}

	if !bytes.Contains(body, []byte("10000")) {
		t.Error("New request body doesn't contain modified amount")
	}

	if !bytes.Contains(body, []byte("0xATTACKER")) {
		t.Error("New request body doesn't contain modified recipient")
	}
}

func TestForwardOriginalRequest(t *testing.T) {
	interceptor := NewMessageInterceptor()

	originalBody := `{"amount": 100}`
	originalReq := httptest.NewRequest("POST", "/payment", bytes.NewBufferString(originalBody))
	originalReq.Header.Set("Content-Type", "application/json")
	originalReq.Header.Set("Authorization", "Bearer token123")

	targetURL := "http://localhost:8091"

	newReq, err := interceptor.ForwardOriginalRequest(originalReq, targetURL)

	if err != nil {
		t.Fatalf("ForwardOriginalRequest() error: %v", err)
	}

	if newReq == nil {
		t.Fatal("ForwardOriginalRequest() returned nil request")
	}

	// Check URL
	expectedURL := targetURL + "/payment"
	if newReq.URL.String() != expectedURL {
		t.Errorf("Forwarded request URL: got %s, want %s", newReq.URL.String(), expectedURL)
	}

	// Check method
	if newReq.Method != "POST" {
		t.Errorf("Forwarded request method: got %s, want POST", newReq.Method)
	}

	// Check headers
	if newReq.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header was not copied")
	}

	if newReq.Header.Get("Authorization") != "Bearer token123" {
		t.Error("Authorization header was not copied")
	}

	// Check body is unchanged
	body, err := io.ReadAll(newReq.Body)
	if err != nil {
		t.Fatalf("Failed to read forwarded request body: %v", err)
	}

	if string(body) != originalBody {
		t.Errorf("Forwarded request body: got %s, want %s", string(body), originalBody)
	}
}

func TestCreateModifiedRequest_InvalidMessage(t *testing.T) {
	interceptor := NewMessageInterceptor()

	originalReq := httptest.NewRequest("POST", "/payment", bytes.NewBufferString(""))

	// Message with unmarshalable content (channel, which can't be marshaled to JSON)
	modifiedMsg := map[string]interface{}{
		"channel": make(chan int),
	}

	targetURL := "http://localhost:8091"

	_, err := interceptor.CreateModifiedRequest(originalReq, modifiedMsg, targetURL)

	// Should return error for unmarshalable message
	if err == nil {
		t.Error("CreateModifiedRequest() should return error for unmarshalable message")
	}
}

func TestForwardOriginalRequest_ReadError(t *testing.T) {
	interceptor := NewMessageInterceptor()

	// Create a reader that always returns an error
	errorReader := &errorReadCloser{err: io.ErrUnexpectedEOF}
	req := httptest.NewRequest("POST", "/payment", errorReader)

	targetURL := "http://localhost:8091"

	_, err := interceptor.ForwardOriginalRequest(req, targetURL)

	// Should return error when body reading fails
	if err == nil {
		t.Error("ForwardOriginalRequest() should return error when body reading fails")
	}
}

// Helper type for testing error scenarios
type errorReadCloser struct {
	err error
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReadCloser) Close() error {
	return nil
}

func TestInterceptRequest_ReadError(t *testing.T) {
	interceptor := NewMessageInterceptor()

	// Create a reader that always returns an error
	errorReader := &errorReadCloser{err: io.ErrUnexpectedEOF}
	req := httptest.NewRequest("POST", "/payment", errorReader)

	_, _, err := interceptor.InterceptRequest(req)

	// Should return error when body reading fails
	if err == nil {
		t.Error("InterceptRequest() should return error when body reading fails")
	}
}

func TestCreateModifiedRequest_InvalidURL(t *testing.T) {
	interceptor := NewMessageInterceptor()

	originalReq := httptest.NewRequest("POST", "/payment", bytes.NewBufferString(""))

	modifiedMsg := map[string]interface{}{
		"amount": 100.0,
	}

	// Use invalid URL with control characters that will cause http.NewRequest to fail
	invalidURL := "http://\x00invalid"

	_, err := interceptor.CreateModifiedRequest(originalReq, modifiedMsg, invalidURL)

	// Should return error for invalid URL
	if err == nil {
		t.Error("CreateModifiedRequest() should return error for invalid URL")
	}
}

func TestForwardOriginalRequest_InvalidURL(t *testing.T) {
	interceptor := NewMessageInterceptor()

	originalReq := httptest.NewRequest("POST", "/payment", bytes.NewBufferString("test"))

	// Use invalid URL with control characters
	invalidURL := "http://\x00invalid"

	_, err := interceptor.ForwardOriginalRequest(originalReq, invalidURL)

	// Should return error for invalid URL
	if err == nil {
		t.Error("ForwardOriginalRequest() should return error for invalid URL")
	}
}
