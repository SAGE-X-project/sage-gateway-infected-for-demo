package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/types"
)

// WebSocketHub interface for broadcasting logs
type WebSocketHub interface {
	BroadcastLog(level, eventType, message string, data map[string]interface{})
	GetClientCount() int
}

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	ATTACK
)

var (
	logLevel     = INFO
	infoLogger   *log.Logger
	errorLogger  *log.Logger
	debugLogger  *log.Logger
	attackLogger *log.Logger
	wsHub        WebSocketHub // WebSocket hub for broadcasting logs
)

func init() {
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
	attackLogger = log.New(os.Stdout, "[ATTACK] ", log.Ldate|log.Ltime)
}

// SetLogLevel sets the logging level
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warn":
		logLevel = WARN
	case "error":
		logLevel = ERROR
	default:
		logLevel = INFO
	}
}

// SetWebSocketHub sets the WebSocket hub for broadcasting logs
func SetWebSocketHub(hub WebSocketHub) {
	wsHub = hub
	if hub != nil {
		log.Printf("[logger] WebSocket broadcasting enabled")
	}
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if logLevel <= DEBUG {
		debugLogger.Printf(format, v...)
		if wsHub != nil {
			message := fmt.Sprintf(format, v...)
			wsHub.BroadcastLog("debug", "debug", message, nil)
		}
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if logLevel <= INFO {
		infoLogger.Printf(format, v...)
		if wsHub != nil {
			message := fmt.Sprintf(format, v...)
			wsHub.BroadcastLog("info", "info", message, nil)
		}
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if logLevel <= WARN {
		infoLogger.Printf("[WARN] "+format, v...)
		if wsHub != nil {
			message := fmt.Sprintf(format, v...)
			wsHub.BroadcastLog("warn", "warn", message, nil)
		}
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if logLevel <= ERROR {
		errorLogger.Printf(format, v...)
		if wsHub != nil {
			message := fmt.Sprintf(format, v...)
			wsHub.BroadcastLog("error", "error", message, nil)
		}
	}
}

// LogAttack logs an attack event with detailed information
func LogAttack(attackLog *types.AttackLog) {
	attackLogger.Println("===== ATTACK DETECTED =====")
	attackLogger.Printf("Type: %s", attackLog.AttackType)
	attackLogger.Printf("Timestamp: %s", attackLog.Timestamp.Format(time.RFC3339))
	attackLogger.Printf("Target Endpoint: %s", attackLog.TargetEndpoint)
	attackLogger.Println("Changes:")

	for _, change := range attackLog.Changes {
		attackLogger.Printf("  - Field: %s", change.Field)
		attackLogger.Printf("    Original: %v", change.OriginalValue)
		attackLogger.Printf("    Modified: %v", change.ModifiedValue)
	}

	attackLogger.Println("===========================")

	// Broadcast attack to WebSocket clients
	if wsHub != nil {
		data := map[string]interface{}{
			"attack_type":     attackLog.AttackType,
			"timestamp":       attackLog.Timestamp.Format(time.RFC3339),
			"target_endpoint": attackLog.TargetEndpoint,
			"original_msg":    attackLog.OriginalMsg,
			"modified_msg":    attackLog.ModifiedMsg,
			"changes":         attackLog.Changes,
		}
		wsHub.BroadcastLog("warn", "attack", "Attack detected: "+attackLog.AttackType, data)
	}
}

// LogAttackSimple logs a simple attack message
func LogAttackSimple(format string, v ...interface{}) {
	attackLogger.Printf(format, v...)
}

// LogAttackBanner prints an attack banner
func LogAttackBanner() {
	banner := `
╔══════════════════════════════════════════════╗
║         ATTACK MODE ENABLED                  ║
║  This gateway will modify messages           ║
║  for demonstration purposes only             ║
╚══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// LogNormalModeBanner prints a normal mode banner
func LogNormalModeBanner() {
	banner := `
╔══════════════════════════════════════════════╗
║         TRANSPARENT PROXY MODE               ║
║  Messages will pass through unmodified       ║
╚══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
