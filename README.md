# SAGE Gateway (Infected) - Demo

중간자 공격(Man-in-the-Middle Attack)을 시뮬레이션하는 Gateway 서버입니다.
SAGE 프로토콜의 보안 효과를 시연하기 위한 데모용 컴포넌트입니다.

## 목적

이 Gateway 서버는 **의도적으로** 메시지를 가로채고 변조하여, SAGE 프로토콜이 없을 때 발생할 수 있는 보안 위협을 시연합니다.

- **SAGE OFF**: 변조된 메시지가 그대로 전달됨 (공격 성공)
- **SAGE ON**: 서명 검증 실패로 변조 탐지 (공격 차단)

## 주요 기능

### 1. HTTP 프록시 서버
- Agent 간 통신을 중계하는 프록시 역할
- 모든 HTTP 요청/응답을 가로채기

### 2. 메시지 변조 (Attack Types)

#### Price Manipulation (금액 변조)
```json
// 원본 메시지
{"amount": 100, "recipient": "0xVENDOR"}

// 변조된 메시지
{"amount": 10000, "recipient": "0xATTACKER"}
```

#### Address Manipulation (주소 변조)
```json
// 원본
{"shipping_address": "서울시 강남구"}

// 변조
{"shipping_address": "공격자 주소"}
```

#### Product Substitution (상품 변조)
```json
// 원본
{"product": "iPhone 15 Pro"}

// 변조
{"product": "iPhone SE"}
```

### 3. 공격 로그 시스템 ✨
- 실시간 변조 로그 출력
- 변조 전/후 비교 표시
- **WebSocket을 통한 Frontend 전송 (완료)**
  - 엔드포인트: `ws://localhost:8090/ws/logs`
  - 실시간 로그 스트리밍
  - HTML 테스트 클라이언트 포함

## 프로젝트 구조

```
sage-gateway-infected-for-demo/
├── main.go                  # 메인 서버
├── config/
│   └── config.go           # 설정 관리
├── handlers/
│   ├── proxy.go            # 프록시 핸들러
│   ├── interceptor.go      # 메시지 가로채기
│   └── modifier.go         # 메시지 변조
├── attacks/
│   ├── price.go            # 금액 변조
│   ├── address.go          # 주소 변조
│   └── product.go          # 상품 변조
├── logger/
│   └── logger.go           # 로그 시스템
├── types/
│   └── message.go          # 메시지 타입
└── README.md
```

## 설치 및 실행

### 빠른 시작 (Makefile 사용 - 권장)

#### 1. 환경 설정
```bash
make setup
```

이 명령은:
- 의존성 다운로드
- `.env.example`을 `.env`로 복사
- 개발 환경 초기화

#### 2. 빌드
```bash
make build
```

#### 3. 실행
```bash
make run
```

#### 4. 정리
```bash
make clean
```

### 상세 가이드

**모든 명령어 보기:**
```bash
make help
```

**빌드 상태 확인:**
```bash
make status
```

**테스트 실행:**
```bash
make test              # 단위 테스트
make test-gateway      # Gateway 통합 테스트
make test-websocket    # WebSocket 테스트
make test-attack       # 공격 시나리오 테스트
```

### 수동 설정 (고급)

#### 1. 의존성 설치
```bash
go mod download
```

#### 2. 환경 변수 설정
```bash
# .env 파일 생성
cp .env.example .env

# 또는 직접 설정
export GATEWAY_PORT=5500
export ATTACK_ENABLED=true
export ATTACK_TYPE=price_manipulation
export AGENT_URLS='{"root":"http://localhost:18080","payment":"http://localhost:19083"}'
export LOG_LEVEL=info
```

#### 3. 빌드 및 실행
```bash
# 빌드
go build -o gateway-infected .

# 실행
./gateway-infected
```

### 연결 포인트

게이트웨이가 실행되면:

1. **Root Agent 연결:** `http://localhost:5500`
   - Agent 간 통신 프록시 엔드포인트
   - 메시지 변조 및 전달

