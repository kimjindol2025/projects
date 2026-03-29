---
title: "Quorum 기반 분산 잠금: 안전하고 빠른 합의"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# Quorum 기반 분산 잠금: 안전하고 빠른 합의
## 요약

- Quorum의 개념과 수학
- Read/Write Quorum 트레이드오프
- 분산 잠금 구현 (Chubby, Zookeeper)
- 성능 vs 가용성 분석
- 실전 사용 패턴

---

## 1. Quorum이란?

### 정의

```
N개 노드 중 과반(> N/2)에 동의 필요

Quorum Q:
├─ 항상 일관성 보장
├─ Q1 ∩ Q2 ≠ ∅ (모든 Quorum이 겹침)
└─ 최소 Quorum = ⌈(N+1)/2⌉

예: 5개 노드
├─ Quorum 크기: 3개 (5 > 2*2)
├─ 최대 장애 노드: 2개
└─ 가용성: 97%+
```

### Quorum의 수학적 성질

```
N개 노드, Q = Quorum 크기

두 Quorum Q1, Q2:
|Q1| + |Q2| > N
→ Q1 ∩ Q2 ≠ ∅

증명:
|Q1| > N/2, |Q2| > N/2
|Q1| + |Q2| > N
→ 최소 1개 겹침

결론: 서로 다른 Quorum은 항상 최소 1개 노드 공유
```

---

## 2. Read/Write Quorum 트레이드오프

### 전략 1: Strong Consistency (읽기 무거움)

```
쓰기: W개 노드 동의 필요
읽기: R개 노드 동의 필요

조건: R + W > N (최신 데이터 보장)

예: N=5, W=3, R=3
├─ 쓰기 지연: 중간
├─ 읽기 지연: 중간
├─ 장애 허용: 2개 노드
```

### 전략 2: 읽기 최적화

```
쓰기: W = N (모든 노드)
읽기: R = 1 (1개 노드만)

조건: R + W = N+1 > N ✓

예: N=5, W=5, R=1
├─ 쓰기 지연: 높음 (모든 노드)
├─ 읽기 지연: 낮음 (1개 노드)
├─ 쓰기 처리량: 낮음
└─ 읽기 처리량: 높음 (캐시)
```

### 전략 3: 쓰기 최적화

```
쓰기: W = 1 (1개 노드만)
읽기: R = N (모든 노드)

예: N=5, W=1, R=5
├─ 쓰기 지연: 매우 낮음
├─ 읽기 지연: 높음 (모든 노드)
├─ 쓰기 처리량: 높음
└─ 읽기 처리량: 낮음
```

### 비교표

| 전략 | W | R | 쓰기 지연 | 읽기 지연 | 일관성 |
|------|---|---|---------|---------|--------|
| 균형 | 3 | 3 | 중간 | 중간 | 강함 |
| 읽기최적 | 5 | 1 | 높음 | 낮음 | 약함 |
| 쓰기최적 | 1 | 5 | 낮음 | 높음 | 약함 |

---

## 3. 분산 잠금 구현

### 기본 아이디어

```
잠금 = Quorum 다수가 "점유" 상태

규칙:
├─ 잠금 획득: Quorum이 "점유 불가" 응답할 때까지 재시도
├─ 잠금 해제: 모든 노드에 해제 요청
└─ 만료: 타임아웃 후 자동 해제
```

### Go 구현

