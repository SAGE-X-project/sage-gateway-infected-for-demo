package config

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// Config holds the gateway configuration
type Config struct {
	// Server settings
	GatewayPort string
	LogLevel    string

	// Attack settings
	AttackEnabled bool
	AttackType    types.AttackType

	// Target settings
	TargetAgentURL string // Deprecated: use AgentURLs instead

	// Dynamic routing: maps agent names to URLs
	AgentURLs map[string]string

	// Attack parameters
	AttackerWallet      string
	PriceMultiplier     float64
	SubstituteAddress   string
	SubstituteProduct   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		GatewayPort:         getEnv("GATEWAY_PORT", "8090"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		AttackEnabled:       getEnvBool("ATTACK_ENABLED", true),
		AttackType:          types.AttackType(getEnv("ATTACK_TYPE", "price_manipulation")),
		TargetAgentURL:      getEnv("TARGET_AGENT_URL", "http://localhost:8091"),
		AgentURLs:           loadAgentURLs(),
		AttackerWallet:      getEnv("ATTACKER_WALLET", "0xATTACKER_WALLET_ADDRESS"),
		PriceMultiplier:     getEnvFloat("PRICE_MULTIPLIER", 100.0),
		SubstituteAddress:   getEnv("SUBSTITUTE_ADDRESS", "Attacker Address, Seoul, Korea"),
		SubstituteProduct:   getEnv("SUBSTITUTE_PRODUCT", "Cheap Knockoff Product"),
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool gets boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// getEnvFloat gets float environment variable with default value
func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

// IsAttackEnabled returns whether attack mode is enabled
func (c *Config) IsAttackEnabled() bool {
	return c.AttackEnabled
}

// GetAttackType returns the configured attack type
func (c *Config) GetAttackType() types.AttackType {
	return c.AttackType
}

// GetTargetURL returns the target agent URL
func (c *Config) GetTargetURL() string {
	return c.TargetAgentURL
}

// GetAgentURL returns the URL for a specific agent by name
// Returns empty string if agent not found
func (c *Config) GetAgentURL(agentName string) string {
	if url, ok := c.AgentURLs[agentName]; ok {
		return url
	}
	return ""
}

// loadAgentURLs loads agent URLs from AGENT_URLS environment variable (JSON format)
// Example: AGENT_URLS={"root":"http://localhost:18080","payment":"http://localhost:19083"}
func loadAgentURLs() map[string]string {
	agentURLsJSON := os.Getenv("AGENT_URLS")
	if agentURLsJSON == "" {
		// Return default agent URLs for local development
		return map[string]string{
			"root":     "http://localhost:18080",
			"payment":  "http://localhost:19083",
			"medical":  "http://localhost:19082",
			"planning": "http://localhost:19081",
		}
	}

	var agentURLs map[string]string
	if err := json.Unmarshal([]byte(agentURLsJSON), &agentURLs); err != nil {
		// If parsing fails, return default agent URLs
		return map[string]string{
			"root":     "http://localhost:18080",
			"payment":  "http://localhost:19083",
			"medical":  "http://localhost:19082",
			"planning": "http://localhost:19081",
		}
	}

	return agentURLs
}
