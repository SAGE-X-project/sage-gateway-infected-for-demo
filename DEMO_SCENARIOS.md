# SAGE Gateway - Demo Scenarios

**목적**: SAGE 프로토콜의 보안 효과를 시연하기 위한 데모 시나리오 가이드

---

## 📋 준비 사항

### 1. Gateway 빌드 및 실행

```bash
# 1. 환경 설정
make setup

# 2. 빌드
make build

# 3. 테스트 (선택사항)
make test

# 4. 실행
make run
```

**확인사항**:
- ✅ Gateway가 포트 8090에서 실행 중
- ✅ WebSocket이 `ws://localhost:8090/ws/logs`에서 대기 중

---

## 🎬 시나리오 1: 기본 공격 시연 (SAGE OFF)

### 목적
보안 프로토콜이 없을 때 MITM 공격이 성공하는 것을 시연

### 설정
```bash
# .env 파일 설정
ATTACK_ENABLED=true
ATTACK_TYPE=price_manipulation
GATEWAY_PORT=8090
```

### 단계

#### 1. Gateway 재시작
```bash
pkill -f gateway-infected
make run
```

#### 2. Mock Target Agent 시작
```bash
# Terminal 2
cd test
python3 -m http.server 9999 &
```

#### 3. 테스트 요청 전송
```bash
# Terminal 3
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{
    "id": "msg-001",
    "from": "root",
    "to": "payment",
    "metadata": {
      "amount": 100,
      "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    }
  }'
```

#### 4. 예상 결과
```
Gateway 로그:
[INFO] Protocol detection: SAGE: ❌ OFF, HPKE: ❌ OFF
[INFO] 📝 No HPKE - applying JSON modification attack
[ATTACK] Field: metadata.amount (100 → 10000)
[ATTACK] Field: metadata.recipient (0x742d... → 0xATTACKER)
```

**결론**: ✅ 공격 성공 - SAGE 없이는 메시지 변조가 감지되지 않음

---

## 🎬 시나리오 2: SAGE 서명으로 공격 차단

### 목적
SAGE 서명이 메시지 변조를 감지하는 것을 시연

### 단계

#### 1. SAGE 서명이 포함된 요청
```bash
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -H "Signature: sig1=:ABC123DEF456:" \
  -H 'Signature-Input: sig1=("@method" "@path");created=1234567890;keyid="test-key"' \
  -d '{
    "id": "msg-002",
    "from": "root",
    "to": "payment",
    "metadata": {
      "amount": 100,
      "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    }
  }'
```

#### 2. 예상 결과
```
Gateway 로그:
[INFO] Protocol detection: SAGE: ✅ ON, HPKE: ❌ OFF
[INFO] ✅ RFC 9421 Signature detected (ID: sig1)
[INFO] 📝 No HPKE - applying JSON modification attack
[WARN] ⚠️ SAGE signature detected - JSON modification will invalidate signature
[ATTACK] Field: metadata.amount (100 → 10000)

Target Agent:
❌ 401 Unauthorized - Signature verification failed
```

**결론**: ✅ 공격 차단 - SAGE 서명이 변조를 감지하여 요청 거부

---

## 🎬 시나리오 3: HPKE 암호화로 공격 차단

### 목적
HPKE 암호화가 비트 플립 공격도 차단하는 것을 시연

### 단계

#### 1. HPKE 암호화된 메시지
```bash
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -d '{
    "encryptedPayload": "dGhpcyBpcyBlbmNyeXB0ZWQgZGF0YQ==",
    "type": "secure"
  }'
```

#### 2. 예상 결과
```
Gateway 로그:
[INFO] Protocol detection: SAGE: ❌ OFF, HPKE: ✅ ON
[INFO] ✅ HPKE encrypted payload detected
[INFO] 🔐 HPKE detected - applying encrypted payload bit-flip attack
[INFO] 🔥 Bit-flip attack on encryptedPayload field
[ATTACK] Field: encryptedPayload (modified)
[WARN] ⚠️ Target agent will FAIL to decrypt this message

Target Agent:
❌ 400 Bad Request - HPKE decryption failed (integrity check)
```

**결론**: ✅ 공격 차단 - HPKE 무결성 검증이 비트 플립 공격 감지

---

## 🎬 시나리오 4: 완전한 보안 (SAGE + HPKE)

### 목적
SAGE 서명과 HPKE 암호화를 모두 사용한 다층 보안 시연

### 단계

#### 1. 완전히 보호된 메시지
```bash
curl -X POST http://localhost:8090/payment \
  -H "Content-Type: application/json" \
  -H "Signature: sig1=:SECURE:" \
  -H 'Signature-Input: sig1=("@method" "@path");created=1234567890;keyid="ecdsa-key"' \
  -d '{
    "encryptedPayload": "dGhpcyBpcyBlbmNyeXB0ZWQgZGF0YQ==",
    "type": "secure"
  }'
```

#### 2. 예상 결과
```
Gateway 로그:
[INFO] Protocol detection: SAGE: ✅ ON, HPKE: ✅ ON
[INFO] ✅ RFC 9421 Signature detected (ID: sig1)
[INFO] ✅ HPKE encrypted payload detected
[INFO] 🔐 HPKE detected - applying encrypted payload bit-flip attack
[WARN] ⚠️ Target agent will REJECT this request due to:
[WARN]    - Signature verification failure
[WARN]    - HPKE decryption failure

Target Agent:
❌ 401 Unauthorized - Multiple security failures
```