```go
type DistributedLock struct {
    Name      string
    Owner     string  // 잠금 소유자 ID
    Timestamp int64   // 획득 시간
    TTL       int     // 유효 기간 (초)
}

type LockManager struct {
    locks map[string]DistributedLock
    nodes []string  // Quorum 노드들
}

func (lm *LockManager) Acquire(lockName, ownerID string, ttl int) bool {
    quorumSize := (len(lm.nodes) / 2) + 1
    successes := 0

    for _, node := range lm.nodes {
        if lm.TryLock(node, lockName, ownerID, ttl) {
            successes++
        }
        if successes >= quorumSize {
            return true
        }
    }

    // Quorum 미달 → 획득 실패
    // 부분 획득한 잠금 해제 (cleanup)
    for _, node := range lm.nodes {
        lm.ReleaseLock(node, lockName, ownerID)
    }
    return false
}

func (lm *LockManager) TryLock(node, lockName, ownerID string, ttl int) bool {
    lock, exists := lm.locks[lockName]

    // 이미 잠김
    if exists && lock.Owner != ownerID {
        // TTL 확인
        if time.Now().Unix() - lock.Timestamp < int64(lock.TTL) {
            return false  // 아직 유효
        }
        // TTL 만료 → 탈취 가능
    }

    // 잠금 설정
    lm.locks[lockName] = DistributedLock{
        Name:      lockName,
        Owner:     ownerID,
        Timestamp: time.Now().Unix(),
        TTL:       ttl,
    }
    return true
}

func (lm *LockManager) Release(lockName, ownerID string) bool {
    for _, node := range lm.nodes {
        lm.ReleaseLock(node, lockName, ownerID)
    }
    return true
}

func (lm *LockManager) ReleaseLock(node, lockName, ownerID string) bool {
    lock, exists := lm.locks[lockName]

    // 소유자 확인
    if !exists || lock.Owner != ownerID {
        return false
    }

    delete(lm.locks, lockName)
    return true
}
```

### 사용 패턴

```go
// 잠금 획득
manager := NewLockManager(5)  // 5개 노드

if !manager.Acquire("payment-123", "worker-1", 30) {
    // 잠금 획득 실패 → 재시도 또는 포기
    return errors.New("could not acquire lock")
}

defer manager.Release("payment-123", "worker-1")

// 임계 영역 (Critical Section)
ProcessPayment("payment-123")  // 한 노드씩만 실행

// 자동 해제
```

---

## 4. Chubby (Google의 잠금 서비스)

### 아키텍처

```
Chubby Master (Leader)
├─ Lock 서버 1개 (모든 쓰기)
└─ Session 유지

Chubby Replicas (Followers)
├─ 읽기 분산
└─ Failover

클라이언트:
├─ Master에서만 쓰기
├─ Replicas에서 읽기
└─ Master 장애 시 자동 failover
```

### 세션 (Session)

```
클라이언트와 Chubby의 계약

규칙:
├─ Heartbeat 주기: 60초
├─ Lease 유효 기간: 180초
├─ Lease 만료 → 세션 종료
└─ 모든 잠금 자동 해제

장점:
├─ 좀비 프로세스 정리
└─ 자동 잠금 해제
```

### Chubby 성능

```
잠금 획득 지연: 10-100ms (Quorum 투표)
잠금 해제 지연: 1-10ms
처리량: 1,000-10,000 locks/sec

네트워크 효율:
├─ Heartbeat: 작은 메시지
├─ Batching: 여러 잠금 함께 처리
└─ 캐싱: 클라이언트 측 캐시
```

---

## 5. Zookeeper의 잠금

### Zookeeper Recipes

```
Zookeeper를 이용한 분산 잠금

아이디어:
├─ Ephemeral Sequential Node 생성
├─ 가장 작은 시퀀스 번호 = 잠금 획득자
└─ Watch로 앞의 노드 감시 (깨어나기)

예:
/locks/my-lock
├─ /locks/my-lock-000001 (현재 소유)
├─ /locks/my-lock-000002 (대기)
└─ /locks/my-lock-000003 (대기)
```

### 구현 (의사코드)

```go
func AcquireLock(path string) (acquired bool) {
    // 1. Sequential ephemeral node 생성
    myNode := Create(path + "-", mode=EPHEMERAL_SEQUENTIAL)
    // 결과: /locks/my-lock-000042

    for {
        // 2. 모든 자식 나열 (정렬됨)
        children := GetChildren(path)
        // [my-lock-000041, my-lock-000042, my-lock-000043, ...]

        // 3. 내 노드가 가장 작은 번호?
        if myNode == children[0] {
            return true  // 잠금 획득!
        }

        // 4. 앞의 노드 감시
        predecessor := children[indexOf(myNode) - 1]
        Watch(predecessor)

        // 5. 앞 노드 삭제 대기 (watched)
        WaitForDeletion(predecessor)  // 블로킹
    }
}

func ReleaseLock(myNode string) {
    Delete(myNode)  // 자동으로 앞 노드의 Watch 트리거
}
```

