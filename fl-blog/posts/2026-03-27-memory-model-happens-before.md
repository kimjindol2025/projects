---
title: "Go 메모리 모델: Happens-Before 관계로 배우는 동시성의 안전성"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["performance", "concurrency"]
toc: true
comments: true
---

# Go 메모리 모델: Happens-Before 관계로 배우는 동시성의 안전성
## 요약

Go에서 동시성 버그는 종종 메모리 가시성(memory visibility) 때문에 발생합니다.
한 고루틴의 쓰기가 다른 고루틴에서 언제 보이는가?
이 글에서는 **Happens-Before 관계**를 통해 Go 메모리 모델의 핵심을 배우고,
실제 코드에서 데이터 레이스를 예방하는 방법을 보여줍니다.

**배울 것**:
- 메모리 배리어(Memory Barrier)의 역할
- Happens-Before 관계 5가지 규칙
- sync.Mutex, sync.Once, sync.WaitGroup의 보장
- Channel 통신의 동기화 의미
- 실제 버그 패턴과 수정 방법

---

## 1. 문제: 데이터 레이스의 미묘함

다음 코드를 실행해보세요:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	var x, y int

	go func() {
		x = 1      // 쓰기 1
		fmt.Print(y) // 읽기 Y
	}()

	go func() {
		y = 1      // 쓰기 2
		fmt.Print(x) // 읽기 X
	}()

	time.Sleep(100 * time.Millisecond)
}
```

이 코드의 출력은?
- `"00"` (둘 다 0)
- `"01"` (x=0, y=1)
- `"10"` (x=1, y=0)
- `"11"` (둘 다 1)

**모두 가능합니다!** 🤯

CPU 캐시와 메모리 쓰기 순서 때문에 고루틴이 다른 고루틴의 변경을 바로 보지 못할 수 있습니다.
이것이 **Happens-Before 관계**가 중요한 이유입니다.

---

## 2. 핵심 개념: Happens-Before 관계란?

**Happens-Before**: 메모리 연산의 **관찰 가능한 순서**를 정의합니다.

```
┌─────────────┐         ┌─────────────┐
│ Goroutine A │         │ Goroutine B │
│             │         │             │
│ write X=1 ──────sync────→ read X    │
│ (Happens)   │         │ (Before)    │
└─────────────┘         └─────────────┘
```

**동기화 없으면**: 연산 순서 보장 없음
**동기화 있으면**: 특정 순서 보장

Go에서 동기화 메커니즘:
1. Mutex/RWMutex
2. Channel send/receive
3. sync.Once, sync.WaitGroup
4. atomic 연산

---

## 3. Go 메모리 모델 5가지 규칙

### 규칙 1: 단일 고루틴 내 프로그램 순서

```go
func singleGoroutine() {
	x := 1
	y := x + 1  // y=2 (반드시 x=1 이후)
	fmt.Println(y)
}
```

**보장**: 단일 고루틴 내에서는 `go tool` 내 읽기/쓰기 순서가 보장됩니다.

### 규칙 2: Mutex 잠금/해제

```go
var mu sync.Mutex
var x int

// Goroutine A
mu.Lock()
x = 1
mu.Unlock()

// Goroutine B
mu.Lock()
fmt.Println(x)  // 반드시 1 출력
mu.Unlock()
```

**보장**: `Unlock()` 이전의 모든 쓰기가 이후의 `Lock()` 이후에 보입니다.

**실제 코드로 검증**:

```go
package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestMutexMemoryModel(t *testing.T) {
	var mu sync.Mutex
	var x int
	done := make(chan bool)

	go func() {
		mu.Lock()
		x = 1
		mu.Unlock()
		done <- true
	}()

	go func() {
		<-done
		mu.Lock()
		if x != 1 {
			t.Fatalf("expected x=1, got x=%d", x)
		}
		mu.Unlock()
	}()

	<-done
}
```

### 규칙 3: Channel Send/Receive

```go
var (
	c = make(chan int)
	x int
)

// Goroutine A
go func() {
	x = 1
	c <- 1    // Send
}()

// Goroutine B
<-c        // Receive (반드시 x=1이 보임)
fmt.Println(x)  // 반드시 1
```

**보장 순서**:
- Send 전의 쓰기 → Send 연산
- Receive 연산 → Receive 후의 읽기

### 규칙 4: Close Channel

```go
var (
	c = make(chan int)
	x int
)

// Goroutine A
go func() {
	x = 1
	close(c)  // Channel 닫기
}()

// Goroutine B
<-c        // Receive on closed channel
fmt.Println(x)  // 반드시 1
```

**Close 전의 모든 쓰기**는 Close 이후의 Receive에 보입니다.

### 규칙 5: WaitGroup.Done()

```go
var wg sync.WaitGroup
var x int

wg.Add(1)
go func() {
	x = 1
	wg.Done()  // 완료 신호
}()

