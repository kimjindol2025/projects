---
title: "모니터링: Prometheus/Grafana로 99.9% SLA 달성하기"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["systems", "devops", "cloud"]
toc: true
comments: true
---

# 모니터링: Prometheus/Grafana로 99.9% SLA 달성하기
## 요약

**배우는 내용**:
- Prometheus: 메트릭 수집 및 저장
- Grafana: 대시보드 시각화
- 알림 규칙 (SLO/SLI)
- 실전: 99.9% SLA 모니터링

---

## 1. Prometheus 아키텍처

```
┌─────────────────────────────────────┐
│ 애플리케이션 (Instrumentation)      │
│ - HTTP 요청 수                      │
│ - 응답 시간                         │
│ - 에러 수                          │
└──────────────┬──────────────────────┘
               │ Scrape (주기적)
┌──────────────▼──────────────────────┐
│ Prometheus Server                   │
│ - 메트릭 저장 (TSDB)                │
│ - 규칙 평가                        │
│ - 알림 생성                        │
└──────────────┬──────────────────────┘
               │
        ┌──────┴──────┐
        ▼             ▼
┌──────────────┐  ┌──────────────┐
│ Alertmanager │  │ Grafana      │
│ (알림 발송)  │  │ (시각화)     │
└──────────────┘  └──────────────┘
```

---

## 2. 메트릭 수집

### Go 앱에서 Prometheus 메트릭

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

// 메트릭 정의
var (
    // Counter: 누적 수
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    // Gauge: 현재값
    goroutinesCount = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "goroutines_count",
            Help: "Number of goroutines",
        },
    )

    // Histogram: 분포
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: []float64{.001, .01, .1, 1, 10},
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(goroutinesCount)
    prometheus.MustRegister(httpDuration)
}

func main() {
    // HTTP 핸들러
    http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // 비즈니스 로직
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))

        // 메트릭 기록
        httpRequestsTotal.WithLabelValues(
            r.Method,
            "/api/users",
            "200",
        ).Inc()

        httpDuration.WithLabelValues(
            r.Method,
            "/api/users",
        ).Observe(time.Since(start).Seconds())
    })

    // Prometheus 엔드포인트
    http.Handle("/metrics", promhttp.Handler())

    http.ListenAndServe(":8080", nil)
}
```

### Prometheus 설정

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'api-server'
    static_configs:
      - targets: ['localhost:8080']

  - job_name: 'database'
    static_configs:
      - targets: ['localhost:9100']  # Node Exporter

  - job_name: 'kubernetes'
    kubernetes_sd_configs:
      - role: pod
```

---

## 3. 쿼리 언어 (PromQL)

### 기본 쿼리

```promql
# 1. 간단한 메트릭
http_requests_total

# 2. 필터링
http_requests_total{status="200"}
http_requests_total{method="GET", status=~"2.."}

# 3. 범위 선택 (5분 동안)
rate(http_requests_total[5m])

# 4. 집계
sum(http_requests_total) by (status)

# 5. 연산
rate(http_requests_total[5m]) / rate(http_requests_total[5m])

# 6. 백분위수 (histogram)
histogram_quantile(0.95, http_request_duration_seconds)

# 7. 증가 감지
increase(http_errors_total[5m]) > 10

# 8. 좌표 조인
http_requests_total / ignoring(endpoint) http_total_duration_seconds
```

### 실전 쿼리

```promql
# 요청 속도 (초당)
rate(http_requests_total[1m])

# 에러율
rate(http_requests_total{status=~"5.."}[5m])
/ rate(http_requests_total[5m])

# P95 응답시간
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 메모리 사용량
process_resident_memory_bytes

# CPU 사용률
rate(process_cpu_seconds_total[1m])

# 동시 연결 수
increase(network_connections_total[1h])
```

---

## 4. 알림 규칙

### SLO/SLI 정의

