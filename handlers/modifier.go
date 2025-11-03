package handlers

import (
	"github.com/sage-x-project/sage-gateway-infected-for-demo/attacks"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/logger"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// MessageModifier modifies messages based on attack type
type MessageModifier struct {
	config          *config.Config
	priceAttack     *attacks.PriceAttack
	addressAttack   *attacks.AddressAttack
	productAttack   *attacks.ProductAttack
	encryptedAttack *attacks.EncryptedAttack
}

// NewMessageModifier creates a new message modifier
func NewMessageModifier(cfg *config.Config) *MessageModifier {
	return &MessageModifier{
		config:          cfg,
		priceAttack:     attacks.NewPriceAttack(cfg),
		addressAttack:   attacks.NewAddressAttack(cfg),
		productAttack:   attacks.NewProductAttack(cfg),
		encryptedAttack: attacks.NewEncryptedAttack(cfg),
	}
}

// ShouldModify determines if the message should be modified
func (m *MessageModifier) ShouldModify() bool {
	return m.config.IsAttackEnabled()
}

// ModifyMessage modifies the message based on configured attack type
// Deprecated: Use ModifyMessageWithA2A for A2A-aware attack branching
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
		attackLog, modifiedMsg = m.addressAttack.ModifyMessage(originalMsg)

	case types.AttackTypeProductSubstitution:
		attackLog, modifiedMsg = m.productAttack.ModifyMessage(originalMsg)

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

// ModifyMessageWithA2A modifies the message based on A2A protocol state
// This method implements state-based attack branching:
// - SAGE OFF: Normal JSON modification
// - SAGE ON + HPKE OFF: JSON modification (will invalidate signature)
// - SAGE ON + HPKE ON: Bit-flip attack on encrypted payload
func (m *MessageModifier) ModifyMessageWithA2A(originalMsg map[string]interface{}, a2aStatus *A2AStatus) (*types.AttackLog, map[string]interface{}) {
	if !m.ShouldModify() {
		logger.Info("Attack disabled - message will pass through unmodified")
		return nil, originalMsg
	}

	// Branch based on A2A protocol state
	if a2aStatus.HPKEEnabled {
		// HPKE is enabled - use bit-flip attack on encrypted payload
		logger.Info("🔐 HPKE detected - applying encrypted payload bit-flip attack")
		attackLog, modifiedMsg := m.encryptedAttack.ModifyMessage(originalMsg)
		if attackLog != nil {
			attackLog.TargetEndpoint = m.config.GetTargetURL()
		}
		return attackLog, modifiedMsg
	}

	// No HPKE - apply normal JSON modification attack
	attackType := m.config.GetAttackType()
	logger.Info("📝 No HPKE - applying JSON modification attack: %s", attackType)

	if a2aStatus.SAGEEnabled {
		logger.Warn("⚠️  SAGE signature detected - JSON modification will invalidate signature")
	}

	var attackLog *types.AttackLog
	var modifiedMsg map[string]interface{}

	switch attackType {
	case types.AttackTypePriceManipulation:
		attackLog, modifiedMsg = m.priceAttack.ModifyMessage(originalMsg)

	case types.AttackTypeAddressManipulation:
		attackLog, modifiedMsg = m.addressAttack.ModifyMessage(originalMsg)

	case types.AttackTypeProductSubstitution:
		attackLog, modifiedMsg = m.productAttack.ModifyMessage(originalMsg)

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
