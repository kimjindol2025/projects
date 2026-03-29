---
title: "컴파일러 최적화: 인라인, 루프 언롤, 불변식 제거"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# 컴파일러 최적화: 인라인, 루프 언롤, 불변식 제거
## 요약

- 컴파일러 최적화 수준 (-O0 ~ -O3)
- 함수 인라인 (Inlining)
- 루프 언롤 (Loop Unrolling)
- 데드 코드 제거 (Dead Code Elimination)
- 실전 프로파일링

---

## 1. 컴파일러 최적화 수준

### GCC/Clang 최적화 플래그

```bash
-O0: 최적화 없음
├─ 디버깅 가능
├─ 컴파일 빠름
└─ 실행 느림

-O1: 기본 최적화
├─ 불필요한 코드 제거
├─ 루프 불변식 제거
└─ 컴파일 시간 증가, 실행 빨라짐

-O2: 중간 최적화 (권장)
├─ -O1의 모든 것
├─ 함수 인라인
├─ 벡터화
└─ 실행 매우 빨라짐

-O3: 적극적 최적화
├─ -O2의 모든 것
├─ 루프 언롤
├─ 함수 복제
└─ 바이너리 크기 증가

-Ofast: 표준 무시
├─ -O3 + unsafe 최적화
├─ IEEE 부동소수점 무시
└─ 위험 (정확성 손실 가능)

-Os: 크기 최적화
├─ 임베디드 환경
├─ 캐시 친화적
└─ 성능 중간
```

### 최적화 결과

```
코드: 루프 내 간단한 계산 (100만 반복)

-O0: 1000ms
-O1: 100ms (10배)
-O2: 50ms (20배)
-O3: 30ms (33배)

트레이드오프:
├─ 컴파일 시간 증가
├─ 바이너리 크기 증가
└─ 디버깅 어려움
```

---

## 2. 함수 인라인 (Inlining)

### 개념

```
함수 호출 오버헤드 제거:

일반 함수 호출:
├─ 인자 스택 저장
├─ return address 저장
├─ 점프 (캐시 미스 가능)
├─ 함수 실행
├─ return
└─ 캐시 오염 (긴 call stack)

인라인:
함수 본문을 호출 위치에 직접 삽입
→ 호출 오버헤드 완전 제거
```

### 예

```c
// 원본
int add(int a, int b) {
    return a + b;
}

int result = add(1, 2) + add(3, 4);

// 인라인 후
int result = (1 + 2) + (3 + 4);

// 최적화 후
int result = 10;  // 컴파일 타임 계산
```

### 인라인 규칙

```c
// ❌ 인라인 안함
void big_function() {
    // 1000줄 코드
    // 바이너리 크기 폭증
}

// ✅ 인라인함
inline int add(int a, int b) {
    return a + b;
}

// ✅ 강제 인라인 (GCC)
__attribute__((always_inline))
int add(int a, int b) {
    return a + b;
}

// ❌ 인라인 금지
__attribute__((noinline))
void expensive() {
    // ...
}
```

### 성능 영향

```
짧은 함수 (1-3줄):
└─ 인라인 → 10배 빨라짐

긴 함수 (100+ 줄):
└─ 인라인 → 바이너리 폭증, 성능 악화

I-cache thrashing (명령어 캐시 오염):
└─ 과도한 인라인 → 캐시 미스 증가
```

---

## 3. 루프 언롤 (Loop Unrolling)

### 개념

```
루프 오버헤드 감소:

원본:
for (int i = 0; i < 100; i++) {
    sum += arr[i];
}

각 반복마다:
├─ i 증가
├─ 경계 체크 (i < 100)
├─ 점프 (캐시 미스 가능)
└─ 계산

오버헤드: 100회 × 3 명령어 = 300 명령어
```

### 언롤된 코드

```c
// 4번 언롤 (-O3)
for (int i = 0; i < 100; i += 4) {
    sum += arr[i];
    sum += arr[i+1];
    sum += arr[i+2];
    sum += arr[i+3];
}

오버헤드: 25회 × 3 명령어 = 75 명령어 (4배 감소!)
계산: 100회 추가 (필요함)
```

### 효과

```
언롤 팩터 = 한 번에 몇 반복 실행?

Factor=1: 오버헤드 100%
Factor=2: 오버헤드 50%
Factor=4: 오버헤드 25%
Factor=8: 오버헤드 12.5%

한계:
├─ 너무 많이 언롤 → 캐시 미스
├─ 루프 본문 크기 증가
└─ 컴파일 시간 증가
```

### 수동 언롤 vs 자동

```c
// 수동 (C++ 템플릿)
template<int N>
struct Loop {
    static void run(int *arr, int &sum) {
        sum += arr[N];
        Loop<N-1>::run(arr, sum);
    }
};

template<>
struct Loop<0> {
    static void run(int *arr, int &sum) {}
};

// 컴파일 타임에 완전 언롤됨

// 자동 (GCC)
gcc -O3 -funroll-loops code.c
// 컴파일러가 자동 판단
```

---

