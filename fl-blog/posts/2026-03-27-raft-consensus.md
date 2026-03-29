---
title: "Raft 합의 알고리즘: 분산 시스템의 심장"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# Raft 합의 알고리즘: 분산 시스템의 심장
## 요약

- Raft 알고리즘의 핵심 개념
- Leader Election (리더 선출)
- Log Replication (로그 복제)
- Safety 보장
- 실전 구현 (etcd, Consul)

---

## 1. Raft란?

### Paxos의 대안

```
Paxos:     이해하기 어려움, 프로덕션 구현 복잡
Raft:      이해하기 쉬움, 프로덕션 준비됨

논문: "In Search of an Understandable Consensus Algorithm" (2014)
저자: Diego Ongaro, John Ousterhout (Stanford)
```

### 3가지 핵심 개념

```
1. Leader Election  → 리더 선출 메커니즘
2. Log Replication  → 상태 머신 복제
3. Safety           → 안전성 보장 (분할 내성)
```

---

## 2. Raft의 5가지 상태

### 노드 상태

```go
type RaftNode struct {
    state      State    // Follower, Candidate, Leader
    term       int      // 현재 term
    votedFor   int      // 누구에게 투표했는가
    log        []Entry  // 로그 항목
    commitIdx  int      // 커밋된 인덱스
    lastApplied int     // 적용된 마지막 인덱스

    // Leader만 유지
    nextIdx    []int    // 다음 전송 인덱스
    matchIdx   []int    // 복제된 인덱스
}

type State int
const (
    Follower = iota
    Candidate
    Leader
)

type Entry struct {
    Term  int
    Index int
    Cmd   []byte  // 상태 머신 명령
}
```

---

## 3. Leader Election (리더 선출)

### 타임아웃 기반 선출

```
Follower 상태:
├─ 심장박동(AppendEntries) 수신 → 타임아웃 리셋
├─ 타임아웃 (150-300ms) → Candidate로 전환
└─ 선출 타임아웃 → 리더 없음 상태

Candidate 상태:
├─ Term 증가
├─ 자신에게 투표
├─ RequestVote RPC 전송 (모든 노드)
├─ 과반 투표 획득 → Leader
├─ 높은 term의 AppendEntries 수신 → Follower
└─ 선출 타임아웃 → 새 Candidate
```

### RequestVote RPC

```go
type RequestVoteArgs struct {
    Term         int  // Candidate의 term
    CandidateId  int  // Candidate ID
    LastLogIdx   int  // 마지막 로그 인덱스
    LastLogTerm  int  // 마지막 로그 term
}

type RequestVoteReply struct {
    Term        int  // 응답자의 current term
    VoteGranted bool // Candidate에게 투표 여부
}

// Follower의 투표 규칙
if args.Term < rf.currentTerm {
    reply.VoteGranted = false  // 낮은 term 거부
}

if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
    if args.LastLogTerm > lastLogTerm ||
       (args.LastLogTerm == lastLogTerm && args.LastLogIdx >= lastLogIdx) {
        rf.votedFor = args.CandidateId
        reply.VoteGranted = true  // 로그가 최신 → 투표
    }
}
```

### "Randomized Election Timeout" 핵심

```
문제: 동시에 여러 Candidate 선출 시도
      → 투표가 분산되어 누구도 과반 못함
      → 계속 재선출

해결책: 각 노드의 타임아웃을 무작위로 설정
├─ Node A: 150-300ms
├─ Node B: 150-300ms (다른 범위)
├─ Node C: 150-300ms (다른 범위)

결과: 확률적으로 한 노드가 먼저 타임아웃
      → 그 노드가 Candidate → 다른 노드의 투표 획득
```

---

## 4. Log Replication (로그 복제)

### AppendEntries RPC

```go
type AppendEntriesArgs struct {
    Term         int     // Leader의 term
    LeaderId     int     // Leader ID
    PrevLogIdx   int     // 이전 로그 인덱스
    PrevLogTerm  int     // 이전 로그 term
    Entries      []Entry // 새 항목 (0개면 heartbeat)
    LeaderCommit int     // Leader의 commitIdx
}

type AppendEntriesReply struct {
    Term    int  // 응답자의 term
    Success bool // PrevLogIdx/Term이 일치?
}

// Leader의 복제 로직
func (rf *Raft) SendAppendEntries(server int) {
    args := &AppendEntriesArgs{
        Term:        rf.currentTerm,
        LeaderId:    rf.me,
        PrevLogIdx:  rf.nextIdx[server] - 1,
        PrevLogTerm: rf.log[rf.nextIdx[server]-1].Term,
        Entries:     rf.log[rf.nextIdx[server]:],
        LeaderCommit: rf.commitIdx,
    }

    var reply AppendEntriesReply
    ok := rf.sendAppendEntries(server, args, &reply)

    if !ok {
        return
    }

    if reply.Term > rf.currentTerm {
        rf.currentTerm = reply.Term
        rf.state = Follower
        rf.votedFor = -1
        return
    }

    if reply.Success {
        rf.nextIdx[server] = args.PrevLogIdx + len(args.Entries) + 1
        rf.matchIdx[server] = args.PrevLogIdx + len(args.Entries)
    } else {
        rf.nextIdx[server]--  // 백트래킹
    }
}
```

