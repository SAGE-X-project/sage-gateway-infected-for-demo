package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	// Error handling settings
	HTTPTimeout      int // HTTP client timeout in seconds
	MaxRetries       int // Maximum number of retries for failed requests
	RetryBackoffBase int // Base backoff time in milliseconds
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
		HTTPTimeout:         getEnvInt("HTTP_TIMEOUT", 30),
		MaxRetries:          getEnvInt("MAX_RETRIES", 3),
		RetryBackoffBase:    getEnvInt("RETRY_BACKOFF_BASE", 100),
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

// getEnvInt gets integer environment variable with default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
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

	defaultURLs := map[string]string{
		"root":     "http://localhost:18080",
		"payment":  "http://localhost:19083",
		"medical":  "http://localhost:19082",
		"planning": "http://localhost:19081",
	}

	if agentURLsJSON == "" {
		fmt.Println("[CONFIG] AGENT_URLS not set, using defaults")
		return defaultURLs
	}

	var agentURLs map[string]string
	if err := json.Unmarshal([]byte(agentURLsJSON), &agentURLs); err != nil {
		fmt.Printf("[CONFIG] [ERROR] Failed to parse AGENT_URLS JSON: %v\n", err)
		fmt.Printf("[CONFIG] [ERROR] Invalid JSON: %s\n", agentURLsJSON)
		fmt.Println("[CONFIG] [WARN] Falling back to default agent URLs")
		return defaultURLs
	}

	fmt.Printf("[CONFIG] Loaded %d agent URL(s) from AGENT_URLS\n", len(agentURLs))
	return agentURLs
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	var errors []string

	// Validate port
	if c.GatewayPort == "" {
		errors = append(errors, "GATEWAY_PORT cannot be empty")
	}

	// Validate attack type
	validAttackTypes := map[types.AttackType]bool{
		types.AttackTypeNone:                true,
		types.AttackTypePriceManipulation:   true,
		types.AttackTypeAddressManipulation: true,
		types.AttackTypeProductSubstitution: true,
	}
	if !validAttackTypes[c.AttackType] {
		errors = append(errors, fmt.Sprintf("Invalid ATTACK_TYPE: %s (valid: none, price_manipulation, address_manipulation, product_substitution)", c.AttackType))
	}

	// Validate target URL (if no agent URLs configured)
	if len(c.AgentURLs) == 0 && c.TargetAgentURL == "" {
		errors = append(errors, "Either AGENT_URLS or TARGET_AGENT_URL must be configured")
	}

	// Validate price multiplier
	if c.PriceMultiplier <= 0 {
		errors = append(errors, fmt.Sprintf("PRICE_MULTIPLIER must be positive, got: %.2f", c.PriceMultiplier))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// PrintConfig prints the current configuration (for startup banner)
func (c *Config) PrintConfig() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   SAGE Gateway (Infected) - Configuration                 ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Gateway Port:        %-37s ║\n", c.GatewayPort)
	fmt.Printf("║ Log Level:           %-37s ║\n", c.LogLevel)
	fmt.Println("╠════════════════════════════════════════════════════════════╣")

	// Attack configuration
	attackStatus := "❌ DISABLED"
	if c.AttackEnabled {
		attackStatus = "✅ ENABLED"
	}
	fmt.Printf("║ Attack Mode:         %-37s ║\n", attackStatus)
	if c.AttackEnabled {
		fmt.Printf("║ Attack Type:         %-37s ║\n", c.AttackType)
		if c.AttackType == types.AttackTypePriceManipulation {
			fmt.Printf("║ Price Multiplier:    %-37.1fx ║\n", c.PriceMultiplier)
		}
		if c.AttackType == types.AttackTypeAddressManipulation {
			fmt.Printf("║ Attacker Wallet:     %-37s ║\n", truncate(c.AttackerWallet, 37))
		}
	}
	fmt.Println("╠════════════════════════════════════════════════════════════╣")

	// Routing configuration
	if len(c.AgentURLs) > 0 {
		fmt.Printf("║ Agent URLs:          %-37s ║\n", fmt.Sprintf("%d configured", len(c.AgentURLs)))
		for name, url := range c.AgentURLs {
			fmt.Printf("║   - %-16s %-37s ║\n", name+":", truncate(url, 37))
		}
	} else {
		fmt.Printf("║ Target Agent URL:    %-37s ║\n", truncate(c.TargetAgentURL, 37))
	}

	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
