package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sage-x-project/sage-gateway-infected-for-demo/config"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/handlers"
	"github.com/sage-x-project/sage-gateway-infected-for-demo/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set log level
	logger.SetLogLevel(cfg.LogLevel)

	// Print banner
	printBanner()

	// Print attack status
	if cfg.IsAttackEnabled() {
		logger.LogAttackBanner()
		logger.Info("Attack Mode: ENABLED")
		logger.Info("Attack Type: %s", cfg.GetAttackType())
		logger.Info("Target URL: %s", cfg.GetTargetURL())
		logger.Info("Attacker Wallet: %s", cfg.AttackerWallet)
		logger.Info("Price Multiplier: %.1fx", cfg.PriceMultiplier)
	} else {
		logger.LogNormalModeBanner()
		logger.Info("Attack Mode: DISABLED")
		logger.Info("Operating in transparent proxy mode")
	}

	// Create proxy handler
	proxyHandler := handlers.NewProxyHandler(cfg)

	// Setup HTTP routes
	http.HandleFunc("/", proxyHandler.HandleRequest)
	http.HandleFunc("/payment", proxyHandler.HandleRequest)
	http.HandleFunc("/order", proxyHandler.HandleRequest)
	http.HandleFunc("/health", proxyHandler.HandleHealth)
	http.HandleFunc("/status", proxyHandler.HandleStatus)

	// Start server
	addr := ":" + cfg.GatewayPort
	logger.Info("Gateway server starting on port %s", cfg.GatewayPort)
	logger.Info("Listening on http://localhost%s", addr)

	// Setup graceful shutdown
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down gateway server...")
	os.Exit(0)
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║   SAGE Gateway (Infected) - Demo                            ║
║                                                              ║
║   WARNING: This gateway simulates Man-in-the-Middle attacks ║
║   for educational and demonstration purposes only            ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
