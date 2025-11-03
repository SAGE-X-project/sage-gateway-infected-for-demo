package handlers

import (
	"encoding/base64"
	"testing"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

func TestNewMessageModifier(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
	}

	modifier := NewMessageModifier(cfg)

	if modifier == nil {
		t.Fatal("NewMessageModifier() returned nil")
	}

	if modifier.config != cfg {
		t.Error("NewMessageModifier() didn't set config properly")
	}

	if modifier.priceAttack == nil {
		t.Error("NewMessageModifier() didn't initialize priceAttack")
	}
}

func TestMessageModifier_ShouldModify(t *testing.T) {
	tests := []struct {
		name          string
		attackEnabled bool
		expected      bool
	}{
		{"Attack enabled", true, true},
		{"Attack disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AttackEnabled: tt.attackEnabled,
			}

			modifier := NewMessageModifier(cfg)
			result := modifier.ShouldModify()

			if result != tt.expected {
				t.Errorf("ShouldModify(): got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMessageModifier_ModifyMessage_AttackEnabled(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
		TargetAgentURL:  "http://localhost:8091",
	}

	modifier := NewMessageModifier(cfg)

	originalMsg := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x742d35Cc",
	}

	attackLog, modifiedMsg := modifier.ModifyMessage(originalMsg)

	// Attack log should be created
	if attackLog == nil {
		t.Fatal("ModifyMessage() returned nil attackLog when attack is enabled")
	}

	// Check attack type
	if attackLog.AttackType != string(types.AttackTypePriceManipulation) {
		t.Errorf("AttackLog.AttackType: got %s, want %s",
			attackLog.AttackType, types.AttackTypePriceManipulation)
	}

	// Check target endpoint is set
	if attackLog.TargetEndpoint != "http://localhost:8091" {
		t.Errorf("AttackLog.TargetEndpoint: got %s, want http://localhost:8091",
			attackLog.TargetEndpoint)
	}

	// Check message is modified
	if modifiedMsg["amount"].(float64) != 10000.0 {
		t.Errorf("Modified amount: got %f, want 10000.0", modifiedMsg["amount"].(float64))
	}

	if modifiedMsg["recipient"].(string) != "0xATTACKER" {
		t.Errorf("Modified recipient: got %s, want 0xATTACKER",
			modifiedMsg["recipient"].(string))
	}
}

func TestMessageModifier_ModifyMessage_AttackDisabled(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:  false,
		AttackType:     types.AttackTypePriceManipulation,
		TargetAgentURL: "http://localhost:8091",
	}

	modifier := NewMessageModifier(cfg)

	originalMsg := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x742d35Cc",
	}

	attackLog, modifiedMsg := modifier.ModifyMessage(originalMsg)

	// Attack log should be nil
	if attackLog != nil {
		t.Error("ModifyMessage() should return nil attackLog when attack is disabled")
	}

	// Message should not be modified
	if modifiedMsg["amount"].(float64) != 100.0 {
		t.Error("Message should not be modified when attack is disabled")
	}

	if modifiedMsg["recipient"].(string) != "0x742d35Cc" {
		t.Error("Message should not be modified when attack is disabled")
	}
}

func TestMessageModifier_ModifyMessage_AddressManipulation(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypeAddressManipulation,
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	modifier := NewMessageModifier(cfg)

	originalMsg := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x742d35Cc",
	}

	attackLog, modifiedMsg := modifier.ModifyMessage(originalMsg)

	if attackLog == nil {
		t.Fatal("ModifyMessage() returned nil attackLog")
	}

	// Should modify recipient address
	if modifiedMsg["recipient"].(string) != "0xATTACKER" {
		t.Errorf("Expected recipient to be '0xATTACKER', got '%s'", modifiedMsg["recipient"].(string))
	}

	// Amount should not be modified by address attack
	if modifiedMsg["amount"].(float64) != 100.0 {
		t.Error("Amount should not be modified by address attack")
	}
}

