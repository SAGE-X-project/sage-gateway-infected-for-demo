package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// A2AStatus represents the detected A2A protocol status
type A2AStatus struct {
	SAGEEnabled bool   // RFC 9421 signature detected
	HPKEEnabled bool   // HPKE encrypted payload detected
	SignatureID string // Signature identifier (e.g., "sig1")
	Algorithm   string // Signature algorithm
}

// DetectA2AProtocol detects if the request uses SAGE (RFC 9421) signatures and/or HPKE encryption
func DetectA2AProtocol(r *http.Request, body []byte) *A2AStatus {
	status := &A2AStatus{
		SAGEEnabled: false,
		HPKEEnabled: false,
	}

	// Check for RFC 9421 Signature headers
	signatureHeader := r.Header.Get("Signature")
	signatureInputHeader := r.Header.Get("Signature-Input")

	if signatureHeader != "" && signatureInputHeader != "" {
		status.SAGEEnabled = true

		// Parse signature ID and algorithm from Signature-Input header
		// Format: sig1=(...);created=...; keyid="..."
		if strings.Contains(signatureInputHeader, "sig1=") {
			status.SignatureID = "sig1"
		}

		// Try to extract algorithm from keyid or other fields
		if strings.Contains(signatureInputHeader, "ecdsa") {
			status.Algorithm = "ecdsa-p256-sha256"
		} else if strings.Contains(signatureInputHeader, "secp256k1") {
			status.Algorithm = "secp256k1"
		}
	}

	// Check for HPKE encrypted payload in the body
	if len(body) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			// Check for HPKE-related fields
			if _, hasEncryptedPayload := bodyMap["encryptedPayload"]; hasEncryptedPayload {
				status.HPKEEnabled = true
			}
			if _, hasEncData := bodyMap["enc_data"]; hasEncData {
				status.HPKEEnabled = true
			}
			if _, hasCiphertext := bodyMap["ciphertext"]; hasCiphertext {
				status.HPKEEnabled = true
			}

			// Check for SAGE transport.SecureMessage structure
			if msgType, ok := bodyMap["type"].(string); ok {
				if msgType == "secure" || msgType == "encrypted" {
					status.HPKEEnabled = true
				}
			}
		}
	}

	return status
}

// GetStatusString returns a human-readable status string
func (s *A2AStatus) GetStatusString() string {
	if s.SAGEEnabled && s.HPKEEnabled {
		return "SAGE: ✅ ON, HPKE: ✅ ON"
	} else if s.SAGEEnabled {
		return "SAGE: ✅ ON, HPKE: ❌ OFF"
	} else if s.HPKEEnabled {
		return "SAGE: ❌ OFF, HPKE: ✅ ON"
	}
	return "SAGE: ❌ OFF, HPKE: ❌ OFF"
}

// GetLogLevel returns the appropriate log level based on security status
func (s *A2AStatus) GetLogLevel() string {
	if s.SAGEEnabled {
		return "info"
	}
	return "warn"
}

// IsSecure returns true if any security protocol is enabled
func (s *A2AStatus) IsSecure() bool {
	return s.SAGEEnabled || s.HPKEEnabled
}

// GetSecurityDetails returns detailed security information
func (s *A2AStatus) GetSecurityDetails() map[string]interface{} {
	return map[string]interface{}{
		"sage_enabled":  s.SAGEEnabled,
		"hpke_enabled":  s.HPKEEnabled,
		"signature_id":  s.SignatureID,
		"algorithm":     s.Algorithm,
		"is_secure":     s.IsSecure(),
		"status_string": s.GetStatusString(),
	}
}
