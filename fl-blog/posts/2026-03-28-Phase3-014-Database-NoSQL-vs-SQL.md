---
layout: post
title: Phase3-014-Database-NoSQL-vs-SQL
date: 2026-03-28
---
# 데이터베이스: NoSQL vs SQL 언제 뭘 쓸까?

## 요약

**배우는 내용**:
- SQL (관계형): ACID 보장, 스키마 엄격
- NoSQL (문서/KV): 스키마 유연, 수평 확장
- 성능 벤치마크 (쓰기/읽기 속도)
- 실제 사례: 언제 어떤 DB를 선택할까

---

## 1. SQL vs NoSQL 비교

### 핵심 차이점

```
항목              | SQL                | NoSQL
-----------------|-------------------|---
스키마            | 엄격 (고정)        | 유연 (동적)
트랜잭션          | ACID 보장          | 최종 일관성
확장              | 수직 (더 큰 서버)  | 수평 (서버 추가)
쿼리              | SQL 표준           | 각각 다름
조인              | 네이티브 지원      | 애플리케이션 단계
일관성            | 즉시 (Strong)      | 약한 (Eventual)
```

---

## 2. SQL: 관계형 데이터베이스

### 특징

```sql
-- 스키마: 미리 정의
CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    age INT
);

-- 데이터 삽입
INSERT INTO users VALUES (1, 'Alice', 'alice@example.com', 30);

-- 조인: 여러 테이블 연결
SELECT u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id
HAVING COUNT(o.id) > 5;

-- 트랜잭션: ACID 보장
BEGIN TRANSACTION;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;  -- 둘 다 성공하거나 둘 다 실패
```

### 예시: PostgreSQL

```go
// Go에서 SQL 쿼리
db, _ := sql.Open("postgres", "postgres://user:pass@localhost/db")

var name string
var age int

row := db.QueryRow("SELECT name, age FROM users WHERE id = ?", 1)
err := row.Scan(&name, &age)

// Prepared Statement (SQL Injection 방지)
stmt, _ := db.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
defer stmt.Close()

result, _ := stmt.Exec("Bob", "bob@example.com")
lastID, _ := result.LastInsertId()
```

### SQL의 강점
✅ **데이터 무결성**: 외래 키 제약, 타입 검증
✅ **복잡한 쿼리**: JOIN, GROUP BY, 집계 쉬움
✅ **트랜잭션**: 다중 업데이트 원자성 보장
✅ **표준화**: SQL 문법 통일

### SQL의 약점
❌ **수평 확장**: 샤딩 복잡
❌ **스키마 변경**: 대량 데이터 시 느림
❌ **준구조 데이터**: JSON 처리 어려움

---

## 3. NoSQL: 비관계형 데이터베이스

### 타입별 특징

#### (1) 문서 DB (Document DB) - MongoDB

```javascript
// 스키마 없음: 언제든 필드 추가 가능
db.users.insertOne({
    _id: 1,
    name: "Alice",
    email: "alice@example.com",
    age: 30,
    tags: ["vip", "early-adopter"],  // 배열
    address: {                        // 중첩 객체
        street: "123 Main St",
        city: "Seoul"
    }
});

// 쿼리: JSON 형식
db.users.find({
    age: { $gt: 25 },
    tags: { $in: ["vip"] }
});

// 집계
db.orders.aggregate([
    { $match: { status: "completed" } },
    { $group: { _id: "$user_id", total: { $sum: "$amount" } } },
    { $sort: { total: -1 } }
]);
```

#### (2) Key-Value DB - Redis

```go
// 메모리 기반, 극단적 성능
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 문자열
client.Set(ctx, "user:1:name", "Alice", 0)
val, _ := client.Get(ctx, "user:1:name").Result()

// 해시
client.HSet(ctx, "user:1", map[string]interface{}{
    "name": "Alice",
    "age": "30",
    "email": "alice@example.com",
})

// 리스트 (큐)
client.RPush(ctx, "job_queue", "job1", "job2", "job3")
job, _ := client.LPop(ctx, "job_queue").Result()

// 집합 (고유값)
client.SAdd(ctx, "user:1:tags", "vip", "early-adopter")
tags, _ := client.SMembers(ctx, "user:1:tags").Result()
```

#### (3) 열 기반 DB - HBase, Cassandra

```
행 기반 (OLTP):          열 기반 (OLAP):
┌─────────────────┐     ┌─────────┬─────────┬─────┐
│ ID │ Name │ Age │     │ ID │ ID │ ID │ │     │
├────┼──────┼─────┤     ├────┼────┼────┼─────┤
│ 1  │ Alice│ 30  │     │ 1  │ 2  │ 3  │ Age │
│ 2  │ Bob  │ 25  │     │ 2  │ 3  │ 4  │ 25 │
└─────────────────┘     └─────────────────────┘

좋음: ID=1 조회 (행 전체)
나쁨: 모든 나이 조회 (전체 스캔)

좋음: 모든 나이 계산 (1개 열만)
나쁨: ID=1 조회 (여러 열 검색)
```

### NoSQL의 강점
✅ **수평 확장**: 노드 추가로 용량 증가
✅ **성능**: 읽기/쓰기 매우 빠름
✅ **스키마 유연성**: 동적 필드
✅ **빅데이터**: 수 TB 데이터 처리

