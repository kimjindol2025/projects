---
layout: post
title: Phase2-006-Profiling-Debugging
date: 2026-03-28
---
# Profiling과 성능 분석: go tool pprof 마스터하기

**작성**: 2026-03-27
**카테고리**: Performance Debugging, DevOps Tools
**읽는 시간**: 약 18분
**난이도**: 중급 개념, 실무 중심
**코드**: 실제 pprof 분석 사례

---

## 들어가며: "왜 느린가?"를 답하는 방법

```
프로덕션 서버 알람:

응답 시간이 갑자기 2배 증가!

매니저: "뭐가 문제야?"
개발자: "글쎄요... 아마도..."
  ├─ 알고리즘?
  ├─ 메모리?
  ├─ 네트워크?
  ├─ 디스크?
  └─ 추측만 하다가 1시간 낭비

✅ 올바른 방법:
1. pprof로 CPU 프로파일 수집
2. flamegraph 생성
3. "함수 X가 전체 시간의 58% 점유" 확인
4. 5분 내에 원인 파악 ✅
```

---

## 문제: 성능 저하의 원인을 모름

### 1.1 추측과 증명의 차이

```
시나리오: API 응답 시간 500ms → 2000ms (4배 악화)

❌ 추측 방식:
   "음... 메모리일까? 아니면 IO?"
   → 아무거나 건드려봄
   → 시간만 낭비

✅ 측정 방식:
   1. go tool pprof로 프로파일링
   2. top 10 함수 확인
   3. flamegraph로 시각화
   4. 정확한 원인 파악
   5. 해결 방안 실행
```

### 1.2 pprof가 답할 수 있는 질문들

```
Q1. "CPU 시간을 어디서 쓰는가?"
A: CPU profile → flamegraph
   "해시 함수 X에서 58% 소비" ✓

Q2. "메모리를 어디서 할당하는가?"
A: Heap profile → 메모리 누수 위치 파악 ✓

Q3. "Goroutine이 몇 개나 있는가?"
A: Goroutine profile → 고루틴 누수 감지 ✓

Q4. "Lock에서 얼마나 대기하는가?"
A: Mutex profile → Lock contention 분석 ✓

Q5. "Garbage Collection이 얼마나 자주 일어나는가?"
A: GC stats → GC 압박 파악 ✓
```

---

## 해결책: pprof 사용

### 2.1 pprof 기본 설정

```go
// pkg/server/server.go
package server

import (
    "net/http"
    _ "net/http/pprof"  // 이 한 줄이 전부!
)

type Server struct {
    // ...
}

func (s *Server) Start() error {
    // 메인 API 서버
    go http.ListenAndServe(":8080", s.handler)

    // pprof 서버 (별도 포트)
    go http.ListenAndServe(":6060", nil)  // pprof는 기본 mux 사용

    return nil
}
```

**이제 pprof 사용 가능:**

```bash
# CPU 프로파일 (30초)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 메모리 프로파일
go tool pprof http://localhost:6060/debug/pprof/heap

# 고루틴 프로파일
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 가용 프로파일 목록
curl http://localhost:6060/debug/pprof/
```

---

## 2.2 실전: CPU 프로파일 분석

### Case 1: API 응답 느려짐

```bash
# Step 1: CPU 프로파일 수집 (30초)
$ go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

Fetching profile over HTTP from http://localhost:6060/debug/pprof/profile?seconds=30
Saved profile in /home/user/pprof/pprof.samples.cpu.001.pb.gz
File: server
Type: cpu
Time: Mar 27 2026 09:15:33 UTC
Duration: 30.00s, Total samples = 15.32s

# Step 2: 상위 함수 확인
(pprof) top 10

Showing nodes accounting for 8.95s out of 15.32s (58.4%)
      flat  flat%   sum%        cum   cum%
    2.85s  18.6%  18.6%     2.85s  18.6%  crypto/sha256.Sum256
    1.95s  12.7%  31.3%     1.95s  12.7%  sync.(*Mutex).Lock
    1.40s   9.1%  40.4%     1.40s   9.1%  runtime.mallocgc
    1.20s   7.8%  48.2%     1.20s   7.8%  strconv.ParseInt
    0.98s   6.4%  54.6%     0.98s   6.4%  encoding/json.Unmarshal
    0.85s   5.5%  60.2%     0.85s   5.5%  reflect.deepEqual
    0.76s   5.0%  65.1%     0.76s   5.0%  time.Sleep (테스트용)
    0.65s   4.2%  69.3%     0.65s   4.2%  net.(*Conn).Read
    0.52s   3.4%  72.7%     0.52s   3.4%  bytes.(*Buffer).Write
    0.45s   2.9%  75.7%     0.45s   2.9%  os.readFile

# Step 3: 특정 함수 확인
(pprof) list hash

ROUTINE ======================== crypto/sha256.Sum256 in crypto/sha256
    2.85s     2.85s (flat, cum) 18.6% of Total
         .          .     45: func Sum256(data []byte) [32]byte {
    2.85s     2.85s     46:     h := New()
         .          .     47:     h.Write(data)
         .          .     48:     h.Sum(nil)
         .          .     49: }

원인 발견! SHA-256 해시가 전체 시간의 18.6% 차지
↓
FNV-1a로 교체하면 10배 빨라짐!
```

