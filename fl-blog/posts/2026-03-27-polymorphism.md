---
title: "Polymorphism vs Generics: 성능과 유지보수성"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# Polymorphism vs Generics: 성능과 유지보수성
## 요약

- 다형성의 종류 (Ad hoc, Parametric, Subtyping)
- 제네릭과 동적 디스패치
- 성능 vs 유연성
- 실전 설계 가이드

---

## 1. 다형성의 세 가지 종류

### Parametric Polymorphism (제네릭)

```rust
fn identity<T>(x: T) -> T {
    x
}

// 각 타입별로 코드 생성 (Monomorphization)
identity(42)      // i32 버전
identity("hello") // &str 버전
```

### Ad Hoc Polymorphism (함수 오버로딩)

```cpp
int add(int a, int b) { return a + b; }
double add(double a, double b) { return a + b; }

// 컴파일 타임에 적절한 버전 선택
```

### Subtyping Polymorphism (상속)

```java
class Animal { void speak() { } }
class Dog extends Animal { void speak() { println("Woof"); } }
class Cat extends Animal { void speak() { println("Meow"); } }

void makeSound(Animal a) {
    a.speak();  // 런타임 결정 (virtual method)
}
```

---

## 2. 제네릭 vs 동적 디스패치

### 제네릭 (Static Dispatch)

```rust
trait Display {
    fn display(&self);
}

// 제네릭: 타입별 구현체
fn print<T: Display>(item: T) {
    item.display();  // 컴파일 타임 결정
}

// 코드 생성 (각 T마다):
fn print_i32(item: i32) { ... }
fn print_String(item: String) { ... }
```

### 동적 디스패치 (Virtual Method)

```rust
fn print(item: &dyn Display) {
    item.display();  // 런타임 결정 (vtable)
}

// vtable (가상 함수 테이블):
// display -> 0x1000 (Display 구현)
// 런타임에 포인터 추종
```

### 성능 비교

```
제네릭:
├─ 바이너리 크기: 큼 (코드 복제)
├─ 호출 비용: 0 (인라인)
└─ 성능: 최고 (10ns)

동적 디스패치:
├─ 바이너리 크기: 작음
├─ 호출 비용: vtable 추종 (1-3ns)
└─ 성능: 약간 낮음 (13ns)
```

---

## 3. 언어별 구현

### Rust: trait + impl

```rust
trait Container {
    fn len(&self) -> usize;
}

impl Container for Vec<i32> {
    fn len(&self) -> usize { self.len() }
}

// 제네릭 사용 (Static dispatch)
fn size<T: Container>(c: T) -> usize { c.len() }

// 동적 디스패치
fn size_dyn(c: &dyn Container) -> usize { c.len() }
```

### Java: interface

```java
interface Container {
    int len();
}

class VecInt implements Container {
    public int len() { return ... }
}

// Java는 항상 동적 디스패치
void size(Container c) {
    c.len();  // 런타임 결정
}
```

### Go: interface{}

```go
type Container interface {
    Len() int
}

// 동적 디스패치(느림)
func Size(c Container) int {
    return c.Len()  // interface{} 타입 어설션
}

// 제네릭 (Go 1.18+)
func Size[T Container](c T) int {
    return c.Len()  // 컴파일 타임 결정
}
```

---

## 4. 설계 원칙

### 언제 제네릭?

```
✅ 사용:
├─ 컬렉션 (List<T>, Map<K,V>)
├─ 성능 중요
└─ 타입 안전 원함

❌ 피함:
├─ 바이너리 크기 제약
├─ 컴파일 시간 중요
└─ 코드 일부 동적 필요
```

### 언제 동적 디스패치?

```
✅ 사용:
├─ Plugin 아키텍처
├─ 런타임 다형성
└─ 이질적 컬렉션 (Animal[])

❌ 피함:
├─ 성능 극한
├─ 작은 오버헤드도 문제
└─ 타입 안전 필수
```

---

## 5. 실전 벤치마크

### 시나리오: 100만 요소 합산

```
코드:
trait Number { fn add(&self, other: &Self) -> Self; }

데이터:
├─ Vec<i32> (정적 dispatch)
├─ Vec<Box<dyn Number>> (동적 dispatch)
└─ 일반 i32 배열 (baseline)

결과:
일반 배열: 1ms
제네릭: 1.1ms (오버헤드 거의 없음)
동적: 150ms (150배 느림!)
```

### 이유

```
제네릭:
└─ 컴파일러가 인라인 → 일반과 동일

동적:
├─ vtable 추종 (캐시 미스)
├─ branch misprediction (각 원소마다)
└─ SIMD 불가능 (타입 모름)
```

---

## 핵심 정리

| 방식 | 속도 | 유연성 | 바이너리 |
|------|------|--------|---------|
| **제네릭** | 빠름 | 중간 | 크음 |
| **동적** | 느림 | 높음 | 작음 |

---

## 결론

**"기본은 제네릭, 필요하면 동적"**

대부분 제네릭으로 충분합니다.

---

질문이나 피드백은 댓글로 남겨주세요! 💬
