---
title: "GC 알고리즘 비교: Mark-Sweep, Generational, Concurrent"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# GC 알고리즘 비교: Mark-Sweep, Generational, Concurrent
## 요약

- 가비지 컬렉션의 개념
- Mark-Sweep 알고리즘
- Generational GC
- Concurrent GC
- 성능 벤치마크

---

## 1. GC가 필요한 이유

### 메모리 관리의 두 가지 방식

```
Manual (C, C++):
├─ malloc/free 직접 호출
├─ 개발자 책임
├─ 메모리 누수, double-free 위험
└─ 성능: 빠르지만 위험

Automatic (Java, Go, Python):
├─ GC가 자동 회수
├─ 개발자 편함
├─ 포즈 타임 (GC 일시 정지)
└─ 성능: 약간 느림, 안전함

타이밍:
├─ Eager: 메모리 부족할 때
└─ Periodic: 시간이 지나면
```

---

## 2. Mark-Sweep 알고리즘

### 원리

```
단계 1: Mark (표시)
├─ Root에서 도달 가능한 모든 객체 표시
└─ DFS/BFS로 그래프 탐색

단계 2: Sweep (정리)
├─ 표시되지 않은 객체 회수
└─ 메모리 반환
```

### 구현

```go
type GC struct {
    objects map[*Object]bool  // 객체 추적
    marked  map[*Object]bool  // 표시된 객체
}

func (gc *GC) Mark(root *Object) {
    if root == nil || gc.marked[root] {
        return
    }
    gc.marked[root] = true

    // 참조된 모든 객체 표시
    for _, ref := range root.references {
        gc.Mark(ref)
    }
}

func (gc *GC) Sweep() {
    for obj := range gc.objects {
        if !gc.marked[obj] {
            delete(gc.objects, obj)  // 회수
            obj.Finalize()
        }
    }
    gc.marked = make(map[*Object]bool)
}

func (gc *GC) Collect() {
    gc.Mark(gc.root)
    gc.Sweep()
}
```

### 문제점

```
Stop-the-World (포즈):
├─ GC 중 모든 스레드 정지
├─ 최악의 경우 수 초 정지
└─ 실시간 애플리케이션 부적합

메모리 단편화:
├─ 회수된 메모리가 분산됨
├─ 큰 객체 할당 실패 가능
└─ 압축(Compaction) 필요 (비용 높음)
```

---

## 3. Generational GC

### 핵심 아이디어

```
"대부분 객체는 어리다"

관찰:
- 새로운 객체: 빠르게 쓰레기 됨 (높은 사망률)
- 오래된 객체: 오래 살아남 (낮은 사망률)

전략:
- Young Generation: 자주 수집 (빠름)
- Old Generation: 드물게 수집 (비용)
- 계층 간 참조: 카드 마킹으로 추적
```

### 메모리 레이아웃

```
Heap:
├─ Young Gen (Eden + Survivor)
│  ├─ 신규 객체 할당 위치
│  ├─ Minor GC: 자주 실행
│  └─ 포즈 타임: ~1-10ms
│
└─ Old Gen
   ├─ Young에서 생존한 객체
   ├─ Major GC: 드물게 실행
   └─ 포즈 타임: 100-1000ms
```

### 성능

```
스칼라 할당 (100만 객체):

Mark-Sweep:
├─ 포즈: 500ms
├─ 처리량: 2M ops/sec

Generational:
├─ Minor GC (Young): 50ms (자주)
├─ Major GC (Old): 500ms (드물게)
├─ 평균 포즈: 100ms
└─ 처리량: 10M ops/sec (5배)
```

---

## 4. Concurrent GC

### Concurrent Mark-Sweep

```
문제: Mark 중 새로운 객체 할당?

해결책: 동시 실행
├─ Mark: 애플리케이션과 병렬
├─ Sweep: 애플리케이션과 병렬
└─ Write Barrier: 참조 변화 추적
```

### Write Barrier

```
애플리케이션이 참조 수정할 때:

일반:
object.field = newValue;

Concurrent GC:
object.field = newValue;
gc.writeBarrier(object, newValue);  // 기록

// GC가 놓친 참조 보정
```

### G1GC (Java)

```
특징:
├─ Heap을 영역(Region)으로 분할
├─ 각 영역 독립적으로 수집
├─ 예측 가능한 포즈 (200ms 이하)
└─ 멀티코어 활용

포즈:
├─ Young-only: 50ms
├─ Mixed: 150ms (Old + Young)
└─ Full GC: 1000ms (드물게)
```

---

## 5. GC 알고리즘 비교

| 알고리즘 | 포즈 시간 | 처리량 | 복잡도 |
|---------|----------|--------|--------|
| **Mark-Sweep** | 500ms | 2M ops/s | 낮음 |
| **Generational** | 100ms | 10M ops/s | 중간 |
| **Concurrent** | 50ms | 20M ops/s | 높음 |
| **G1GC** | 150ms | 15M ops/s | 높음 |

---

## 6. 튜닝 예시

### Java (-Xmx2g)

```bash
# 기본 (Parallel GC)
java -Xmx2g App
# 포즈: ~500ms, 처리량: 높음

# G1GC (권장)
java -Xmx2g -XX:+UseG1GC -XX:MaxGCPauseMillis=200 App
# 포즈: ~200ms, 처리량: 중간

# Low-latency (ZGC)
java -Xmx2g -XX:+UseZGC App
# 포즈: <10ms, 처리량: 중간
```

---

## 핵심 정리

```
Mark-Sweep: 간단, 느림
Generational: 빠름, 복잡
Concurrent: 낮은 포즈, 고복잡도
```

---

## 결론

**"GC는 트레이드오프다"**

포즈 vs 처리량 vs 메모리 선택

---

질문이나 피드백은 댓글로 남겨주세요! 💬
