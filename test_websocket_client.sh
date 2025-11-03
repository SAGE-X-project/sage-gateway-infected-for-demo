#!/bin/bash

# WebSocket Test Client Script
# Tests the gateway's WebSocket log streaming functionality

echo "🧪 WebSocket Log Streaming Test"
echo "================================"
echo ""

# 1. Start a simple mock target server
echo "1️⃣ Starting mock target server on port 9999..."
python3 -c "
from http.server import BaseHTTPRequestHandler, HTTPServer
import json
import sys

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length)
        data = json.loads(body)
        print(f'[Mock Server] Received: {json.dumps(data, indent=2)}', file=sys.stderr)

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()

        response = {'status': 'success', 'received': data}
        self.wfile.write(json.dumps(response).encode())

    def log_message(self, format, *args):
        pass  # Suppress default logging

print('[Mock Server] Starting on port 9999...', file=sys.stderr)
HTTPServer(('', 9999), Handler).serve_forever()
" 2>&1 &

MOCK_PID=$!
echo "   Mock server PID: $MOCK_PID"
sleep 1

# 2. Send a test payment request
echo ""
echo "2️⃣ Sending test payment request (should trigger price manipulation attack)..."
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "currency": "USD",
    "product": "iPhone 15 Pro",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }' \
  -s | jq .

echo ""
echo "✅ Test completed!"
echo ""
echo "📋 To view WebSocket logs in real-time:"
echo "   Open test_websocket.html in your browser"
echo "   or use: websocat ws://localhost:8090/ws/logs"
echo ""
echo "🧹 Cleanup:"
echo "   kill $MOCK_PID  # Stop mock server"

# Cleanup
kill $MOCK_PID 2>/dev/null
