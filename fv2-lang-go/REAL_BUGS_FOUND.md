# 🐛 FV 2.0 Go 실제 발견된 버그

**발견 시간**: 2026-03-20 19:00
**방법**: 실제 코드 실행 & 생성된 C 코드 검증

---

## 🔴 P0: CRITICAL BUGS

### Bug #1: Match 문 코드 생성 완전히 잘못됨

**파일**: `internal/codegen/generator.go:311`
**심각도**: 🔴 CRITICAL

**문제**:
```go
// 현재 코드 (잘못됨)
condition := "1" // 모든 arm에서 항상 true!

// 결과
if (1) { ... } else if (1) { ... } else { ... }
// → 모든 조건이 true이므로 첫 번째만 실행됨
```

**예시**:
```fv
match x {
    1 => { let msg = "one" },
    2 => { let msg = "two" },
    3 => { let msg = "three" },
    _ => { let msg = "other" }
}
```

**생성되는 잘못된 C 코드**:
```c
if (1) {           // ← 항상 true
    auto msg = "one";
} else if (1) {    // ← 절대 실행 안 됨 (이미 if에서 true)
    auto msg = "two";
} else if (1) {    // ← 절대 실행 안 됨
    auto msg = "three";
} else {           // ← 절대 실행 안 됨
    auto msg = "other";
}
```

**올바른 구현**:
```go
// 패턴값과 매칭할 표현식의 값을 비교해야 함
switch p := arm.Pattern.(type) {
case *ast.LiteralPattern:
    // pattern의 리터럴 값을 구해서 비교
    patternValue := g.generateExpression(p.Value)
    condition = fmt.Sprintf("%s == %s", expr, patternValue)
case *ast.IdentifierPattern:
    // 변수 바인딩
    condition = "1"
case *ast.WildcardPattern:
    // 기본 case
    condition = "1"
}
```

**영향도**: ⚠️ Match 문이 제대로 작동하지 않음 (완전 버그)

---

## 🟠 P1: HIGH BUGS

### Bug #2: For 루프에서 배열 크기 계산 안 됨

**파일**: `internal/codegen/generator.go:255`
**심각도**: 🟠 HIGH

**문제**:
```go
// 현재 (잘못됨)
"for (int _i = 0; _i < sizeof(%s)/sizeof(%s[0]); _i++)"

// sizeof(배열 포인터) = 8 바이트
// sizeof(int) = 4 바이트
// → 8/4 = 2만 반복 (배열 실제 크기 무관!)
```

**예시**:
```fv
for i in arr {  // arr = [1,2,3,4,5]
    let x = i
}
```

**생성되는 잘못된 C 코드**:
```c
for (int _i = 0; _i < sizeof(arr)/sizeof(arr[0]); _i++) {
    // sizeof(arr*) / sizeof(int) = 8/4 = 2
    // → 2번만 반복 (배열 5개 원소 중 2개만!)
    int i = _i;
}
```

**원인**:
- 배열이 포인터로 변환되면서 크기 정보 손실
- AST에 배열 길이 정보 저장 안 됨

**해결책**:
1. AST에 배열 길이 정보 저장
2. 또는 for-in을 반복자 패턴으로 변경
3. 또는 배열에 길이 필드 추가 (slice 패턴)

---

### Bug #3: 생성된 C 코드 구문 오류

**문제**:
```c
else if (1) { ... } else { ... }
// ↑ else 앞에 ; 또는 정확한 브레이스 필요
```

**심각도**: 🟠 (gcc는 관대하지만, 표준 C 아님)

---

## 🟡 P2: MEDIUM ISSUES

### Issue #1: 타입 캐스트 부재

**관찰**:
```c
auto x = 42;      // C11만 지원 (auto)
auto y = 3.14;    // auto 타입 추론
auto name = "FV"; // 포인터 타입 불명확
```

**문제**:
- `auto` 는 C11 고급 기능
- 이식성 문제 (C99, C89 미지원)
- 타입 명시 필요

**개선**:
```c
long long x = 42;
double y = 3.14;
char* name = "FV";  // 명시적 타입
```

---

### Issue #2: 메모리 관리 부재

**관찰**:
```fv
let arr = [1, 2, 3, 4, 5]
```

**생성되는 C 코드**:
```c
auto arr = {1, 2, 3, 4, 5};  // 스택 배열
// 문제: 스택 배열은 블록 범위만 유지
```

**문제**:
- 함수에서 반환 불가능
- 전역 배열 필요시 처리 안 됨
- free() 호출 무관

---

### Issue #3: 함수 반환 타입 부재

**관찰**:
```c
void main(void) {
  ...
  return;
}

int main() {
  return 0;
}
```

**문제**:
- 원본 main은 `void`
- 호출 main은 `int`
- 중복 정의

---

## 📊 테스트 결과

### 컴파일 통과하는 것들 ✅
- 기본 타입: ✅
- 산술 연산: ✅
- 논리 연산: ✅
- 배열 (작은 크기): ✅
- 함수 정의 & 호출: ✅
- If-else: ✅
- For-range: ✅

### 작동하지 않는 것들 ❌
- **Match 문**: ❌ (모든 arm이 첫 번째 실행)
- **For-in 루프**: ❌ (배열 크기 불일치)
- **복합 배열**: ❌ (크기 계산 오류)
- **포인터 배열**: ❌ (메모리 관리 없음)

---

## 🎯 즉시 수정 필요

### 우선순위 1: Match 문 수정
```
영향: 모든 match 표현식
예상 시간: 30분
중요도: CRITICAL
```

### 우선순위 2: 배열 크기 처리
```
영향: 모든 for-in 루프
예상 시간: 45분
중요도: HIGH
```

### 우선순위 3: C 코드 생성 개선
```
영향: 모든 출력 코드
예상 시간: 30분
중요도: MEDIUM
```

---

## 다음 단계

1. **Match 버그 수정** (30분)
   - 패턴 타입 확인하고 올바른 조건 생성

2. **배열 처리 개선** (45분)
   - AST에 길이 정보 저장하거나
   - 런타임 길이 추적 추가

3. **C 코드 정리** (30분)
   - 타입 명시화
   - 메모리 관리 개선

4. **재검증** (30분)
   - 생성된 C 코드 실제 컴파일 & 실행
   - 모든 테스트 케이스 다시 검증

**총 예상 시간**: 2-2.5시간

