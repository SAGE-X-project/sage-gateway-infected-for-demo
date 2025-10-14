package config

import (
	"os"
	"testing"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all environment variables
	os.Clearenv()

	cfg := LoadConfig()

	// Test default values
	if cfg.GatewayPort != "8090" {
		t.Errorf("GatewayPort default: got %s, want 8090", cfg.GatewayPort)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel default: got %s, want info", cfg.LogLevel)
	}
	if !cfg.AttackEnabled {
		t.Error("AttackEnabled default: got false, want true")
	}
	if cfg.AttackType != types.AttackTypePriceManipulation {
		t.Errorf("AttackType default: got %s, want price_manipulation", cfg.AttackType)
	}
	if cfg.TargetAgentURL != "http://localhost:8091" {
		t.Errorf("TargetAgentURL default: got %s, want http://localhost:8091", cfg.TargetAgentURL)
	}
	if cfg.PriceMultiplier != 100.0 {
		t.Errorf("PriceMultiplier default: got %f, want 100.0", cfg.PriceMultiplier)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	// Set custom environment variables
	os.Setenv("GATEWAY_PORT", "9090")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ATTACK_ENABLED", "false")
	os.Setenv("ATTACK_TYPE", "address_manipulation")
	os.Setenv("TARGET_AGENT_URL", "http://localhost:9091")
	os.Setenv("ATTACKER_WALLET", "0xCUSTOM_WALLET")
	os.Setenv("PRICE_MULTIPLIER", "200.5")
	os.Setenv("SUBSTITUTE_ADDRESS", "Custom Address")
	os.Setenv("SUBSTITUTE_PRODUCT", "Custom Product")

	cfg := LoadConfig()

	// Test custom values
	if cfg.GatewayPort != "9090" {
		t.Errorf("GatewayPort: got %s, want 9090", cfg.GatewayPort)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel: got %s, want debug", cfg.LogLevel)
	}
	if cfg.AttackEnabled {
		t.Error("AttackEnabled: got true, want false")
	}
	if cfg.AttackType != "address_manipulation" {
		t.Errorf("AttackType: got %s, want address_manipulation", cfg.AttackType)
	}
	if cfg.TargetAgentURL != "http://localhost:9091" {
		t.Errorf("TargetAgentURL: got %s, want http://localhost:9091", cfg.TargetAgentURL)
	}
	if cfg.AttackerWallet != "0xCUSTOM_WALLET" {
		t.Errorf("AttackerWallet: got %s, want 0xCUSTOM_WALLET", cfg.AttackerWallet)
	}
	if cfg.PriceMultiplier != 200.5 {
		t.Errorf("PriceMultiplier: got %f, want 200.5", cfg.PriceMultiplier)
	}
	if cfg.SubstituteAddress != "Custom Address" {
		t.Errorf("SubstituteAddress: got %s, want Custom Address", cfg.SubstituteAddress)
	}
	if cfg.SubstituteProduct != "Custom Product" {
		t.Errorf("SubstituteProduct: got %s, want Custom Product", cfg.SubstituteProduct)
	}

	// Cleanup
	os.Clearenv()
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{"With env value", "TEST_KEY", "default", "custom", "custom"},
		{"Without env value", "TEST_KEY_MISSING", "default", "", "default"},
		{"Empty string env", "TEST_KEY_EMPTY", "default", "", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(): got %s, want %s", result, tt.expected)
			}
		})
	}

	os.Clearenv()
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{"True string", "TEST_BOOL", false, "true", true},
		{"False string", "TEST_BOOL", true, "false", false},
		{"1 string", "TEST_BOOL", false, "1", true},
		{"0 string", "TEST_BOOL", true, "0", false},
		{"Invalid string", "TEST_BOOL", true, "invalid", true},
		{"Empty string", "TEST_BOOL", true, "", true},
		{"Missing env", "TEST_BOOL_MISSING", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool(): got %v, want %v", result, tt.expected)
			}
		})
	}

	os.Clearenv()
}

func TestGetEnvFloat(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue float64
		envValue     string
		expected     float64
	}{
		{"Valid float", "TEST_FLOAT", 100.0, "200.5", 200.5},
		{"Integer", "TEST_FLOAT", 100.0, "300", 300.0},
		{"Invalid float", "TEST_FLOAT", 100.0, "invalid", 100.0},
		{"Empty string", "TEST_FLOAT", 100.0, "", 100.0},
		{"Missing env", "TEST_FLOAT_MISSING", 50.0, "", 50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnvFloat(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvFloat(): got %f, want %f", result, tt.expected)
			}
		})
	}

	os.Clearenv()
}

func TestConfig_IsAttackEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"Attack enabled", true},
		{"Attack disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{AttackEnabled: tt.enabled}
			result := cfg.IsAttackEnabled()
			if result != tt.enabled {
				t.Errorf("IsAttackEnabled(): got %v, want %v", result, tt.enabled)
			}
		})
	}
}

func TestConfig_GetAttackType(t *testing.T) {
	tests := []struct {
		name       string
		attackType types.AttackType
	}{
		{"Price manipulation", types.AttackTypePriceManipulation},
		{"Address manipulation", types.AttackTypeAddressManipulation},
		{"Product substitution", types.AttackTypeProductSubstitution},
		{"None", types.AttackTypeNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{AttackType: tt.attackType}
			result := cfg.GetAttackType()
			if result != tt.attackType {
				t.Errorf("GetAttackType(): got %s, want %s", result, tt.attackType)
			}
		})
	}
}

func TestConfig_GetTargetURL(t *testing.T) {
	tests := []struct {
		name      string
		targetURL string
	}{
		{"Default URL", "http://localhost:8091"},
		{"Custom URL", "http://custom-host:9000"},
		{"HTTPS URL", "https://secure-host.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{TargetAgentURL: tt.targetURL}
			result := cfg.GetTargetURL()
			if result != tt.targetURL {
				t.Errorf("GetTargetURL(): got %s, want %s", result, tt.targetURL)
			}
		})
	}
}

func TestLoadConfig_Integration(t *testing.T) {
	// Set realistic environment
	os.Setenv("GATEWAY_PORT", "8090")
	os.Setenv("ATTACK_ENABLED", "true")
	os.Setenv("ATTACK_TYPE", "price_manipulation")
	os.Setenv("TARGET_AGENT_URL", "http://localhost:8091")
	os.Setenv("PRICE_MULTIPLIER", "100")

	cfg := LoadConfig()

	// Verify integrated configuration
	if !cfg.IsAttackEnabled() {
		t.Error("Attack should be enabled")
	}
	if cfg.GetAttackType() != types.AttackTypePriceManipulation {
		t.Error("Attack type should be price_manipulation")
	}
	if cfg.GetTargetURL() != "http://localhost:8091" {
		t.Error("Target URL mismatch")
	}

	os.Clearenv()
}
