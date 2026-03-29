---
title: "타입 시스템: Hindley-Milner vs Gradual Typing"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# 타입 시스템: Hindley-Milner vs Gradual Typing
## 요약

- 타입 시스템의 종류 (Static/Dynamic)
- Hindley-Milner 타입 추론
- Gradual Typing의 장점과 한계
- 실전: Go, Rust, TypeScript의 타입 시스템
- 성능과 안전성 트레이드오프

---

## 1. 타입 시스템의 분류

### Static vs Dynamic

```
Static (컴파일 타임):
├─ Go, Rust, Java, C++
├─ 모든 타입 미리 결정
├─ 컴파일러 검증
└─ 성능: 빠름, 안전성: 높음

Dynamic (런타임):
├─ Python, JavaScript, Ruby
├─ 값에 따라 타입 결정
├─ 런타임 검사
└─ 성능: 느림, 유연성: 높음

Gradual (혼합):
├─ TypeScript, Dart
├─ Static + Dynamic 선택
└─ 유연성 + 안전성 균형
```

### Strong vs Weak

```
Strong (엄격):
├─ 명시적 변환 필수
├─ Go, Rust, Python

Weak (관대):
├─ 암묵적 변환 가능
├─ C, JavaScript
└─ 버그 위험 높음
```

---

## 2. Hindley-Milner 타입 추론

### 개념

```
프로그래머가 타입을 명시하지 않아도
컴파일러가 자동으로 타입 추론

예: Haskell, OCaml, ML

let add x y = x + y
// 컴파일러가 자동으로 추론:
// add :: Int -> Int -> Int
// (또는 Num a => a -> a -> a)
```

### 추론 알고리즘

```
1. 각 표현식의 제약 조건 수집
2. 제약 조건 통합 (Unification)
3. 일관성 있는 할당 찾기

예:
add x y = x + y

제약 조건:
├─ x + y가 유효 → x는 Num 타입
├─ y가 x와 더함 → x, y 같은 타입
└─ 결과가 x+y → 반환 타입 = x+y 타입

결론:
add :: Num a => a -> a -> a
```

### 한계

```
❌ 상위 등급 함수 (Higher-Rank Types)
├─ forall와 명시적 타입 필요

❌ 복잡한 제약 조건
├─ 컴파일러가 해석 못함
└─ 명시 필요

예: Haskell
forall a. (forall b. b -> b) -> a -> a
```

---

## 3. Gradual Typing (TypeScript)

### 개념

```
"필요한 만큼만 타입 명시"

Option 1: 완전 정적
function add(x: number, y: number): number {
    return x + y;
}

Option 2: 부분 정적
function add(x: number, y) {
    return x + y;  // y: any
}

Option 3: 동적 (일반 JS)
function add(x, y) {
    return x + y;
}

시간에 따라 마이그레이션 가능!
```

### 타입 안전성 vs 유연성

```
안전성 높음 ← ─ ─ ─ ─ ─ ─ → 유연성 높음

Rust (100% static)
  ↓
Go (98% static, 2% interface{})
  ↓
TypeScript (혼합)
  ↓
Python (type hints 선택사항)
  ↓
JavaScript (완전 동적)
```

---

## 4. 실전: Go의 Interface{}

### 문제

```go
// Go는 명시적 제네릭 없음
func printValue(v interface{}) {
    switch v.(type) {
    case int:
        fmt.Println("int:", v.(int))
    case string:
        fmt.Println("string:", v.(string))
    default:
        fmt.Println("unknown")
    }
}

// 타입 안전성 없음 (런타임 패닉 위험)
```

### 해결책: Generics (Go 1.18+)

```go
// 이제 제네릭 가능!
func printValue[T any](v T) {
    fmt.Println(v)  // T는 컴파일 타임에 결정
}

// 호출
printValue(42)       // T = int
printValue("hello")  // T = string
```

---

## 5. Rust의 타입 시스템

### 특징

```
1. 소유권 (Ownership)
   ├─ 메모리 안전성 보장
   └─ 컴파일 타임 검사

2. 트레이트 바운드 (Trait Bounds)
   fn print<T: Display>(v: T) {
       println!("{}", v);  // Display 구현 필수
   }

3. Lifetime (생명주기)
   fn borrow<'a>(x: &'a str) -> &'a str {
       x  // 참조 유효성 검증
   }
```

### 성능 영향

```
Rust 타입 시스템:
├─ Zero-cost abstractions
├─ 런타임 오버헤드 없음
├─ 컴파일 타임 검증
└─ 메모리 안전 + 속도 = 최고

C++ 템플릿:
├─ 유사한 성능
├─ 하지만 복잡성 높음
└─ 컴파일 시간 오래 걸림
```

---

## 6. TypeScript의 Structural Typing

### Nominal vs Structural

```
Nominal (이름 기반):
class Dog { name: string; }
class Cat { name: string; }

let dog: Dog = { name: "Max" };
let cat: Cat = dog;  // ❌ 에러 (다른 클래스)

Structural (구조 기반):
interface Named { name: string; }

let dog = { name: "Max" };
let cat: Named = dog;  // ✅ OK (구조 동일)
```

### 이점

```
✅ 더 유연함
❌ 의도치 않은 호환성
```

---

## 7. 성능 영향

### 컴파일 시간

```
JavaScript (완전 동적):
├─ 컴파일 시간 0
└─ 런타임 10초

TypeScript (점진적):
├─ 컴파일 시간 3초
├─ 번들링 2초
└─ 런타임 10초
└─ 합계: 15초 (50% 증가)

Rust (완전 정적):
├─ 컴파일 시간 30초
├─ 링킹 2초
└─ 런타임 0.1초 (JavaScript 대비 100배)
```

### 런타임 성능

```
동적 타입 (Python):
├─ 연산: 1000ms
└─ 메모리: 높음

점진적 타입 (TypeScript):
├─ 타입 정보 런타임 제거
├─ 연산: 100ms (10배)
└─ 메모리: 중간

정적 타입 (Go):
├─ 컴파일 타임 최적화
├─ 연산: 10ms (100배)
└─ 메모리: 낮음
```

---

## 8. 타입 시스템 선택 가이드

| 언어 | 타입 시스템 | 용도 | 성능 |
|------|----------|------|------|
| **Rust** | Static + Ownership | 시스템 프로그래밍 | ⭐⭐⭐⭐⭐ |
| **Go** | Static + Simple | 백엔드 | ⭐⭐⭐⭐ |
| **Java** | Static + OOP | 엔터프라이즈 | ⭐⭐⭐ |
| **TypeScript** | Gradual | 프론트엔드 | ⭐⭐⭐ |
| **Python** | Dynamic + Hints | 스크립팅 | ⭐⭐ |
| **JavaScript** | Dynamic | 웹 | ⭐⭐ |

---

## 핵심 정리

```
강한 타입 → 안전성 높음, 성능 좋음, 개발 느림
약한 타입 → 개발 빠름, 버그 많음, 성능 낮음
점진적 → 균형 잡힌 선택
```

---

## 결론

**"타입 시스템은 트레이드오프다"**

완벽한 시스템은 없다. 필요에 맞게 선택하세요!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