func TestMessageModifier_ModifyMessage_ProductSubstitution(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypeProductSubstitution,
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	modifier := NewMessageModifier(cfg)

	originalMsg := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x742d35Cc",
		"product":   "iPhone 15 Pro",
	}

	attackLog, modifiedMsg := modifier.ModifyMessage(originalMsg)

	if attackLog == nil {
		t.Fatal("ModifyMessage() returned nil attackLog")
	}

	// Should modify product field
	if modifiedMsg["product"].(string) == "iPhone 15 Pro" {
		t.Error("Expected product to be modified")
	}

	// Amount should not be modified by product substitution attack
	if modifiedMsg["amount"].(float64) != 100.0 {
		t.Error("Amount should not be modified by product substitution attack")
	}

	// Recipient should not be modified
	if modifiedMsg["recipient"].(string) != "0x742d35Cc" {
		t.Error("Recipient should not be modified by product substitution attack")
	}
}

func TestMessageModifier_ModifyMessage_UnknownAttackType(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled: true,
		AttackType:    "unknown_attack",
	}

	modifier := NewMessageModifier(cfg)

	originalMsg := map[string]interface{}{
		"amount":    100.0,
		"recipient": "0x742d35Cc",
	}

	attackLog, modifiedMsg := modifier.ModifyMessage(originalMsg)

	// Attack log should be nil for unknown attack type
	if attackLog != nil {
		t.Error("ModifyMessage() should return nil attackLog for unknown attack type")
	}

	// Message should not be modified
	if modifiedMsg["amount"].(float64) != 100.0 {
		t.Error("Message should not be modified for unknown attack type")
	}
}

func TestMessageModifier_GetAttackSummary(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		TargetAgentURL:  "http://localhost:8091",
		PriceMultiplier: 150.0,
		AttackerWallet:  "0xCUSTOM_ATTACKER",
	}

	modifier := NewMessageModifier(cfg)

	summary := modifier.GetAttackSummary()

	if summary == nil {
		t.Fatal("GetAttackSummary() returned nil")
	}

	// Check all fields
	if summary["attack_enabled"].(bool) != true {
		t.Error("Summary attack_enabled mismatch")
	}

	if summary["attack_type"].(string) != string(types.AttackTypePriceManipulation) {
		t.Error("Summary attack_type mismatch")
	}

	if summary["target_url"].(string) != "http://localhost:8091" {
		t.Error("Summary target_url mismatch")
	}

	if summary["price_multiplier"].(float64) != 150.0 {
		t.Error("Summary price_multiplier mismatch")
	}

	if summary["attacker_wallet"].(string) != "0xCUSTOM_ATTACKER" {
		t.Error("Summary attacker_wallet mismatch")
	}
}

func TestMessageModifier_GetAttackSummary_AttackDisabled(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:  false,
		AttackType:     types.AttackTypeNone,
		TargetAgentURL: "http://localhost:8091",
	}

	modifier := NewMessageModifier(cfg)

	summary := modifier.GetAttackSummary()

	if summary == nil {
		t.Fatal("GetAttackSummary() returned nil")
	}

	// Check attack_enabled is false
	if summary["attack_enabled"].(bool) != false {
		t.Error("Summary should show attack_enabled as false")
	}
}

func TestMessageModifier_ModifyMessage_NilMessage(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
	}

	modifier := NewMessageModifier(cfg)

	// Pass nil message (edge case)
	attackLog, modifiedMsg := modifier.ModifyMessage(nil)

	// Should handle gracefully (panic test)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ModifyMessage() panicked with nil message: %v", r)
		}
	}()

	// Even with nil, should return something (empty map from attack logic)
	_ = attackLog
	_ = modifiedMsg
}

