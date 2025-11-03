package attacks

import (
	"encoding/base64"
	"testing"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
)

func TestEncryptedAttack_ModifyMessage_EncryptedPayload(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: true,
		AttackType:    "encrypted_bitflip",
	}

	attack := NewEncryptedAttack(cfg)

	// Create test message with encryptedPayload
	originalPayload := base64.StdEncoding.EncodeToString([]byte("this is a secret encrypted message"))
	originalMsg := map[string]interface{}{
		"encryptedPayload": originalPayload,
		"type":             "secure",
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Check that attack was applied
	if attackLog == nil {
		t.Fatal("Expected attack log, got nil")
	}

	if len(attackLog.Changes) == 0 {
		t.Fatal("Expected at least one change in attack log")
	}

	// Check that encryptedPayload was modified
	modifiedPayload, ok := modifiedMsg["encryptedPayload"].(string)
	if !ok {
		t.Fatal("Expected encryptedPayload to be string")
	}

	if modifiedPayload == originalPayload {
		t.Error("Expected encryptedPayload to be modified, but it's unchanged")
	}

	// Check that type field is preserved
	if modifiedMsg["type"] != "secure" {
		t.Error("Expected type field to be preserved")
	}

	// Verify the change was logged
	if attackLog.Changes[0].Field != "encryptedPayload" {
		t.Errorf("Expected change field to be 'encryptedPayload', got '%s'", attackLog.Changes[0].Field)
	}
}

func TestEncryptedAttack_ModifyMessage_Ciphertext(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: true,
		AttackType:    "encrypted_bitflip",
	}

	attack := NewEncryptedAttack(cfg)

	// Create test message with ciphertext field
	originalCiphertext := base64.StdEncoding.EncodeToString([]byte("encrypted data here"))
	originalMsg := map[string]interface{}{
		"ciphertext": originalCiphertext,
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Check that attack was applied
	if attackLog == nil {
		t.Fatal("Expected attack log, got nil")
	}

	// Check that ciphertext was modified
	modifiedCiphertext, ok := modifiedMsg["ciphertext"].(string)
	if !ok {
		t.Fatal("Expected ciphertext to be string")
	}

	if modifiedCiphertext == originalCiphertext {
		t.Error("Expected ciphertext to be modified, but it's unchanged")
	}
}

func TestEncryptedAttack_ModifyMessage_EncData(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: true,
		AttackType:    "encrypted_bitflip",
	}

	attack := NewEncryptedAttack(cfg)

	// Create test message with enc_data field
	originalEncData := base64.StdEncoding.EncodeToString([]byte("more encrypted stuff"))
	originalMsg := map[string]interface{}{
		"enc_data": originalEncData,
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Check that attack was applied
	if attackLog == nil {
		t.Fatal("Expected attack log, got nil")
	}

	// Check that enc_data was modified
	modifiedEncData, ok := modifiedMsg["enc_data"].(string)
	if !ok {
		t.Fatal("Expected enc_data to be string")
	}

	if modifiedEncData == originalEncData {
		t.Error("Expected enc_data to be modified, but it's unchanged")
	}
}

func TestEncryptedAttack_ModifyMessage_NoEncryptedFields(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: true,
		AttackType:    "encrypted_bitflip",
	}

	attack := NewEncryptedAttack(cfg)

	// Create test message without encrypted fields
	originalMsg := map[string]interface{}{
		"amount":    100,
		"recipient": "0x123",
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Should return nil attack log when no encrypted fields found
	if attackLog != nil {
		t.Error("Expected nil attack log when no encrypted fields found")
	}

	// Message should be unchanged
	if modifiedMsg["amount"] != 100 {
		t.Error("Expected message to be unchanged")
	}
}

func TestEncryptedAttack_BitFlipPayload(t *testing.T) {
	cfg := &config.Config{}
	attack := NewEncryptedAttack(cfg)

	// Test with base64-encoded payload
	originalPayload := base64.StdEncoding.EncodeToString([]byte("test payload for bit flipping"))
	modifiedPayload := attack.bitFlipPayload(originalPayload)

	// Should be different
	if modifiedPayload == originalPayload {
		t.Error("Expected payload to be modified")
	}

	// Should still be valid base64
	_, err := base64.StdEncoding.DecodeString(modifiedPayload)
	if err != nil {
		t.Errorf("Expected modified payload to be valid base64, got error: %v", err)
	}

	// Decoded bytes should be different
	originalBytes, _ := base64.StdEncoding.DecodeString(originalPayload)
	modifiedBytes, _ := base64.StdEncoding.DecodeString(modifiedPayload)

	if len(originalBytes) != len(modifiedBytes) {
		t.Error("Expected same length after bit flip")
	}

	// At least one byte should be different
	different := false
	for i := range originalBytes {
		if originalBytes[i] != modifiedBytes[i] {
			different = true
			break
		}
	}

	if !different {
		t.Error("Expected at least one byte to be different after bit flip")
	}
}

func TestEncryptedAttack_BitFlipString(t *testing.T) {
	cfg := &config.Config{}
	attack := NewEncryptedAttack(cfg)

	// Test with non-base64 string
	originalString := "plain text payload"
	modifiedString := attack.bitFlipString(originalString)

	// Should be different
	if modifiedString == originalString {
		t.Error("Expected string to be modified")
	}

	// Should have same length
	if len(modifiedString) != len(originalString) {
		t.Error("Expected same length after bit flip")
	}
}

func TestEncryptedAttack_MultipleFields(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: true,
		AttackType:    "encrypted_bitflip",
	}

	attack := NewEncryptedAttack(cfg)

	// Create message with multiple encrypted fields
	originalMsg := map[string]interface{}{
		"encryptedPayload": base64.StdEncoding.EncodeToString([]byte("payload1")),
		"ciphertext":       base64.StdEncoding.EncodeToString([]byte("payload2")),
		"enc_data":         base64.StdEncoding.EncodeToString([]byte("payload3")),
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Should modify the first field found (encryptedPayload)
	if attackLog == nil || len(attackLog.Changes) == 0 {
		t.Fatal("Expected attack to be applied")
	}

	// At least one field should be modified
	encPayloadModified := modifiedMsg["encryptedPayload"] != originalMsg["encryptedPayload"]
	ciphertextModified := modifiedMsg["ciphertext"] != originalMsg["ciphertext"]
	encDataModified := modifiedMsg["enc_data"] != originalMsg["enc_data"]

	if !encPayloadModified && !ciphertextModified && !encDataModified {
		t.Error("Expected at least one encrypted field to be modified")
	}
}
