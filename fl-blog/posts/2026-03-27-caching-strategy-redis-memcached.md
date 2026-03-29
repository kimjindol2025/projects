---
title: "캐싱 전략: Redis vs Memcached 실전 비교"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["systems", "devops", "cloud"]
toc: true
comments: true
---

# 캐싱 전략: Redis vs Memcached 실전 비교
## 요약

**배우는 내용**:
- Memcached: 간단한 KV 캐시 (메모리만)
- Redis: 고급 자료구조 + 영속성
- 캐시 전략: Cache-Aside, Write-Through, Write-Behind
- 실제 벤치마크와 선택 기준

---

## 1. Memcached vs Redis

### 핵심 비교

```
항목              | Memcached    | Redis
-----------------|--------------|----------
메모리 효율       | 매우 높음    | 높음
자료구조          | String만     | 5+ 타입
영속성            | 없음         | RDB/AOF
복제              | 없음         | 마스터-슬레이브
스크립트          | 없음         | Lua
성능              | 매우 빠름    | 매우 빠름
복잡도            | 낮음         | 중간
메모리사용        | 100MB/1M keys| 250MB/1M keys
```

---

## 2. 사용 패턴

### Memcached

```python
import memcache

mc = memcache.Client(['127.0.0.1:11211'])

# 저장 (300초 TTL)
mc.set('user:1:name', 'Alice', 300)

# 조회
name = mc.get('user:1:name')

# 삭제
mc.delete('user:1:name')

# 증가 (원자성 보장)
mc.incr('counter', 1)

# Batch 조회 (효율적)
keys = ['user:1', 'user:2', 'user:3']
users = mc.get_multi(keys)
```

### Redis

```python
import redis

r = redis.Redis(host='localhost', port=6379, decode_responses=True)

# 문자열
r.set('user:1:name', 'Alice', ex=300)
name = r.get('user:1:name')

# 해시 (객체)
r.hset('user:1', mapping={'name': 'Alice', 'age': 30})
user = r.hgetall('user:1')

# 리스트 (큐)
r.rpush('jobs', 'job1', 'job2')
job = r.lpop('jobs')

# 집합 (고유값)
r.sadd('tags:python', 'async', 'fast', 'simple')

# 정렬집합 (순위)
r.zadd('leaderboard', {'alice': 100, 'bob': 50, 'charlie': 75})
top10 = r.zrevrange('leaderboard', 0, 9, withscores=True)

# Lua 스크립트
script = r.register_script("""
    local current = redis.call('get', KEYS[1])
    if current then
        return redis.call('incr', KEYS[1])
    else
        return redis.call('set', KEYS[1], 1)
    end
""")
result = script(keys=['counter'])
```

---

## 3. 캐시 전략

### (1) Cache-Aside (Lazy Loading)

```python
def get_user(user_id):
    # 1. 캐시 확인
    cached = cache.get(f'user:{user_id}')
    if cached:
        return cached

    # 2. DB 조회
    user = db.get_user(user_id)

    # 3. 캐시에 저장
    cache.set(f'user:{user_id}', user, ex=3600)
    return user

# 장점: 필요한 데이터만 캐싱
# 단점: 첫 조회 시간 길어짐, 일관성 관리 필요
```

### (2) Write-Through

```python
def update_user(user_id, data):
    # 1. DB 업데이트
    db.update_user(user_id, data)

    # 2. 캐시 업데이트 (DB 이후)
    cache.set(f'user:{user_id}', data)
    return True

# 장점: 캐시-DB 일관성 보장
# 단점: 쓰기 지연 증가
```

### (3) Write-Behind (Write-Back)

```python
def batch_update(updates):
    # 1. 캐시만 업데이트 (즉시)
    for user_id, data in updates:
        cache.set(f'user:{user_id}', data)

    # 2. 배경 작업: DB 업데이트 (비동기)
    background_task.queue({
        'type': 'db_sync',
        'data': updates,
        'delay': 60  # 60초 후
    })

# 장점: 극도로 빠른 응답 (캐시만)
# 단점: DB 지연, 장애 시 데이터 손실
```

---

## 4. 실전 구현: 블로그 조회 수

### 시나리오

```python
# 매분 1000개 조회 발생
# DB 업데이트가 병목

class BlogService:
    def __init__(self, db, cache):
        self.db = db
        self.cache = cache

    def increment_view(self, post_id):
        # 1. Redis에 카운트 증가 (매우 빠름)
        current = self.cache.incr(f'post:{post_id}:views')

        # 2. 100의 배수마다 DB 동기화
        if current % 100 == 0:
            self.db.update_views(post_id, current)

    def get_views(self, post_id):
        # 캐시에서 조회 (거의 항상)
        return self.cache.get(f'post:{post_id}:views') or 0

# 성능:
# - 1 req: Redis <1ms
# - vs SQL: 50ms
# - 결과: 50배 빠름
```

