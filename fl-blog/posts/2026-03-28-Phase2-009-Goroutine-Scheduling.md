---
layout: post
title: Phase2-009-Goroutine-Scheduling
date: 2026-03-28
---
# Go 런타임 스케줄링: 100만 고루틴을 관리하는 방법

## 요약

Go는 "병렬 프로그래밍을 쉽게" 한다고 광고합니다.
하지만 고루틴이 어떻게 OS 스레드 위에서 스케줄되는지 이해하지 못하면,
성능 최적화나 버그 디버깅에서 막히게 됩니다.

이 글에서는 **M:N 스케줄러**의 아키텍처, 작업 도둑질(Work Stealing),
컨텍스트 스위칭의 비용을 배웁니다.
그리고 100만 고루틴을 만들 때 어떤 일이 일어나는지 봅시다.

**배울 것**:
- M:N 스케줄러의 핵심 (Goroutine, Machine, Processor)
- Work Stealing으로 로드 밸런싱
- 컨텍스트 스위칭과 성능
- 고루틴 누수와 메모리 프로파일링
- 실제 성능 튜닝 예제

---

## 1. 문제: 고루틴은 정말 "가볍다"?

다음 코드를 실행해보세요:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()

	for i := 0; i < 1_000_000; i++ {
		go func(id int) {
			time.Sleep(1 * time.Second)
		}(i)
	}

	// 모든 고루틴이 완료될 때까지 대기
	time.Sleep(2 * time.Second)

	elapsed := time.Since(start)
	fmt.Printf("100만 고루틴 생성: %v\n", elapsed)
	fmt.Println("메모리 사용량은?")
}
```

**예상 결과**:
- 생성 시간: ~1초
- 메모리: ~3-4GB (고루틴당 ~2-4KB)

**Why?** 고루틴은 OS 스레드가 아니라 **User-level 태스크**입니다.

---

## 2. Go 스케줄러 아키텍처

```
┌─────────────────────────────────────────────────────┐
│              Go Scheduler                           │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐      │
│  │ P0     │ │ P1     │ │ P2     │ │ P3     │      │  Logical Processors
│  │ ┌────┐ │ │ ┌────┐ │ │ ┌────┐ │ │ ┌────┐ │      │  (GOMAXPROCS 개)
│  │ │ G1 │ │ │ │ G5 │ │ │ │ G7 │ │ │ │    │ │      │
│  │ │ G2 │ │ │ │ G6 │ │ │ │ G8 │ │ │ │    │ │      │  Local Run Queue
│  │ └────┘ │ │ └────┘ │ │ └────┘ │ │ │    │ │      │  (per processor)
│  │    ↓   │ │    ↓   │ │    ↓   │ │ └────┘ │      │
│  │   M0   │ │   M1   │ │   M2   │ │   M3   │      │  OS Threads
│  └────────┘ └────────┘ └────────┘ └────────┘      │
│       ↓          ↓          ↓          ↓            │
└───────────────────────────────────────────────────┬─┘
                                                    │
                    ┌───────────────────────────────┘
                    │
              ┌─────▼──────┐
              │ OS Kernel  │
              │ Scheduler  │
              └────────────┘
```

**핵심 개념**:

| 용어 | 뜻 | 개수 |
|------|------|------|
| **G** (Goroutine) | 사용자 작업 | 백만 개 가능 |
| **M** (Machine) | OS 스레드 | GOMAXPROCS (기본: CPU 코어 수) |
| **P** (Processor) | 논리 프로세서 | GOMAXPROCS 개 |

---

## 3. 스케줄링 과정

### 단계 1: 고루틴 생성

```go
go myFunc()  // G 객체 생성
```

**실제로는**:
1. 새 G 객체 할당 (메모리 풀에서)
2. 현재 P의 **Local Run Queue**에 추가
3. 스케줄러가 나중에 실행

```
┌──────────────┐
│ Current P0   │
│ Local Queue: │
│  [G1, G2, G3] ← 새로운 G 추가
└──────────────┘
```

### 단계 2: Work Stealing (작업 도둑질)

P0의 로컬 큐가 비어있고, P1의 큐가 가득 차 있으면?

```
┌──────────────┐  "나 일해! 😩"  ┌──────────────┐
│ Current P0   │◄────────────────│ Current P1   │
│ Local Queue: │  Work Stealing  │ Local Queue: │
│  []          │   (P0이 P1의   │  [G10, G11, │
│              │  작업 절반 가져감) │   G12, ...]  │
└──────────────┘                 └──────────────┘
                ┌──────────────┐
                │ Stole work:  │
                │ [G10, G11]   │
                └──────────────┘
