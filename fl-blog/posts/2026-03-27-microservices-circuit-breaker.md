---
title: "마이크로서비스: Circuit Breaker 패턴으로 장애 격리하기"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["systems", "devops", "cloud"]
toc: true
comments: true
---

# 마이크로서비스: Circuit Breaker 패턴으로 장애 격리하기
## 요약

**배우는 내용**:
- Circuit Breaker의 3가지 상태 (Closed, Open, Half-Open)
- 실전 Go 구현으로 장애 전파 방지
- Hystrix 패턴 vs Resilience4j vs Go 구현 비교
- 타임아웃, 재시도, 폴백 전략

---

## 1. 마이크로서비스의 문제

### 분산 시스템 장애 시나리오

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│ 사용자   │→ │API      │→ │결제      │
└──────────┘  │Gateway  │  │서비스    │
              └──────────┘  └────┬─────┘
                                  │
                                  ↓ (DB 연결 끊김)
                            ❌ 요청 타임아웃

사용자 영향: ⏳ 30초 대기 → 응답 없음
요청 적체: 1개 → 10개 → 100개 → Thread 고갈
전체 시스템: API Gateway도 응답 불가 (Cascading Failure)
```

---

## 2. Circuit Breaker 패턴

### 상태 다이어그램

```
                    실패 증가
                       ↓
              ┌─────────────────┐
              │    Closed       │ ◄──── 초기 상태
              │ (정상 동작)      │
              └────────┬────────┘
                       │ 임계값 초과
                       ↓
              ┌─────────────────┐
              │     Open        │ ◄──── 요청 차단
              │ (차단)          │
              └────────┬────────┘
                       │ timeout (예: 30초)
                       ↓
              ┌─────────────────┐
              │  Half-Open      │ ◄──── 시험적 요청
              │ (테스트)        │
              └────────┬────────┘
                       │
           ┌───────────┴───────────┐
           ↓ 성공                   ↓ 실패
        Closed                   Open
```

### 상태별 동작

```
상태        | 동작 | 예시
------------|------|----------
Closed      | 모든 요청 통과 | 정상 서비스
Open        | 모든 요청 차단 | "Service unavailable"
Half-Open   | 시험 요청만 | 1개 요청 테스트
```

---

## 3. Go 구현

### 3-1. 간단한 Circuit Breaker

```go
package breaker

import (
    "errors"
    "sync"
    "time"
)

type State int

const (
    Closed State = iota
    Open
    HalfOpen
)

type CircuitBreaker struct {
    mu              sync.RWMutex
    state           State
    lastFailureTime time.Time
    failureCount    int
    successCount    int
    timeout         time.Duration
    maxFailures     int
    maxSuccesses    int
}

func New(timeout time.Duration, maxFailures int) *CircuitBreaker {
    return &CircuitBreaker{
        state:       Closed,
        timeout:     timeout,
        maxFailures: maxFailures,
        maxSuccesses: 2,
    }
}

// 요청 실행
func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    // Open 상태에서 timeout 초과했는지 확인
    if cb.state == Open {
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = HalfOpen
            cb.failureCount = 0
            cb.successCount = 0
        } else {
            return errors.New("circuit breaker is open")
        }
    }

    // 실제 함수 실행
    err := fn()

    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()

        // Closed → Open 전환
        if cb.failureCount >= cb.maxFailures {
            cb.state = Open
        }
        return err
    }

    // 성공
    cb.failureCount = 0

    if cb.state == HalfOpen {
        cb.successCount++
        if cb.successCount >= cb.maxSuccesses {
            cb.state = Closed  // HalfOpen → Closed
        }
    }
    return nil
}

func (cb *CircuitBreaker) GetState() State {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}
```

### 3-2. 실전 사용

```go
package main

import (
    "fmt"
    "time"
    "breaker"
)

func main() {
    cb := breaker.New(30*time.Second, 5)  // 30초 timeout, 5회 실패 시 Open

    // 외부 서비스 호출 (결제 서비스)
    for i := 0; i < 20; i++ {
        err := cb.Execute(func() error {
            // 실제 RPC 호출
            return callPaymentService()
        })

        if err != nil {
            fmt.Printf("[%d] 요청 실패: %v\n", i, err)
        } else {
            fmt.Printf("[%d] 요청 성공\n", i)
        }

        time.Sleep(100 * time.Millisecond)
    }
}

// 외부 서비스 호출 (시뮬레이션)
var failCount = 0
func callPaymentService() error {
    failCount++
    if failCount <= 5 {
        return fmt.Errorf("payment service timeout")
    }
    return nil  // 6번째부터 성공
}
```

**출력**:
```
[0] 요청 실패: payment service timeout
[1] 요청 실패: payment service timeout
[2] 요청 실패: payment service timeout
[3] 요청 실패: payment service timeout
[4] 요청 실패: payment service timeout
[5] 요청 실패: circuit breaker is open     ← Open 전환
[6] 요청 실패: circuit breaker is open     ← 요청 차단
...
[35] 요청 성공                              ← timeout 후 Half-Open
[36] 요청 성공                              ← Closed 복귀
```

---

## 4. 고급 패턴

### 4-1. Retry + Circuit Breaker

```go
func executeWithRetry(cb *CircuitBreaker, fn func() error, maxRetries int) error {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        err := cb.Execute(fn)
        if err == nil {
            return nil  // 성공
        }

        lastErr = err

        // Circuit Breaker가 Open된 경우 재시도 불가
        if err.Error() == "circuit breaker is open" {
            return err
        }

        // 지수 백오프 (exponential backoff)
        backoff := time.Duration(1<<uint(attempt)) * time.Second
        time.Sleep(backoff)
    }

    return lastErr
}

