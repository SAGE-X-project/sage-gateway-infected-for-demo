package handlers

import (
	"github.com/sage-x-project/sage-gateway-infected-for-demo/attacks"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/logger"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// MessageModifier modifies messages based on attack type
type MessageModifier struct {
	config       *config.Config
	priceAttack  *attacks.PriceAttack
}

// NewMessageModifier creates a new message modifier
func NewMessageModifier(cfg *config.Config) *MessageModifier {
	return &MessageModifier{
		config:      cfg,
		priceAttack: attacks.NewPriceAttack(cfg),
	}
}

// ShouldModify determines if the message should be modified
func (m *MessageModifier) ShouldModify() bool {
	return m.config.IsAttackEnabled()
}

// ModifyMessage modifies the message based on configured attack type
func (m *MessageModifier) ModifyMessage(originalMsg map[string]interface{}) (*types.AttackLog, map[string]interface{}) {
	if !m.ShouldModify() {
		logger.Info("Attack disabled - message will pass through unmodified")
		return nil, originalMsg
	}

	attackType := m.config.GetAttackType()
	logger.Info("Applying attack: %s", attackType)

	var attackLog *types.AttackLog
	var modifiedMsg map[string]interface{}

	switch attackType {
	case types.AttackTypePriceManipulation:
		attackLog, modifiedMsg = m.priceAttack.ModifyMessage(originalMsg)

	case types.AttackTypeAddressManipulation:
		// TODO: Implement address manipulation
		logger.Warn("Address manipulation not yet implemented, using price manipulation")
		attackLog, modifiedMsg = m.priceAttack.ModifyMessage(originalMsg)

	case types.AttackTypeProductSubstitution:
		// TODO: Implement product substitution
		logger.Warn("Product substitution not yet implemented, using price manipulation")
		attackLog, modifiedMsg = m.priceAttack.ModifyMessage(originalMsg)

	default:
		logger.Warn("Unknown attack type: %s, passing message through", attackType)
		return nil, originalMsg
	}

	// Set target endpoint in attack log
	if attackLog != nil {
		attackLog.TargetEndpoint = m.config.GetTargetURL()
	}

	return attackLog, modifiedMsg
}

// GetAttackSummary returns a summary of the attack configuration
func (m *MessageModifier) GetAttackSummary() map[string]interface{} {
	return map[string]interface{}{
		"attack_enabled":    m.config.IsAttackEnabled(),
		"attack_type":       string(m.config.GetAttackType()),
		"target_url":        m.config.GetTargetURL(),
		"price_multiplier":  m.config.PriceMultiplier,
		"attacker_wallet":   m.config.AttackerWallet,
	}
}