```

**구현**: P가 자신의 큐 일반 절반을 다른 P의 큐에서 가져옵니다.

### 단계 3: 실행 및 컨텍스트 스위칭

```go
// P0는 M0(OS 스레드)에서 실행 중
P0: M0 실행 중 → G1 실행 → G1 일시 중단 → G2 실행 → ...
```

**컨텍스트 스위칭은 언제?**
- 고루틴이 `<-c` (채널 대기)
- 고루틴이 `time.Sleep()`
- 고루틴이 `sync.Mutex.Lock()` (잠금 대기)
- 고루틴이 함수 종료

**중요**: OS 스레드 스위칭이 아니라 **고루틴 스위칭**입니다!

---

## 4. 동기화 호출과 성능

고루틴이 블로킹되면 어떻게 될까?

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func blockingIO() {
	// net.Dial, os.Open, time.Sleep 등
	// 이것은 OS 레벨 블로킹 호출
	time.Sleep(100 * time.Millisecond)
}

func main() {
	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			blockingIO()  // ⚠️ 여기서 뭐가 일어날까?
		}()
	}

	wg.Wait()
	fmt.Printf("소요 시간: %v\n", time.Since(start))
}
```

**실행 흐름**:

```
Time 0ms:   P0 실행, G1~100 생성, G1이 sleep() 호출
            → M0는 G1과 함께 대기 (블로킹!)

            ⚠️ 문제: M0이 블로킹되면, P0에 다른 작업을 할 수 없나?

Time 0ms:   Go 스케줄러가 감지: M0가 블로킹됨
            → 새 OS 스레드 M1 생성
            → P0을 M1에 연결
            → M1이 P0의 큐에서 G101 실행

Time 100ms: M0이 깨어남 (sleep 완료)
            → G1 실행 완료, P0의 큐에서 다음 작업 가져옴
```

**결과**:
- 순차 처리: 100s (각 고루틴 100ms × 1000개)
- Go 스케줄러: ~100ms (모든 고루틴이 병렬 대기)

---

## 5. 컨텍스트 스위칭 비용 분석

**User-level 스위칭** (고루틴 ↔ 고루틴):
- CPU 비용: ~100 ns (매우 빠름)
- 캐시 영향: 거의 없음

**OS-level 스위칭** (스레드 ↔ 스레드):
- CPU 비용: ~1-10 µs (1,000-10,000 ns!)
- 캐시 영향: 심각 (전체 L1/L2 캐시 무효화)

**예제로 확인**:

```go
package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 1. 고루틴 컨텍스트 스위칭
func BenchmarkGoroutineSwitch(b *testing.B) {
	const workers = 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan int)
		var wg sync.WaitGroup

		// 수신자
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ch {
				// CPU 작업 없음
			}
		}()

		// 발신자
		for j := 0; j < 1000; j++ {
			ch <- j  // 고루틴 스위칭
		}
		close(ch)
		wg.Wait()
	}
}

// 2. OS 스레드 스위칭
func BenchmarkThreadSwitch(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		// 2개 스레드를 컨텍스트 스위칭
		for t := 0; t < 2; t++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// OS 시스템 호출 (컨텍스트 스위칭 강제)
				time.Sleep(0)  // 스케줄러에게 양보
			}()
		}
		wg.Wait()
	}
}

func TestSwitchingCosts(t *testing.T) {
	// 단순 벤치마크
	start := time.Now()
	for i := 0; i < 10000; i++ {
		ch := make(chan int)
		go func() {
			for range ch {
			}
		}()
		ch <- 1
		close(ch)
	}
	fmt.Printf("10k 고루틴 스위칭: %v\n", time.Since(start))
}

func main() {
	// 예상:
	// - 고루틴 스위칭: ~10ms
	// - OS 스레드 스위칭: ~50ms
	// 비율: 5배 차이
}
```

