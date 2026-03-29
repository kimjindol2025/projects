---
title: "스트림 처리(Stream Processing): 실시간 데이터 파이프라인의 핵심"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# 스트림 처리(Stream Processing): 실시간 데이터 파이프라인의 핵심
## 소개

배치 처리는 대량의 데이터를 모아서 한꺼번에 처리합니다. 반면 **스트림 처리**는 데이터가 들어오는 즉시 처리하여 실시간 인사이트를 제공합니다.

```
배치: [1시간 데이터 모음] → 분석 → 결과 (1시간 지연)
스트림: 데이터 → 즉시 처리 → 실시간 결과 (ms 단위 지연)
```

## 스트림 처리의 필요성

### 실제 사례
1. **주식 거래**: 가격 변동 감지 (ms 단위)
2. **사기 탐지**: 거래 패턴 즉시 분석
3. **IoT 센서**: 온도, 습도 실시간 모니터링
4. **추천 시스템**: 사용자 행동에 즉시 반응

## 핵심 개념

### 1. Event Time vs Processing Time

```
이벤트: [11:00] [11:05] [11:10] [11:15]  (실제 발생 시간)
          ↓        ↓        ↓        ↓
전송:   [11:02] [11:08] [11:12] [11:20]  (처리 시간)

문제: 네트워크 지연, 버퍼링으로 인한 순서 변동
```

### 2. Watermark (워터마크)

```
[event @ 11:00] [event @ 11:05] [event @ 11:10]
              ↓
         watermark @ 11:07
         (11:05까지의 모든 데이터가 도착했음을 보장)
         → 윈도우 닫기
```

### 3. Windowing (윈도우)

```
Tumbling Window (겹치지 않음):
[1-5분] [6-10분] [11-15분]

Sliding Window (겹침):
[1-10분] [6-15분] [11-20분]  (5분 슬라이딩)

Session Window (이벤트 기반):
[활동] --- [5분 휴지] --- [활동]
 └─ Window 종료
```

## Apache Kafka + Flink 예제

### 환경 설정
```bash
# Kafka 토픽 생성
kafka-topics.sh --create --topic user-events \
  --partitions 3 --replication-factor 1

# Flink 설정
flink/bin/start-cluster.sh
```

### Flink 스트림 처리 파이프라인
```java
StreamExecutionEnvironment env = 
  StreamExecutionEnvironment.getExecutionEnvironment();

// Kafka에서 읽기
FlinkKafkaConsumer<String> kafkaConsumer = 
  new FlinkKafkaConsumer<>(
    "user-events",
    new SimpleStringSchema(),
    properties
  );

DataStream<UserEvent> events = env
  .addSource(kafkaConsumer)
  .map(json -> parseJSON(json))
  .filter(event -> event.value > 100);  // 필터링

// 5분 윈도우로 집계
DataStream<EventStats> stats = events
  .keyBy("userId")
  .window(TumblingEventTimeWindow.of(Time.minutes(5)))
  .aggregate(new AggregateFunction<...>() {
    public EventStats createAccumulator() { ... }
    public EventStats add(UserEvent e, EventStats acc) { ... }
    public EventStats getResult(EventStats acc) { ... }
    public EventStats merge(EventStats a, EventStats b) { ... }
  });

// 결과를 Elasticsearch로 저장
stats.addSink(new ElasticsearchSink<>(...));

env.execute("User Event Processing");
```

## At-Least-Once vs Exactly-Once

### At-Least-Once (최소 1회)
```
[메시지 처리] → 저장 → ACK 송신
          ↓ (네트워크 오류)
          재처리 (중복 가능)

영향: 같은 메시지가 2번 처리될 수 있음
장점: 빠르고 간단
```

### Exactly-Once (정확히 1회)
```
[메시지 처리] → [트랜잭션] ← ACK 수신 대기
                ↓ (커밋 또는 롤백)
         저장 & ACK 송신

영향: 정확한 처리 보장
단점: 더 복잡하고 느림
```

## 상태 관리 (State Management)

### Keyed State
```
각 key마다 독립적인 상태 유지:
userId_1: sum=1000, count=10, avg=100
userId_2: sum=2000, count=20, avg=100
```

### Operator State
```
전체 operator 차원의 상태:
파티션_1: [버퍼], 상태 저장소
파티션_2: [버퍼], 상태 저장소
```

### Fault Tolerance
```
[체크포인트] @ T=100
상태 저장: {userId_1: {...}, userId_2: {...}}
         (분산 저장소에 백업)

[장애 발생] @ T=150
복구: T=100 이후 메시지 재처리
```

## 실전 패턴

### 1. 이상 탐지 (Anomaly Detection)
```
input: 센서 데이터 스트림
       [온도: 25°C] [습도: 60%] [압력: 1013hPa]

rule: 온도가 평균±2σ를 벗어나면 알림

output: [⚠️ 온도 이상: 45°C 감지!]
```

### 2. 사용자 여정 추적
```
events: [로그인] → [페이지_A] → [상품_검색] → [구매]
                     ↓ (2분 후)
        윈도우 분석: 전환율 80%

결과: 사용자 행동 분석, 개인화 추천
```

### 3. 실시간 대시보드
```
메트릭 계산:
- 분당 이벤트 수
- 에러율
- 평균 응답시간

Kafka → Flink → Redis (메트릭 캐시) → 웹 대시보드
```

## 도구 비교

| 도구 | 강점 | 약점 | 용도 |
|------|------|------|------|
| **Kafka Streams** | 간단, 낮은 지연 | 복잡 로직 제한 | 간단한 필터링, 집계 |
| **Apache Flink** | 강력한 window, 정확성 | 높은 복잡도 | 복잡한 이벤트 처리 |
| **Spark Streaming** | Spark 생태계 통합 | 마이크로배치 지연 | 배치+스트림 혼합 |
| **ksqlDB** | SQL 문법 | 기능 제한 | SQL로 실시간 분석 |

## 성능 최적화

### 1. Parallelism (병렬화)
```
operator parallelism = 데이터 파티션 수
많을수록 처리량 증가, 지연시간 감소
```

### 2. Backpressure (역압)
```
upstream이 빠르면, downstream이 밀려서
자동으로 upstream 속도 조절
```

### 3. 상태 백엔드
```
In-Memory: 빠르지만 실패 시 손실
RocksDB: 느리지만 안정적
External Store (Redis): 중간 선택지
```

## 결론

스트림 처리는 **실시간 데이터 가치 추출**의 핵심입니다. Exactly-Once 보장, 상태 관리, 윈도우 처리 등 복잡한 개념이 있지만, 올바르게 구현하면 강력한 실시간 시스템을 만들 수 있습니다.

대규모 시스템에서는 Kafka (메시징) + Flink (처리) 조합이 업계 표준이 되었습니다.
