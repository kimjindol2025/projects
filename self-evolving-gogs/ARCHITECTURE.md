# Self-Evolving Gogs — FreeLang Native Architecture

## 철학
> "언어가 스스로를 진화시킨다"
> 모든 지능은 .fl 안에 있다. C는 오직 시간 측정 런타임만.

## 진화 루프
```
[Gogs Push]
    ↓
[webhook_server.fl]     — HTTP webhook 수신
    ↓
[hook_injector.fl]      — .fl 소스에 __fl_profile_enter/exit 주입
    ↓
[빌드 & 실행]            — 주입된 .fl → C → 바이너리 실행
    ↓
[metrics_reader.fl]     — /tmp/fl_metrics.csv 파싱
    ↓
[complexity_analyzer.fl]— 소스에서 O(n²) 패턴 탐지
    ↓
[evolution_decider.fl]  — "진화가 필요한가?" 판단
    ↓ YES
[code_mutator.fl]       — AST 레벨 변형 (소스 재작성)
    ↓
[gogs_client.fl]        — 자동 커밋 (시스템 자아의 메시지)
    ↓
[다시 루프]
```

## 파일 구조
```
self-evolving-gogs/
├── src/
│   ├── hook_injector.fl       # .fl 소스 → 프로파일링 버전 생성
│   ├── metrics_reader.fl      # 실행 메트릭 파싱
│   ├── complexity_analyzer.fl # O(n²) 등 복잡도 탐지
│   ├── evolution_decider.fl   # 진화 여부 결정
│   ├── code_mutator.fl        # 코드 변형 엔진
│   ├── gogs_client.fl         # Gogs HTTP API 커밋
│   ├── webhook_server.fl      # Webhook 수신기
│   └── main.fl                # 진화 루프 오케스트레이터
├── runtime/
│   ├── fl_profile.h           # __fl_profile_enter/exit 선언
│   └── fl_profile.c           # 실제 시간 측정 (유일한 C 파일)
├── fl_sources/
│   └── sample.fl              # 진화 대상 샘플
└── Makefile
```

## 커밋 메시지 형식 (시스템 자아)
```
[SELF-EVOLVE 0x{id}] {pattern} detected

Observation: {func_name} avg={avg_ns}ns, calls={count}
Root cause:  {source_file}:{line} — {pattern_name}
Confidence:  {confidence}%

Action: {mutation_description}
Proof:  before={before_ns}ns → after={after_ns}ns (estimated)

This commit was not written by a human.
```