---

## 6. 실전: 고루틴 누수 감지

고루틴이 끝나지 않으면?

```go
❌ 잘못된 코드:

func leakyFunction(ch chan int) {
	go func() {
		// ch를 절대 받지 않는 고루틴
		for {
			time.Sleep(1 * time.Second)
		}
	}()
}

func main() {
	for i := 0; i < 10000; i++ {
		leakyFunction(make(chan int))
	}

	// 10,000개 고루틴이 메모리에 남아있음!
	time.Sleep(100 * time.Second)
}
```

**메모리 누수 감지**:

```go
import "runtime"

func TestGoroutineLeaks(t *testing.T) {
	startGorutines := runtime.NumGoroutine()
	fmt.Printf("시작 고루틴: %d\n", startGorutines)

	// 테스트 코드 실행
	for i := 0; i < 1000; i++ {
		leakyFunction(make(chan int))
	}

	time.Sleep(1 * time.Second)

	endGoroutines := runtime.NumGoroutine()
	fmt.Printf("종료 고루틴: %d\n", endGoroutines)

	if endGoroutines > startGorutines+100 {
		t.Errorf("고루틴 누수 의심: %d개 증가",
			endGoroutines-startGorutines)
	}
}
```

**pprof로 상세 분석**:

```bash
# 고루틴 프로파일 생성
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 상위 누수자 확인
(pprof) top
```

---

## 7. GOMAXPROCS와 성능

`GOMAXPROCS`는 동시에 실행할 P(논리 프로세서) 개수를 제한합니다.

```go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func cpuBoundWork(duration time.Duration) int {
	count := 0
	end := time.Now().Add(duration)
	for time.Now().Before(end) {
		count++
	}
	return count
}

func BenchmarkGOMAXPROCS() {
	tests := []int{1, 2, 4, 8}

	for _, procs := range tests {
		runtime.GOMAXPROCS(procs)

		start := time.Now()
		var wg sync.WaitGroup

		// CPU-bound 작업 실행
		for i := 0; i < procs; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cpuBoundWork(100 * time.Millisecond)
			}()
		}

		wg.Wait()
		elapsed := time.Since(start)

		fmt.Printf("GOMAXPROCS=%d: %v\n", procs, elapsed)
	}
}

func main() {
	BenchmarkGOMAXPROCS()

	// 예상 결과:
	// GOMAXPROCS=1: ~400ms (순차)
	// GOMAXPROCS=2: ~200ms (병렬 2배)
	// GOMAXPROCS=4: ~100ms (병렬 4배)
	// GOMAXPROCS=8: ~100ms (프로세서 부족)
}
```

---

## 8. 성능 튜닝: 채널 vs 뮤텍스

고루틴 간 통신은 채널이 항상 빠를까?

```go
package main

import (
	"fmt"
	"sync"
	"testing"
)

// 1. 채널 기반 (동기화)
type ChannelCounter struct {
	ch    chan int
	value int
}

func (c *ChannelCounter) Increment() {
	c.ch <- 1
}

// 2. 뮤텍스 기반
type MutexCounter struct {
	mu    sync.Mutex
	value int64
}

func (c *MutexCounter) Increment() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func BenchmarkChannelCounter(b *testing.B) {
	c := &ChannelCounter{
		ch: make(chan int, 100),
	}

	go func() {
		for range c.ch {
			c.value++
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Increment()
	}
	close(c.ch)
}

func BenchmarkMutexCounter(b *testing.B) {
	c := &MutexCounter{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Increment()
		}
	})
}

func main() {
	fmt.Println("채널 vs 뮤텍스 성능:")
	fmt.Println("(go test -bench=. -benchmem)")
	// 예상: Mutex가 채널보다 10-100배 빠름 (contention 많을 때)
}
```