### NoSQL의 약점
❌ **약한 일관성**: 최종 일관성만 보장
❌ **복잡한 조인**: 애플리케이션 코드 필요
❌ **트랜잭션**: 단일 문서만 원자성 보장
❌ **학습곡선**: 각 DB마다 다른 API

---

## 4. 성능 벤치마크

### 벤치마크 1: 쓰기 성능 (초당 요청 수)

```
데이터베이스      | 1MB 문서 | 100B 문서 | 메모리
-----------------|---------|----------|------
PostgreSQL       | 5K req/s | 50K req/s | 500MB
MongoDB          | 8K req/s | 80K req/s | 800MB
Redis            | 100K req/s | 1M req/s | 100MB
Cassandra        | 20K req/s | 200K req/s | 1GB
```

### 벤치마크 2: 읽기 성능 (95 percentile latency)

```
작업              | PostgreSQL | MongoDB | Redis | Cassandra
-----------------|-----------|---------|-------|----------
단일 문서 읽기    | 2ms       | 3ms     | <1ms  | 5ms
100개 문서 쿼리  | 50ms      | 100ms   | 10ms  | 150ms
집계 (1M)         | 500ms     | 1000ms  | -     | 2000ms
전체 스캔 (1M)   | 5000ms    | 10000ms | -     | 20000ms
```

### 벤치마크 3: 메모리 사용 (100만 문서)

```
{"id": 1, "name": "Alice", "email": "alice@example.com", "age": 30}

DB        | 메모리 | 인덱스 | 메타 | 총합
----------|--------|--------|------|------
PostgreSQL| 500MB  | 200MB  | 50MB | 750MB
MongoDB   | 800MB  | 300MB  | 200MB| 1.3GB
Redis     | 1.2GB  | 0      | 100MB| 1.3GB
```

---

## 5. 실제 선택 기준

### SQL을 선택할 때

```
사용 사례: 금융 시스템 (은행)

특징:
- 계좌 간 송금 (트랜잭션 필수)
- 정규화 스키마 (계좌, 사용자, 거래)
- 복잡한 리포트 (JOIN 다용)
- 데이터 무결성 (외래 키)

구현:
PostgreSQL + Connection Pool
┌──────────────┐
│ App Server   │
├──────────────┤
│ - 트랜잭션   │
│ - 리포트     │
│ - 검증       │
└──────┬───────┘
       │
    ┌──┴──┐
    │     │
┌───┴───┬─┴───┬────┐
│ Conn  │ Conn│    │
└───┬───┴─────┴────┘
    │
┌───┴──────────────┐
│ PostgreSQL       │
│ - ACID           │
│ - Foreign Keys   │
└──────────────────┘
```

### NoSQL을 선택할 때

```
사용 사례: 소셜 미디어 피드

특징:
- 대량 쓰기 (수백만 포스트/일)
- 유연 스키마 (사진, 비디오, 링크 등)
- 최종 일관성 OK (약간의 지연 허용)
- 수평 확장 필수

구현:
MongoDB + Sharding
┌──────────────────┐
│ App Server       │
├──────────────────┤
│ - 피드 조회      │
│ - 포스트 작성    │
│ - 댓글 추가      │
└──────┬───────────┘
       │
    ┌──┴──┬──┬───┐
    │     │  │   │
┌───┴───┬─┴──┴─┬─┴────┬────┐
│Shard 1│Shard│Shard │Shard│
└───────┴──────┴──────┴────┘
     MongoDB Cluster
     (자동 밸런싱)
```

### 하이브리드

```
실전 아키텍처: 전자상거래

┌──────────────────────────────────────┐
│ API Gateway                          │
└──────────────────┬───────────────────┘
       │           │              │
    읽기         쓰기          분석
       │           │              │
┌─────┴─────┐  ┌──┴──┐      ┌────┴──┐
│ Redis     │  │MySQL│      │Hadoop │
│ (캐시)    │  │ (주)│      │ (분석)│
└───────────┘  └─────┘      └───────┘
              │
         ┌────┴────┐
         │          │
      ┌──┴──┐  ┌───┴──┐
      │MongoDB│  │Cassandra
      │ (로그)│  │ (시계열)
      └──────┘  └──────┘
```

---

## 6. 마이그레이션 전략

### SQL → NoSQL

```sql
-- SQL: 정규화된 스키마
CREATE TABLE users (id INT, name VARCHAR);
CREATE TABLE orders (id INT, user_id INT, amount DECIMAL);

SELECT u.name, COUNT(o.id) as count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id;
```

```javascript
// NoSQL: 비정규화 (denormalization)
db.users.insertOne({
    _id: 1,
    name: "Alice",
    orders: [
        { id: 101, amount: 10000 },
        { id: 102, amount: 20000 }
    ],
    order_count: 2
})

// 쿼리: 조인 불필요
db.users.findOne({ _id: 1 });
```

---

## 핵심 정리

| 특징 | SQL | NoSQL |
|------|-----|-------|
| **데이터 무결성** | ✅ | ❌ |
| **확장성** | 어려움 | 쉬움 |
| **성능** | 중간 | 매우 빠름 |
| **복잡한 쿼리** | ✅ | 어려움 |
| **학습곡선** | 완만 | 가파름 |

---

## 결론

**"둘 다 배워야 한다"**

- SQL: 데이터 무결성이 필수라면
- NoSQL: 확장성과 성능이 우선이라면
- 현실: 둘 다 사용하는 하이브리드

규칙이 아닌 **비즈니스 요구사항**이 결정합니다! 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
