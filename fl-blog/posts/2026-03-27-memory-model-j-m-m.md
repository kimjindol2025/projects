---
title: "메모리 모델과 JMM: 멀티스레드 프로그래밍의 기초"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# 메모리 모델과 JMM: 멀티스레드 프로그래밍의 기초
## 요약

- 메모리 모델의 개념 (SC vs TSO vs DRF)
- Java Memory Model (JMM) 상세 분석
- Happens-Before 관계
- Volatile, Synchronized, Atomic
- 실전 데이터 레이스 탐지

---

## 1. 메모리 모델이란?

### 정의

```
메모리 모델 = CPU와 컴파일러가 어떤 메모리 최적화를
              허용하는지 정의하는 계약

목표:
├─ 프로그래머: 예측 가능한 멀티스레드 동작
├─ 컴파일러/CPU: 최적화 자유도
└─ 트레이드오프 균형

예시:
메모리 모델 없음 → 컴파일러가 아무렇게나 재정렬
메모리 모델 있음 → 규칙에 맞춰 최적화만 허용
```

### 3가지 주요 메모리 모델

```
1. SC (Sequential Consistency)
   └─ 모든 연산이 순서대로 실행
   └─ 구현 비용 높음 (최적화 거의 불가)

2. TSO (Total Store Order)
   └─ 읽기 먼저, 쓰기는 지연 가능
   └─ x86/SPARC 아키텍처

3. DRF (Data Race Free)
   └─ 데이터 레이스 없으면 SC처럼 동작
   └─ 레이스 있으면 동작 정의 안함
   └─ Java, C++11 모델
```

---

## 2. Java Memory Model (JMM)

### JMM의 핵심 약속

```
DRF (Data Race Free) 보장:

데이터 레이스 없음 → SC처럼 동작
├─ 모든 스레드 동작 예측 가능
└─ 컴파일러 최적화로도 안전

데이터 레이스 있음 → 미정의 동작
└─ 아무것도 보장 안함
```

### Happens-Before 관계

```
A happens-before B의 의미:
"A의 모든 메모리 쓰기가 B에 가시적"

예:
Thread 1: x = 1          // A
Thread 1: unlock(lock)   // B가 이 unlock 이전에 시작
Thread 2: lock(lock)     // B
Thread 2: print(x)       // x = 1 보장

JMM 규칙:
1. 프로그램 순서: 같은 스레드 내 순서
2. Monitor Lock: unlock → lock
3. Volatile Write → Read
4. Thread Start: start() → 스레드 내 첫 명령
5. Thread Join: 스레드 마지막 → join() 후
6. Transitivity: A hb B, B hb C → A hb C
```

---

## 3. Synchronized vs Volatile

### Synchronized (모니터 잠금)

```go
// Java
synchronized void increment() {
    count++;  // 원자적 + 가시성
}

// 메커니즘:
// 1. Lock acquisition (happens-before)
// 2. 임계 영역 실행
// 3. Lock release (happens-before)

비용: 높음 (Lock 오버헤드)
가시성: 모든 메모리
성능: 최적화 제한적
```

### Volatile

```go
// Java
volatile int flag = 0;

void setFlag() {
    flag = 1;  // 쓰기가 즉시 가시적
}

int getFlag() {
    return flag;  // 최신 값 읽음
}

// 메커니즘:
// Volatile Write happens-before Volatile Read
// (중간 스레드의 다른 메모리는 보장 안함)

비용: 낮음 (일반 쓰기 + 배리어)
가시성: volatile 변수만
성능: 최적화 많음
```

### 비교

```
작업           Synchronized  Volatile
─────────────────────────────────────
Lock/Unlock    필요           불필요
메모리 가시성  전체           volatile만
성능           낮음           높음
사용 케이스    여러 변수      단일 플래그
```

---

## 4. Atomic 변수 (CAS)

### Compare-And-Swap (CAS)

```go
// Java
AtomicInteger counter = new AtomicInteger(0);

// CAS: 원자적 비교 후 교환
boolean success = counter.compareAndSet(0, 1);
// if (counter == 0) counter = 1; return true else return false

// 또는 get-and-increment
counter.incrementAndGet();  // 원자적 + happens-before
```

### 구현

```c
// x86 기계어 (Lock Prefix)
lock cmpxchg rax, [rbx]

// 메커니즘:
// 1. RBX 주소의 값을 RAX와 비교
// 2. 같으면 새 값으로 교환
// 3. CPU 캐시 라인 Lock (다른 CPU 접근 차단)
```

### 성능

```
방식              처리량      경합 시
─────────────────────────────────
Synchronized      1K ops/s    100K ops/s (경합)
Volatile write    10K ops/s   동일
Atomic (uncontended) 10K ops/s 동일
Atomic (contended) 10K ops/s   1K ops/s (Lock free 아님)
```