### Zookeeper의 이점

```
Chubby vs Zookeeper:
├─ Chubby: Session 기반 (복잡함)
├─ Zookeeper: Node 기반 (간단함)
├─ Zookeeper: Open source (Hadoop 생태계)
└─ Zookeeper: 더 확장성 좋음
```

---

## 6. 성능 분석

### Quorum 크기 vs 가용성

```
N=5 노드:
├─ Quorum 크기 3 (60%)
├─ 장애 허용 2개
├─ 가용성 99%+

N=7 노드:
├─ Quorum 크기 4 (57%)
├─ 장애 허용 3개
├─ 가용성 99.9%

N=9 노드:
├─ Quorum 크기 5 (56%)
├─ 장애 허용 4개
├─ 가용성 99.99%

패턴: N이 커질수록 상대 Quorum 크기 감소
```

### 지연시간 분석

```
시나리오: 5개 노드, RTT 1ms

Serial 접근 (모두 기다림):
├─ 최악: 5 × 1ms = 5ms
├─ 평균: 3 × 1ms = 3ms (Quorum도달)

Parallel 접근 (병렬 요청):
├─ 최악: 1ms (가장 느린 노드, 3/5는 응답)
└─ 실제: 1-2ms

지연시간 = Quorum 크기와 무관 (병렬화 시)
```

### 처리량

```
잠금 서버 처리 능력:
├─ 1개 노드: 10,000 locks/sec
├─ Quorum 3 (병렬): 10,000 locks/sec (동일!)
├─ 병렬화로 처리량 보존 가능

병목: 네트워크 대역폭, not Quorum
```

---

## 7. 한계와 개선

### 문제 1: Byzantine Failures

```
Quorum은 정직한 노드만 가정

Byzantine (악의적) 노드:
├─ "yes"라고 거짓말
├─ 데이터 손상
└─ 일관성 위반

해결책: BFT (Byzantine Fault Tolerance)
├─ 2f+1개 서명 필요 (f = Byzantine 노드)
├─ Quorum보다 복잡함
```

### 문제 2: 네트워크 분할

```
5개 노드 → 3:2 분할

그룹 A (3개):
├─ Quorum = 3
├─ 잠금 획득 가능

그룹 B (2개):
├─ Quorum = 3
├─ 잠금 획득 불가

두 그룹이 서로 다른 잠금 획득?
→ 일관성 위반 ❌
```

### 해결책: Fencing Token

```
각 잠금에 Token 부여

규칙:
├─ Token은 monotonic (증가만 함)
├─ 자원 접근 시 Token 확인
└─ 구 Token은 거부

예:
Lock1: token=100
Lock2: token=101 (새 잠금)

리소스 서버:
├─ token=100 요청 거부 (구형)
└─ token=101 요청 수락
```

---

## 8. 실전 팁

### Deadlock 방지

```
규칙:
├─ 항상 같은 순서로 잠금 획득
├─ Timeout 설정 필수
└─ 순환 의존성 제거

예:
Lock A ← Lock B → Lock A (×)
A, B 순서 고정 (○)
```

### Starvation 방지

```
문제: 높은 경합에서 특정 스레드가 계속 대기

Zookeeper 해결책:
└─ Sequential node → FIFO 순서 보장

Chubby 해결책:
└─ Random delay로 재시도
```

---

## 핵심 정리

| 개념 | 용도 | 복잡도 |
|------|------|--------|
| **Quorum** | 기본 합의 | 낮음 |
| **Chubby** | 프로덕션 | 중간 |
| **Zookeeper** | 프로덕션 | 중간 |
| **BFT** | 극단적 안전성 | 높음 |

---

## 결론

**"Quorum은 간단하지만 강력하다"**

Google, Yahoo, Airbnb 등이 기반으로 사용합니다.

분산 시스템의 일관성 = Quorum 이해 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