**결론**: ✅ 최고 수준 보안 - 다층 보안으로 모든 공격 차단

---

## 🎬 시나리오 5: 실시간 모니터링 (WebSocket)

### 목적
Frontend에서 실시간으로 공격 모니터링

### 단계

#### 1. WebSocket 테스트 클라이언트 열기
```bash
open test_websocket.html
```

또는 커맨드라인:
```bash
./test_websocket_client.sh
```

#### 2. 여러 공격 시나리오 실행
위의 시나리오 1-4를 순차적으로 실행

#### 3. 실시간 로그 확인
HTML 클라이언트에서 다음을 실시간으로 확인:
- 🔵 **INFO**: 프로토콜 감지 상태
- 🟡 **WARN**: 보안 경고
- 🔴 **ATTACK**: 공격 감지 및 상세 정보
- ❌ **ERROR**: 에러 발생

**화면 예시**:
```
[INFO] 03:15:19  Protocol detection: SAGE: ✅ ON, HPKE: ✅ ON
[ATTACK] 03:15:19  Attack detected: encrypted payload bit-flip
  Changes:
    - encryptedPayload: <modified>
[WARN] 03:15:19  ⚠️ Both SAGE and HPKE will reject this request
```

---

## 🎬 시나리오 6: sage-multi-agent 통합 (고급)

### 목적
실제 sage-multi-agent 시스템과 통합하여 end-to-end 데모

### 준비

#### 1. Gateway 설정
```bash
# .env 파일
GATEWAY_PORT=5500
ATTACK_ENABLED=true
ATTACK_TYPE=price_manipulation
AGENT_URLS={"root":"http://localhost:18080","payment":"http://localhost:19083","medical":"http://localhost:19082"}
```

#### 2. sage-multi-agent 빌드
```bash
cd ../sage-multi-agent
make build
```

#### 3. Agent 실행
```bash
# Terminal 1: Gateway
cd ../sage-gateway-infected-for-demo
make run

# Terminal 2: Payment Agent
cd ../sage-multi-agent
./build/bin/payment

# Terminal 3: Medical Agent
./build/bin/medical

# Terminal 4: Root Agent
./build/bin/root --port=18080

# Terminal 5: Client API
./build/bin/client
```

#### 4. 테스트 요청
```bash
# Terminal 6
curl -X POST http://localhost:8086/api/request \
  -H "Content-Type: application/json" \
  -H "X-SAGE-Enabled: false" \
  -H "X-HPKE-Enabled: false" \
  -d '{"prompt":"send 100 KRW to merchant for payment"}'
```

#### 5. 로그 모니터링
```bash
# Gateway 로그
tail -f logs/gateway.log

# Payment Agent 로그
tail -f ../sage-multi-agent/logs/payment.log
```

---

## 📊 자동화된 테스트 스크립트

### 모든 공격 시나리오 자동 테스트
```bash
./test_attack_scenarios.sh
```

이 스크립트는 다음을 자동으로 테스트:
1. ✅ No security + attack
2. ✅ SAGE only + attack
3. ✅ HPKE only + attack
4. ✅ SAGE + HPKE + attack
5. ✅ Alternative encrypted fields
6. ✅ Multiple attack types

### A2A 프로토콜 감지 테스트
```bash
./test_a2a_detection.sh
```

---

## 🎯 데모 핵심 메시지

### 1. SAGE 서명의 필요성
- ❌ **SAGE OFF**: 메시지 변조 성공 (100 KRW → 10000 KRW)
- ✅ **SAGE ON**: 서명 검증 실패로 공격 차단

### 2. HPKE 암호화의 필요성
- ❌ **HPKE OFF**: 메시지 내용 노출 및 변조 가능
- ✅ **HPKE ON**: 암호화 + 무결성 검증으로 공격 차단

### 3. 다층 보안의 중요성
- 🛡️ **SAGE + HPKE**: 두 가지 독립적인 보안 계층
- 🔒 **Defense in Depth**: 하나가 뚫려도 다른 계층이 방어

### 4. 실시간 모니터링
- 👁️ WebSocket으로 모든 공격 실시간 가시화
- 📊 공격 유형, 변조 내용, 보안 상태 즉시 확인

---

## 🔧 문제 해결

### Gateway가 시작되지 않음
```bash
# 포트 사용 중 확인
lsof -i :8090

# 프로세스 종료
pkill -f gateway-infected

# 재시작
make run
```

### 테스트 실패
```bash
# 전체 클린 빌드
make clean
make build
make test
```

### WebSocket 연결 실패
```bash
# Gateway가 실행 중인지 확인
ps aux | grep gateway-infected

# 로그 확인
tail -f gateway.log
```

---

## 📚 추가 리소스

- **README.md** - 프로젝트 개요 및 기능 설명
- **BUILD_AND_RUN.md** - 상세 빌드 및 실행 가이드
- **MAKEFILE_IMPLEMENTATION.md** - Makefile 명령어 참조
- **todolist.md** - 프로젝트 진행 상황

---

**작성일**: 2025-11-04
**프로젝트**: sage-gateway-infected-for-demo
**버전**: 1.0.0
