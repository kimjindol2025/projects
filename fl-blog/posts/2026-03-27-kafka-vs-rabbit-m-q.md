---
title: "메시징: Kafka vs RabbitMQ 완전 비교"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["systems", "devops", "cloud"]
toc: true
comments: true
---

# 메시징: Kafka vs RabbitMQ 완전 비교
## 요약

- Kafka: 높은 처리량, 이벤트 스트림 처리
- RabbitMQ: 신뢰성, 메시지 큐
- 벤치마크 및 실전 사례
- 언제 뭘 쓸까?

---

## 1. 핵심 비교

| 항목 | Kafka | RabbitMQ |
|------|-------|----------|
| 처리량 | 100K+ msg/s | 50K msg/s |
| 지연 | 낮음 (<100ms) | 매우 낮음 (<10ms) |
| 보증 | At-least-once | Exactly-once |
| 구조 | Pub-Sub + Queue | Queue |
| 영속성 | 필수 | 선택 |
| 커뮤니티 | 크고 활발 | 성숙함 |

---

## 2. Kafka 설치 및 사용

### Docker

```yaml
docker-compose up -d kafka zookeeper
```

### Producer

```python
from kafka import KafkaProducer
import json

producer = KafkaProducer(
    bootstrap_servers=['localhost:9092'],
    value_serializer=lambda v: json.dumps(v).encode('utf-8')
)

for i in range(1000):
    producer.send('orders', {'id': i, 'amount': 100})

producer.flush()
```

### Consumer (Group)

```python
from kafka import KafkaConsumer

consumer = KafkaConsumer(
    'orders',
    bootstrap_servers=['localhost:9092'],
    group_id='order-processors',
    auto_offset_reset='earliest'
)

for message in consumer:
    print(f"Order: {message.value}")
```

---

## 3. RabbitMQ 설치 및 사용

### Docker

```bash
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

### Producer

```python
import pika
import json

connection = pika.BlockingConnection(
    pika.ConnectionParameters('localhost')
)
channel = connection.channel()

channel.queue_declare(queue='orders', durable=True)

for i in range(1000):
    channel.basic_publish(
        exchange='',
        routing_key='orders',
        body=json.dumps({'id': i, 'amount': 100})
    )

connection.close()
```

### Consumer

```python
def callback(ch, method, properties, body):
    print(f"Order: {body}")
    ch.basic_ack(delivery_tag=method.delivery_tag)

channel.basic_qos(prefetch_count=1)
channel.basic_consume(
    queue='orders',
    on_message_callback=callback
)

channel.start_consuming()
```

---

## 4. 성능 벤치마크

### 처리량 (msg/sec)

```
메시지 크기: 1KB
복제: 1개

Kafka:
- 1개 Producer: 100K msg/s
- 10개 Producers: 500K msg/s

RabbitMQ:
- 1개 Producer: 50K msg/s
- 10개 Producers: 150K msg/s

승자: Kafka (3배 높음)
```

### 지연시간 (P99)

```
Kafka:
- 발행 → 소비: 100ms

RabbitMQ:
- 발행 → 소비: 5ms

승자: RabbitMQ (20배 빠름)
```

---

## 5. 선택 기준

### Kafka 선택

```
- 높은 처리량 필요
- 이벤트 히스토리 유지
- 스트림 처리 (복잡한 변환)
- 예: 실시간 로그, 클릭 이벤트, 센서 데이터
```

### RabbitMQ 선택

```
- 낮은 지연 필수
- 메시지 신뢰성 중요
- 복잡한 라우팅 필요
- 예: 주문 처리, 결제, 이메일 발송
```

---

## 6. 실전 패턴

### Kafka: 이벤트 소싱

```python
# 모든 이벤트를 저장 (히스토리)
events = [
    {'event': 'order_created', 'id': 1},
    {'event': 'payment_completed', 'id': 1},
    {'event': 'order_shipped', 'id': 1},
]

for event in events:
    producer.send('events', event)

# 소비자 1: 미니 물리화
consumer1 -> orders_view (현재 상태)

# 소비자 2: 분석
consumer2 -> analytics (트렌드)

# 소비자 3: 감사
consumer3 -> audit_log (전체 히스토리)
```

### RabbitMQ: Task Queue

```python
# 작업 큐
channel.queue_declare(queue='tasks', durable=True)

# 생성
channel.basic_publish(
    exchange='',
    routing_key='tasks',
    body=json.dumps({'task': 'send_email', 'to': 'user@example.com'}),
    properties=pika.BasicProperties(delivery_mode=2)  # 영속성
)

# 소비 (여러 워커)
worker1, worker2, worker3 처리
```

---

## 7. 하이브리드

```
아키텍처:
┌─────────────────────────────────┐
│ 이벤트 스트림 (Kafka)            │
│ - 모든 이벤트 저장               │
│ - 히스토리 추적                 │
└──────────────┬────────────────────┘
               │
        ┌──────┴──────┐
        ▼             ▼
    ┌────────┐   ┌─────────────┐
    │RabbitMQ│   │분석 파이프라인│
    │(즉시)  │   │(배치)       │
    └────────┘   └─────────────┘
```

---

## 핵심 정리

- **Kafka**: 대규모, 느림
- **RabbitMQ**: 빠름, 작은 규모

---

## 결론

**문제에 맞는 도구를 선택하세요!** 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
