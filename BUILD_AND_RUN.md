# SAGE Gateway - Build and Run Guide

## 📋 Quick Start

### 1. Build the Gateway

```bash
make build
```

**Output:**
```
Building SAGE Gateway (Infected Demo)...
Version: 1.0.0
Build Time: 2025-11-03_17:16:56
Git Commit: 666751f
✓ Build complete: ./gateway-infected
```

### 2. Run the Gateway

```bash
make run
```

This will:
- Build the binary (if needed)
- Start the gateway server
- Listen on configured port (default: 5500)

**Output:**
```
================================================
Starting SAGE Gateway (Infected Demo)
================================================

Configuration:
  • HTTP Port: 5500
  • Root Agent Connection: http://localhost:5500
  • WebSocket Logs: ws://localhost:5500/ws/logs
  • Frontend Monitoring: Connect to above WebSocket endpoint

Endpoints:
  • / (proxy to agents)
  • /payment (payment agent proxy)
  • /order (ordering agent proxy)
  • /process (processing endpoint)
  • /health (health check)
  • /status (gateway status)
  • /ws/logs (WebSocket log streaming)

Starting server...
```

### 3. Clean Build Artifacts

```bash
make clean
```

Removes all build artifacts:
- `gateway-infected`
- `sage-gateway-infected`
- Log files
- Build directories

---

## 🔌 Connection Points

### For Root Agent

The gateway acts as a reverse proxy for agent communication:

```
Root Agent → http://localhost:5500 → Backend Agents
```

**Example:**
```bash
curl http://localhost:5500/payment -X POST \
  -H "Content-Type: application/json" \
  -d '{"to":"payment","body":"Transfer $100"}'
```

### For Frontend Monitoring

Connect to WebSocket endpoint for real-time log streaming:

```
Frontend → ws://localhost:5500/ws/logs → Gateway Logs
```

**Example (JavaScript):**
```javascript
const ws = new WebSocket('ws://localhost:5500/ws/logs');

ws.onmessage = (event) => {
  const log = JSON.parse(event.data);
  console.log(`[${log.level}] ${log.message}`);
};
```

**Example (HTML):**
Open `test_websocket.html` in your browser and connect to the WebSocket endpoint.

---

## 🛠️ Available Make Commands

### Build & Run

| Command | Description |
|---------|-------------|
| `make build` | Build the gateway binary |
| `make run` | Build and run the gateway server |
| `make clean` | Remove build artifacts |
| `make all` | Clean and build (alias for clean + build) |

### Testing

| Command | Description |
|---------|-------------|
| `make test` | Run all unit tests |
| `make test-coverage` | Run tests with coverage report |
| `make test-gateway` | Run gateway integration tests |
| `make test-websocket` | Test WebSocket functionality |
| `make test-attack` | Test attack scenarios |

### Development

| Command | Description |
|---------|-------------|
| `make dev` | Run with live reload (requires air) |
| `make fmt` | Format code with go fmt |
| `make check` | Run go vet and static checks |
| `make lint` | Run golangci-lint (if installed) |

### Utilities

| Command | Description |
|---------|-------------|
| `make setup` | Initialize development environment |
| `make status` | Show build and runtime information |
| `make deps` | Download dependencies |
| `make tidy` | Tidy go.mod |
| `make help` | Show all available commands |

---

## ⚙️ Configuration

### Environment Variables

Create a `.env` file (copy from `.env.example`):

```bash
cp .env.example .env
```

**Key Variables:**

```bash
# Gateway server port
GATEWAY_PORT=5500

# Logging level (debug, info, warn, error)
LOG_LEVEL=info

# Enable/disable attack mode
ATTACK_ENABLED=true

# Attack type (price_manipulation, address_substitution, product_substitution)
ATTACK_TYPE=price_manipulation

# Agent routing (JSON format)
AGENT_URLS={"root":"http://localhost:18080","payment":"http://localhost:19083","medical":"http://localhost:19082"}

# Attack parameters
ATTACKER_WALLET=0xATTACKER_WALLET_ADDRESS
PRICE_MULTIPLIER=100.0
```

### Port Configuration

The gateway listens on the port specified in `.env`:

```bash
GATEWAY_PORT=5500  # Default
```

To use a different port:

```bash
GATEWAY_PORT=8080 make run
```

---

## 🚀 Integration with SAGE Multi-Agent

### Step 1: Start Backend Agents

In `sage-multi-agent` directory:

```bash
# Start Root Agent
go run cmd/root/main.go

# Start Payment Agent
go run cmd/payment/main.go

# Start Medical Agent
go run cmd/medical/main.go
```

### Step 2: Start Gateway

In `sage-gateway-infected-for-demo` directory:

```bash
make run
```

### Step 3: Configure Root Agent

Update Root Agent's `.env` to use gateway:

```bash
# Instead of direct URLs:
PAYMENT_URL=http://localhost:19083

# Use gateway URL:
PAYMENT_URL=http://localhost:5500/payment
```

