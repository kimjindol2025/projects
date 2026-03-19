# 🚀 FreeLang Julia Compiler

[![Language](https://img.shields.io/badge/language-FreeLang-blue.svg)](#)
[![Status](https://img.shields.io/badge/status-Development-blue.svg)](#)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

**FreeLang으로 구현된 Julia 컴파일러**

---

## 📋 개요

Julia 컴파일러를 Go에서 FreeLang으로 재구현한 프로젝트입니다.

**목표**:
- ✅ Julia 구문 파싱
- ✅ 타입 추론
- ✅ 중간 표현 (IR) 생성
- ✅ 바이트코드 생성 및 실행

**특징**:
- 8단계 컴파일 파이프라인 (Lexing → VM Execution)
- 동적 타입 시스템
- 다중 디스패치 지원
- 고성능 바이트코드 VM

---

## 🏗️ 아키텍처

```
Source Code (.jl)
      ↓
   [Lexer] → Tokens
      ↓
  [Parser] → AST
      ↓
[Semantic Analyzer] → Type-Checked AST
      ↓
  [IR Builder] → IR Module
      ↓
[Type Inference] → Typed IR
      ↓
 [Optimizer] → Optimized IR
      ↓
[Code Generator] → Bytecode
      ↓
   [VM] → Result
```

---

## 🚀 빠른 시작

### 빌드

```bash
# FreeLang으로 컴파일
freelang build -o julia src/main.fl

# 또는 기존 바이너리 사용
./julia
```

### 실행

```bash
# 간단한 코드
julia hello.jl

# 디버그 모드
julia -debug hello.jl

# 최적화 빌드
julia -O hello.jl
```

### 예제

**hello.jl**:
```julia
function add(a, b)
    return a + b
end

result = add(2, 3)
print(result)  # 5
```

```bash
julia hello.jl
# 출력: 5
```

---

## 📂 프로젝트 구조

```
freelang-julia/
├── src/
│   ├── lexer.fl          # 토크나이제이션 (420줄)
│   ├── parser.fl         # AST 파싱 (550줄)
│   ├── types.fl          # 타입 시스템 (280줄)
│   ├── sema.fl           # 의미 분석 (620줄)
│   ├── ir.fl             # IR 정의 (200줄)
│   ├── ir_builder.fl     # IR 생성 (320줄)
│   ├── optimizer.fl      # 최적화 (300줄)
│   ├── codegen.fl        # 코드 생성 (500줄)
│   ├── vm.fl             # VM 실행 (400줄)
│   └── main.fl           # 진입점 (150줄)
├── tests/
│   ├── lexer_test.fl     (18 tests)
│   ├── parser_test.fl    (14 tests)
│   ├── types_test.fl     (12 tests)
│   ├── sema_test.fl      (20 tests)
│   ├── ir_test.fl        (15 tests)
│   ├── codegen_test.fl   (12 tests)
│   ├── e2e_test.fl       (8 tests)
│   └── benchmark.fl      (성능 벤치마크)
├── examples/
│   ├── hello.jl
│   ├── arithmetic.jl
│   ├── fibonacci.jl
│   └── ...
├── docs/
│   ├── ARCHITECTURE.md
│   ├── API.md
│   └── BUILD.md
└── README.md
```

---

## ✅ 개발 진행 상황

### Phase 1: 설계 & 분석 (Week 1)
- [x] Julia 컴파일러 구조 분석
- [ ] FreeLang 기능 평가
- [ ] 이식 로드맵 작성

**상태**: 🟢 진행 중 (계획 문서 완성)

### Phase 2: 핵심 모듈 이식 (Week 2-3)
- [ ] Lexer 이식
- [ ] Parser 이식
- [ ] Type System 이식

### Phase 3: 분석 & 생성 (Week 4-5)
- [ ] Semantic Analyzer 이식
- [ ] IR Builder 이식
- [ ] Code Generator 이식

### Phase 4: 통합 & 최적화 (Week 6)
- [ ] 통합 테스트
- [ ] 성능 최적화
- [ ] 문서화

---

## 📊 통계

| 지표 | 예상 |
|------|------|
| **코드량** | 3,590줄 (FreeLang) |
| **테스트** | 90+ 테스트 |
| **모듈** | 10개 |
| **완료 예정** | 2026-04-30 |

---

## 🔗 관련 프로젝트

- **원본**: [Julia Compiler (Go)](../julia-compiler) - v0.2.0
- **기반**: [FreeLang](../freelang-to-c) - 자체호스팅 증명
- **계획**: [freelang-julia-porting-plan.md](../.claude/projects/-data-data-com-termux-files-home/memory/freelang-julia-porting-plan.md)

---

## 📝 라이선스

MIT License © 2026

---

## 📞 문의

- **메모리**: [freelang-julia-porting-plan.md]
- **상태**: 🟢 계획 수립 완료 (2026-03-19)
- **다음**: Phase 1 작업 시작

---

**현재 버전**: 0.1.0 (계획)
**최종 업데이트**: 2026-03-19
**상태**: 🟢 Development