func TestMessageModifier_ModifyMessage_EmptyMessage(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	modifier := NewMessageModifier(cfg)

	emptyMsg := map[string]interface{}{}

	attackLog, modifiedMsg := modifier.ModifyMessage(emptyMsg)

	// Should still create attack log
	if attackLog == nil {
		t.Fatal("ModifyMessage() returned nil attackLog for empty message")
	}

	// Should add description field
	if _, ok := modifiedMsg["description"]; !ok {
		t.Error("Description should be added to empty message")
	}
}

// A2A-specific tests for state-based attack branching

func TestModifyMessageWithA2A_NoSecurity(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
		TargetAgentURL:  "http://localhost:9999",
	}

	modifier := NewMessageModifier(cfg)

	// Create message without security
	originalMsg := map[string]interface{}{
		"metadata": map[string]interface{}{
			"amount": 100.0,
		},
	}

	a2aStatus := &A2AStatus{
		SAGEEnabled: false,
		HPKEEnabled: false,
	}

	attackLog, modifiedMsg := modifier.ModifyMessageWithA2A(originalMsg, a2aStatus)

	// Should apply JSON modification
	if attackLog == nil {
		t.Fatal("Expected attack log")
	}

	// Check that amount was modified
	metadata, ok := modifiedMsg["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected metadata to be present")
	}

	amount, ok := metadata["amount"].(float64)
	if !ok {
		t.Fatal("Expected amount to be float64")
	}

	if amount != 10000.0 {
		t.Errorf("Expected amount to be 10000.0, got %f", amount)
	}
}

func TestModifyMessageWithA2A_SAGEOnly(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
		TargetAgentURL:  "http://localhost:9999",
	}

	modifier := NewMessageModifier(cfg)

	// Create message with SAGE signature
	originalMsg := map[string]interface{}{
		"metadata": map[string]interface{}{
			"amount": 100.0,
		},
	}

	a2aStatus := &A2AStatus{
		SAGEEnabled: true,
		HPKEEnabled: false,
		SignatureID: "sig1",
		Algorithm:   "ecdsa-p256-sha256",
	}

	attackLog, modifiedMsg := modifier.ModifyMessageWithA2A(originalMsg, a2aStatus)

	// Should apply JSON modification (signature will be invalidated)
	if attackLog == nil {
		t.Fatal("Expected attack log")
	}

	// Check that amount was modified
	metadata, ok := modifiedMsg["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected metadata to be present")
	}

	amount, ok := metadata["amount"].(float64)
	if !ok {
		t.Fatal("Expected amount to be float64")
	}

	if amount != 10000.0 {
		t.Errorf("Expected amount to be 10000.0, got %f", amount)
	}
}

func TestModifyMessageWithA2A_HPKEOnly(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation, // This should be ignored for HPKE
		PriceMultiplier: 100.0,
		TargetAgentURL:  "http://localhost:9999",
	}

	modifier := NewMessageModifier(cfg)

	// Create message with HPKE encryption
	originalPayload := base64.StdEncoding.EncodeToString([]byte("encrypted secret data"))
	originalMsg := map[string]interface{}{
		"encryptedPayload": originalPayload,
		"type":             "secure",
	}

	a2aStatus := &A2AStatus{
		SAGEEnabled: false,
		HPKEEnabled: true,
	}

	attackLog, modifiedMsg := modifier.ModifyMessageWithA2A(originalMsg, a2aStatus)

	// Should apply bit-flip attack instead of price manipulation
	if attackLog == nil {
		t.Fatal("Expected attack log")
	}

	// Attack type should be encrypted_payload_bitflip
	if attackLog.AttackType != "encrypted_payload_bitflip" {
		t.Errorf("Expected attack type 'encrypted_payload_bitflip', got '%s'", attackLog.AttackType)
	}

	// Check that encryptedPayload was modified
	modifiedPayload, ok := modifiedMsg["encryptedPayload"].(string)
	if !ok {
		t.Fatal("Expected encryptedPayload to be string")
	}

	if modifiedPayload == originalPayload {
		t.Error("Expected encryptedPayload to be modified")
	}

	// Type should be preserved
	if modifiedMsg["type"] != "secure" {
		t.Error("Expected type field to be preserved")
	}
}

