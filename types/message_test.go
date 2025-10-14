package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPaymentMessage_JSON(t *testing.T) {
	msg := PaymentMessage{
		Amount:      100.0,
		Currency:    "USD",
		Product:     "Sunglasses",
		Recipient:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		Sender:      "0x1234567890abcdef",
		Description: "Test payment",
		Timestamp:   time.Now().Unix(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal PaymentMessage: %v", err)
	}

	// Test JSON unmarshaling
	var decoded PaymentMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal PaymentMessage: %v", err)
	}

	// Verify fields
	if decoded.Amount != msg.Amount {
		t.Errorf("Amount mismatch: got %f, want %f", decoded.Amount, msg.Amount)
	}
	if decoded.Currency != msg.Currency {
		t.Errorf("Currency mismatch: got %s, want %s", decoded.Currency, msg.Currency)
	}
	if decoded.Product != msg.Product {
		t.Errorf("Product mismatch: got %s, want %s", decoded.Product, msg.Product)
	}
	if decoded.Recipient != msg.Recipient {
		t.Errorf("Recipient mismatch: got %s, want %s", decoded.Recipient, msg.Recipient)
	}
}

func TestOrderMessage_JSON(t *testing.T) {
	msg := OrderMessage{
		OrderID:         "ORDER-123",
		Product:         "iPhone 15",
		Quantity:        2,
		Amount:          2000.0,
		ShippingAddress: "123 Main St, Seoul, Korea",
		Recipient:       "John Doe",
		Timestamp:       time.Now().Unix(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal OrderMessage: %v", err)
	}

	// Test JSON unmarshaling
	var decoded OrderMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal OrderMessage: %v", err)
	}

	// Verify fields
	if decoded.OrderID != msg.OrderID {
		t.Errorf("OrderID mismatch: got %s, want %s", decoded.OrderID, msg.OrderID)
	}
	if decoded.Amount != msg.Amount {
		t.Errorf("Amount mismatch: got %f, want %f", decoded.Amount, msg.Amount)
	}
	if decoded.Quantity != msg.Quantity {
		t.Errorf("Quantity mismatch: got %d, want %d", decoded.Quantity, msg.Quantity)
	}
}

func TestAttackLog(t *testing.T) {
	log := AttackLog{
		Timestamp:   time.Now(),
		AttackType:  string(AttackTypePriceManipulation),
		OriginalMsg: map[string]interface{}{"amount": 100.0},
		ModifiedMsg: map[string]interface{}{"amount": 10000.0},
		Changes: []Change{
			{
				Field:         "amount",
				OriginalValue: 100.0,
				ModifiedValue: 10000.0,
			},
		},
		TargetEndpoint: "http://localhost:8091",
	}

	// Test JSON marshaling
	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Failed to marshal AttackLog: %v", err)
	}

	// Test JSON unmarshaling
	var decoded AttackLog
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal AttackLog: %v", err)
	}

	// Verify fields
	if decoded.AttackType != log.AttackType {
		t.Errorf("AttackType mismatch: got %s, want %s", decoded.AttackType, log.AttackType)
	}
	if decoded.TargetEndpoint != log.TargetEndpoint {
		t.Errorf("TargetEndpoint mismatch: got %s, want %s", decoded.TargetEndpoint, log.TargetEndpoint)
	}
	if len(decoded.Changes) != len(log.Changes) {
		t.Errorf("Changes length mismatch: got %d, want %d", len(decoded.Changes), len(log.Changes))
	}
}

func TestChange(t *testing.T) {
	change := Change{
		Field:         "amount",
		OriginalValue: 100.0,
		ModifiedValue: 10000.0,
	}

	// Test JSON marshaling
	data, err := json.Marshal(change)
	if err != nil {
		t.Fatalf("Failed to marshal Change: %v", err)
	}

	// Test JSON unmarshaling
	var decoded Change
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Change: %v", err)
	}

	// Verify fields
	if decoded.Field != change.Field {
		t.Errorf("Field mismatch: got %s, want %s", decoded.Field, change.Field)
	}
}

func TestAttackType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		attack   AttackType
		expected string
	}{
		{"Price Manipulation", AttackTypePriceManipulation, "price_manipulation"},
		{"Address Manipulation", AttackTypeAddressManipulation, "address_manipulation"},
		{"Product Substitution", AttackTypeProductSubstitution, "product_substitution"},
		{"None", AttackTypeNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.attack) != tt.expected {
				t.Errorf("AttackType mismatch: got %s, want %s", tt.attack, tt.expected)
			}
		})
	}
}

func TestProxyResponse(t *testing.T) {
	resp := ProxyResponse{
		Success:        true,
		OriginalMsg:    map[string]interface{}{"amount": 100.0},
		ModifiedMsg:    map[string]interface{}{"amount": 10000.0},
		AttackDetected: true,
		AttackType:     string(AttackTypePriceManipulation),
		TargetResponse: map[string]interface{}{"status": "success"},
		Error:          "",
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ProxyResponse: %v", err)
	}

	// Test JSON unmarshaling
	var decoded ProxyResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProxyResponse: %v", err)
	}

	// Verify fields
	if decoded.Success != resp.Success {
		t.Errorf("Success mismatch: got %v, want %v", decoded.Success, resp.Success)
	}
	if decoded.AttackDetected != resp.AttackDetected {
		t.Errorf("AttackDetected mismatch: got %v, want %v", decoded.AttackDetected, resp.AttackDetected)
	}
	if decoded.AttackType != resp.AttackType {
		t.Errorf("AttackType mismatch: got %s, want %s", decoded.AttackType, resp.AttackType)
	}
}

func TestGenericMessage(t *testing.T) {
	msg := GenericMessage{
		"amount":    100.0,
		"product":   "Sunglasses",
		"recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal GenericMessage: %v", err)
	}

	// Test JSON unmarshaling
	var decoded GenericMessage
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal GenericMessage: %v", err)
	}

	// Verify fields
	if decoded["amount"].(float64) != msg["amount"].(float64) {
		t.Errorf("Amount mismatch")
	}
	if decoded["product"].(string) != msg["product"].(string) {
		t.Errorf("Product mismatch")
	}
}
