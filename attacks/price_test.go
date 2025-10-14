package attacks

import (
	"testing"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

func TestNewPriceAttack(t *testing.T) {
	cfg := &config.Config{
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	attack := NewPriceAttack(cfg)

	if attack == nil {
		t.Fatal("NewPriceAttack() returned nil")
	}
	if attack.config != cfg {
		t.Error("NewPriceAttack() didn't set config properly")
	}
}

func TestPriceAttack_ModifyMessage_WithAmount(t *testing.T) {
	cfg := &config.Config{
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER_WALLET",
	}

	attack := NewPriceAttack(cfg)

	originalMsg := map[string]interface{}{
		"amount":    100.0,
		"product":   "Sunglasses",
		"recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Check attack log
	if attackLog == nil {
		t.Fatal("ModifyMessage() returned nil attackLog")
	}
	if attackLog.AttackType != string(types.AttackTypePriceManipulation) {
		t.Errorf("AttackLog.AttackType: got %s, want %s",
			attackLog.AttackType, types.AttackTypePriceManipulation)
	}

	// Check modified message - amount should be multiplied
	if modifiedMsg["amount"].(float64) != 10000.0 {
		t.Errorf("Modified amount: got %f, want 10000.0", modifiedMsg["amount"].(float64))
	}

	// Check modified message - recipient should be attacker's wallet
	if modifiedMsg["recipient"].(string) != "0xATTACKER_WALLET" {
		t.Errorf("Modified recipient: got %s, want 0xATTACKER_WALLET",
			modifiedMsg["recipient"].(string))
	}

	// Check that original message is preserved
	if originalMsg["amount"].(float64) != 100.0 {
		t.Error("Original message was modified")
	}

	// Check changes in attack log
	if len(attackLog.Changes) < 2 {
		t.Errorf("AttackLog.Changes: got %d changes, want at least 2", len(attackLog.Changes))
	}

	// Verify amount change
	foundAmountChange := false
	for _, change := range attackLog.Changes {
		if change.Field == "amount" {
			foundAmountChange = true
			if change.OriginalValue.(float64) != 100.0 {
				t.Errorf("Amount change original value: got %v, want 100.0",
					change.OriginalValue)
			}
			if change.ModifiedValue.(float64) != 10000.0 {
				t.Errorf("Amount change modified value: got %v, want 10000.0",
					change.ModifiedValue)
			}
		}
	}
	if !foundAmountChange {
		t.Error("AttackLog doesn't contain amount change")
	}
}

func TestPriceAttack_ModifyMessage_WithoutAmount(t *testing.T) {
	cfg := &config.Config{
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER_WALLET",
	}

	attack := NewPriceAttack(cfg)

	// Message without amount field
	originalMsg := map[string]interface{}{
		"product":   "Sunglasses",
		"recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Attack log should still be created
	if attackLog == nil {
		t.Fatal("ModifyMessage() returned nil attackLog")
	}

	// Modified message should have recipient changed
	if modifiedMsg["recipient"].(string) != "0xATTACKER_WALLET" {
		t.Error("Recipient should be modified even without amount field")
	}

	// Description should be added
	if _, ok := modifiedMsg["description"]; !ok {
		t.Error("Description field should be added")
	}
}

func TestPriceAttack_ModifyMessage_WithDescription(t *testing.T) {
	cfg := &config.Config{
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER_WALLET",
	}

	attack := NewPriceAttack(cfg)

	originalMsg := map[string]interface{}{
		"amount":      100.0,
		"recipient":   "0x742d35Cc",
		"description": "Original description",
	}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Description should be modified
	if modifiedMsg["description"].(string) != "HACKED - Redirected to attacker" {
		t.Errorf("Description not modified: got %s", modifiedMsg["description"])
	}

	// Attack log should contain description change
	foundDescChange := false
	for _, change := range attackLog.Changes {
		if change.Field == "description" {
			foundDescChange = true
			if change.OriginalValue.(string) != "Original description" {
				t.Error("Description change original value mismatch")
			}
		}
	}
	if !foundDescChange {
		t.Error("AttackLog doesn't contain description change")
	}
}

func TestPriceAttack_ModifyMessage_DifferentMultipliers(t *testing.T) {
	tests := []struct {
		name       string
		multiplier float64
		original   float64
		expected   float64
	}{
		{"100x multiplier", 100.0, 100.0, 10000.0},
		{"10x multiplier", 10.0, 100.0, 1000.0},
		{"2x multiplier", 2.0, 100.0, 200.0},
		{"1.5x multiplier", 1.5, 100.0, 150.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				PriceMultiplier: tt.multiplier,
				AttackerWallet:  "0xATTACKER",
			}

			attack := NewPriceAttack(cfg)

			originalMsg := map[string]interface{}{
				"amount": tt.original,
			}

			_, modifiedMsg := attack.ModifyMessage(originalMsg)

			if modifiedMsg["amount"].(float64) != tt.expected {
				t.Errorf("Modified amount: got %f, want %f",
					modifiedMsg["amount"].(float64), tt.expected)
			}
		})
	}
}

func TestPriceAttack_GetAttackType(t *testing.T) {
	cfg := &config.Config{}
	attack := NewPriceAttack(cfg)

	attackType := attack.GetAttackType()

	if attackType != types.AttackTypePriceManipulation {
		t.Errorf("GetAttackType(): got %s, want %s",
			attackType, types.AttackTypePriceManipulation)
	}
}

func TestPriceAttack_ModifyMessage_PreservesOtherFields(t *testing.T) {
	cfg := &config.Config{
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	attack := NewPriceAttack(cfg)

	originalMsg := map[string]interface{}{
		"amount":    100.0,
		"product":   "Sunglasses",
		"recipient": "0x742d35Cc",
		"custom":    "custom_value",
		"count":     5,
	}

	_, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Check that non-attacked fields are preserved
	if modifiedMsg["product"].(string) != "Sunglasses" {
		t.Error("Product field was not preserved")
	}
	if modifiedMsg["custom"].(string) != "custom_value" {
		t.Error("Custom field was not preserved")
	}
	if modifiedMsg["count"].(int) != 5 {
		t.Error("Count field was not preserved")
	}
}

func TestPriceAttack_ModifyMessage_EmptyMessage(t *testing.T) {
	cfg := &config.Config{
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	attack := NewPriceAttack(cfg)

	originalMsg := map[string]interface{}{}

	attackLog, modifiedMsg := attack.ModifyMessage(originalMsg)

	// Should still create attack log
	if attackLog == nil {
		t.Fatal("ModifyMessage() returned nil attackLog for empty message")
	}

	// Should add description field
	if _, ok := modifiedMsg["description"]; !ok {
		t.Error("Description should be added to empty message")
	}
}

func TestPriceAttack_ModifyMessage_AttackLogTimestamp(t *testing.T) {
	cfg := &config.Config{
		PriceMultiplier: 100.0,
		AttackerWallet:  "0xATTACKER",
	}

	attack := NewPriceAttack(cfg)

	originalMsg := map[string]interface{}{
		"amount": 100.0,
	}

	attackLog, _ := attack.ModifyMessage(originalMsg)

	// Check that timestamp is set
	if attackLog.Timestamp.IsZero() {
		t.Error("AttackLog.Timestamp is zero")
	}

	// Check that timestamp is recent (within last second)
	// This is a simple check to ensure timestamp is being set
	if attackLog.Timestamp.Unix() == 0 {
		t.Error("AttackLog.Timestamp was not set properly")
	}
}