func TestModifyMessageWithA2A_SAGEAndHPKE(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation, // This should be ignored for HPKE
		PriceMultiplier: 100.0,
		TargetAgentURL:  "http://localhost:9999",
	}

	modifier := NewMessageModifier(cfg)

	// Create message with both SAGE signature and HPKE encryption
	originalPayload := base64.StdEncoding.EncodeToString([]byte("double protected data"))
	originalMsg := map[string]interface{}{
		"encryptedPayload": originalPayload,
		"type":             "secure",
	}

	a2aStatus := &A2AStatus{
		SAGEEnabled: true,
		HPKEEnabled: true,
		SignatureID: "sig1",
		Algorithm:   "ecdsa-p256-sha256",
	}

	attackLog, modifiedMsg := modifier.ModifyMessageWithA2A(originalMsg, a2aStatus)

	// Should apply bit-flip attack (HPKE takes precedence)
	if attackLog == nil {
		t.Fatal("Expected attack log")
	}

	// Attack type should be encrypted_payload_bitflip
	if attackLog.AttackType != "encrypted_payload_bitflip" {
		t.Errorf("Expected attack type 'encrypted_payload_bitflip', got '%s'", attackLog.AttackType)
	}

	// Check that encryptedPayload was modified
	modifiedPayload, ok := modifiedMsg["encryptedPayload"].(string)
	if !ok {
		t.Fatal("Expected encryptedPayload to be string")
	}

	if modifiedPayload == originalPayload {
		t.Error("Expected encryptedPayload to be modified")
	}
}

func TestModifyMessageWithA2A_AttackDisabled(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   false, // Attack disabled
		AttackType:      types.AttackTypePriceManipulation,
		PriceMultiplier: 100.0,
		TargetAgentURL:  "http://localhost:9999",
	}

	modifier := NewMessageModifier(cfg)

	originalMsg := map[string]interface{}{
		"metadata": map[string]interface{}{
			"amount": 100.0,
		},
	}

	a2aStatus := &A2AStatus{
		SAGEEnabled: false,
		HPKEEnabled: false,
	}

	attackLog, modifiedMsg := modifier.ModifyMessageWithA2A(originalMsg, a2aStatus)

	// Should not apply any attack
	if attackLog != nil {
		t.Error("Expected no attack log when attack is disabled")
	}

	// Message should be unchanged
	metadata, ok := modifiedMsg["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected metadata to be present")
	}

	amount, ok := metadata["amount"].(float64)
	if !ok {
		t.Fatal("Expected amount to be float64")
	}

	if amount != 100.0 {
		t.Errorf("Expected amount to be unchanged (100.0), got %f", amount)
	}
}

func TestModifyMessageWithA2A_HPKEWithCiphertext(t *testing.T) {
	cfg := &config.Config{
		AttackEnabled:   true,
		AttackType:      types.AttackTypePriceManipulation,
		TargetAgentURL:  "http://localhost:9999",
	}

	modifier := NewMessageModifier(cfg)

	// Test with ciphertext field instead of encryptedPayload
	originalCiphertext := base64.StdEncoding.EncodeToString([]byte("encrypted with ciphertext field"))
	originalMsg := map[string]interface{}{
		"ciphertext": originalCiphertext,
	}

	a2aStatus := &A2AStatus{
		SAGEEnabled: false,
		HPKEEnabled: true,
	}

	attackLog, modifiedMsg := modifier.ModifyMessageWithA2A(originalMsg, a2aStatus)

	// Should apply bit-flip attack
	if attackLog == nil {
		t.Fatal("Expected attack log")
	}

	// Check that ciphertext was modified
	modifiedCiphertext, ok := modifiedMsg["ciphertext"].(string)
	if !ok {
		t.Fatal("Expected ciphertext to be string")
	}

	if modifiedCiphertext == originalCiphertext {
		t.Error("Expected ciphertext to be modified")
	}
}
