# Julia Compiler - 빌드 & 테스트 가이드

**최종 업데이트**: 2026-03-19
**상태**: ✅ 완료 (v0.2.0)

---

## 📋 빌드 요구사항

- **Go**: 1.18+
- **OS**: Linux, macOS, Windows

---

## 🔨 빌드 방법

### 1. 개발 빌드

```bash
go build -o jcc ./cmd/jcc
```

### 2. 릴리스 빌드

```bash
go build -ldflags="-s -w" -o jcc ./cmd/jcc
```

### 3. 크로스 컴파일 (Linux x86_64)

```bash
GOOS=linux GOARCH=amd64 go build -o jcc-linux ./cmd/jcc
```

### 4. 크로스 컴파일 (macOS ARM64)

```bash
GOOS=darwin GOARCH=arm64 go build -o jcc-mac ./cmd/jcc
```

---

## 🧪 테스트 실행

### 전체 테스트

```bash
go test ./...
```

### 상세 테스트 (verbose)

```bash
go test -v ./...
```

### E2E 테스트만 실행

```bash
go test -run E2E ./test
```

### 특정 테스트만 실행

```bash
go test -run TestLexer ./internal/lexer
```

### 테스트 커버리지

```bash
go test -cover ./...
```

### 상세 커버리지 리포트

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📊 성능 벤치마크

### 렉서 벤치마크

```bash
go test -bench=BenchmarkLexer -benchmem ./test
```

**예상 결과**:
```
BenchmarkLexer-8    50000    24000 ns/op    8192 B/op    12 allocs/op
```

### 파서 벤치마크

```bash
go test -bench=BenchmarkParser -benchmem ./test
```

**예상 결과**:
```
BenchmarkParser-8   30000    42000 ns/op   12288 B/op    18 allocs/op
```

### 전체 컴파일 파이프라인 벤치마크

```bash
go test -bench=BenchmarkFullCompilation -benchmem ./test
```

**예상 결과**:
```
BenchmarkFullCompilation-8    1000    1200000 ns/op   65536 B/op   156 allocs/op
```

### 모든 벤치마크 실행

```bash
go test -bench=. -benchmem ./test
```

---

## 🚀 사용 예제

### 간단한 Julia 코드 컴파일

**hello.jl**:
```julia
function add(a, b)
    return a + b
end

result = add(2, 3)
print(result)
```

**컴파일 및 실행**:
```bash
./jcc hello.jl
# 또는 디버그 모드
./jcc -debug hello.jl
```

### Fibonacci 계산

**fibonacci.jl**:
```julia
function fib(n)
    if n <= 1
        return n
    end
    return fib(n-1) + fib(n-2)
end

result = fib(10)
```

```bash
./jcc fibonacci.jl
```

---

## 🐛 디버그 모드

디버그 모드에서는 각 컴파일 단계의 상세 정보를 출력합니다:

```bash
./jcc -debug hello.jl
```

**출력 예**:
```
🐛 디버그 모드 활성화
Phase: Lexing...
  ✓ 42 tokens
Phase: Parsing...
  ✓ Parsed 3 statements
Phase: Semantic Analysis...
  ✓ Semantic analysis passed
...
✅ 컴파일 성공: hello.out
```

---

## 📁 프로젝트 구조

```
julia-compiler/
├── cmd/jcc/
│   └── main.go          # CLI 진입점 (리팩토링 완료 ✅)
├── internal/
│   ├── ast/             # AST 정의
│   ├── codegen/         # 코드 생성 & VM
│   ├── ir/              # 중간 표현 (리팩토링 완료 ✅)
│   ├── lexer/           # 토크나이제이션
│   ├── optimizer/       # 최적화
│   ├── parser/          # 파싱
│   ├── sema/            # 의미 분석
│   ├── typeinf/         # 타입 추론
│   └── types/           # 타입 시스템
├── test/
│   ├── e2e_test.go      # E2E 테스트 (새로 추가 ✅)
│   └── ...
└── docs/
    └── BUILD.md         # 이 파일
```

---

## ✅ 개선사항 (v0.2.0)

### Code Review 이슈 수정 (7개 중 4개)

**완료**:
1. ✅ **Issue #1**: PhaseLogger 헬퍼 추상화
   - Phase 로깅 반복 제거 (40줄 → 15줄)
   - `compile()` 함수 간결화

2. ✅ **Issue #2**: readSourceFile 함수 추출
   - 파일 읽기 에러 처리 중복 제거
   - 일관된 에러 메시지

3. ✅ **Issue #3**: buildFunctionParameters 헬퍼
   - Parameter 루프 추출 (ir/builder.go)
   - 매개변수 처리 통일

4. ✅ **Issue #4**: buildCallArguments 헬퍼
   - Call argument 루프 추출
   - 재사용 가능한 인터페이스

**선택적** (Issue #5-7):
- resolveType 헬퍼 (sema.go) - 복잡도 분석 필요
- TokenStream 래퍼 (parser.go) - 선택적 최적화
- emitOp 헬퍼 (codegen.go) - 선택적 최적화

### 테스트 추가

- ✅ **e2e_test.go**: E2E 통합 테스트 (3개 케이스)
- ✅ **BenchmarkLexer**: 렉서 성능 벤치마크
- ✅ **BenchmarkParser**: 파서 성능 벤치마크
- ✅ **BenchmarkFullCompilation**: 전체 파이프라인 성능 벤치마크

### 문서화

- ✅ **BUILD.md**: 빌드 & 테스트 가이드 (완성)
- ✅ **README.md**: 기본 정보 (유지)

---

## 🎯 다음 단계 (v0.3.0)

1. **복잡한 Julia 기능 지원**
   - 다중 디스패치 (Multiple Dispatch) 완성
   - 고급 타입 시스템
   - 메타프로그래밍

2. **성능 최적화**
   - JIT 컴파일 고도화
   - 메모리 최적화
   - 캐싱 메커니즘

3. **표준 라이브러리**
   - Collections (Arrays, Dicts)
   - Math 라이브러리
   - I/O 지원

4. **도구 개선**
   - REPL (대화형 셸)
   - IDE 통합 (LSP)
   - 프로파일러

---

## 📞 문제 해결

### 빌드 오류: "cannot find module"

```bash
# go.mod 초기화
go mod init juliacc

# 의존성 다운로드
go mod tidy
```

### 테스트 오류: "undefined function"

모든 internal 패키지가 export된 인터페이스를 제공하는지 확인:

```bash
go test -v ./...
```

### 성능이 느린 경우

벤치마크를 실행하여 병목 지점 확인:

```bash
go test -bench=BenchmarkFullCompilation -cpuprofile=cpu.prof -memprofile=mem.prof ./test
go tool pprof cpu.prof
```

---

## 📚 참고 자료

- [Go Testing 문서](https://golang.org/doc/effective_go#testing)
- [Benchmarking Guide](https://golang.org/doc/effective_go#benchmark)
- [Julia Language Docs](https://docs.julialang.org)

---

**현재 버전**: 0.2.0
**마지막 업데이트**: 2026-03-19
**관리자**: Kim
