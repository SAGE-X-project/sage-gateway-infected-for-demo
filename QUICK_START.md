# Gateway Server - Quick Start Guide

## 빠른 시작

### 1. 빌드
```bash
go build -o gateway-server
```

### 2. 실행
```bash
# Attack mode enabled (default)
export ATTACK_ENABLED=true
export ATTACK_TYPE=price_manipulation
export GATEWAY_PORT=8090
export TARGET_AGENT_URL=http://localhost:8091

./gateway-server
```

### 3. 테스트
```bash
# 자동 테스트 스크립트 실행
chmod +x test_gateway.sh
./test_gateway.sh
```

---

## 사용 예시

### Health Check
```bash
curl http://localhost:8090/health
```

**응답:**
```json
{
  "status": "healthy",
  "attack_config": {
    "attack_enabled": true,
    "attack_type": "price_manipulation",
    "target_url": "http://localhost:8091",
    "price_multiplier": 100,
    "attacker_wallet": "0xATTACKER_WALLET_ADDRESS"
  }
}
```

### Payment Request (공격 시뮬레이션)
```bash
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "currency": "USD",
    "product": "Sunglasses",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }'
```

**로그 출력 (Attack Enabled):**
```
[ATTACK] ===== ATTACK DETECTED =====
[ATTACK] Type: price_manipulation
[ATTACK] Changes:
[ATTACK]   - Field: amount
[ATTACK]     Original: 100
[ATTACK]     Modified: 10000 (100x)
[ATTACK]   - Field: recipient
[ATTACK]     Original: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
[ATTACK]     Modified: 0xATTACKER_WALLET_ADDRESS
[ATTACK] ===========================
```

---

## 환경 변수

| 변수 | 설명 | 기본값 |
|-----|------|-------|
| `GATEWAY_PORT` | 서버 포트 | `8090` |
| `ATTACK_ENABLED` | 공격 활성화 | `true` |
| `ATTACK_TYPE` | 공격 유형 | `price_manipulation` |
| `TARGET_AGENT_URL` | 타겟 Agent URL | `http://localhost:8091` |
| `LOG_LEVEL` | 로그 레벨 | `info` |
| `ATTACKER_WALLET` | 공격자 지갑 | `0xATTACKER_WALLET_ADDRESS` |
| `PRICE_MULTIPLIER` | 금액 배율 | `100` |

---

## 데모 시나리오

### 시나리오 1: Attack Enabled (SAGE OFF)
1. Gateway 시작 (`ATTACK_ENABLED=true`)
2. Payment Agent 시작 (`SAGE_ENABLED=false`)
3. Frontend에서 "$100 결제" 요청
4. Gateway가 금액을 $10,000로 변조
5. Payment Agent가 변조된 금액으로 처리
6. **결과**: 공격 성공

### 시나리오 2: Attack Blocked (SAGE ON)
1. Gateway 시작 (`ATTACK_ENABLED=true`)
2. Payment Agent 시작 (`SAGE_ENABLED=true`)
3. Frontend에서 "$100 결제" 요청 (서명 포함)
4. Gateway가 금액을 $10,000로 변조
5. Payment Agent가 서명 검증 실패 → 거부
6. **결과**: 공격 차단

---

## 통합 테스트

### 전체 시스템 테스트 (Gateway + Payment Agent)

#### Terminal 1: Gateway Server
```bash
export ATTACK_ENABLED=true
export ATTACK_TYPE=price_manipulation
./gateway-server
```

#### Terminal 2: Payment Agent (Mock)
```bash
# 간단한 Mock Payment Agent
python3 -c "
from http.server import BaseHTTPRequestHandler, HTTPServer
import json

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length)
        data = json.loads(body)
        print(f'Received: {data}')
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        response = {'status': 'success', 'data': data}
        self.wfile.write(json.dumps(response).encode())
    def log_message(self, format, *args):
        print(f'[Payment Agent] {format % args}')

print('[Payment Agent] Starting on port 8091...')
HTTPServer(('', 8091), Handler).serve_forever()
"
```

#### Terminal 3: Test Request
```bash
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "product": "Sunglasses", "recipient": "0x742d35..."}'
```

**결과 확인:**
- Terminal 1 (Gateway): 공격 로그 표시
- Terminal 2 (Payment Agent): 변조된 메시지 수신 (amount: 10000)

---

## 문제 해결

### 포트 충돌
```bash
# 포트 사용 중인 프로세스 확인
lsof -i :8090

# 프로세스 종료
kill -9 <PID>
```

### 빌드 오류
```bash
# 의존성 정리
go mod tidy

# 재빌드
go clean
go build -o gateway-server
```

### Target Agent 연결 실패
```bash
# Target Agent URL 확인
echo $TARGET_AGENT_URL

# Target Agent 실행 여부 확인
curl http://localhost:8091/health

# URL 변경
export TARGET_AGENT_URL=http://localhost:8091
```

---

## 다음 단계

1. **Payment Agent 개발**: sage-payment-agent 프로젝트
2. **Frontend 연동**: sage-fe 프로젝트와 통합
3. **블록체인 연동**: Sepolia 테스트넷 배포
4. **AWS 배포**: 프로덕션 환경 구성

---

**작성일**: 2025-01-27
**버전**: 1.0.0
**상태**: 개발 완료
