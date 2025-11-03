# SAGE Gateway (Infected Demo) - Makefile
# Educational MitM attack simulation gateway

# Binary name
BINARY_NAME=gateway-infected
BINARY_PATH=./$(BINARY_NAME)

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Build info
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Colors for output
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
YELLOW=\033[0;33m
NC=\033[0m # No Color

.PHONY: all build clean run test help install deps check fmt lint

## all: Clean and build the gateway
all: clean build

## build: Build the gateway binary
build:
	@echo "$(BLUE)Building SAGE Gateway (Infected Demo)...$(NC)"
	@echo "$(YELLOW)Version: $(VERSION)$(NC)"
	@echo "$(YELLOW)Build Time: $(BUILD_TIME)$(NC)"
	@echo "$(YELLOW)Git Commit: $(GIT_COMMIT)$(NC)"
	@go build $(LDFLAGS) -o $(BINARY_PATH) .
	@echo "$(GREEN)✓ Build complete: $(BINARY_PATH)$(NC)"
	@echo ""
	@echo "$(BLUE)To run the gateway:$(NC)"
	@echo "  make run"
	@echo "  or"
	@echo "  ./$(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_PATH)
	@rm -f sage-gateway-infected
	@rm -f gateway
	@rm -rf $(GOBIN)
	@rm -f *.log
	@echo "$(GREEN)✓ Clean complete$(NC)"

## run: Build and run the gateway server
run: build
	@echo ""
	@echo "$(BLUE)================================================$(NC)"
	@echo "$(BLUE)Starting SAGE Gateway (Infected Demo)$(NC)"
	@echo "$(BLUE)================================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Configuration:$(NC)"
	@echo "  • HTTP Port: $${GATEWAY_PORT:-5500}"
	@echo "  • Root Agent Connection: http://localhost:$${GATEWAY_PORT:-5500}"
	@echo "  • WebSocket Logs: ws://localhost:$${GATEWAY_PORT:-5500}/ws/logs"
	@echo "  • Frontend Monitoring: Connect to above WebSocket endpoint"
	@echo ""
	@echo "$(YELLOW)Endpoints:$(NC)"
	@echo "  • / (proxy to agents)"
	@echo "  • /payment (payment agent proxy)"
	@echo "  • /order (ordering agent proxy)"
	@echo "  • /process (processing endpoint)"
	@echo "  • /health (health check)"
	@echo "  • /status (gateway status)"
	@echo "  • /ws/logs (WebSocket log streaming)"
	@echo ""
	@echo "$(GREEN)Starting server...$(NC)"
	@echo ""
	@$(BINARY_PATH)

## dev: Run with live reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		echo "$(BLUE)Running in development mode with air...$(NC)"; \
		air; \
	else \
		echo "$(RED)air not found. Install with: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(YELLOW)Falling back to normal run...$(NC)"; \
		$(MAKE) run; \
	fi

## test: Run all tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "$(BLUE)Generating coverage report...$(NC)"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

## test-gateway: Run gateway integration tests
test-gateway:
	@echo "$(BLUE)Running gateway integration tests...$(NC)"
	@if [ -f ./test_gateway.sh ]; then \
		chmod +x ./test_gateway.sh && ./test_gateway.sh; \
	else \
		echo "$(RED)test_gateway.sh not found$(NC)"; \
	fi

## test-websocket: Test WebSocket functionality
test-websocket:
	@echo "$(BLUE)Running WebSocket tests...$(NC)"
	@if [ -f ./test_websocket_client.sh ]; then \
		chmod +x ./test_websocket_client.sh && ./test_websocket_client.sh; \
	else \
		echo "$(RED)test_websocket_client.sh not found$(NC)"; \
	fi

## test-attack: Test attack scenarios
test-attack:
	@echo "$(BLUE)Running attack scenario tests...$(NC)"
	@if [ -f ./test_attack_scenarios.sh ]; then \
		chmod +x ./test_attack_scenarios.sh && ./test_attack_scenarios.sh; \
	else \
		echo "$(RED)test_attack_scenarios.sh not found$(NC)"; \
	fi

## install: Install the binary to $GOPATH/bin
install: build
	@echo "$(BLUE)Installing $(BINARY_NAME) to $$GOPATH/bin...$(NC)"
	@mkdir -p $(GOBIN)
	@cp $(BINARY_PATH) $(GOBIN)/
	@echo "$(GREEN)✓ Installed to $(GOBIN)/$(BINARY_NAME)$(NC)"

## deps: Download dependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

## tidy: Tidy go.mod
tidy:
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

## check: Run go vet and static checks
check:
	@echo "$(BLUE)Running static checks...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Checks passed$(NC)"

## fmt: Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## lint: Run golangci-lint (if installed)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "$(BLUE)Running linter...$(NC)"; \
		golangci-lint run ./...; \
		echo "$(GREEN)✓ Lint complete$(NC)"; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Skipping...$(NC)"; \
		echo "Install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi

## docker-build: Build Docker image
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t sage-gateway-infected:$(VERSION) .
	@docker tag sage-gateway-infected:$(VERSION) sage-gateway-infected:latest
	@echo "$(GREEN)✓ Docker image built$(NC)"

## docker-run: Run Docker container
docker-run:
	@echo "$(BLUE)Running Docker container...$(NC)"
	@docker run -p 5500:5500 --env-file .env sage-gateway-infected:latest

## status: Show build and runtime information
status:
	@echo "$(BLUE)SAGE Gateway Status$(NC)"
	@echo "$(YELLOW)────────────────────────────────────────$(NC)"
	@echo "Binary: $(BINARY_PATH)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo ""
	@if [ -f $(BINARY_PATH) ]; then \
		echo "$(GREEN)✓ Binary exists$(NC)"; \
		ls -lh $(BINARY_PATH); \
	else \
		echo "$(RED)✗ Binary not found$(NC)"; \
		echo "Run 'make build' to create it"; \
	fi
	@echo ""
	@echo "$(YELLOW)Environment Configuration:$(NC)"
	@if [ -f .env ]; then \
		echo "$(GREEN)✓ .env file found$(NC)"; \
		echo "Gateway Port: $${GATEWAY_PORT:-5500}"; \
	else \
		echo "$(YELLOW)⚠ .env file not found$(NC)"; \
		echo "Copy .env.example to .env and configure"; \
	fi

## setup: Setup development environment
setup: deps
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Creating .env from .env.example...$(NC)"; \
		cp .env.example .env; \
		echo "$(GREEN)✓ .env created - please edit with your configuration$(NC)"; \
	else \
		echo "$(GREEN)✓ .env already exists$(NC)"; \
	fi
	@echo ""
	@echo "$(BLUE)Setup complete!$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Edit .env with your configuration"
	@echo "  2. Run 'make build' to build the gateway"
	@echo "  3. Run 'make run' to start the server"

## help: Show this help message
help:
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║  SAGE Gateway (Infected Demo) - Makefile Commands           ║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Build & Run:$(NC)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /' | grep -E "build|run|clean|all|install|dev"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /' | grep "test"
	@echo ""
	@echo "$(YELLOW)Code Quality:$(NC)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /' | grep -E "check|fmt|lint|deps|tidy"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /' | grep "docker"
	@echo ""
	@echo "$(YELLOW)Utilities:$(NC)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /' | grep -E "setup|status|help"
	@echo ""
	@echo "$(BLUE)Quick Start:$(NC)"
	@echo "  1. $(GREEN)make setup$(NC)    - Initialize development environment"
	@echo "  2. $(GREEN)make build$(NC)    - Build the gateway binary"
	@echo "  3. $(GREEN)make run$(NC)      - Start the gateway server"
	@echo ""
	@echo "$(BLUE)Connection Points:$(NC)"
	@echo "  • Root Agent:       http://localhost:$${GATEWAY_PORT:-5500}"
	@echo "  • Frontend Monitor: ws://localhost:$${GATEWAY_PORT:-5500}/ws/logs"
	@echo ""

# Default target
.DEFAULT_GOAL := help