wg.Wait()   // Done() 이전의 쓰기가 보임
fmt.Println(x)  // 반드시 1
```

---

## 4. 실전: sync.Once의 메모리 보장

`sync.Once`는 단 한 번만 실행하는 함수를 보장합니다.
내부적으로 이것이 어떻게 메모리 안전성을 보장할까요?

```go
type Once struct {
	m    sync.Mutex
	done uint32
}

func (o *Once) Do(f func()) {
	if atomic.LoadUint32(&o.done) == 0 {  // 빠른 읽기
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}
```

**핵심**:
1. 처음: Mutex로 보호된 느린 경로 (`doSlow`)
2. 나중: Atomic Load로 빠른 경로 (잠금 없음)
3. **Happens-Before**: `StoreUint32`는 다른 고루틴의 `LoadUint32` 이전에 발생

**실제 테스트**:

```go
func TestOnceMemoryModel(t *testing.T) {
	var once sync.Once
	var x int
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			once.Do(func() {
				x = 1
			})
			if x != 1 {
				t.Errorf("x should be 1, got %d", x)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
```

---

## 5. 실제 버그 패턴: 메모리 가시성 문제

### 패턴 1: Unsafe 플래그

❌ **잘못된 코드**:

```go
package main

import (
	"fmt"
	"time"
)

var ready bool
var msg string

func main() {
	go func() {
		msg = "Hello"
		ready = true  // 다른 고루틴이 이를 언제 볼까?
	}()

	for !ready {
		// Busy loop - ready=true를 볼 때까지 대기
		// 하지만 컴파일러가 루프를 최적화해서 무한 루프가 될 수도!
	}

	fmt.Println(msg)
}
```

**문제**:
- CPU 캐시 때문에 `ready=true`를 보지 못할 수 있음
- 컴파일러가 `ready` 읽기를 루프 밖으로 최적화할 수 있음

✅ **올바른 코드**:

```go
var ready bool
var msg string
var mu sync.Mutex

go func() {
	mu.Lock()
	msg = "Hello"
	ready = true
	mu.Unlock()
}()

mu.Lock()
for !ready {
	mu.Unlock()
	time.Sleep(1 * time.Millisecond)  // Busy loop 피함
	mu.Lock()
}
mu.Unlock()

fmt.Println(msg)
```

또는 **Channel 사용** (더 나음):

```go
done := make(chan string)

go func() {
	done <- "Hello"
}()

msg := <-done
fmt.Println(msg)
```

### 패턴 2: Publish-Subscribe Race

❌ **잘못된 코드**:

```go
type Cache struct {
	data map[string]string
}

func (c *Cache) Set(key, value string) {
	c.data[key] = value  // 잠금 없음!
}

func (c *Cache) Get(key string) string {
	return c.data[key]    // 데이터 레이스!
}
```

Set과 Get이 동시에 실행되면 패닉이 발생할 수 있습니다.

✅ **올바른 코드**:

```go
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Cache) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}
```

**검증 테스트**:

```go
func TestCacheMemorySafety(t *testing.T) {
	cache := &Cache{data: make(map[string]string)}
	done := make(chan bool, 100)

	for i := 0; i < 50; i++ {
		go func(idx int) {
			cache.Set(fmt.Sprintf("key%d", idx), fmt.Sprintf("val%d", idx))
			done <- true
		}(i)
	}

	for i := 0; i < 50; i++ {
		go func(idx int) {
			_ = cache.Get(fmt.Sprintf("key%d", idx))
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
```

---

## 6. Atomic 연산과 Happens-Before

Atomic 연산도 메모리 보장을 제공합니다.

```go
var x int32
var flag int32

// Goroutine A
atomic.StoreInt32(&x, 1)     // Write barrier 포함
atomic.StoreInt32(&flag, 1)  // x의 쓰기가 먼저 보임

// Goroutine B
if atomic.LoadInt32(&flag) == 1 {  // Read barrier 포함
	fmt.Println(atomic.LoadInt32(&x))  // 반드시 1
}
```

**Atomic Load/Store는**:
- `Load`: Acquire semantic (이후 연산 이전에 발생)
- `Store`: Release semantic (이전 연산 이후에 발생)

**실제 코드**:

```go
func TestAtomicMemoryModel(t *testing.T) {
	var x int32
	var flag int32
	const iterations = 10000

	for iter := 0; iter < iterations; iter++ {
		x = 0
		flag = 0
		done := make(chan bool, 2)

		go func() {
			atomic.StoreInt32(&x, 1)
			atomic.StoreInt32(&flag, 1)
			done <- true
		}()

		go func() {
			for atomic.LoadInt32(&flag) == 0 {
				// Spin until flag=1
			}
			if atomic.LoadInt32(&x) != 1 {
				t.Fatalf("iteration %d: x should be 1", iter)
			}
			done <- true
		}()

		<-done
		<-done
	}
}
```

---

## 7. 성능 임팩트: Mutex vs Atomic vs Lock-Free

각 메커니즘의 성능을 비교해봅시다:

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 1. Mutex 기반
type MutexCounter struct {
	mu    sync.Mutex
	value int64
}

func (c *MutexCounter) Inc() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

// 2. Atomic 기반
type AtomicCounter struct {
	value int64
}

func (c *AtomicCounter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// 3. Lock-Free (TAS - Test And Set)
type LockFreeCounter struct {
	value int64
	lock  int32
}

func (c *LockFreeCounter) Inc() {
	for {
		if atomic.CompareAndSwapInt32(&c.lock, 0, 1) {
			c.value++
			atomic.StoreInt32(&c.lock, 0)
			break
		}
	}
}

func BenchmarkMutexCounter(b *testing.B) {
	c := &MutexCounter{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
	b.ReportMetric(float64(c.value), "ops")
}

func BenchmarkAtomicCounter(b *testing.B) {
	c := &AtomicCounter{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
	b.ReportMetric(float64(atomic.LoadInt64(&c.value)), "ops")
}

func BenchmarkLockFreeCounter(b *testing.B) {
	c := &LockFreeCounter{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
	b.ReportMetric(float64(c.value), "ops")
}

func TestCounterCorrectness(t *testing.T) {
	const workers = 100
	const opsPerWorker = 1000

	tests := []struct {
		name string
		fn   func() int64
	}{
		{
			name: "Mutex",
			fn: func() int64 {
				c := &MutexCounter{}
				var wg sync.WaitGroup
				for i := 0; i < workers; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for j := 0; j < opsPerWorker; j++ {
							c.Inc()
						}
					}()
				}
				wg.Wait()
				return c.value
			},
		},
		{
			name: "Atomic",
			fn: func() int64 {
				c := &AtomicCounter{}
				var wg sync.WaitGroup
				for i := 0; i < workers; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for j := 0; j < opsPerWorker; j++ {
							c.Inc()
						}
					}()
				}
				wg.Wait()
				return atomic.LoadInt64(&c.value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := int64(workers * opsPerWorker)
			got := tt.fn()
			if got != expected {
				t.Fatalf("expected %d, got %d", expected, got)
			}
		})
	}
}

func main() {
	fmt.Println("메모리 모델 성능 벤치마크:")
	fmt.Println("(go test -bench=. -benchmem 으로 실행)")
}
```

**벤치마크 결과 (예상)**:
```
BenchmarkMutexCounter-8         850K     1,200 ns/op    (Contention 높음)
BenchmarkAtomicCounter-8      1.2M       950 ns/op     (Lock-free, 더 나음)
BenchmarkLockFreeCounter-8    1.0M     1,100 ns/op     (수동 구현은 느림)
```

---

## 8. 실무 가이드: 언제 뭘 쓸까?

| 상황 | 추천 | 이유 |
|------|------|------|
| **여러 필드 보호** | Mutex | 원자성 필요 |
| **단일 정수 카운터** | atomic | 더 빠름 |
| **Key-Value 맵** | sync.Map 또는 RWMutex | 복잡도 vs 성능 |
| **일회성 초기화** | sync.Once | 최적화됨 |
| **고루틴 대기** | sync.WaitGroup | 직관적 |
| **순서 보장 필요** | Channel | 깔끔한 코드 |

**실제 예시**:

```go
// ❌ 잘못된 선택
type Service struct {
	count int64  // 잠금 없이 사용 - 데이터 레이스!
}

// ✅ 좋은 선택
type Service struct {
	count int64  // atomic 사용
}

func (s *Service) Increment() {
	atomic.AddInt64(&s.count, 1)
}
```

---

## 9. 핵심 정리

1. **메모리 가시성**: CPU 캐시 때문에 다른 고루틴의 변경이 즉시 보이지 않을 수 있음
2. **Happens-Before**: 메모리 연산의 **관찰 가능한 순서**를 정의하는 메커니즘
3. **동기화 메커니즘**:
   - Mutex/RWMutex: 여러 필드 보호
   - Channel: 순서와 값 전달
   - sync.Once/WaitGroup: 특정 패턴
   - Atomic: 단일 정수/포인터

4. **검증**: `go test -race`로 데이터 레이스 감지

---

## 10. 다음 읽을 거리

- [Go Memory Model 공식 문서](https://golang.org/ref/mem)
- [Java Memory Model](https://docs.oracle.com/javase/specs/jls/se11/html/jls-17.html) (유사한 개념)
- [x86-64 Memory Ordering](https://www.1024cores.net/) (하드웨어 수준)

---

## 11. 피드백

이 글로 도움이 되었나요?
- 더 깊이 다루고 싶은 주제?
- 코드 예시가 부족한가?
- 실제 프로젝트에서 경험한 메모리 버그?

댓글로 남겨주세요! 🚀

---

**만든이**: FreeLang 마케팅 팀
**기술 검수**: Go 런타임 메모리 모델 + go test -race
**최종 수정**: 2026-03-27

---

## 벤치마크 실행 방법

```bash
# 현재 디렉토리에 memory_model_test.go 저장

# 모든 테스트 실행 (데이터 레이스 감지)
go test -v -race

# 벤치마크 실행
go test -bench=. -benchmem -benchtime=1s

# 프로파일 생성
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```