---

## 5. 실전 예: 싱글톤 (Double-Checked Locking)

### ❌ 나쁜 예 (데이터 레이스)

```java
class Singleton {
    private static Singleton instance;

    static Singleton getInstance() {
        if (instance == null) {  // ❌ 레이스!
            synchronized (Singleton.class) {
                if (instance == null) {
                    instance = new Singleton();  // ❌ 보이지 않음
                }
            }
        }
        return instance;
    }
}

문제:
Thread A: instance 체크 (null)
Thread B: instance 체크 (null)
Thread A: lock 획득, 초기화, unlock
Thread B: lock 획득하려 대기
Thread B: lock 획득, 초기화 (다시!)

또한:
new Singleton() = 할당 → 초기화 → 참조 할당
중간에 재정렬 가능 → 미완성 객체 참조!
```

### ✅ 좋은 예 (Volatile)

```java
class Singleton {
    private static volatile Singleton instance;  // ← volatile!

    static Singleton getInstance() {
        if (instance == null) {
            synchronized (Singleton.class) {
                if (instance == null) {
                    instance = new Singleton();  // 안전
                }
            }
        }
        return instance;
    }
}

이유:
volatile write happens-before volatile read
→ 초기화가 완료된 객체만 보임
```

### ✅ 더 나은 예 (Eager Initialization)

```java
class Singleton {
    private static final Singleton instance = new Singleton();

    static Singleton getInstance() {
        return instance;  // Lock free!
    }
}

이유:
클래스 로딩 = 스레드 안전 (JVM 보장)
→ 동기화 불필요
```

---

## 6. 데이터 레이스 탐지

### ThreadSanitizer (Go)

```bash
# Go race detector
go test -race ./...

# 출력
==================
WARNING: DATA RACE
Write at 0x00c000100000 by goroutine 2:
  main.main.func1()
      main.go:10 +0x44

Previous read at 0x00c000100000 by goroutine 1:
  main.main()
      main.go:8 +0x6c
==================
```

### Java: ThreadSanitizer, Checkstyle

```bash
# Maven Checkstyle
mvn checkstyle:check

# Thread safety annotation
@GuardedBy("lock")
private int counter;  // lock으로 보호됨
```

### 탐지 원리

```
동적 분석:
├─ 모든 메모리 접근 추적
├─ 동기화 관계 추적
└─ 경합하는 접근 탐지

비용: 3-5배 느림 (테스트 용도만)
```

---

## 7. 메모리 배리어

### CPU 수준 배리어

```
메모리 배리어 = CPU 명령
"이 점 전후로 메모리 연산 재정렬 금지"

종류:
├─ Store Barrier: 이전 store가 먼저 실행
├─ Load Barrier: 이후 load가 나중 실행
└─ Full Barrier: 모든 연산 순서 보장

x86 명령어:
MFENCE  ; 모든 메모리 연산 동기화
SFENCE  ; Store barrier
LFENCE  ; Load barrier

비용:
├─ MFENCE: 10-50 사이클 (비쌈)
├─ SFENCE: 1 사이클
└─ LFENCE: 1 사이클
```

### JMM 배리어

```
Java 컴파일러가 CPU 배리어 삽입:

volatile write:
  ... code ...
  MOV [rax], rbx   ; volatile 쓰기
  MFENCE           ; ← 배리어 삽입
  ... code ...

volatile read:
  ... code ...
  LFENCE           ; ← 배리어 삽입
  MOV rax, [rbx]   ; volatile 읽기
  ... code ...
```

---

## 8. 메모리 모델 비교

### JMM vs C++11 vs Rust

```
메트릭          Java    C++11   Rust
────────────────────────────────────
DRF 보장        ○       ○       ○
Undefined Behavior  제한적  광범위  거의 없음
Atomic          AtomicInt std::atomic atomic
Volatile        유사    유사    한정적
```

### JMM vs x86 TSO

```
JMM: 더 강함 (SC와 유사)
├─ Write → Read 순서 보장
└─ volatile으로 세밀 제어

x86 TSO: 더 약함
├─ Read can pass Write
└─ MFENCE 필요
```

---

## 핵심 정리

| 개념 | 용도 | 비용 |
|------|------|------|
| **Synchronized** | 여러 변수 보호 | 높음 |
| **Volatile** | 플래그/상태 | 낮음 |
| **Atomic** | 카운터 증가 | 중간 |
| **Lock Free** | 극도의 성능 | 복잡함 |

---

## 결론

**"메모리 모델을 이해하면 멀티스레드 버그를 피한다"**

Java, C++, Rust 모두 메모리 모델 정의:
- DRF이면 SC처럼 동작 보장
- 데이터 레이스 = 악의 근원

메모리 모델 이해 = 안전한 병렬 프로그래밍 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
