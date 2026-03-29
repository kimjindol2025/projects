---
layout: post
title: Phase4-034-SIMD-Vectorization
date: 2026-03-28
---
# SIMD와 벡터화: AVX-512로 10배 가속화

## 요약

- SIMD (Single Instruction Multiple Data) 개념
- AVX-512 아키텍처
- 벡터화 컴파일러 최적화
- 실전 사례 (이미지 처리, 데이터베이스)
- 성능 벤치마크

---

## 1. SIMD란?

### CPU 병렬화의 세 가지 방식

```
1. ILP (Instruction Level Parallelism)
   └─ 스칼라 명령 병렬화 (순서 없는 실행)
   └─ 1 clock에 1-4개 명령 (3-4 instructions/cycle)

2. 스레드 병렬화
   └─ 여러 스레드 (코어)
   └─ 멀티코어: 코어당 4-8개 명령/cycle

3. SIMD (벡터화) ← 가장 강력함
   └─ 1 명령이 여러 데이터 처리
   └─ 8 요소 동시 처리 (8x 가속)
```

### 예: 100만 요소 배열 합산

```c
// 스칼라 (일반 루프)
int sum = 0;
for (int i = 0; i < 1000000; i++) {
    sum += array[i];
}
시간: 1000000 cycles

// SIMD (벡터화)
for (int i = 0; i < 1000000; i += 8) {
    v256 = _mm256_load_ps(&array[i]);      // 8개 로드
    v_sum = _mm256_add_ps(v_sum, v256);    // 8개 동시 더하기
}
시간: 1000000/8 = 125000 cycles (8배!)
```

---

## 2. AVX-512 아키텍처

### 진화 과정

```
SSE (2000)
└─ 128-bit 레지스터
└─ 4개 float32 또는 2개 float64

AVX (2011)
└─ 256-bit 레지스터 (2배)
└─ 8개 float32 또는 4개 float64

AVX2 (2013)
└─ 256-bit 정수 연산

AVX-512 (2017)
└─ 512-bit 레지스터 (4배)
└─ 16개 float32 또는 8개 float64
└─ Mask 레지스터 (조건부 실행)
```

### AVX-512 레지스터

```
zmm0~zmm31: 512-bit 벡터 레지스터 (32개)

zmm0:
[float0 | float1 | float2 | float3 | ... | float15]
 f32     f32     f32     f32           f32

또는:

[int0 | int1 | int2 | ... | int15]
 i32   i32   i32         i32
```

### 주요 명령어

```
_mm512_load_ps(ptr)       // 16개 float32 로드
_mm512_store_ps(ptr, v)   // 16개 float32 저장
_mm512_add_ps(a, b)       // 16개 더하기
_mm512_mul_ps(a, b)       // 16개 곱하기
_mm512_fmadd_ps(a, b, c)  // a*b+c (Fused Multiply-Add)

_mm512_mask_add_ps(zmm0, mask, a, b)  // 조건부 덧셈
```

---

## 3. SIMD 벡터화 구현

### 예 1: 배열 정규화 (Normalization)

```c
// 스칼라 버전
void normalize_scalar(float *data, int n) {
    float max = 0.0f;

    // Pass 1: 최대값 찾기
    for (int i = 0; i < n; i++) {
        if (data[i] > max) max = data[i];
    }

    // Pass 2: 정규화
    for (int i = 0; i < n; i++) {
        data[i] = data[i] / max;
    }
}

// SIMD 버전 (AVX-512)
#include <immintrin.h>

void normalize_simd(float *data, int n) {
    __m512 v_max = _mm512_setzero_ps();  // 0으로 초기화

    // Pass 1: 병렬로 최대값 찾기
    for (int i = 0; i < n; i += 16) {
        __m512 v = _mm512_load_ps(&data[i]);
        v_max = _mm512_max_ps(v_max, v);  // 16개 병렬 비교
    }

    // 16개 결과 중 실제 최대값 찾기 (reduce)
    float max_values[16];
    _mm512_storeu_ps(max_values, v_max);
    float max = max_values[0];
    for (int i = 1; i < 16; i++) {
        if (max_values[i] > max) max = max_values[i];
    }

    // Pass 2: 정규화
    __m512 v_max_broadcast = _mm512_set1_ps(max);
    for (int i = 0; i < n; i += 16) {
        __m512 v = _mm512_load_ps(&data[i]);
        __m512 normalized = _mm512_div_ps(v, v_max_broadcast);
        _mm512_store_ps(&data[i], normalized);
    }
}
```

