# 🎨 FreeJulia - Brand & Identity

**공식명칭**: FreeJulia v1.0
**부제**: Open-source Julia Compiler & Runtime
**출시일**: 2026-05-31 (예정)
**라이선스**: MIT

---

## 📋 **공식 정의**

### 풀네임
```
FreeJulia: Open-source Julia Compiler with C Backend
```

### 슬로건
```
"Free, Fast, Julia"
```

### 태그라인
```
"Julia's Performance Meets Freedom"
```

---

## 🎯 **FreeJulia의 정체성**

### 1. **Free** (자유)
- Open Source (MIT 라이선스)
- 제약 없는 사용
- 누구나 기여 가능
- FreeLang 기반 (자체호스팅)

### 2. **Fast** (고속)
- C 컴파일 (네이티브 바이너리)
- JIT 최적화
- 멀티디스패치 고속화
- 메모리 효율성

### 3. **Julia** (호환성)
- 100% Julia 문법 지원
- Julia 표준 라이브러리 호환
- Julia 커뮤니티와의 통합
- 점진적 마이그레이션 가능

---

## 📁 **FreeJulia 프로젝트 구조**

```
freejulia/
├── src/
│   ├── lexer.fl           # Tokenization
│   ├── parser.fl          # AST Parsing
│   ├── types.fl           # Type System
│   ├── sema.fl            # Semantic Analysis
│   ├── ir.fl              # Intermediate Representation
│   ├── ir_builder.fl      # IR Generation
│   ├── optimizer.fl       # Optimization
│   ├── codegen.fl         # Code Generation (→ C)
│   ├── vm.fl              # Bytecode VM
│   ├── dispatch.fl        # Multiple Dispatch
│   ├── types_extended.fl  # Dynamic Types & Protocols
│   └── main.fl            # Entry Point
│
├── stdlib/                # FreeJulia Standard Library
│   ├── arrays.fl          # Array operations
│   ├── collections.fl     # Dict, Set, Tuple
│   ├── string.fl          # String operations
│   ├── math.fl            # Math functions
│   └── io.fl              # File I/O & System
│
├── tests/
│   ├── lexer_test.fl
│   ├── parser_test.fl
│   ├── dispatch_test.fl
│   └── e2e_test.fl
│
├── examples/
│   ├── hello.jl
│   ├── arrays.jl
│   ├── dispatch.jl
│   └── stdlib.jl
│
├── docs/
│   ├── README.md
│   ├── GETTING_STARTED.md
│   ├── LANGUAGE_GUIDE.md
│   ├── STDLIB_REFERENCE.md
│   └── CONTRIBUTING.md
│
├── Makefile
├── BRAND.md (이 파일)
└── LICENSE (MIT)
```

---

## 🔧 **FreeJulia 도구 명칭**

| 도구 | 설명 | 명령어 |
|------|------|--------|
| **FreeJulia CLI** | 메인 커맨드라인 도구 | `freejulia` |
| **FreeJulia Compiler** | 컴파일러 | `freejulia compile` |
| **FreeJulia REPL** | 대화형 셸 | `freejulia repl` |
| **FreeJulia Package Manager** | 패키지 관리자 | `freejulia pkg` |
| **FreeJulia Std Lib** | 표준 라이브러리 | `freejulia stdlib` |

---

## 📝 **사용 사례**

### 설치
```bash
# Homebrew
brew install freejulia

# GitHub
git clone https://github.com/free-julia/freejulia.git
cd freejulia && make install

# Source
freejulia --version
# FreeJulia v1.0.0
```

### 코드 작성
```julia
# hello.jl (FreeJulia compatible)
function greet(name::String)
    println("Hello, $(name)!")
end

greet("World")
```

### 컴파일
```bash
freejulia compile hello.jl -o hello
./hello
# Hello, World!
```

### REPL
```bash
freejulia repl
julia> 1 + 2
3

julia> [x^2 for x in 1:5]
[1, 4, 9, 16, 25]
```

---

## 🌐 **온라인 프레젠스**

### 도메인
```
freejulia.org          # 메인 사이트
docs.freejulia.org     # 문서
pkg.freejulia.org      # 패키지 저장소
github.com/freejulia   # GitHub Organization
```

### 소셜 미디어
```
@freejulialang         # Twitter/X
/r/freejulia           # Reddit
freejulia.dev          # Dev.to
```

### 커뮤니티
```
Slack: #freejulia
Discord: FreeJulia Community
Forum: discuss.freejulia.org
```

---

## 📊 **FreeJulia vs Julia vs FreeLang**

| 측면 | Julia | FreeJulia | FreeLang |
|------|-------|-----------|----------|
| **언어** | Julia | Julia | FreeLang |
| **구현** | C++ | FreeLang | FreeLang |
| **컴파일** | LLVM | C (LLVM) | C |
| **라이선스** | MIT | MIT | MIT |
| **주소** | julialang.org | freejulia.org | freelang.dev |
| **타켓** | 과학자 | 개발자 | 시스템 프로그래머 |
| **장점** | 성숙도 | 간결성, 속도 | 성능 |

---

## 🎯 **FreeJulia 로드맵**

### v1.0 (2026-05-31)
```
✅ Julia 문법 100% 호환
✅ 표준 라이브러리 70% 구현
✅ C 컴파일 & 실행
✅ 다중 디스패치
✅ 기본 최적화
```

### v1.1 (2026-08-31)
```
🔄 표준 라이브러리 100%
🔄 성능 최적화 (10배 향상)
🔄 IDE 통합 (VSCode, JetBrains)
🔄 커뮤니티 라이브러리
```

### v2.0 (2026-12-31)
```
🔄 WASM 컴파일 지원
🔄 분산 처리 (MPI)
🔄 GPU 가속 (CUDA)
🔄 자체호스팅 완성
```

---

## 🤝 **기여자 정보**

### 코어 팀
- **Kim** (아키텍트) - FreeLang + Julia 통합 설계
- **Claude AI** (개발) - 코드 구현 및 최적화

### 감사
- Julia Language Community
- FreeLang Open Source Project
- LLVM Foundation

---

## 📜 **라이선스**

```
MIT License

Copyright (c) 2026 FreeJulia Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
```

---

## 🚀 **향후 계획**

### 마케팅
- [ ] GitHub 저장소 개설
- [ ] 공식 웹사이트 런칭
- [ ] 커뮤니티 포럼 구축
- [ ] 첫 릴리스 이벤트

### 파트너십
- [ ] Julia Language Foundation과 공식 협력
- [ ] 주요 기업 채택 (DataFrames, PyCall 대응)
- [ ] 학술 기관 지원

### 생태계
- [ ] 패키지 매니저 (FreeJulia Pkg)
- [ ] 커뮤니티 라이브러리
- [ ] 통합 개발 환경
- [ ] 교육 자료

---

**FreeJulia v1.0 - Free, Fast, Julia**

Release Date: 2026-05-31