---

## 5. 성능 벤치마크

### 벤치마크 1: 쓰기 처리량

```
작업                    | Memcached | Redis
------------------------|-----------|-------
set 연산 (1KB)          | 200K/s    | 180K/s
set + incrby            | 100K/s    | 150K/s
list push              | -          | 120K/s
sorted set add         | -          | 80K/s
```

### 벤치마크 2: 메모리 효율 (100만 key)

```
저장 데이터: key=16B, value=1KB

Memcached: 1.2GB (오버헤드 200MB)
Redis:     1.4GB (메타 200MB + 자료구조 오버헤드)

효율: Memcached > Redis
```

### 벤치마크 3: 지연시간 분포

```
Percentile | Memcached | Redis
-----------|-----------|-------
P50        | <0.1ms    | <0.1ms
P95        | 0.5ms     | 0.5ms
P99        | 2ms       | 2ms
P99.9      | 10ms      | 10ms
```

---

## 6. 실전 문제

### 문제 1: Cache Stampede (동시 갱신)

```python
# ❌ 문제: 동시에 여러 요청이 캐시 미스 발생
# 1000개 요청 → 모두 DB 조회 (DB 과부하)

def get_user_slow(user_id):
    user = cache.get(f'user:{user_id}')
    if not user:
        user = db.get_user(user_id)  # 1000x 호출!
        cache.set(f'user:{user_id}', user, ex=60)
    return user

# ✅ 해결 1: Lock 기반
def get_user_lock(user_id):
    user = cache.get(f'user:{user_id}')
    if not user:
        lock = cache.lock(f'user:{user_id}:lock', timeout=5)
        if lock.acquire(blocking=False):
            try:
                user = db.get_user(user_id)
                cache.set(f'user:{user_id}', user, ex=60)
            finally:
                lock.release()
        else:
            # 다른 스레드가 로드 중, 대기
            time.sleep(0.1)
            user = cache.get(f'user:{user_id}')
    return user

# ✅ 해결 2: Probabilistic early expiration
def get_user_prob(user_id):
    key = f'user:{user_id}'
    user, ttl = cache.get_with_ttl(key)
    if not user:
        return db.get_user(user_id)

    # TTL 90% 이상 경과 시 배경 갱신
    if ttl < 6:  # 60초 중 6초 남음 (90%)
        background_task.queue({
            'task': 'refresh_user',
            'user_id': user_id
        })
    return user
```

### 문제 2: 캐시 무효화

```python
# ❌ 문제: 사용자 정보 변경 → 캐시 갱신 지연

# ✅ 해결 1: 명시적 삭제
def update_user(user_id, data):
    db.update_user(user_id, data)
    cache.delete(f'user:{user_id}')  # 강제 무효화

# ✅ 해결 2: TTL 설정
def get_user(user_id):
    return cache.get_or_set(
        f'user:{user_id}',
        lambda: db.get_user(user_id),
        ex=300  # 5분 자동 만료
    )

# ✅ 해결 3: Tag 기반
def update_user(user_id, data):
    db.update_user(user_id, data)
    cache.delete_tag(f'user:{user_id}')
```

---

## 7. 선택 기준

### Memcached 선택

```
- 간단한 KV 캐시만 필요
- 극도의 성능 중요
- 메모리 제약 있음
- 예: 세션 스토어, HTML 조각 캐싱
```

### Redis 선택

```
- 복잡한 자료구조 필요
- 영속성/복제 필요
- Lua 스크립트 필요
- 예: 랭킹, 세션, 실시간 분석
```

---

## 8. 모니터링

```python
# Redis INFO
info = r.info()
print(f"Used Memory: {info['used_memory_human']}")
print(f"Connected Clients: {info['connected_clients']}")
print(f"Commands/sec: {info['instantaneous_ops_per_sec']}")

# Memcached STATS
stats = mc.get_stats()
print(f"Bytes: {stats[0][1]['bytes']}")
print(f"Evictions: {stats[0][1]['evictions']}")
```

---

## 핵심 정리

| 측면 | Memcached | Redis |
|------|-----------|-------|
| **속도** | 매우 빠름 | 매우 빠름 |
| **메모리** | 더 효율 | 약간 더 많음 |
| **기능** | 기본 | 풍부 |
| **복잡도** | 낮음 | 중간 |

---

## 결론

**"캐싱은 예술이다"**

- 올바른 전략 선택
- 일관성 유지
- 모니터링과 튜닝

🚀 캐싱으로 10배 빠른 애플리케이션을 만드세요!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