### 성능 비교

```
데이터: 100만 요소 float32

스칼라: 10ms
AVX-512: 1.5ms (6.7배 가속)

이론: 16x (AVX-512 레지스터)
실제: 6.7x (메모리 대역폭 병목)
```

### 예 2: 이미지 필터 (컨볼루션)

```c
// 3x3 Gaussian blur (AVX-512)
void blur_simd(uint8_t *src, uint8_t *dst, int width, int height) {
    float kernel[9] = {1, 2, 1, 2, 4, 2, 1, 2, 1};
    float divisor = 16.0f;

    // 커널을 SIMD 상수로 변환
    __m512 k0 = _mm512_set1_ps(kernel[0]);
    __m512 k1 = _mm512_set1_ps(kernel[1]);
    // ... k2~k8

    for (int y = 1; y < height - 1; y++) {
        for (int x = 1; x < width - 16; x += 16) {
            // 16개 픽셀 병렬 처리
            __m512 sum = _mm512_setzero_ps();

            // 3x3 영역의 9개 픽셀 로드 및 곱하기
            __m512 v0 = _mm512_cvtepu8_epi32(...);  // 캐스팅
            sum = _mm512_fmadd_ps(v0, k0, sum);  // fused multiply-add

            // ... 8개 더 반복

            __m512 result = _mm512_div_ps(sum, _mm512_set1_ps(divisor));
            _mm512_storeu_epi32(..., result);
        }
    }
}
```

---

## 4. 컴파일러 자동 벡터화

### GCC/Clang의 자동 벡터화

```bash
# 벡터화 활성화
gcc -O3 -mavx512f mycode.c

# 벡터화 리포트 보기
gcc -O3 -mavx512f -fopt-info-vec mycode.c

# SIMD 레벨별 컴파일
gcc -O3 -msse2 mycode.c        # SSE2 (128-bit)
gcc -O3 -mavx2 mycode.c        # AVX2 (256-bit)
gcc -O3 -mavx512f mycode.c     # AVX-512 (512-bit)
```

### 자동 벡터화 가능한 패턴

```c
// ✅ 벡터화됨 (독립적 루프)
for (int i = 0; i < n; i++) {
    c[i] = a[i] + b[i];  // 각 반복 독립적
}

// ❌ 벡터화 불가 (의존성)
for (int i = 1; i < n; i++) {
    a[i] = a[i-1] + b[i];  // i가 i-1에 의존
}

// ✅ 벡터화됨 (조건문 단순)
for (int i = 0; i < n; i++) {
    if (a[i] > 0) {
        c[i] = a[i] * 2;
    }
}

// ❌ 벡터화 불가 (함수 호출)
for (int i = 0; i < n; i++) {
    c[i] = expensive_function(a[i]);
}
```

---

## 5. 메모리 대역폭 분석

### Roofline Model

```
성능 = min(Peak Flops/s, Peak BW × Arithmetic Intensity)

Peak Flops/s = CPU 주파수 × 코어 수 × 연산/사이클
Peak BW = 메모리 버스 × 주파수

Arithmetic Intensity = 연산 수 / 메모리 접근 수

예: Skylake-X (AVX-512)
├─ Peak Flops: 768 GFlops (3.8 GHz × 2 × 512-bit FMA)
├─ Peak BW: 76 GB/s (메모리 대역폭)
└─ Roofline knee: 10.1 (연산:메모리 비율)

케이스 1: 덧셈 (AI=1)
└─ 제한: 76 GFlops (메모리 대역폭)

케이스 2: 내적 (AI=16)
└─ 제한: 768 GFlops (연산)
```

