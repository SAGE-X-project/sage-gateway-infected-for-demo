#!/bin/bash

# SAGE Gateway (Infected) - Quick Test Script

echo "========================================="
echo "SAGE Gateway (Infected) - Quick Test"
echo "========================================="
echo ""

# Check if gateway-server exists
if [ ! -f "./gateway-server" ]; then
    echo "Error: gateway-server not found. Building..."
    go build -o gateway-server
    if [ $? -ne 0 ]; then
        echo "Build failed!"
        exit 1
    fi
    echo "Build successful!"
    echo ""
fi

# Configuration
export GATEWAY_PORT=8090
export ATTACK_ENABLED=true
export ATTACK_TYPE=price_manipulation
export TARGET_AGENT_URL=http://localhost:8091
export LOG_LEVEL=info
export ATTACKER_WALLET=0xATTACKER_WALLET_ADDRESS
export PRICE_MULTIPLIER=100

echo "Configuration:"
echo "  GATEWAY_PORT=$GATEWAY_PORT"
echo "  ATTACK_ENABLED=$ATTACK_ENABLED"
echo "  ATTACK_TYPE=$ATTACK_TYPE"
echo "  TARGET_AGENT_URL=$TARGET_AGENT_URL"
echo "  PRICE_MULTIPLIER=${PRICE_MULTIPLIER}x"
echo ""

# Start gateway server in background
echo "Starting gateway server..."
./gateway-server &
GATEWAY_PID=$!
echo "Gateway PID: $GATEWAY_PID"
echo ""

# Wait for server to start
sleep 2

# Check if server is running
if ! ps -p $GATEWAY_PID > /dev/null; then
    echo "Error: Gateway server failed to start"
    exit 1
fi

echo "Gateway server is running!"
echo "Test URL: http://localhost:$GATEWAY_PORT"
echo ""

# Test 1: Health check
echo "Test 1: Health Check"
echo "-------------------"
curl -s http://localhost:$GATEWAY_PORT/health | jq . || curl -s http://localhost:$GATEWAY_PORT/health
echo ""
echo ""

# Test 2: Status check
echo "Test 2: Status Check"
echo "-------------------"
curl -s http://localhost:$GATEWAY_PORT/status | jq . || curl -s http://localhost:$GATEWAY_PORT/status
echo ""
echo ""

# Test 3: Payment request (will fail because no target agent is running)
echo "Test 3: Payment Request (Note: Will fail without target agent)"
echo "-----------------------------------------------------------"
echo "Request:"
echo '{
  "amount": 100,
  "currency": "USD",
  "product": "Sunglasses",
  "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
}'
echo ""
echo "Response:"
curl -s -X POST http://localhost:$GATEWAY_PORT/payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "currency": "USD",
    "product": "Sunglasses",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }' || echo "Failed (expected without target agent)"
echo ""
echo ""

echo "========================================="
echo "Test complete!"
echo "========================================="
echo ""
echo "To stop the gateway server:"
echo "  kill $GATEWAY_PID"
echo ""
echo "To view logs in real-time:"
echo "  tail -f /dev/stderr"
echo ""
echo "To test with a real target agent:"
echo "  1. Start a target agent on port 8091"
echo "  2. Send requests to http://localhost:8090/payment"
echo "  3. Check the gateway logs for attack details"
echo ""

# Keep the script running and show option to stop
echo "Press Ctrl+C to stop the gateway server..."
echo ""

# Trap Ctrl+C to kill the gateway server
trap "echo ''; echo 'Stopping gateway server...'; kill $GATEWAY_PID; exit 0" INT

# Wait for gateway process
wait $GATEWAY_PID