### Case 2: 메모리 누수

```bash
# Step 1: 메모리 프로파일 수집
$ go tool pprof http://localhost:6060/debug/pprof/heap

Fetching profile over HTTP from http://localhost:6060/debug/pprof/heap
File: server
Type: inuse_space
Time: Mar 27 2026 09:20:15 UTC
Showing nodes accounting for 1.2GB out of 1.8GB (67%)

# Step 2: 상위 할당자 확인
(pprof) top 5

Showing nodes accounting for 809.60MB out of 1228.53MB (65.9%)
      flat  flat%   sum%        cum   cum%
   524.30MB 42.7%  42.7%   524.30MB 42.7%  time.After
   185.40MB 15.1%  57.8%   185.40MB 15.1%  runtime.newTimerObj
   85.23MB  6.9%  64.8%    85.23MB  6.9%   encoding/json.Unmarshal
   10.50MB  0.9%  65.6%    10.50MB  0.9%   bytes.makeSlice
    5.20MB  0.4%  66.1%     5.20MB  0.4%   net.Pipe

# Step 3: time.After 호출 지점 확인
(pprof) list main.Client.CallWithTimeout

원인 발견! time.After가 메모리 누수 (524MB!)
↓
time.NewTimer + defer Stop로 교체하면 95% 절감
```

### Case 3: Goroutine 누수

```bash
# Step 1: 고루틴 프로파일
$ go tool pprof http://localhost:6060/debug/pprof/goroutine

Fetching profile over HTTP from http://localhost:6060/debug/pprof/goroutine
File: server
Type: goroutine
Time: Mar 27 2026 09:22:45 UTC
Showing nodes accounting for 48523 goroutines

# Step 2: 스택 추적
(pprof) traces Database.Query

goroutine 1234: [stack trace]
  runtime/time.Sleep
  main.Database.Query.func1
  └─ 48,523개 고루틴이 모두 여기서 대기!

원인 발견! 채널 수신 대기 상태의 고루틴이 방치됨
↓
채널 정리 로직 추가 필요
```

---

## 2.3 Flamegraph 시각화

### 단계 1: Flamegraph 생성

```bash
# 프로파일 수집 (첫 번째 방법)
$ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# 자동으로 웹 브라우저 열림
# http://localhost:8081에서 인터랙티브 flamegraph 확인 가능
```

### 단계 2: Flamegraph 해석

```
높이: 콜스택 깊이 (위로 갈수록 깊음)
너비: 시간 점유율 (넓을수록 많이 소비)

예시:
┌─────────────────────────────────────────┐
│           main.main (루트)              │
├──────────────────┬──────────────────┬──┤
│   handler       │   database       │  │
│  (전체 40%)    │   (전체 50%)     │  │
├─────────────────┼──────────────────┤  │
│  sha256(18%) │ lock(8%) │ io(24%)│  │
└─────────────────┴──────────────────┴──┘

클릭으로 상세 확인:
- "sha256" 클릭 → SHA-256 함수의 모든 호출 지점 표시
- "lock" 클릭 → Lock contention 분석
```

---

## 3. 실제 사례: KVStore 성능 개선

### 3.1 Before 프로파일

```bash
$ go test ./pkg/kvstore -bench=BenchmarkCluster -cpuprofile=cpu.prof
$ go tool pprof cpu.prof

(pprof) list

Total: 4850ms (100%)

함수별 시간 분포:
1. crypto/sha256.Sum256:     2850ms (58.8%) ← SHA-256 문제!
2. sync.(*Mutex).Lock:        950ms (19.6%) ← Lock contention!
3. runtime.mallocgc:          560ms (11.5%) ← 메모리 할당!
4. encoding/json.Unmarshal:   340ms ( 7.0%)
5. Others:                     160ms ( 3.1%)

병목 명확:
1순위: SHA-256 → FNV-1a로 교체 (10배 개선)
2순위: Lock → sync.Pool로 메모리 재사용 (4배 개선)
3순위: 메모리 할당 최소화
```

### 3.2 After 프로파일

```bash
$ go tool pprof cpu.prof  # 최적화 후

(pprof) top

Total: 1200ms (100%)

함수별 시간 분포:
1. sync.(*RWMutex).RLock:     420ms (35%) ← 읽기 Lock (정상)
2. net.(*Conn).Read:          350ms (29%) ← 네트워크 (정상)
3. encoding/json.Unmarshal:   180ms (15%) ← JSON 파싱
4. strconv.ParseInt:          150ms (12%)
5. Others:                     100ms ( 8%)

개선:
- 전체 시간: 4850ms → 1200ms (4배 빨라짐!) ✅
- SHA-256 제거 ✅
- 다음 병목은 네트워크 (개선 어려움)

결론: "더 이상 개선은 diminishing returns"
```

