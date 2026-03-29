---
layout: post
title: Phase4-035-Cache-Line-Optimization
date: 2026-03-28
---
# CPU 캐시 라인 최적화: L1에서 메모리까지 1000배 차이

## 요약

- CPU 캐시 계층 구조 (L1/L2/L3)
- 캐시 라인 (64 bytes) 개념
- False Sharing 문제와 해결책
- 실전 최적화 (배열 레이아웃, 구조체 정렬)
- 성능 벤치마크

---

## 1. CPU 메모리 계층

### 접근 시간과 용량

```
Tier     용량    지연시간   대역폭      예측 성능
─────────────────────────────────────────────────
L1D      32KB    4ns       64GB/s      극빠름
L2       256KB   12ns      32GB/s      매우빠름
L3       8MB     40ns      16GB/s      빠름
메인     32GB    100ns     8GB/s       느림
```

### 지연시간 비교

```
L1 캐시 히트:  4ns        1 사이클
L2 캐시 히트:  12ns       3 사이클
L3 캐시 히트:  40ns       10 사이클
메인 메모리:   100ns      25+ 사이클

결론: 메모리 접근 = 25배 느림! (캐시 미스 시)
```

### CPU 아키텍처 (Skylake 예)

```
CPU 코어 (3.8 GHz)
├─ L1I 캐시 (32KB, instruction)
├─ L1D 캐시 (32KB, data)
│  ├─ 4-way associative
│  └─ 64-byte 라인
│
├─ L2 캐시 (256KB, private per core)
│  ├─ 8-way associative
│  └─ 64-byte 라인
│
└─ L3 캐시 (8MB, shared)
   ├─ 20-way associative
   └─ 64-byte 라인

메인 메모리 (DDR4, 3200MHz)
```

---

## 2. 캐시 라인과 False Sharing

### 캐시 라인이란?

```
메모리의 최소 로드 단위 = 64 bytes

주소 0:     [================================] 64 bytes = 1 라인
주소 64:    [================================] 64 bytes = 1 라인
주소 128:   [================================] 64 bytes = 1 라인

접근: addr[0] → 라인 전체 (0~63) 로드
접근: addr[10] → 이미 캐시에 있음 (히트)
접근: addr[65] → 새 라인 로드 (미스)
```

### False Sharing 문제

```go
// ❌ 나쁜 예: 두 스레드가 같은 캐시 라인 접근
type Counter struct {
    A int64  // offset 0
    B int64  // offset 8 (같은 라인!)
}

// 메모리: [A     B     ....] (한 라인)

// Thread 1: A++
// Thread 2: B++

메커니즘:
1. Thread 1이 라인을 수정
2. Thread 2의 L1 캐시 무효화 (Invalidate)
3. Thread 2가 라인 다시 로드
4. Thread 1의 L1 캐시 무효화
...
계속 캐시 라인 왕복 → "False Sharing"
```

### 성능 영향

```
Single threaded: A++ 만 실행
시간: 100ms

Two threads (shared cache line):
├─ 캐시 라인 공유
├─ 계속 무효화/재로드
└─ 시간: 5000ms (50배 느림!)

경합 수준:
├─ 2 스레드, 같은 라인: 50배 느림
├─ 8 스레드, 같은 라인: 200배 느림
└─ 32 스레드, 같은 라인: 1000배 느림
```

### ✅ 해결책: Padding (패딩)

```go
// ✅ 좋은 예: 캐시 라인 분리
type Counter struct {
    A int64
    _pad0 [56]byte  // 패딩 (56 = 64 - 8)
    B int64
    _pad1 [56]byte  // 패딩
}

// 메모리: [A + 56bytes] [B + 56bytes]
//         라인 1         라인 2

// Thread 1이 A 접근 → 라인 1 로드
// Thread 2가 B 접근 → 라인 2 로드
// 서로 다른 라인 → False Sharing 해결!
```

### Go 라이브러리 예

```go
// sync/atomic의 Uint64 구현
type Uint64 struct {
    noCopy noCopy
    pad    [8 - 1%8]uint64  // 64바이트 경계 정렬
    v      uint64
}

// sync.Pool의 비우기
type poolLocal struct {
    private interface{}
    shared  poolLocalInternal
    pad     [128 - unsafe.Sizeof(poolLocalInternal{})%128]byte
}
```