## 4. 루프 불변식 제거 (Loop Invariant Code Motion)

### 문제

```c
// ❌ 루프 안에서 매번 계산
for (int i = 0; i < n; i++) {
    int len = strlen(str);  // 매번 전체 문자열 스캔!
    printf("%c\n", str[i]);
}

// 시간: O(n²)
```

### 해결책

```c
// ✅ 루프 전에 한 번만
int len = strlen(str);
for (int i = 0; i < len; i++) {
    printf("%c\n", str[i]);
}

// 시간: O(n)
```

### 컴파일러 수준

```c
// 컴파일러가 자동으로 감지 (-O2)
for (int i = 0; i < 100; i++) {
    int x = 10 + 20;  // 불변식!
    sum += x;
}

// 최적화 후
int x = 30;
for (int i = 0; i < 100; i++) {
    sum += x;
}
```

### 어려운 경우

```c
// 컴파일러가 못 감지 (함수 호출)
for (int i = 0; i < 100; i++) {
    int len = get_length();  // get_length()가 부작용 있을 수 있음
    sum += len;
}

// 해결책: 수동으로 제거
int len = get_length();
for (int i = 0; i < 100; i++) {
    sum += len;
}
```

---

## 5. 데드 코드 제거 (Dead Code Elimination)

### 개념

```
도달 불가능하거나 사용되지 않는 코드 제거

유형 1: 도달 불가능 코드
└─ if (0) { ... }

유형 2: 미사용 변수
└─ int x = 5;  // x 사용 안함

유형 3: 미사용 함수
└─ void unused() { ... }
```

### 예

```c
// ❌ 미적화
int result = expensive_function();  // 결과 미사용
return 42;

// ✅ 최적화 후
return 42;  // 함수 호출 제거됨
```

### 제약

```c
// 문제: 컴파일러가 부작용 감지하지 못함
void side_effect() {
    printf("hi");  // 부작용!
}

int main() {
    side_effect();  // 반환값 미사용
}

// -O2 컴파일러
// side_effect() 호출 제거 가능 (버그!)

// 해결책: volatile
void volatile_side_effect() volatile {
    // volatile = "부작용 있음" 힌트
}
```

---

## 6. 기타 최적화

### 상수 폴딩 (Constant Folding)

```c
int x = 10 + 20 + 30;
// → int x = 60;

int y = 1000 * 1000;
// → int y = 1000000;
```

### 함수 복제 (Function Cloning)

```c
// 원본
void process(int *arr, int size) {
    for (int i = 0; i < size; i++) {
        if (cond) arr[i] *= 2;
    }
}

// 복제 (최적화 버전)
void process_optimized(int *arr, int size) {
    // if 제거, cond = true로 가정
    for (int i = 0; i < size; i++) {
        arr[i] *= 2;
    }
}

// 호출 위치에 따라 적절한 버전 선택
```

### 테일 콜 최적화 (Tail Call Optimization)

```c
// 원본 (스택 오버플로우 위험)
int factorial(int n) {
    if (n <= 1) return 1;
    return n * factorial(n-1);
}

// TCO 후 (스택 안전)
int factorial(int n, int acc = 1) {
    if (n <= 1) return acc;
    return factorial(n-1, n * acc);  // 점프로 변환
}
```

---

## 7. 프로파일링으로 최적화 검증

### GCC 프로파일 가이드 최적화 (PGO)

```bash
# 1단계: 프로파일 정보 수집
gcc -fprofile-generate -O2 code.c -o code

# 2단계: 실행 (프로파일 데이터 수집)
./code < input.txt

# 3단계: 프로파일 기반 최적화로 재컴파일
gcc -fprofile-use -O2 code.c -o code-optimized

# 결과: 10-20% 추가 성능 향상
```

### perf로 최적화 검증

```bash
perf stat -e instructions,branches,branch-misses ./program

# 결과
Performance counter stats:
  5,000,000  instructions              # 명령어 수
   500,000   branches                  # 분기 수
    10,000   branch-misses             # 분기 미스 (2%)

// 낮을수록 좋음
```

---

## 8. 벤치마크

### 최적화별 성능 향상

| 최적화 | 효과 | 난이도 |
|--------|------|--------|
| **-O2** | 10-20배 | 매우낮음 |
| **인라인** | 2-5배 | 낮음 |
| **루프 언롤** | 2-4배 | 중간 |
| **불변식 제거** | 2-10배 | 낮음 |
| **PGO** | 1.1-1.3배 | 높음 |

---

## 핵심 정리

**컴파일러 최적화 체크리스트**

```
☐ -O2 또는 -O3 사용
☐ 불필요한 함수 호출 제거
☐ 루프 불변식 제거
☐ 데드 코드 확인
☐ 캐시 친화적 배치
☐ 프로파일링으로 검증
☐ -funroll-loops 시도
```

---

## 결론

**"컴파일러는 당신보다 똑똑하다"**

대부분의 최적화는 컴파일러가 자동으로 수행합니다.

-O2/O3 + 좋은 알고리즘 = 최고의 성능 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
