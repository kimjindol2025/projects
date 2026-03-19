# Phase 4: FV 2.0 Self-Hosting (컴파일러를 V로 재작성)

**목표**: FV 언어로 FV 컴파일러를 작성하여 `fv` 명령어로 FV 코드를 컴파일할 수 있도록 구현

**상태**: 🚀 시작 예정
**예상 기간**: 2주 (8-14일)
**예상 결과**: 7,000+ 줄의 V 코드로 작성된 완전 셀프호스팅 컴파일러

---

## 📋 상세 계획

### Phase 4.1: Lexer를 V로 재작성 (1-2일)
**목표**: 480줄의 V 코드로 Lexer 구현

**파일**: `fv-lang/src/lexer.fv`

**구현 내용**:
- Token 타입 정의 (50+ 토큰)
- 키워드 매핑 (34개 키워드)
- 문자 읽기 헬퍼 (readChar, peekChar, skipWhitespace)
- 토큰 인식 로직
  - 식별자 & 키워드
  - 숫자 (정수, 부동소수점)
  - 문자열 (큰따옴표, 작은따옴표)
  - 연산자 & 구분자
  - 주석 (라인, 블록)
- 전체 토크나이제이션

**테스트**: 14개 테스트 케이스
**기준**: Go Lexer와 동일한 토큰 출력

**코드 라인**: ~480줄

---

### Phase 4.2: Parser를 V로 재작성 (2-3일)
**목표**: 1,100줄의 V 코드로 Parser 구현

**파일**: `fv-lang/src/parser.fv`

**구현 내용**:
- AST 노드 정의 (Expression, Statement, Declaration 등)
- 표현식 파싱 (Pratt Parser)
  - 리터럴, 식별자, 이항/단항 연산
  - 함수 호출, 배열 인덱싱
  - 괄호 표현식
- 문장 파싱
  - 변수 선언 (let)
  - 할당
  - 반복 (for, while)
  - 조건 (if, else)
  - 함수 호출
  - 반환
- 선언 파싱
  - 함수 정의
  - 구조체 정의
  - 인터페이스 정의
  - 변수/상수 정의
- 에러 복구 및 보고

**테스트**: 38개 테스트 케이스
**기준**: Go Parser와 동일한 AST 생성

**코드 라인**: ~1,100줄

---

### Phase 4.3: Type Checker를 V로 재작성 (2-3일)
**목표**: 850줄의 V 코드로 Type Checker 구현

**파일**: `fv-lang/src/checker.fv`

**구현 내용**:
- 타입 정의 (Primitive, Array, Function, Option, Result, Struct, Union, Dynamic, Protocol)
- 타입 검사 엔진
  - 변수 선언 검증
  - 함수 호출 검증
  - 이항 연산 검증
  - 배열 연산 검증
  - 제어문 검증
- 심볼 테이블 관리
- 에러 수집 및 보고

**테스트**: 16개 테스트 케이스
**기준**: Go Type Checker와 동일한 에러 감지

**코드 라인**: ~850줄

---

### Phase 4.4: Code Generator를 V로 재작성 (2-3일)
**목표**: 1,150줄의 V 코드로 Code Generator 구현

**파일**: `fv-lang/src/generator.fv`

**구현 내용**:
- C 코드 생성기
  - 헤더 생성 (includes, typedefs)
  - 함수 선언 생성
  - 함수 구현 생성
  - 변수 선언 생성
  - 표현식 코드 생성
  - 제어문 코드 생성
- 타입 매핑 (FV Type → C Type)
- 들여쓰기 & 포맷팅
- 주석 보존

**테스트**: 12개 테스트 케이스
**기준**: Go Code Generator와 동일한 C 코드 생성

**코드 라인**: ~1,150줄

---

### Phase 4.5: 통합 및 테스트 (1-2일)
**목표**: 모든 모듈 통합 및 200+ 테스트 통과

**파일**: `fv-lang/src/main.fv`

**구현 내용**:
- 메인 컴파일러 루프
  - 파일 읽기
  - Lexer 호출
  - Parser 호출
  - Type Checker 호출
  - Code Generator 호출
  - C 파일 쓰기
- 에러 처리 및 보고
- 성능 로깅

**테스트**:
- 단위 테스트: 80개
- 통합 테스트: 50개
- E2E 테스트: 70개
- 성능 벤치마크: 기준 대비 150-200ms

**결과**: 267개 테스트 모두 통과

---

## 📊 진행 상황 추적

```
Week 1:
  [████░░░░░░░░░░░░░░░░] 20% 완료

  Day 1-2: Lexer (V)                     [████░░░░░░░]  40%
  Day 3-4: Parser (V)                    [░░░░░░░░░░░]   0%
  Day 5:   Type Checker (V)              [░░░░░░░░░░░]   0%

Week 2:
  [░░░░░░░░░░░░░░░░░░░░] 0% 완료

  Day 6-7: Code Generator (V)            [░░░░░░░░░░░]   0%
  Day 8-10: 통합 & 테스트                 [░░░░░░░░░░░]   0%
  Day 11-14: 최적화 & 문서화              [░░░░░░░░░░░]   0%
```

