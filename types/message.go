package types

import "time"

// PaymentMessage represents a payment request message
type PaymentMessage struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency,omitempty"`
	Product     string  `json:"product,omitempty"`
	Recipient   string  `json:"recipient"`
	Sender      string  `json:"sender,omitempty"`
	Description string  `json:"description,omitempty"`
	Timestamp   int64   `json:"timestamp,omitempty"`
}

// OrderMessage represents an order request message
type OrderMessage struct {
	OrderID         string  `json:"order_id"`
	Product         string  `json:"product"`
	Quantity        int     `json:"quantity"`
	Amount          float64 `json:"amount"`
	ShippingAddress string  `json:"shipping_address"`
	Recipient       string  `json:"recipient"`
	Timestamp       int64   `json:"timestamp,omitempty"`
}

// GenericMessage represents any message type
type GenericMessage map[string]interface{}

// AttackLog represents an attack log entry
type AttackLog struct {
	Timestamp      time.Time              `json:"timestamp"`
	AttackType     string                 `json:"attack_type"`
	OriginalMsg    map[string]interface{} `json:"original_message"`
	ModifiedMsg    map[string]interface{} `json:"modified_message"`
	Changes        []Change               `json:"changes"`
	TargetEndpoint string                 `json:"target_endpoint"`
}

// Change represents a single field modification
type Change struct {
	Field         string      `json:"field"`
	OriginalValue interface{} `json:"original_value"`
	ModifiedValue interface{} `json:"modified_value"`
}

// AttackType represents the type of attack
type AttackType string

const (
	AttackTypePriceManipulation   AttackType = "price_manipulation"
	AttackTypeAddressManipulation AttackType = "address_manipulation"
	AttackTypeProductSubstitution AttackType = "product_substitution"
	AttackTypeNone                AttackType = "none"
)

// ProxyResponse represents the response from the proxy
type ProxyResponse struct {
	Success        bool                   `json:"success"`
	OriginalMsg    map[string]interface{} `json:"original_message,omitempty"`
	ModifiedMsg    map[string]interface{} `json:"modified_message,omitempty"`
	AttackDetected bool                   `json:"attack_detected"`
	AttackType     string                 `json:"attack_type,omitempty"`
	TargetResponse interface{}            `json:"target_response,omitempty"`
	Error          string                 `json:"error,omitempty"`
}
