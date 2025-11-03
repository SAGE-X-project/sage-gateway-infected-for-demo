package handlers

import (
	"bytes"
	"net/http"
	"testing"
)

func TestDetectA2AProtocol_NoSecurity(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"test":"data"}`))
	body := []byte(`{"test":"data"}`)

	status := DetectA2AProtocol(req, body)

	if status.SAGEEnabled {
		t.Error("Expected SAGE to be disabled")
	}
	if status.HPKEEnabled {
		t.Error("Expected HPKE to be disabled")
	}
	if status.IsSecure() {
		t.Error("Expected IsSecure to be false")
	}

	expectedString := "SAGE: ❌ OFF, HPKE: ❌ OFF"
	if status.GetStatusString() != expectedString {
		t.Errorf("Expected status '%s', got '%s'", expectedString, status.GetStatusString())
	}
}

func TestDetectA2AProtocol_SAGEOnly(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"test":"data"}`))
	req.Header.Set("Signature", "sig1=:ABC123:")
	req.Header.Set("Signature-Input", `sig1=("@method" "@path");created=1234567890;keyid="test-key"`)
	body := []byte(`{"test":"data"}`)

	status := DetectA2AProtocol(req, body)

	if !status.SAGEEnabled {
		t.Error("Expected SAGE to be enabled")
	}
	if status.HPKEEnabled {
		t.Error("Expected HPKE to be disabled")
	}
	if !status.IsSecure() {
		t.Error("Expected IsSecure to be true")
	}
	if status.SignatureID != "sig1" {
		t.Errorf("Expected signature ID 'sig1', got '%s'", status.SignatureID)
	}

	expectedString := "SAGE: ✅ ON, HPKE: ❌ OFF"
	if status.GetStatusString() != expectedString {
		t.Errorf("Expected status '%s', got '%s'", expectedString, status.GetStatusString())
	}
}

func TestDetectA2AProtocol_HPKEOnly(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", nil)
	body := []byte(`{"encryptedPayload":"base64encodeddata"}`)

	status := DetectA2AProtocol(req, body)

	if status.SAGEEnabled {
		t.Error("Expected SAGE to be disabled")
	}
	if !status.HPKEEnabled {
		t.Error("Expected HPKE to be enabled")
	}
	if !status.IsSecure() {
		t.Error("Expected IsSecure to be true")
	}

	expectedString := "SAGE: ❌ OFF, HPKE: ✅ ON"
	if status.GetStatusString() != expectedString {
		t.Errorf("Expected status '%s', got '%s'", expectedString, status.GetStatusString())
	}
}

func TestDetectA2AProtocol_SAGEAndHPKE(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Signature", "sig1=:ABC123:")
	req.Header.Set("Signature-Input", `sig1=("@method" "@path");created=1234567890;keyid="ecdsa-key"`)
	body := []byte(`{"encryptedPayload":"base64data","type":"secure"}`)

	status := DetectA2AProtocol(req, body)

	if !status.SAGEEnabled {
		t.Error("Expected SAGE to be enabled")
	}
	if !status.HPKEEnabled {
		t.Error("Expected HPKE to be enabled")
	}
	if !status.IsSecure() {
		t.Error("Expected IsSecure to be true")
	}
	if status.Algorithm != "ecdsa-p256-sha256" {
		t.Logf("Algorithm detected: %s", status.Algorithm)
	}

	expectedString := "SAGE: ✅ ON, HPKE: ✅ ON"
	if status.GetStatusString() != expectedString {
		t.Errorf("Expected status '%s', got '%s'", expectedString, status.GetStatusString())
	}
}

func TestDetectA2AProtocol_HPKEWithCiphertext(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", nil)
	body := []byte(`{"ciphertext":"encrypted_data_here"}`)

	status := DetectA2AProtocol(req, body)

	if !status.HPKEEnabled {
		t.Error("Expected HPKE to be enabled with ciphertext field")
	}
}

func TestDetectA2AProtocol_HPKEWithEncData(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", nil)
	body := []byte(`{"enc_data":"encrypted_payload"}`)

	status := DetectA2AProtocol(req, body)

	if !status.HPKEEnabled {
		t.Error("Expected HPKE to be enabled with enc_data field")
	}
}

func TestDetectA2AProtocol_InvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", nil)
	body := []byte(`{invalid json}`)

	status := DetectA2AProtocol(req, body)

	// Should not panic, just return false for HPKE
	if status.HPKEEnabled {
		t.Error("Expected HPKE to be disabled with invalid JSON")
	}
}

func TestA2AStatus_GetSecurityDetails(t *testing.T) {
	status := &A2AStatus{
		SAGEEnabled: true,
		HPKEEnabled: true,
		SignatureID: "sig1",
		Algorithm:   "ecdsa-p256-sha256",
	}

	details := status.GetSecurityDetails()

	if details["sage_enabled"] != true {
		t.Error("Expected sage_enabled to be true")
	}
	if details["hpke_enabled"] != true {
		t.Error("Expected hpke_enabled to be true")
	}
	if details["signature_id"] != "sig1" {
		t.Error("Expected signature_id to be sig1")
	}
	if details["algorithm"] != "ecdsa-p256-sha256" {
		t.Error("Expected algorithm to be ecdsa-p256-sha256")
	}
	if details["is_secure"] != true {
		t.Error("Expected is_secure to be true")
	}
}

func TestA2AStatus_GetLogLevel(t *testing.T) {
	secureStatus := &A2AStatus{SAGEEnabled: true}
	if secureStatus.GetLogLevel() != "info" {
		t.Error("Expected log level 'info' for secure status")
	}

	insecureStatus := &A2AStatus{SAGEEnabled: false}
	if insecureStatus.GetLogLevel() != "warn" {
		t.Error("Expected log level 'warn' for insecure status")
	}
}