2. **Frontend 모니터링:** `ws://localhost:5500/ws/logs`
   - 실시간 로그 스트리밍
   - WebSocket 연결

**상세 가이드:** [BUILD_AND_RUN.md](./BUILD_AND_RUN.md) 참조

## 사용 예시

### 정상 요청 (공격 비활성화)
```bash
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "product": "Sunglasses",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }'
```

### 공격 활성화
```bash
export ATTACK_ENABLED=true
export ATTACK_TYPE=price_manipulation

# 동일한 요청 → 금액이 100배 증가
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100,
    "product": "Sunglasses",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }'

# 출력 (로그):
# [ATTACK] Original amount: 100 → Modified: 10000
# [ATTACK] Original recipient: 0x742d35... → Modified: 0xATTACKER...
```

## API 엔드포인트

### WebSocket: /ws/logs ✨ NEW
실시간 로그 스트리밍을 제공합니다.

**연결 방법**:
```javascript
const ws = new WebSocket('ws://localhost:8090/ws/logs');

ws.onmessage = (event) => {
  const logEvent = JSON.parse(event.data);
  console.log(logEvent);
};
```

**로그 이벤트 포맷**:
```json
{
  "type": "attack",
  "timestamp": "2025-11-04T12:34:56.789Z",
  "level": "warn",
  "message": "Attack detected: price_manipulation",
  "data": {
    "attack_type": "price_manipulation",
    "original_msg": {"amount": 100},
    "modified_msg": {"amount": 10000},
    "changes": [...]
  }
}
```

**HTML 테스트 클라이언트**:
```bash
# 브라우저에서 열기
open test_websocket.html
```

### POST /payment
결제 요청을 프록시하고 변조합니다.

**요청:**
```json
{
  "amount": 100,
  "currency": "USD",
  "product": "Sunglasses",
  "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
}
```

**응답 (ATTACK_ENABLED=true):**
```json
{
  "amount": 10000,
  "currency": "USD",
  "product": "Sunglasses",
  "recipient": "0xATTACKER_WALLET_ADDRESS"
}
```

### POST /order
주문 요청을 프록시하고 변조합니다.

### GET /health
서버 상태 확인

## 환경 변수

| 변수 | 설명 | 기본값 | 예시 |
|-----|------|-------|------|
| `GATEWAY_PORT` | Gateway 서버 포트 | `8090` | `8090` |
| `ATTACK_ENABLED` | 공격 활성화 여부 | `true` | `true`, `false` |
| `ATTACK_TYPE` | 공격 유형 | `price_manipulation` | `price_manipulation`, `address_manipulation`, `product_substitution` |
| `TARGET_AGENT_URL` | 타겟 Agent URL | `http://localhost:8091` | `http://localhost:8091` |
| `LOG_LEVEL` | 로그 레벨 | `info` | `debug`, `info`, `warn`, `error` |
| `ATTACKER_WALLET` | 공격자 지갑 주소 | `0xATTACKER...` | `0x...` |

## 테스트

### 단위 테스트
```bash
go test ./...
```

### 통합 테스트
```bash
# Terminal 1: Gateway 실행
export ATTACK_ENABLED=true
go run main.go

# Terminal 2: Payment Agent 실행
cd ../sage-payment-agent
go run main.go

# Terminal 3: 테스트 요청
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{"amount": 100}'
```

## 로그 예시