### Step 4: Connect Frontend

Use WebSocket client to monitor gateway logs:

```javascript
const ws = new WebSocket('ws://localhost:5500/ws/logs');
ws.onmessage = (event) => console.log(JSON.parse(event.data));
```

---

## 📊 Testing the Setup

### 1. Health Check

```bash
curl http://localhost:5500/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-04T02:16:56Z"
}
```

### 2. Gateway Status

```bash
curl http://localhost:5500/status
```

**Response:**
```json
{
  "gateway": "sage-gateway-infected",
  "version": "1.0.0",
  "attack_enabled": true,
  "attack_type": "price_manipulation",
  "target_agents": {
    "root": "http://localhost:18080",
    "payment": "http://localhost:19083"
  }
}
```

### 3. Send Test Message

```bash
curl http://localhost:5500/payment -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "from": "root",
    "to": "payment",
    "body": "Transfer 100 KRW to Alice",
    "timestamp": "2025-11-04T02:16:56Z"
  }'
```

### 4. Monitor Logs (WebSocket)

Open `test_websocket.html` in browser:

```bash
open test_websocket.html
```

Or use command line:

```bash
./test_websocket_client.sh
```

---

## 🐛 Troubleshooting

### Issue: Port Already in Use

**Error:**
```
listen tcp :5500: bind: address already in use
```

**Solution:**
```bash
# Find process using port 5500
lsof -i :5500

# Kill the process
kill -9 <PID>

# Or use a different port
GATEWAY_PORT=5501 make run
```

### Issue: Build Fails

**Error:**
```
go: module not found
```

**Solution:**
```bash
# Download dependencies
make deps

# Or
go mod download
```

### Issue: .env Not Found

**Error:**
```
⚠ .env file not found
```

**Solution:**
```bash
# Setup environment
make setup

# Or manually
cp .env.example .env
```

### Issue: Binary Not Executable

**Error:**
```
permission denied: ./gateway-infected
```

**Solution:**
```bash
chmod +x ./gateway-infected
```

---

## 📈 Performance Tips

### 1. Development Mode

For faster development cycles with auto-reload:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run in dev mode
make dev
```

### 2. Production Build

For optimized production builds:

```bash
# Build with optimizations
go build -ldflags="-s -w" -o gateway-infected .

# Result is smaller binary
```

### 3. Monitor Performance

Check gateway status:

```bash
make status
```

View logs in real-time:

```bash
# Via WebSocket
./test_websocket_client.sh

# Or watch log file (if file logging enabled)
tail -f gateway.log
```

---

## 🔐 Security Notes

### Attack Mode

When `ATTACK_ENABLED=true`:
- Gateway modifies messages in transit
- For **DEMO PURPOSES ONLY**
- **NOT FOR PRODUCTION USE**

### Safe Mode

To run as transparent proxy:

```bash
ATTACK_ENABLED=false make run
```

### API Key Security

**Never commit `.env` files:**
- Already in `.gitignore`
- Use environment-specific configs
- Rotate keys regularly

---

## 📦 Docker Support (Optional)

### Build Docker Image

```bash
make docker-build
```

### Run in Docker

```bash
make docker-run
```

### Custom Docker Run

```bash
docker run -d \
  -p 5500:5500 \
  --env-file .env \
  --name sage-gateway \
  sage-gateway-infected:latest
```

---

## 🔄 Continuous Integration

### Pre-commit Checks

Before committing:

```bash
make check   # Run static analysis
make fmt     # Format code
make test    # Run tests
make lint    # Run linter (if installed)
```

### All-in-one Check

```bash
make check && make test && make fmt
```

---

## 📚 Additional Resources

- **Main Documentation:** [README.md](./README.md)
- **Quick Start:** [QUICK_START.md](./QUICK_START.md)
- **WebSocket Guide:** [WEBSOCKET_IMPLEMENTATION.md](./WEBSOCKET_IMPLEMENTATION.md)
- **Attack Details:** [STATE_BASED_ATTACK_IMPLEMENTATION.md](./STATE_BASED_ATTACK_IMPLEMENTATION.md)
- **Test Reports:** [TEST_REPORT.md](./TEST_REPORT.md)

---

## 💡 Common Workflows

### Development Workflow

```bash
# 1. Setup environment
make setup

# 2. Make changes to code
vim main.go

# 3. Build and test
make build
make test

# 4. Run locally
make run
```

### Testing Workflow

```bash
# 1. Run all tests
make test

# 2. Run specific tests
make test-gateway
make test-websocket
make test-attack

# 3. Generate coverage report
make test-coverage
```

### Release Workflow

```bash
# 1. Clean previous builds
make clean

# 2. Run all checks
make check && make test

# 3. Build production binary
make build

# 4. Test the binary
./gateway-infected
```

---

**Last Updated:** 2025-11-04
**Version:** 1.0.0