---

## 3. 배열 레이아웃 최적화

### SoA vs AoS

```
데이터: 100만 점 (x, y, z 좌표)

AoS (Array of Structures):
memory: [x1 y1 z1 x2 y2 z2 x3 y3 z3 ...]

조회: 모든 x 값만 필요
└─ 메모리 낭비 (y, z도 로드)
└─ 캐시 효율: 33% (1/3만 사용)

SoA (Structure of Arrays):
memory: [x1 x2 x3 ... x1000000][y1 y2 y3 ... y1000000][...]

조회: 모든 x 값만 필요
└─ 순차 접근 (prefetch 가능)
└─ 캐시 효율: 100% (필요한 것만)
```

### 구현 비교

```go
// ❌ AoS
type Point struct {
    X, Y, Z float32
}

func ProcessAoS(points []Point) {
    for i := range points {
        points[i].X *= 2  // Y, Z도 로드되지만 미사용
    }
}

// ✅ SoA
type Points struct {
    X []float32
    Y []float32
    Z []float32
}

func ProcessSoA(p Points) {
    for i := range p.X {
        p.X[i] *= 2  // X만 로드 (효율적)
    }
}
```

### 벤치마크

```
데이터: 100만 점

AoS (Sequential X 접근):
├─ 캐시 미스: 높음
├─ 처리량: 100ms

SoA (X-only):
├─ 캐시 미스: 낮음
├─ 처리량: 20ms (5배 향상)
```

---

## 4. 구조체 정렬

### 필드 순서의 중요성

```go
// ❌ 비효율적 정렬
type Bad struct {
    A bool       // 1 byte  (offset 0)
    B int64      // 8 bytes (offset 8, 패딩 7 bytes)
    C bool       // 1 byte  (offset 16)
    D int32      // 4 bytes (offset 20, 패딩 4 bytes)
}
// 메모리: [A(1) pad(7) | B(8) | C(1) pad(4) D(4)]
// 크기: 32 bytes

// ✅ 효율적 정렬 (큰 것부터)
type Good struct {
    B int64      // 8 bytes (offset 0)
    D int32      // 4 bytes (offset 8)
    A bool       // 1 byte  (offset 12)
    C bool       // 1 byte  (offset 13)
}
// 메모리: [B(8) | D(4) | A(1) C(1) pad(2)]
// 크기: 16 bytes (50% 절약!)
```

### 도구: Go의 fieldalignment

```bash
# 필드 정렬 권장 확인
go vet -structcheck ./...

# 또는
go build -v 2>&1 | grep "fieldalignment"
```

---

## 5. 캐시 친화적 알고리즘

### 예 1: 행렬 곱셈

```c
// ❌ 캐시 비친화적 (열 우선 접근)
void matmul_bad(float A[N][N], float B[N][N], float C[N][N]) {
    for (int i = 0; i < N; i++) {
        for (int j = 0; j < N; j++) {
            float sum = 0;
            for (int k = 0; k < N; k++) {
                sum += A[i][k] * B[k][j];  // B[k][j] = 불규칙 접근
            }
            C[i][j] = sum;
        }
    }
}
// 캐시 미스: 높음 (열 접근)

// ✅ 캐시 친화적 (행 우선 + 블로킹)
#define BLOCK_SIZE 64

void matmul_good(float A[N][N], float B[N][N], float C[N][N]) {
    // B 전치 (행 우선 접근)
    float B_T[N][N];
    for (int i = 0; i < N; i++) {
        for (int j = 0; j < N; j++) {
            B_T[i][j] = B[j][i];
        }
    }

    // 블록 단위 곱셈 (L3 캐시에 맞음)
    for (int bi = 0; bi < N; bi += BLOCK_SIZE) {
        for (int bj = 0; bj < N; bj += BLOCK_SIZE) {
            for (int bk = 0; bk < N; bk += BLOCK_SIZE) {
                for (int i = bi; i < bi + BLOCK_SIZE; i++) {
                    for (int j = bj; j < bj + BLOCK_SIZE; j++) {
                        float sum = C[i][j];
                        for (int k = bk; k < bk + BLOCK_SIZE; k++) {
                            sum += A[i][k] * B_T[j][k];
                        }
                        C[i][j] = sum;
                    }
                }
            }
        }
    }
}

성능:
일반: 1ms (200x200 행렬)
최적화: 0.1ms (10배 향상)
```