### 메모리 대역폭 병목 확인

```bash
# perf로 메모리 접근 분석
perf stat -e memory_port_utilized ./program

# 높으면 SIMD가 효과적
# 낮으면 메모리 대역폭이 병목
```

---

## 6. 실전 사례

### Case 1: 데이터베이스 필터링 (Apache Arrow)

```c
// 조건 필터 (x > threshold)를 SIMD로 가속

// 스칼라
for (int i = 0; i < n; i++) {
    if (data[i] > threshold) {
        output[out_count++] = data[i];
    }
}

// SIMD (AVX-512)
__m512 v_threshold = _mm512_set1_ps(threshold);

for (int i = 0; i < n; i += 16) {
    __m512 v = _mm512_load_ps(&data[i]);
    __mmask16 mask = _mm512_cmp_ps_mask(v, v_threshold, _CMP_GT_OQ);

    // Mask에 따라 선택적으로 저장
    int count = _mm_popcnt_u32(mask);  // 비트 개수
    _mm512_mask_compressstoreu_ps(&output[out_count], mask, v);
    out_count += count;
}

성능: 스칼라 10ms → SIMD 1.5ms (6.7배)
```

### Case 2: 이미지 처리 (OpenCV)

```python
import cv2
import numpy as np

# OpenCV 내부 SIMD 활용
img = cv2.imread('image.jpg')

# 모든 연산이 자동으로 SIMD 활용
result = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
blurred = cv2.GaussianBlur(result, (5, 5), 0)
edges = cv2.Canny(blurred, 100, 200)

# 성능: 풀 HD (1920x1080)
# 스칼라: 50ms
# SIMD: 8-10ms (5배)
```

---

## 7. 포트빌리티와 한계

### CPU 지원 확인

```bash
# 현재 CPU 지원 확인
cat /proc/cpuinfo | grep avx

# 결과
flags: fpu vme de pse tsc avx avx2 avx512f avx512cd ...

# 없으면 SIMD 실행 불가
```

### 이식성 전략

```c
#ifdef __AVX512F__
    // AVX-512 코드
    result = _mm512_add_ps(a, b);
#elif __AVX2__
    // AVX2 코드
    result = _mm256_add_ps(a, b);
#else
    // 스칼라 코드
    result = a + b;
#endif
```

### 한계

```
1. CPU 마다 다름 (Intel vs AMD)
2. 모바일/ARM에서 미지원
3. 컴파일러마다 결과 다름
4. 디버깅 어려움
```

---

## 8. 벤치마크

### 작업별 예상 성능 향상

| 작업 | 스칼라 | SIMD | 향상도 |
|------|--------|------|--------|
| 덧셈/곱셈 | 1ms | 0.1ms | 10배 |
| 필터링 | 10ms | 1.5ms | 6.7배 |
| 이미지 블러 | 50ms | 8ms | 6배 |
| 정렬 (비교 기반) | 100ms | 20ms | 5배 |
| 문자열 검색 | 50ms | 10ms | 5배 |

### 제약

```
메모리 대역폭 병목:
└─ 연산:메모리 비율이 낮으면 효과 미미

분기문 많음:
└─ if/else가 많으면 마스킹 오버헤드

함수 호출:
└─ 컴파일러 벡터화 실패
```

---

## 핵심 정리

| 레벨 | 비트폭 | 요소수 (float32) | 가속도 |
|------|--------|-----------------|--------|
| **SSE** | 128 | 4 | 2-4x |
| **AVX** | 256 | 8 | 4-6x |
| **AVX2** | 256 | 8 | 4-6x |
| **AVX-512** | 512 | 16 | 6-10x |

---

## 결론

**"SIMD는 수학 집약적 코드를 10배 빠르게 한다"**

데이터베이스, 이미지 처리, 머신러닝 모두 SIMD 활용 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
