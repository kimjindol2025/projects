---
layout: post
title: Phase4-032-Vector-Clock
date: 2026-03-28
---
# Vector Clock: 분산 시스템의 인과관계 추적

## 요약

- Lamport Clock의 한계
- Vector Clock 개념과 구현
- Happens-Before 관계
- 충돌 감지 (Version Control)
- 메모리 최적화 (Interval Tree Clocks)

---

## 1. 왜 시계가 필요한가?

### 분산 시스템의 시간 문제

```
문제 1: 물리적 시간 불신뢰
├─ 네트워크 지연 (수ms ~ 수초)
├─ NTP 동기화 오류 (≤100ms)
└─ 클라우드 환경의 클럭 드리프트

문제 2: 인과관계 추적 불가
Event A → Network → Event B
시간만으로는 A가 B 원인인지 알 수 없음
```

### 예시: 병렬 업데이트

```
프로세스 P1               프로세스 P2
────────────────────────────────────
T=10ms write("x=1")
                          T=11ms write("x=2")

어느 것이 "최신"인가?
시간만으로 결정 불가능 → 버전 충돌 발생
```

---

## 2. Lamport Clock (1978)

### 아이디어

```
각 이벤트에 단조증가하는 숫자를 붙이기

규칙:
1. 로컬 이벤트 → LC 증가
2. 메시지 수신 → LC = max(self, received) + 1

결과: LC(A) < LC(B) ⟹ A happens-before B
```

### 구현

```go
type Process struct {
    id int
    lc int
}

func (p *Process) LocalEvent() int {
    p.lc++
    return p.lc
}

func (p *Process) ReceiveMessage(msg Message) int {
    p.lc = max(p.lc, msg.LC) + 1
    return p.lc
}

type Message struct {
    Sender int
    LC     int
    Data   string
}
```

### 문제: 동시성을 구분할 수 없음

```
Process P1          Process P2
─────────────────────────────
LC(P1): 1           LC(P2): 1
event1
                    event2

LC(event1) = 1
LC(event2) = 1

1 < 1? No
1 > 1? No
→ "동시"라고 판단 (맞음)

하지만 연쇄적으로:

P1: event3 (LC=2) ← P2로부터 메시지 수신
                    P2: event4 (LC=2)

LC(event3) = max(2, 2) + 1 = 3
LC(event4) = 2

이제 LC(event4) < LC(event3)

하지만 event4 → event3 인과관계 없음! (동시)
Lamport Clock으로는 구분 불가
```

---

## 3. Vector Clock 해결책

### 아이디어

```
Lamport Clock: 숫자 1개
Vector Clock:  숫자 N개 (N = 프로세스 수)

VC[P] = [t1, t2, t3, ..., tn]
        각 프로세스별 로컬 시간
```

### 규칙

```
1. 로컬 이벤트:
   VC[P] 증가 (자신의 위치만)

2. 메시지 전송:
   메시지에 VC 포함

3. 메시지 수신:
   VC[i] = max(VC[i], msg.VC[i]) for all i
   VC[P] += 1 (수신 프로세스만)

비교:
VC1 < VC2 ⟺ VC1[i] ≤ VC2[i] for all i,
             VC1[j] < VC2[j] for some j

VC1 || VC2 ⟺ VC1 < VC2 아님 AND VC2 < VC1 아님
             (동시)
```

### 구현

```go
type VectorClock []int

type Process struct {
    id int
    vc VectorClock
}

func (p *Process) LocalEvent() VectorClock {
    p.vc[p.id]++
    return append([]int{}, p.vc...)  // 복사 반환
}

func (p *Process) SendMessage(msg Message) {
    msg.VC = p.LocalEvent()
}

func (p *Process) ReceiveMessage(msg Message) VectorClock {
    for i := range p.vc {
        p.vc[i] = max(p.vc[i], msg.VC[i])
    }
    p.vc[p.id]++
    return append([]int{}, p.vc...)
}

func Compare(vc1, vc2 VectorClock) string {
    less := false
    greater := false

    for i := range vc1 {
        if vc1[i] < vc2[i] {
            less = true
        }
        if vc1[i] > vc2[i] {
            greater = true
        }
    }

    if !less && !greater {
        return "equal"
    } else if !greater {
        return "less"
    } else if !less {
        return "greater"
    }
    return "concurrent"
}
```

### 예시: 3개 프로세스

```
P1          P2          P3
────────────────────────────
[1,0,0]
event1
│
├─ msg ──→          [0,0,0] → [0,1,0]
   [1,0,0]          event2     event3
               │
               └─ msg ──────→ [0,1,1]
                  [0,1,0]      event4

최종:
event1: [1,0,0]
event2: [0,1,0]
event3: [0,1,0]
event4: [0,1,1]

event1 happens-before event4?
[1,0,0] < [0,1,1]? No (1 > 0 at index 0)
결과: 동시 (concurrent) ✓

event3 happens-before event4?
[0,1,0] < [0,1,1]? Yes (0≤0, 1≤1, 0<1)
결과: event3 → event4 ✓
```

---

## 4. 실전 응용: 버전 관리

### 동시 편집 감지 (Git, CRDT)

