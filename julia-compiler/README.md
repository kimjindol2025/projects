# 🔧 Julia Compiler

[![Language](https://img.shields.io/badge/language-Rust-orange.svg)](#)
[![Status](https://img.shields.io/badge/status-Development-blue.svg)](#)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-kimjindol2025%2Fjulia--compiler-blue?logo=github)](https://github.com/kimjindol2025/julia-compiler)

**Julia 언어의 고성능 컴파일러 구현**

---

## 📋 목차

- [개요](#개요)
- [특징](#특징)
- [빠른 시작](#빠른-시작)
- [지원 기능](#지원-기능)
- [예제](#예제)
- [아키텍처](#아키텍처)
- [빌드](#빌드)
- [라이선스](#라이선스)

---

## 개요

**Julia Compiler**는 Rust로 구현된 Julia 언어의 컴파일러입니다. Julia의 다중 디스패치, JIT 컴파일, 성능 최적화를 목표로 합니다.

**목표**:
- ✅ Julia 구문 파싱
- ✅ 타입 추론
- ✅ LLVM 코드 생성
- ✅ JIT 컴파일
- ✅ 고성능 실행

---

## 특징

- 동적 타입 시스템
- 다중 디스패치 (Multiple Dispatch)
- JIT 컴파일
- LLVM 백엔드
- 표준 라이브러리 지원

---

## 빠른 시작

```bash
git clone https://github.com/kimjindol2025/julia-compiler.git
cd julia-compiler
cargo build --release
```

---

## 빌드

```bash
cargo build --release
cargo test
```

---

## 라이선스

MIT License © 2026

---

**현재 버전**: 0.1.0
**최종 업데이트**: 2026-03-16
**상태**: 🟡 개발 중