### 로그 복제 타이밍

```
시간    Leader          Follower1       Follower2
────────────────────────────────────────────────
T0      [cmd1]          []              []
T1      [cmd1] →        [cmd1]
T2      [cmd1] →                        [cmd1]
T3      [cmd1,cmd2] →   [cmd1,cmd2]     [cmd1]
T4      [cmd1,cmd2] →                   [cmd1,cmd2]
T5      commit(1) →     commit(1) →
T6                                      commit(1) →

과반(2/3)이 복제 → Leader가 commit
```

---

## 5. Safety 보장

### 3가지 안전성 보장

#### (1) Election Safety
```
한 term에 최대 1명의 리더만 선출됨

이유: 리더 선출에 과반 투표 필요
      → 최대 1명만 과반 투표 획득 가능
```

#### (2) Leader Append-Only
```
로그 항목은 Leader에서만 추가됨
Leader는 기존 항목을 덮어쓰거나 삭제하지 않음

이유: Log Replication에서 Follower의 로그만 변경
```

#### (3) Log Matching Property
```
두 로그가 같은 index/term을 가지면:
├─ 그 이전의 모든 항목도 동일
└─ 그 이후의 모든 항목도 동일

구현: AppendEntries의 PrevLogIdx/PrevLogTerm 검사
```

### "State Machine Safety"

```
문제: 일부 서버만 특정 로그 항목을 적용한 후
      그 서버가 offline → 새 리더 선출
      → 그 로그 항목 손실 가능

해결책: 높은 term의 명령만 commit
├─ Leader는 자신의 term에서 과반 복제된 항목만 commit
└─ 이전 term의 항목은 현재 term 항목이 commit될 때만 commit
```

```go
// Leader가 commit 결정
if rf.matchIdx[i] >= rf.commitIdx {
    // 과반이 인덱스 N을 복제
    if N > rf.commitIdx && rf.log[N].Term == rf.currentTerm {
        rf.commitIdx = N
    }
}
```

---

## 6. 실전: etcd (Kubernetes 기초)

### etcd 아키텍처

```
etcd (10개 노드)
├─ 1 Leader
└─ 9 Followers

Kubernetes kube-apiserver (모든 노드)
└─ etcd와 통신 (read/write)

모든 상태가 etcd에 저장됨:
├─ Pod 정의
├─ Service 설정
├─ ConfigMap
└─ Secret
```

### etcd 성능

```
쓰기 처리량: 1,000 ops/sec (1MB/s)
읽기 처리량: 100,000 ops/sec (캐시)

지연시간:
├─ 로컬 읽기: <1ms
├─ 분산 쓰기: 50-100ms (Raft 복제)
├─ Quorum 읽기: 10-50ms
```

---

## 7. 실전 구현 팁

### (1) "Prevote" 최적화

```
문제: Network partition 복구 시
      isolated Candidate의 높은 term이 모든 노드을 reset

해결책: Prevote phase 추가
├─ 실제 term 증가 전에 투표 연습
└─ Prevote 과반 획득 후만 term 증가
```

### (2) "Snapshotting"

```
문제: 오래된 로그로 인한 메모리 폭증

해결책: 주기적으로 상태 머신 스냅샷 저장
├─ 스냅샷 이전 로그 제거
└─ 복구 시 스냅샷 로드 + 이후 로그 재생
```

### (3) "Batching" 최적화

```
매번 1개씩 append → RPC 오버헤드 큼

대신:
├─ 버퍼에 여러 명령 모음
├─ 일정 시간 또는 크기 도달
└─ 배치로 전송
```

---

## 8. 벤치마크

### Raft vs Paxos

```
메트릭               Raft    Paxos
────────────────────────────────
이해도               쉬움    어려움
구현 복잡도          낮음    높음
성능                 같음    같음
프로덕션 준비        준비됨  미준비
커뮤니티 채택        광범위  제한적
```

### 실전 성능 (5개 노드)

```
지연시간 (커밋):
├─ LAN (1ms RTT):        10-50ms
├─ WAN (50ms RTT):       100-500ms
└─ 느린 네트워크:         >1초

처리량:
├─ 로컬:     1,000-10,000 ops/sec
├─ 원격:     100-1,000 ops/sec
└─ Batch:    10,000-100,000 ops/sec (배치 크기 1000)
```

---

## 핵심 정리

| 개념 | 설명 | 중요도 |
|------|------|--------|
| **Term** | 선출 에포크 | ⭐⭐⭐ |
| **Randomized timeout** | 리더 선출 수렴 보장 | ⭐⭐⭐ |
| **Log Matching** | 안전성 기초 | ⭐⭐⭐ |
| **Quorum** | 과반 복제 = 안전 | ⭐⭐⭐ |
| **Prevote** | 분할 내성 개선 | ⭐⭐ |

---

## 결론

**"Raft는 이해할 수 있는 합의 알고리즘이다"**

Kubernetes의 etcd, Consul, TiDB 등 수백 개 시스템이 Raft에 의존합니다.

분산 시스템을 배우면 Raft를 배우세요! 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