**결과 분석**:
- **Low contention**: 채널 빠름 (고루틴 스위칭 최소)
- **High contention**: 뮤텍스 빠름 (채널 오버헤드 크다)

---

## 9. 최적화 기법

### 기법 1: 고루틴 풀 (Worker Pool)

❌ **비효율적**:
```go
for i := 0; i < 10_000_000; i++ {
	go doWork(item)  // 1000만 고루틴 생성!
}
```

✅ **효율적**:
```go
const workers = runtime.NumCPU()
jobs := make(chan Work, 100)
results := make(chan Result, 100)

for i := 0; i < workers; i++ {
	go worker(jobs, results)  // 재사용 가능
}

for _, item := range items {
	jobs <- Work{item}
}
```

### 기법 2: 배치 처리

```go
// 개별 처리
for item := range items {
	process(item)  // 고루틴 오버헤드
}

// 배치 처리
for i := 0; i < len(items); i += batchSize {
	batch := items[i:min(i+batchSize, len(items))]
	go processBatch(batch)  // 훨씬 적은 고루틴
}
```

### 기법 3: 제한된 동시성

```go
sem := make(chan struct{}, maxConcurrency)

for item := range items {
	sem <- struct{}{}  // 토큰 획득
	go func(item Item) {
		defer func() { <-sem }()  // 토큰 반환
		process(item)
	}(item)
}
```

---

## 10. 메모리 프로파일링

고루틴과 메모리의 관계는?

```go
import (
	"fmt"
	"runtime"
	"time"
)

func analyzeMemory() {
	var m runtime.MemStats

	// 측정 전
	runtime.ReadMemStats(&m)
	fmt.Printf("Goroutines: %d, Memory: %v MB\n",
		runtime.NumGoroutine(), m.Alloc/1024/1024)

	// 고루틴 생성
	for i := 0; i < 100_000; i++ {
		go func() {
			time.Sleep(10 * time.Second)
		}()
	}

	// 측정 후
	runtime.ReadMemStats(&m)
	fmt.Printf("Goroutines: %d, Memory: %v MB\n",
		runtime.NumGoroutine(), m.Alloc/1024/1024)

	// 예상: 100만 고루틴 ≈ 500MB-1GB
}
```

---

## 11. 핵심 정리

1. **M:N 스케줄러**: 백만 고루틴을 수십 개 OS 스레드 위에서 관리
2. **Work Stealing**: 로드 불균형 해결
3. **컨텍스트 스위칭**: 고루틴은 ~100ns, 스레드는 ~1µs
4. **동기화 호출**: OS 블로킹은 자동으로 새 스레드 생성
5. **메모리**: 고루틴당 ~2-4KB (100만 개 = ~2-4GB)
6. **최적화**: 워커 풀, 배치, 세마포어 사용

---

## 12. 다음 읽을 거리

- [Go 스케줄러 설계](https://golang.org/doc/go-faq.html)
- [Work Stealing 논문](https://en.wikipedia.org/wiki/Work_stealing)
- [GOMAXPROCS 가이드](https://github.com/golang/go/wiki/CodeReviewComments#goroutine-lifetimes)

---

## 13. 벤치마크 실행

```bash
go test -bench=. -benchmem -benchtime=5s
```

---

**만든이**: FreeLang 마케팅 팀
**기술 검수**: Go 1.21+ 런타임
**최종 수정**: 2026-03-27

---

## 추가: 100만 고루틴 성능 테스트

```go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func millionGoroutines() {
	var m runtime.MemStats

	start := time.Now()
	runtime.ReadMemStats(&m)

	var wg sync.WaitGroup
	for i := 0; i < 1_000_000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)
	runtime.ReadMemStats(&m)

	fmt.Printf("시간: %v, 메모리: %v MB\n",
		elapsed, m.Alloc/1024/1024)
}

func main() {
	millionGoroutines()
	// 예상:
	// 시간: ~100ms
	// 메모리: ~500MB-1GB
}
```
