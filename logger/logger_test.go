package logger

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel LogLevel
	}{
		{"Debug level", "debug", DEBUG},
		{"Info level", "info", INFO},
		{"Warn level", "warn", WARN},
		{"Error level", "error", ERROR},
		{"Debug uppercase", "DEBUG", DEBUG},
		{"Unknown level", "unknown", INFO}, // defaults to INFO
		{"Empty string", "", INFO},         // defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLogLevel(tt.level)
			if logLevel != tt.expectedLevel {
				t.Errorf("SetLogLevel(%s): got %v, want %v", tt.level, logLevel, tt.expectedLevel)
			}
		})
	}

	// Reset to default
	SetLogLevel("info")
}

func TestDebug(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	debugLogger = log.New(&buf, "[DEBUG] ", log.Ldate|log.Ltime)

	// Set log level to DEBUG
	logLevel = DEBUG

	// Test debug logging
	Debug("Test debug message: %s", "hello")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Debug() should produce output when log level is DEBUG")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Test debug message: hello")) {
		t.Errorf("Debug() output doesn't contain expected message: %s", output)
	}

	// Test with log level higher than DEBUG
	buf.Reset()
	logLevel = INFO
	Debug("Should not appear")

	if buf.Len() > 0 {
		t.Error("Debug() should not produce output when log level is INFO")
	}

	// Reset
	debugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
	logLevel = INFO
}

func TestInfo(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)

	// Set log level to INFO
	logLevel = INFO

	// Test info logging
	Info("Test info message: %d", 123)

	output := buf.String()
	if len(output) == 0 {
		t.Error("Info() should produce output when log level is INFO")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Test info message: 123")) {
		t.Errorf("Info() output doesn't contain expected message: %s", output)
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
}

func TestWarn(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)

	// Set log level to WARN
	logLevel = WARN

	// Test warn logging
	Warn("Test warn message")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Warn() should produce output when log level is WARN")
	}
	if !bytes.Contains(buf.Bytes(), []byte("[WARN]")) {
		t.Errorf("Warn() output doesn't contain [WARN] prefix: %s", output)
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	logLevel = INFO
}

func TestError(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	errorLogger = log.New(&buf, "[ERROR] ", log.Ldate|log.Ltime)

	// Set log level to ERROR
	logLevel = ERROR

	// Test error logging
	Error("Test error message")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Error() should produce output when log level is ERROR")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Test error message")) {
		t.Errorf("Error() output doesn't contain expected message: %s", output)
	}

	// Reset
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
}

func TestLogAttack(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	attackLogger = log.New(&buf, "[ATTACK] ", log.Ldate|log.Ltime)

	// Create attack log
	attackLog := &types.AttackLog{
		Timestamp:  time.Now(),
		AttackType: "price_manipulation",
		Changes: []types.Change{
			{
				Field:         "amount",
				OriginalValue: 100.0,
				ModifiedValue: 10000.0,
			},
			{
				Field:         "recipient",
				OriginalValue: "0x742d35Cc",
				ModifiedValue: "0xATTACKER",
			},
		},
		TargetEndpoint: "http://localhost:8091",
	}

	// Test attack logging
	LogAttack(attackLog)

	output := buf.String()
	if len(output) == 0 {
		t.Error("LogAttack() should produce output")
	}

	// Verify output contains key information
	expectedStrings := []string{
		"ATTACK DETECTED",
		"price_manipulation",
		"amount",
		"recipient",
		"http://localhost:8091",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("LogAttack() output doesn't contain '%s': %s", expected, output)
		}
	}

	// Reset
	attackLogger = log.New(os.Stdout, "[ATTACK] ", log.Ldate|log.Ltime)
}

func TestLogAttackSimple(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	attackLogger = log.New(&buf, "[ATTACK] ", log.Ldate|log.Ltime)

	// Test simple attack logging
	LogAttackSimple("Simple attack message: %s", "test")

	output := buf.String()
	if len(output) == 0 {
		t.Error("LogAttackSimple() should produce output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Simple attack message: test")) {
		t.Errorf("LogAttackSimple() output doesn't contain expected message: %s", output)
	}

	// Reset
	attackLogger = log.New(os.Stdout, "[ATTACK] ", log.Ldate|log.Ltime)
}

func TestLogAttackBanner(t *testing.T) {
	// This function prints directly to stdout, so we just test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogAttackBanner() panicked: %v", r)
		}
	}()

	LogAttackBanner()
}

func TestLogNormalModeBanner(t *testing.T) {
	// This function prints directly to stdout, so we just test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogNormalModeBanner() panicked: %v", r)
		}
	}()

	LogNormalModeBanner()
}

func TestLogLevel_Hierarchy(t *testing.T) {
	// Test that log levels are properly ordered
	if DEBUG >= INFO {
		t.Error("DEBUG should be less than INFO")
	}
	if INFO >= WARN {
		t.Error("INFO should be less than WARN")
	}
	if WARN >= ERROR {
		t.Error("WARN should be less than ERROR")
	}
	if ERROR >= ATTACK {
		t.Error("ERROR should be less than ATTACK")
	}
}

func TestLogLevel_Filtering(t *testing.T) {
	tests := []struct {
		name        string
		setLevel    LogLevel
		testLevel   LogLevel
		shouldLog   bool
		logFunction func()
	}{
		{
			"DEBUG logs at DEBUG level",
			DEBUG,
			DEBUG,
			true,
			func() { Debug("test") },
		},
		{
			"DEBUG doesn't log at INFO level",
			INFO,
			DEBUG,
			false,
			func() { Debug("test") },
		},
		{
			"INFO logs at INFO level",
			INFO,
			INFO,
			true,
			func() { Info("test") },
		},
		{
			"INFO logs at DEBUG level",
			DEBUG,
			INFO,
			true,
			func() { Info("test") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer
			infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)
			debugLogger = log.New(&buf, "[DEBUG] ", log.Ldate|log.Ltime)

			// Set log level
			logLevel = tt.setLevel

			// Execute log function
			tt.logFunction()

			// Check if output was produced
			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("Expected log output: %v, got: %v", tt.shouldLog, hasOutput)
			}
		})
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
	logLevel = INFO
}
