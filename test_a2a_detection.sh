#!/bin/bash

echo "🧪 A2A Protocol Detection Test"
echo "================================"
echo ""

# Start gateway
echo "1️⃣ Starting Gateway..."
ATTACK_ENABLED=false GATEWAY_PORT=8090 TARGET_AGENT_URL=http://localhost:9999 ./gateway-infected > /tmp/gateway_a2a.log 2>&1 &
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
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 1: No Security (SAGE OFF, HPKE OFF)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "recipient": "0x123"}' \
  -s > /dev/null

echo ""
echo "Expected log: SAGE: ❌ OFF, HPKE: ❌ OFF"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 2: SAGE Only (SAGE ON, HPKE OFF)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -H "Signature: sig1=:ABC123DEF456:" \
  -H 'Signature-Input: sig1=("@method" "@path");created=1234567890;keyid="test-key"' \
  -d '{"amount": 100}' \
  -s > /dev/null

echo ""
echo "Expected log: SAGE: ✅ ON, HPKE: ❌ OFF"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 3: HPKE Only (SAGE OFF, HPKE ON)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -d '{"encryptedPayload": "base64encodeddata"}' \
  -s > /dev/null

echo ""
echo "Expected log: SAGE: ❌ OFF, HPKE: ✅ ON"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 4: Full Security (SAGE ON, HPKE ON)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl -X POST http://localhost:8090/test \
  -H "Content-Type: application/json" \
  -H "Signature: sig1=:SECURE:" \
  -H 'Signature-Input: sig1=("@method" "@path");created=1234567890;keyid="ecdsa-key"' \
  -d '{"encryptedPayload": "encrypted_data", "type": "secure"}' \
  -s > /dev/null

echo ""
echo "Expected log: SAGE: ✅ ON, HPKE: ✅ ON"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 Gateway Logs"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat /tmp/gateway_a2a.log | grep -E "(Protocol detection|RFC 9421|HPKE|No RFC 9421)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧹 Cleanup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kill $GATEWAY_PID 2>/dev/null
kill $MOCK_PID 2>/dev/null
echo "✅ Test completed!"