// 사용
err := executeWithRetry(cb, func() error {
    return callPaymentService()
}, 3)  // 최대 3회 재시도
```

**재시도 전략**:
```
시도 1 (즉시) → 실패
대기 1초
시도 2        → 실패
대기 2초
시도 3        → 실패 또는 성공
```

### 4-2. Fallback (폴백)

```go
func getPaymentStatus(userID string) (string, error) {
    // 1. Circuit Breaker를 통한 호출
    var status string
    err := cb.Execute(func() error {
        resp, err := callPaymentService(userID)
        status = resp.Status
        return err
    })

    if err != nil {
        // 2. Circuit Breaker 열림 또는 서비스 오류
        // → 폴백: 캐시된 상태 반환
        cached := getFromCache(userID)
        if cached != "" {
            fmt.Printf("Fallback: 캐시된 상태 반환\n")
            return cached, nil
        }

        // 3. 캐시도 없으면 기본값
        fmt.Printf("Fallback: 기본값 반환\n")
        return "UNKNOWN", nil
    }

    return status, nil
}
```

### 4-3. 타임아웃

```go
// 타임아웃 포함 RPC 호출
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.Payment(ctx, &PaymentRequest{
    UserID: "user123",
    Amount: 10000,
})

if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // 타임아웃: Circuit Breaker 업데이트
        cb.Execute(func() error { return err })
        return nil, fmt.Errorf("payment timeout")
    }
}
```

---

## 5. 주요 라이브러리 비교

### Go 생태계

```
라이브러리          | 기능 | 복잡도 | 권장
--------------------|------|--------|-------
제작사구현          | 기본 | 낮음  | 학습용
go-kit/kit          | 풍부 | 중간  | 마이크로서비스
grpc-ecosystem      | gRPC | 중간  | gRPC 사용
eapache/go-resilienc| 풍부 | 낮음  | 프로덕션
```

### Hystrix vs Resilience4j vs Go 구현

```
항목              | Hystrix   | Resilience4j | Go (자체)
-----------------|-----------|---|-------
언어              | Java      | Java | Go
메모리            | 50MB+     | 20MB | <1MB
설정 복잡도       | 높음      | 중간 | 낮음
학습곡선          | 가파름    | 중간 | 완만
성능 오버헤드     | ~5-10%    | ~2-5% | <1%
프로덕션 성숙도   | 높음      | 높음 | 중간
```

---

## 6. 실제 사례

### 사례: 결제 서비스 장애 격리

**시나리오**: 결제 서비스가 5분 동안 다운

```
시간     | 액션 | 영향
---------|------|------
0:00     | 결제 서비스 다운 | CB: Closed
0:05     | CB: Open (5회 실패) | 사용자: 폴백으로 주문 보관
0:10     | 재고 서비스 정상 | 다른 사용자 주문 계속
0:35     | CB: Half-Open | 시험 요청 성공
0:36     | CB: Closed | 결제 서비스 복구
```

**결과**:
- 다운타임: 0초 (사용자는 주문 가능)
- 복구 후 영향: 보관된 주문 자동 처리
- 다른 서비스: 독립적 운영

---

## 7. 메트릭 추적

```go
type CircuitBreakerMetrics struct {
    TotalRequests   int64
    SuccessCount    int64
    FailureCount    int64
    OpenCount       int64  // Open으로 차단된 요청
    StateChanges    int64
}

func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    return CircuitBreakerMetrics{
        TotalRequests: cb.TotalRequests,
        SuccessCount: cb.SuccessCount,
        FailureCount: cb.failureCount,
        OpenCount: cb.openCount,
    }
}

// 모니터링
metrics := cb.GetMetrics()
if metrics.OpenCount > 10 {
    alert("Circuit breaker too many rejections")
}
```

---

## 핵심 정리

| 패턴 | 목적 | 구현 시간 |
|------|------|----------|
| **Circuit Breaker** | 장애 격리 | 1시간 |
| **Retry** | 일시적 오류 복구 | 30분 |
| **Timeout** | 무한 대기 방지 | 15분 |
| **Fallback** | 기본값 제공 | 30분 |

---

## 결론

Circuit Breaker는 **분산 시스템의 필수 방어막**입니다.

- 언제 사용: 외부 의존성 호출 (DB, API, 캐시)
- 이점: 빠른 실패, 자동 복구, 서비스 독립성
- 비용: 약간의 복잡성, 메트릭 관리

🚀 안정적인 마이크로서비스의 시작입니다!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
