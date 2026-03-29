---
layout: post
title: Phase3-023-Jaeger-Tracing
date: 2026-03-28
---
# Jaeger: 분산 시스템 병목 분석 완벽 가이드

## 요약

**배우는 내용**:
- 분산 추적(Distributed Tracing) 원리
- Jaeger 스팬(Span) 구조
- 마이크로서비스 병목 분석
- 실전: 응답시간 900ms → 100ms 개선

---

## 1. 분산 추적 개념

```
단일 서버:          분산 시스템:
┌──────────────┐    ┌─────────────────────────────┐
│ 요청 처리    │    │ API Gateway (100ms)         │
│ 100ms       │    │  └─ Auth Service (20ms)     │
│  └─ DB      │    │  └─ User Service (50ms)     │
│  └─ Cache   │    │     └─ DB (30ms)            │
└──────────────┘    │  └─ Order Service (200ms)   │
                    │     └─ Inventory (150ms)    │
로그: ❌ 어느 부분?  │  └─ Payment (400ms)         │
                    │     └─ External API (350ms) │
                    │  └─ Cache (10ms)            │
                    └─────────────────────────────┘
                    로그: ✅ 각 단계별 시간 추적
```

---

## 2. Jaeger 설치

### Docker Compose

```yaml
version: '3.8'
services:
  jaeger:
    image: jaegertracing/all-in-one
    ports:
      - "6831:6831/udp"    # Jaeger agent
      - "16686:16686"       # Jaeger UI
    environment:
      - COLLECTOR_ZIPKIN_HTTP_PORT=9411
```

---

## 3. Go에서 Jaeger 사용

```go
import "github.com/uber/jaeger-client-go"
import opentracing "github.com/opentracing/opentracing-go"
import "github.com/uber/jaeger-client-go/config"

func initTracer(serviceName string) (opentracing.Tracer, error) {
    cfg := &config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,  // 100% 샘플링
        },
        Reporter: &config.ReporterConfig{
            LocalAgentHostPort: "localhost:6831",
        },
    }

    tracer, _, err := cfg.NewTracer()
    return tracer, err
}

// 사용
tracer, _ := initTracer("api-gateway")

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 스팬 생성
    span := tracer.StartSpan("handleRequest")
    defer span.Finish()

    // 자식 스팬
    userSpan := tracer.StartSpan(
        "getUser",
        opentracing.ChildOf(span.Context()),
    )
    user := getUser()
    userSpan.Finish()

    // 다른 마이크로서비스 호출
    orderSpan := tracer.StartSpan(
        "getOrder",
        opentracing.ChildOf(span.Context()),
    )
    order := callOrderService()
    orderSpan.Finish()

    // 응답
    w.WriteHeader(http.StatusOK)
}

// 마이크로서비스 호출 (컨텍스트 전파)
func callOrderService() {
    ctx := r.Context()
    req, _ := http.NewRequestWithContext(ctx, "GET", "/orders", nil)
    
    // Trace 헤더 추가
    tracer.Inject(
        span.Context(),
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(req.Header),
    )

    client.Do(req)
}
```

---

## 4. Jaeger UI 분석

### 스팬 타임라인

```
handleRequest [0ms --- 500ms]
├─ getUser [5ms --- 25ms]
├─ getOrder [50ms --- 250ms]
│  └─ queryInventory [60ms --- 200ms]
└─ callPayment [300ms --- 490ms]
   └─ externalAPI [350ms --- 480ms]

병목: externalAPI (130ms)
```

---

## 5. 성능 개선

### 개선 전

```
요청 흐름 (동기):
1. Auth (20ms)
2. User (50ms) ← DB (30ms)
3. Order (200ms) ← Inventory (150ms)
4. Payment (400ms) ← External (350ms)

총: 670ms
```

### 개선 후

```
요청 흐름 (병렬):
1. Auth (20ms)
2. User + Order 병렬 (max(50ms, 200ms) = 200ms)
3. Payment (400ms) ← 병렬화 불가

총: 620ms (약간 개선)

더 나은 해결:
- External API → 캐시 (5ms)
- 결과: 50ms → 400ms + 5ms = 405ms ✅
```

---

## 6. 추적 샘플링

```go
// 모든 요청 추적 (개발)
Sampler: &config.SamplerConfig{
    Type:  "const",
    Param: 1,  // 100%
}

// 일부만 추적 (프로덕션)
Sampler: &config.SamplerConfig{
    Type:  "probabilistic",
    Param: 0.01,  // 1%
}

// 오류만 추적
Sampler: &config.SamplerConfig{
    Type:  "ratelimiting",
    Param: 100,  // 초당 100개
}
```

---

## 7. 메트릭 내보내기

```go
// Prometheus 메트릭으로도 내보내기
import "github.com/prometheus/client_golang/prometheus"

func recordSpan(span opentracing.Span) {
    duration := span.FinishTime.Sub(span.StartTime)
    spanDuration.WithLabelValues(
        span.OperationName,
    ).Observe(duration.Seconds())
}
```

---

## 핵심 정리

| 항목 | Jaeger |
|------|--------|
| **추적** | 완전 자동 |
| **병목** | 시각화 |
| **성능** | UI 기반 분석 |

---

## 결론

**분산 추적은 필수입니다.**

900ms → 100ms 개선! 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