### ATTACK_ENABLED=true
```
[INFO] 2025-01-27 10:30:15 Gateway server starting on port 8090
[INFO] 2025-01-27 10:30:15 Attack mode: ENABLED (price_manipulation)
[INFO] 2025-01-27 10:30:15 Target agent: http://localhost:8091

[INFO] 2025-01-27 10:30:20 Incoming request: POST /payment
[DEBUG] 2025-01-27 10:30:20 Original message: {"amount":100,"recipient":"0x742d35..."}

[ATTACK] 2025-01-27 10:30:20 ===== ATTACK DETECTED =====
[ATTACK] 2025-01-27 10:30:20 Type: price_manipulation
[ATTACK] 2025-01-27 10:30:20 Original amount: 100
[ATTACK] 2025-01-27 10:30:20 Modified amount: 10000 (100x)
[ATTACK] 2025-01-27 10:30:20 Original recipient: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
[ATTACK] 2025-01-27 10:30:20 Modified recipient: 0xATTACKER_WALLET_ADDRESS
[ATTACK] 2025-01-27 10:30:20 ===========================

[INFO] 2025-01-27 10:30:20 Forwarding modified message to http://localhost:8091/payment
[INFO] 2025-01-27 10:30:21 Response from target agent: 200 OK
```

### ATTACK_ENABLED=false
```
[INFO] 2025-01-27 10:35:15 Gateway server starting on port 8090
[INFO] 2025-01-27 10:35:15 Attack mode: DISABLED
[INFO] 2025-01-27 10:35:15 Transparent proxy mode

[INFO] 2025-01-27 10:35:20 Incoming request: POST /payment
[INFO] 2025-01-27 10:35:20 Forwarding original message (no modification)
[INFO] 2025-01-27 10:35:21 Response from target agent: 200 OK
```

## 보안 경고

⚠️ **이 서버는 교육 및 데모 목적으로만 사용되어야 합니다.**

- 실제 프로덕션 환경에서 사용 금지
- 악의적인 목적으로 사용 금지
- 권한 없는 시스템에 대한 테스트 금지

이 코드는 SAGE 프로토콜의 필요성을 입증하기 위한 시연 도구입니다.

## 데모 시나리오

### 시나리오 1: SAGE OFF (공격 성공)
1. Gateway 서버 시작 (ATTACK_ENABLED=true)
2. Payment Agent 시작 (SAGE_ENABLED=false)
3. Frontend에서 "$100 결제" 요청
4. Gateway가 금액을 $10,000로 변조
5. Payment Agent가 변조된 금액으로 처리
6. **결과**: 공격자가 100배 많은 금액을 탈취

### 시나리오 2: SAGE ON (공격 차단)
1. Gateway 서버 시작 (ATTACK_ENABLED=true)
2. Payment Agent 시작 (SAGE_ENABLED=true)
3. Frontend에서 "$100 결제" 요청 (RFC-9421 서명 포함)
4. Gateway가 금액을 $10,000로 변조
5. Payment Agent가 서명 검증 → **실패** (메시지 변조 탐지)
6. Payment Agent가 변조된 요청 거부
7. **결과**: 공격 차단, 사용자 보호

## 기술 스택

- **Go 1.21+**: 메인 언어
- **net/http**: HTTP 프록시 서버
- **encoding/json**: JSON 메시지 파싱
- **log**: 로그 시스템

## 개발 로드맵

- [x] 기본 프록시 서버
- [x] 메시지 가로채기
- [x] 금액 변조 (price_manipulation)
- [x] 주소 변조 (address_manipulation)
- [x] 상품 변조 (product_substitution)
- [x] WebSocket 로그 전송 ✨ **NEW**
- [ ] A2A 프로토콜 인식 (RFC 9421, HPKE)
- [ ] 대시보드 통합

## 라이선스

MIT License - 교육 및 데모 목적으로만 사용

## 기여

이 프로젝트는 SAGE 시연 목적이므로 Pull Request는 받지 않습니다.
문의 사항은 GitHub Issues를 통해 주세요.

## 관련 프로젝트

- [sage](../sage) - SAGE 핵심 라이브러리
- [sage-multi-agent](../sage-multi-agent) - 멀티 에이전트 시스템
- [sage-payment-agent](../sage-payment-agent) - 결제 Agent (AP2 통합)
- [sage-fe](../sage-fe) - Frontend 데모

---

**작성일**: 2025-01-27
**목적**: SAGE 프로토콜 보안 효과 시연
**상태**: 개발 중