---

## 4. pprof 명령어 레퍼런스

### 4.1 대화형 명령어

```bash
(pprof) help

Commands:
  top [N]           - 상위 N개 함수 (기본: 10)
  list <name>       - 함수 <name>의 소스코드 + 시간
  web               - 그래프를 DOT로 출력
  pdf               - PDF 생성
  disasm <name>     - <name>의 어셈블리코드
  callgrind <name>  - Callgrind 형식으로 출력
  go tool pprof -http=:8081 ... - 웹 UI 시작
```

### 4.2 자주 사용하는 패턴

```bash
# CPU 프로파일 + 웹 UI
$ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# 메모리 프로파일 (heap alloc)
$ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap

# 메모리 프로파일 (in-use)
$ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap?debug=0

# Goroutine 프로파일
$ go tool pprof http://localhost:6060/debug/pprof/goroutine

# Mutex contention
$ go tool pprof http://localhost:6060/debug/pprof/mutex

# 파일로 저장
$ go tool pprof -raw http://localhost:6060/debug/pprof/profile > cpu.pprof
$ go tool pprof cpu.pprof
```

---

## 5. pprof 통합: 지속적 성능 모니터링

### 5.1 자동화된 profiling

```go
// main.go
package main

import (
    "flag"
    "net/http"
    _ "net/http/pprof"
    "os"
)

func main() {
    cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
    memProfile := flag.String("memprofile", "", "write mem profile to file")
    flag.Parse()

    if *cpuProfile != "" {
        f, _ := os.Create(*cpuProfile)
        defer f.Close()
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

    // ... 앱 실행 ...

    if *memProfile != "" {
        f, _ := os.Create(*memProfile)
        defer f.Close()
        runtime.GC()
        pprof.WriteHeapProfile(f)
    }
}

// 사용:
// $ go run main.go -cpuprofile=cpu.pprof -memprofile=mem.pprof
```

### 5.2 CI/CD 통합

```yaml
# .github/workflows/benchmark.yml
name: Performance Benchmark

on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Run benchmarks
        run: |
          go test ./... -bench=. -benchmem -cpuprofile=cpu.pprof

      - name: Generate pprof report
        run: |
          go tool pprof -text cpu.pprof > benchmark-report.txt

      - name: Upload artifacts
        uses: actions/upload-artifact@v2
        with:
          name: benchmark-report
          path: benchmark-report.txt
```

---

## 6. 성능 모니터링 팁

### 6.1 실전 체크리스트

```
성능 저하 발생 시:

☐ pprof CPU 프로파일 수집 (30초)
  └─ go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

☐ Top 함수 확인
  └─ (pprof) top 10

☐ 상위 2-3개 함수 분석
  └─ (pprof) list <function>

☐ Flamegraph 시각화
  └─ go tool pprof -http=:8081 ...

☐ 메모리 프로파일도 수집
  └─ go tool pprof http://localhost:6060/debug/pprof/heap

☐ Goroutine 상태 확인
  └─ go tool pprof http://localhost:6060/debug/pprof/goroutine

☐ 결과 저장 및 비교
  └─ 최적화 전후 비교 (Before: 4850ms, After: 1200ms)
```

### 6.2 pprof 해석 가이드

```
flat:   이 함수에서 직접 소비한 시간
cum:    이 함수 + 호출한 함수들의 총 시간
flat%:  전체 대비 이 함수의 비율
cum%:   전체 대비 누적 비율

예:
  flat   flat%   cum   cum%
  2.85s  18.6%  2.85s 18.6%  crypto/sha256.Sum256

의미:
  - Sum256 함수: 2.85s 직접 소비 (다른 함수 호출 없음)
  - 전체의 18.6% 차지
  - 가장 우선순위 높은 병목
```

---

## 학습 요점

### 핵심 개념

| 도구 | 용도 |
|------|------|
| **CPU profile** | 어디서 시간을 쓰나? |
| **Heap profile** | 어디서 메모리를 할당하나? |
| **Goroutine profile** | 고루틴 누수는 없나? |
| **Flamegraph** | 콜스택을 시각화 |
| **pprof web UI** | 인터랙티브 분석 |

### 성능 개선 프로세스

1. **측정**: pprof로 정확한 병목 파악
2. **분석**: 상위 함수 확인
3. **개선**: 목표한 함수 최적화
4. **검증**: 벤치마크로 개선 확인
5. **반복**: 다음 병목으로

---

## 다음 글 추천

1. **"go test -bench으로 성능 테스트 자동화"**
   - 벤치마크 작성
   - 성능 회귀 감지

2. **"trace 분석: go tool trace로 동시성 문제 찾기"**
   - goroutine 스케줄링 분석
   - 블로킹 감지

---

**Made in Korea 🇰🇷**
**FreeLang Marketing Team**