```yaml
groups:
  - name: slo_rules
    rules:
      # SLI: 가용성 99.9%
      - alert: HighErrorRate
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total[5m]))
          ) > 0.001  # 0.1% (99.9% 반대)
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "에러율 {{ $value | humanizePercentage }} 초과"

      # SLI: 응답시간 P95 < 100ms
      - alert: SlowResponseTime
        expr: |
          histogram_quantile(
            0.95,
            rate(http_request_duration_seconds_bucket[5m])
          ) > 0.1
        for: 5m
        labels:
          severity: warning

      # SLI: 가용성
      - alert: LowAvailability
        expr: |
          sum(rate(http_requests_total{status="200"}[5m]))
          /
          sum(rate(http_requests_total[5m]))
          < 0.999
        for: 15m
        labels:
          severity: critical
```

---

## 5. Grafana 대시보드

### 대시보드 예시 (JSON)

```json
{
  "dashboard": {
    "title": "API Server - SLA 99.9%",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total[1m])"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Response Time P95",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "SLA Status",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{status=\"200\"}[5m])) / sum(rate(http_requests_total[5m]))"
          }
        ],
        "thresholds": [0.999]  # 99.9% 임계값
      }
    ]
  }
}
```

---

## 6. 성능 벤치마크

### 메트릭 저장소 크기

```
메트릭당 데이터:
- 타임스탬프: 8B
- 값: 8B
- 레이블: ~100B
- 총: ~120B

보존 기간:
- 1000 메트릭
- 1초마다 수집
- 15일 보존

계산:
1000 × 120B × 86,400초/일 × 15일 = 155GB
```

### 쿼리 성능

```
쿼리: rate(http_requests_total[5m])
데이터 포인트: 300개 (5분 × 60초)
실행 시간: 50ms
```

---

## 7. 99.9% SLA 달성

### 계산

```
99.9% SLA = 99.9% 가용성
= 999000ms / 1000000ms
= 월간 다운타임 허용: 43.2초 (30일 기준)

SLI 목표:
1. 에러율 < 0.1%
2. 응답시간 P95 < 100ms
3. 가용성 > 99.9%
```

### 모니터링 전략

```
실시간 (1분):
- 에러율 상승 감지
- 응답시간 증가 감지

단기 (1시간):
- 패턴 분석
- 임계값 조정

장기 (1달):
- SLA 준수율 계산
- 추세 분석
```

---

## 8. 베스트 프랙티스

### (1) 메트릭 네이밍

```
규칙: <namespace>_<subsystem>_<name>_<unit>

예시:
- http_requests_total (카운터, 단위: 개)
- http_request_duration_seconds (히스토그램, 단위: 초)
- process_resident_memory_bytes (게이지, 단위: 바이트)
```

### (2) 레이블 설계

```go
// ❌ 나쁜 예: 너무 많은 레이블
httpDuration.WithLabelValues(
    r.Method,
    r.RequestURI,  // 무한 카디널리티!
    r.RemoteAddr,  // 무한 카디널리티!
).Observe(dur)

// ✅ 좋은 예: 제한된 레이블
httpDuration.WithLabelValues(
    r.Method,
    "/api/users",  // 고정 레이블
    r.Header.Get("User-Agent"),  // 제한된 값
).Observe(dur)
```

### (3) 알림 최소화

```yaml
# ❌ 너무 많은 알림
- alert: AnyError
  expr: http_errors_total > 0

# ✅ 의미있는 알림
- alert: ErrorRateHigh
  expr: rate(http_errors_total[5m]) / rate(http_requests_total[5m]) > 0.01
  for: 5m
```

---

## 핵심 정리

| 항목 | Prometheus | Grafana |
|------|-----------|---------|
| **수집** | ✅ | ❌ |
| **저장** | ✅ | ❌ |
| **쿼리** | ✅ | ❌ |
| **시각화** | ❌ | ✅ |
| **알림** | ✅ | ✅ |

---

## 결론

Prometheus + Grafana는 **모니터링의 표준**입니다.

- 99.9% SLA 달성 가능
- 실시간 분석
- 자동 알림

🚀 메트릭으로 시스템을 통제하세요!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