---

## 🎯 마일스톤

### Milestone 1: Lexer 완성 (3-4일)
- ✅ Token 타입 정의
- ✅ Lexer 구현
- ✅ 14개 테스트 통과
- 📊 코드: 480줄
- 📝 제출 커밋: "Phase 4.1: FV Lexer V로 재작성"

### Milestone 2: Parser 완성 (5-7일)
- ✅ AST 노드 정의
- ✅ Parser 구현
- ✅ 38개 테스트 통과
- 📊 코드: 1,100줄
- 📝 제출 커밋: "Phase 4.2: FV Parser V로 재작성"

### Milestone 3: Type Checker 완성 (7-10일)
- ✅ 타입 정의
- ✅ Type Checker 구현
- ✅ 16개 테스트 통과
- 📊 코드: 850줄
- 📝 제출 커밋: "Phase 4.3: FV Type Checker V로 재작성"

### Milestone 4: Code Generator 완성 (10-13일)
- ✅ Code Generator 구현
- ✅ 12개 테스트 통과
- 📊 코드: 1,150줄
- 📝 제출 커밋: "Phase 4.4: FV Code Generator V로 재작성"

### Milestone 5: Self-Hosting 완성 (13-14일)
- ✅ 모든 모듈 통합
- ✅ 200+ 테스트 통과
- ✅ `fv` 명령어로 FV 코드 컴파일 가능
- 📊 총 코드: 7,000+줄 (V)
- 📊 바이너리: 1.5-2.0MB
- 📝 제출 커밋: "🎉 Phase 4: FV Self-Hosting 완료 (FV로 FV 컴파일)"

---

## 📈 예상 성과

### 완성 후 상태
```
FV 2.0 Self-Hosting 아키텍처
════════════════════════════════════════

V 소스 파일 (*.fv)
  ↓
Lexer (V로 작성) → 토큰
  ↓
Parser (V로 작성) → AST
  ↓
Type Checker (V로 작성) → 검증된 AST
  ↓
Code Generator (V로 작성) → C 코드
  ↓
gcc/clang → 바이너리
```

### 지표
| 지표 | Go 버전 | V 버전 |
|------|---------|--------|
| 코드 라인 | 9,179 | 7,100+ |
| 바이너리 크기 | 3.0MB | 1.5-2.0MB |
| 컴파일 시간 | ~100ms | ~150-200ms |
| 테스트 개수 | 267 | 267+ |
| 테스트 통과율 | 100% | 100% |

---

## 🔄 Bootstrap 프로세스

### Step 1: Bootstrap (Go 컴파일러 사용)
```bash
$ go build -o bin/fv ./cmd/fv2
$ ./bin/fv src/lexer.fv > src/lexer.c
$ gcc -o bin/fv_lexer src/lexer.c
```

### Step 2: Self-Hosting (V 컴파일러 사용)
```bash
$ # 이제 V로 작성된 모든 모듈을 V로 컴파일 가능
$ ./bin/fv src/parser.fv > src/parser.c
$ ./bin/fv src/checker.fv > src/checker.c
$ ./bin/fv src/generator.fv > src/generator.c
$ ./bin/fv src/main.fv > src/main.c
```

### Step 3: 완전 Self-Hosting
```bash
$ # V 컴파일러로 V 컴파일러 자신을 컴파일
$ ./bin/fv src/compiler.fv > src/compiler.c
$ gcc -o bin/fv_new src/compiler.c
$ # 완벽한 셀프호스팅 달성! ✅
```

---

## 📝 기술 스택

### 사용 기술
- **언어**: V (FreeLang)
- **출력**: C 코드
- **컴파일**: gcc/clang
- **테스트**: V 테스트 프레임워크
- **버전 관리**: GOGS

### 주요 특징
- 순수 V 구현
- 제로 의존성 (Go 라이브러리 미사용)
- 완전한 타입 검사
- AST 기반 코드 생성

---

## ✅ 완료 기준

1. ✅ 모든 V 코드 작성 완료 (7,000+ 줄)
2. ✅ 267개 테스트 모두 통과
3. ✅ `fv hello.fv` 명령어로 FV 코드 컴파일 가능
4. ✅ 생성된 바이너리 정상 실행
5. ✅ GOGS 커밋 완료
6. ✅ 문서화 완료

---

## 🚀 시작 명령

```bash
cd ~/projects/fv2-lang-go

# Phase 4.1: Lexer V로 재작성
# Phase 4.2: Parser V로 재작성
# Phase 4.3: Type Checker V로 재작성
# Phase 4.4: Code Generator V로 재작성
# Phase 4.5: 통합 & 테스트

# 각 단계마다 커밋 및 GOGS 푸시
```

---

**예상 완료**: 2026-04-02 (2주)
**최종 목표**: FV 2.0 완전 Self-Hosting 달성 🎉