### 예 2: 링크드 리스트 vs 배열

```go
// ❌ 캐시 비친화적 (링크드 리스트)
type Node struct {
    Value int
    Next  *Node
}

list := &Node{1, &Node{2, &Node{3, ...}}}

for n := list; n != nil; n = n.Next {
    sum += n.Value  // 포인터 추종 = 캐시 미스
}
// 캐시 미스: 매우 높음 (매 노드마다)

// ✅ 캐시 친화적 (배열)
arr := []int{1, 2, 3, ...}

for i := 0; i < len(arr); i++ {
    sum += arr[i]  // 순차 접근 = 캐시 히트
}
// 캐시 미스: 거의 없음 (prefetch)

성능 (100만 요소):
링크드 리스트: 500ms
배열: 5ms (100배!)
```

---

## 6. Prefetching

### 하드웨어 Prefetcher

```
CPU가 자동으로 다음 데이터 미리 로드

규칙:
├─ 순차 접근 감지 → 자동 prefetch
├─ 스트라이드 패턴 감지 (매 N번째)
└─ 캐시 미스 예측 → 미리 로드

효과:
├─ 순차 접근: 거의 캐시 미스 없음
├─ 불규칙: 효과 미미
```

### 소프트웨어 Prefetch

```c
// 명시적 prefetch
#include <immintrin.h>

void process_array(int *arr, int n) {
    for (int i = 0; i < n; i++) {
        _mm_prefetch(&arr[i + 16], _MM_HINT_T0);  // 16칸 앞 미리 로드

        // 현재 i 처리
        process(arr[i]);
    }
}

효과:
└─ 불규칙 패턴에서 캐시 미스 30-50% 감소
```

---

## 7. 측정과 프로파일링

### perf로 캐시 분석

```bash
# 캐시 미스 측정
perf stat -e cache-references,cache-misses ./program

# 결과
Performance counter stats:
  1,234,567  cache-references  (캐시 접근)
     89,012  cache-misses      (미스율: ~7%)

# L3 캐시 미스 자세히
perf stat -e LLC-loads,LLC-load-misses ./program
```

### Go 벤치마크로 확인

```go
func BenchmarkSoA(b *testing.B) {
    points := Points{
        X: make([]float32, 1000000),
        Y: make([]float32, 1000000),
        Z: make([]float32, 1000000),
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        for j := range points.X {
            points.X[j] *= 2
        }
    }
}

// 결과
// BenchmarkSoA-8   1000    1005493 ns/op  (~1ms)
```

---

## 8. 실전 팁

### 1. Bandwidth 계산

```
Memory BW 계산:
전송 바이트 / 시간

예: 100MB 배열, 1초
└─ BW = 100 MB/s (매우 낮음)

예상 BW:
├─ L1: 64 GB/s
├─ L2: 32 GB/s
├─ L3: 16 GB/s
└─ 메모리: 8 GB/s

BW 낮음 → 캐시 최적화 여지 있음
```

### 2. 캐시 라인 크기 확인

```bash
# Linux에서 캐시 라인 크기 확인
cat /sys/devices/system/cpu/cpu0/cache/index0/coherency_line_size
# 결과: 64

# 또는
grep cache_alignment /proc/cpuinfo
```

### 3. NUMA 고려

```
멀티소켓 서버에서 NUMA (Non-Uniform Memory Access)

같은 소켓의 메모리: 빠름 (~30ns)
다른 소켓의 메모리: 느림 (~70ns)

최적화:
├─ Thread를 NUMA 노드에 고정
├─ 데이터도 같은 노드 배치
└─ numactl 도구 사용
```

---

## 핵심 정리

| 최적화 | 효과 | 난이도 |
|--------|------|--------|
| **False Sharing 제거** | 50-100배 | 낮음 |
| **SoA vs AoS** | 5배 | 낮음 |
| **필드 정렬** | 2배 | 매우낮음 |
| **캐시친화 알고리즘** | 10배 | 높음 |
| **Prefetch** | 1.5배 | 중간 |

---

## 결론

**"캐시를 이해하면 프로그램이 보인다"**

메모리 접근 = 가장 큰 성능 병목 (1000배 차이!)

캐시 최적화 = 가장 효율적인 성능 개선 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
