#!/bin/bash

echo "🧪 Attack Scenarios Integration Test"
echo "====================================="
echo ""

# Start gateway
echo "1️⃣ Starting Gateway..."
ATTACK_ENABLED=true ATTACK_TYPE=price_manipulation PRICE_MULTIPLIER=100 GATEWAY_PORT=8090 TARGET_AGENT_URL=http://localhost:9999 ./gateway-infected > /tmp/gateway_attack.log 2>&1 &
GATEWAY_PID=$!
echo "   Gateway PID: $GATEWAY_PID"
sleep 2

# Start mock target
echo ""
echo "2️⃣ Starting mock target agent..."
python3 -c "
from http.server import BaseHTTPRequestHandler, HTTPServer
import json

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'status': 'ok'}).encode())
    def log_message(self, format, *args):
        pass

HTTPServer(('', 9999), Handler).serve_forever()
" 2>/dev/null &
MOCK_PID=$!
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Scenario 1: No Security + Price Attack"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "amount": 100
    }
  }' \
  -s > /dev/null

echo ""
echo "Expected: JSON modification (amount: 100 -> 10000)"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Scenario 2: SAGE Only + Price Attack"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -H "Signature: sig1=:ABC123DEF456:" \
  -H 'Signature-Input: sig1=("@method" "@path");created=1234567890;keyid="test-key"' \
  -d '{
    "metadata": {
      "amount": 100
    }
  }' \
  -s > /dev/null

echo ""
echo "Expected: JSON modification + signature invalidation warning"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Scenario 3: HPKE Only + Attack (switches to bit-flip)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -d '{
    "encryptedPayload": "dGhpcyBpcyBlbmNyeXB0ZWQgZGF0YQ==",
    "type": "secure"
  }' \
  -s > /dev/null

echo ""
echo "Expected: Bit-flip attack on encrypted payload (ignores price attack config)"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Scenario 4: SAGE + HPKE + Attack (bit-flip)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -H "Signature: sig1=:SECURE:" \
  -H 'Signature-Input: sig1=("@method" "@path");created=1234567890;keyid="ecdsa-key"' \
  -d '{
    "encryptedPayload": "ZG91YmxlIHByb3RlY3RlZCBkYXRh",
    "type": "secure"
  }' \
  -s > /dev/null

echo ""
echo "Expected: Bit-flip attack + warnings about signature AND HPKE invalidation"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Scenario 5: HPKE with ciphertext field"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -d '{
    "ciphertext": "Y2lwaGVydGV4dCBkYXRh"
  }' \
  -s > /dev/null

echo ""
echo "Expected: Bit-flip attack on ciphertext field"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Scenario 6: HPKE with enc_data field"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -d '{
    "enc_data": "ZW5jX2RhdGEgaGVyZQ=="
  }' \
  -s > /dev/null

echo ""
echo "Expected: Bit-flip attack on enc_data field"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 Gateway Logs (Attack-related)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat /tmp/gateway_attack.log | grep -E "(Protocol detection|Applying|SAGE|HPKE|Bit-flip|modification)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 Attack Logs"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat /tmp/gateway_attack.log | grep -E "(🚨 ATTACK|Changes:|Field:)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧹 Cleanup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kill $GATEWAY_PID 2>/dev/null
kill $MOCK_PID 2>/dev/null
echo "✅ Test completed!"
