package attacks

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/logger"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// EncryptedAttack handles attacks on encrypted payloads
type EncryptedAttack struct {
	config *config.Config
}

// NewEncryptedAttack creates a new encrypted attack handler
func NewEncryptedAttack(cfg *config.Config) *EncryptedAttack {
	return &EncryptedAttack{
		config: cfg,
	}
}

// ModifyMessage performs bit-flip attack on encrypted payload
func (a *EncryptedAttack) ModifyMessage(originalMsg map[string]interface{}) (*types.AttackLog, map[string]interface{}) {
	attackLog := &types.AttackLog{
		AttackType: "encrypted_payload_bitflip",
		Timestamp:  time.Now(),
		Changes:    []types.Change{},
	}

	// Make a copy of the original message
	modifiedMsg := make(map[string]interface{})
	for k, v := range originalMsg {
		modifiedMsg[k] = v
	}

	// Try to find and modify encrypted payload fields
	modified := false

	// Check for encryptedPayload field
	if encPayload, ok := originalMsg["encryptedPayload"].(string); ok && encPayload != "" {
		modifiedPayload := a.bitFlipPayload(encPayload)
		modifiedMsg["encryptedPayload"] = modifiedPayload
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "encryptedPayload",
			OriginalValue: fmt.Sprintf("<%d bytes>", len(encPayload)),
			ModifiedValue: fmt.Sprintf("<%d bytes, bit-flipped>", len(modifiedPayload)),
		})
		modified = true
		logger.Info("🔥 Bit-flip attack on encryptedPayload field")
	}

	// Check for ciphertext field
	if ciphertext, ok := originalMsg["ciphertext"].(string); ok && ciphertext != "" {
		modifiedPayload := a.bitFlipPayload(ciphertext)
		modifiedMsg["ciphertext"] = modifiedPayload
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "ciphertext",
			OriginalValue: fmt.Sprintf("<%d bytes>", len(ciphertext)),
			ModifiedValue: fmt.Sprintf("<%d bytes, bit-flipped>", len(modifiedPayload)),
		})
		modified = true
		logger.Info("🔥 Bit-flip attack on ciphertext field")
	}

	// Check for enc_data field
	if encData, ok := originalMsg["enc_data"].(string); ok && encData != "" {
		modifiedPayload := a.bitFlipPayload(encData)
		modifiedMsg["enc_data"] = modifiedPayload
		attackLog.Changes = append(attackLog.Changes, types.Change{
			Field:         "enc_data",
			OriginalValue: fmt.Sprintf("<%d bytes>", len(encData)),
			ModifiedValue: fmt.Sprintf("<%d bytes, bit-flipped>", len(modifiedPayload)),
		})
		modified = true
		logger.Info("🔥 Bit-flip attack on enc_data field")
	}

	if !modified {
		logger.Warn("No encrypted payload field found for bit-flip attack")
		return nil, originalMsg
	}

	return attackLog, modifiedMsg
}

// bitFlipPayload performs bit-flip on encrypted payload
// Strategy: Flip random bits in the payload to break HPKE integrity
func (a *EncryptedAttack) bitFlipPayload(payload string) string {
	// Try to decode as base64
	decodedBytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		// If not base64, treat as raw string and flip random bytes
		return a.bitFlipString(payload)
	}

	// Flip random bits in the decoded bytes
	modifiedBytes := make([]byte, len(decodedBytes))
	copy(modifiedBytes, decodedBytes)

	// Flip 3-5 random bits to break HPKE authentication
	numFlips := 3 + (randInt() % 3) // 3-5 flips
	for i := 0; i < numFlips; i++ {
		if len(modifiedBytes) > 0 {
			// Pick random byte position
			bytePos := randInt() % len(modifiedBytes)
			// Pick random bit position (0-7)
			bitPos := randInt() % 8
			// Flip the bit
			modifiedBytes[bytePos] ^= (1 << bitPos)
			logger.Debug("Bit flip at byte %d, bit %d", bytePos, bitPos)
		}
	}

	// Re-encode as base64
	return base64.StdEncoding.EncodeToString(modifiedBytes)
}

// bitFlipString flips random bytes in a string
func (a *EncryptedAttack) bitFlipString(s string) string {
	if len(s) == 0 {
		return s
	}

	bytes := []byte(s)
	numFlips := 3 + (randInt() % 3) // 3-5 flips

	for i := 0; i < numFlips && len(bytes) > 0; i++ {
		bytePos := randInt() % len(bytes)
		bitPos := randInt() % 8
		bytes[bytePos] ^= (1 << bitPos)
	}

	return string(bytes)
}

// randInt returns a random integer (helper function)
func randInt() int {
	var b [4]byte
	rand.Read(b[:])
	// Convert to positive int
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}