```
Document state: [version, content]

User A edits:
├─ VC_A = [1, 0]
├─ "hello" → "hello world"
└─ commit

User B edits (오프라인):
├─ VC_B = [0, 1]
├─ "hello" → "hello everyone"
└─ commit

병합:
VC_A [1,0] vs VC_B [0,1]
→ concurrent → 충돌! 👇

해결책 (Merge):
├─ User A win: "hello world"
├─ User B win: "hello everyone"
└─ 3-way merge: "hello world everyone"
```

### Riak (NoSQL) 사용 사례

```go
// Riak KV Store
type RiakObject struct {
    Key       string
    Value     string
    VectorClock []int  // 버전 추적
    Siblings  []RiakObject  // 충돌된 버전들
}

// 쓰기
func Put(key, value string) {
    obj := Get(key)
    obj.Value = value
    obj.VectorClock[myNode]++
    Store(obj)
}

// 읽기
func Get(key string) RiakObject {
    obj := Fetch(key)
    if len(obj.Siblings) > 0 {
        // 충돌 감지! → 애플리케이션이 병합
        return ResolveConflict(obj.Siblings)
    }
    return obj
}
```

---

## 5. 문제점: 메모리 폭증

### 시간 이동에 따른 Vector Clock 크기

```
프로세스 수: 100개
시간경과: 1개월

Vector Clock 크기:
[1,000,000, 2,500,000, ..., 50,000,000]
= 100 × 8 bytes × 50,000,000 = 40GB

각 메시지마다 40GB... 불가능!
```

### 해결책 1: Interval Tree Clock (ITC)

```
Vector Clock: [t1, t2, ..., tn] = O(n) 공간

Interval Tree Clock:
├─ 트리 구조 (O(log n) 공간)
├─ 인터벌로 범위 표현
└─ 병합 시 인터벌 압축

결과: 크기 1/10 이하
```

### 해결책 2: Pruning (오래된 항목 제거)

```
규칙:
├─ 모든 프로세스가 수신한 이벤트 → 제거 가능
└─ 공통 접두사 (common prefix) 유지

예: [5, 5, 5, 100, 120]
    → [100, 120] (처음 5는 모든 노드가 알고있음)
```

### 해결책 3: Dotted Version Vectors

```
Vector Clock의 경량판

대신 "점(dot)"으로 변화 추적
[P1:5, P2:3, P3:7] → 동시 업데이트 감지

Riak의 실제 구현
크기: 수백 bytes (수MB 대신)
```

---

## 6. 비교: Lamport vs Vector vs Hybrid

### 메트릭 비교

| 메트릭 | Lamport | Vector | Hybrid |
|--------|---------|--------|--------|
| **공간** | O(1) | O(n) | O(log n) |
| **시간** | O(1) | O(n) | O(log n) |
| **인과관계** | 근사치 | 정확 | 정확 |
| **실전** | 추적용 | 충돌감지 | 권장 |

### 선택 기준

```
Lamport Clock:
├─ 대략적 순서 필요
├─ 메모리 제약 심함
└─ 예: 로그 순서화

Vector Clock:
├─ 정확한 인과관계
├─ 충돌 감지 필수
└─ 예: 분산 DB 버전 관리

Hybrid (ITC/DVV):
├─ 실전 프로덕션
├─ 메모리 효율성
└─ 예: Riak, DynamoDB 스타일
```

---

## 7. 벤치마크

### Vector Clock 오버헤드

```
메시지 크기 (프로세스 100개):

Lamport Clock: 8 bytes
Vector Clock: 800 bytes (100 × 8)
ITC: 80 bytes (추정)

네트워크 대역:
├─ Lamport: 1개 메시지 = 8 bytes
├─ Vector: 1개 메시지 = 800 bytes
├─ 초당 1000 메시지 시
  - Lamport: 8 KB/s
  - Vector: 800 KB/s (100배!)
```

### 실전 영향

```
Riak (버전 충돌 감지):
├─ Vector Clock 활성화: 처리량 95%
├─ 비활성화: 처리량 100%
└─ 트레이드오프: 정확성 vs 성능
```

---

## 8. 고급: Causally Consistent Storage

### Vector Clock을 이용한 인과성 일관성

```
규칙: 쓰기는 읽은 모든 버전을 포함해야 함

예:
Client A: Read(x) → VC=[1,0,0]
Client A: Write(y) ← 반드시 VC=[1,0,0] 전송
          (다른 클라이언트의 변화 반영)

Server: 인과성 추적으로 순서 보장
```

### 구현

```go
type CausalWrite struct {
    Key        string
    Value      string
    DependsOn  []VectorClock  // 읽은 버전들
}

func Write(w CausalWrite) error {
    for _, vc := range w.DependsOn {
        if !AllWritesApplied(vc) {
            return ErrCausalityViolation
        }
    }
    // 안전하게 쓰기
    return Store(w)
}
```

---

## 핵심 정리

| 개념 | 용도 | 크기 |
|------|------|------|
| **Lamport** | 전체 순서 | O(1) |
| **Vector** | 인과관계 | O(n) |
| **ITC/DVV** | 인과관계 + 효율 | O(log n) |

---

## 결론

**"벡터 클럭은 분산 시스템의 인과관계를 수학적으로 표현한다"**

Riak, Cassandra, DynamoDB 등 실전 DB들이 사용합니다.

분산 시스템의 동시성 이해 = 벡터 클럭 이해 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
