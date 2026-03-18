# Genspark Clone v3.0 업그레이드 계획

## 현재 상태 분석

**v2.0 (현재)**:
- 기본 5단계 파이프라인 구현
- 4개 에이전트 + 합의 엔진
- 캐싱 + 위젯 렌더링
- **문제**: 실제 Genspark의 핵심 차별성 미포함

**실제 Genspark의 차별성** (외부 분석):
1. **Routing Agent** - 쿼리 의존성 분석 + 병렬/순차 실행 최적화
2. **Cross-Check Agent** - 시맨틱 검증 (벡터 임베딩)
3. **Dynamic Sparkpage** - 정적 HTML이 아닌 SPA (Single Page Application)
4. **Sandboxed Execution** - Docker 격리 환경에서 코드 실행
5. **Autonomous Agent** - Multi-step 작업 자동 실행 (이메일, 스케줄링 등)
6. **Media-Rich Integration** - 이미지/비디오 동기화 생성

---

## v3.0 구현 전략

### Phase 5: RoutingAgent (의존성 기반 작업 그래프)

**목표**: 쿼리를 분석하여 병렬/순차 실행 최적화

```
의존성 분석:
  - 각 sub_query의 의존 관계 파악
  - 독립적인 쿼리: 병렬 실행
  - 종속적인 쿼리: 순차 실행 (chain-of-thought)

예시:
"Python async 라이브러리 비교" →
  - 병렬: "asyncio란", "aiohttp란", "trio란" (독립)
  - 순차: "성능 비교" (위 결과 필요)

구현: 120줄
  - QueryDependencyAnalyzer
  - ExecutionPlanner
  - ParallelExecutor + SequentialExecutor
```

### Phase 6: CrossCheckAgent (시맨틱 검증)

**목표**: 벡터 임베딩으로 정보 충돌 감지 + 신뢰도 개선

```
벡터 검증:
  - 각 소스의 핵심 내용을 벡터로 인코딩
  - 코사인 유사도로 일관성 검사
  - 이상치 탐지 (outlier detection)
  - 충돌 메시지 자동 생성

예시:
"Python 속도" →
  벡터1: "Python은 느리다" (0.8)
  벡터2: "Numba로 C속도 가능" (0.2) → 충돌!
  벡터3: "Cython으로 최적화" (0.3) → 충돌!
  → "성능 관점에서 상충하는 정보 발견"

구현: 150줄
  - VectorEmbedder (SentenceTransformer / OpenAI Embeddings)
  - SemanticValidator
  - ConflictDetector
```

### Phase 7: DynamicSparkpage (SPA 기반 렌더링)

**목표**: 정적 HTML → 동적 SPA로 전환

```
구조:
  - React/Vue 없음 (경량)
  - Vanilla JS + 웹 컴포넌트
  - 세션 기반 상태 관리
  - 실시간 업데이트 (WebSocket)

기능:
  - 토글 가능한 섹션 (expand/collapse)
  - 실시간 검색 결과 추가
  - 사용자 메모 + 북마크
  - 버전 관리 (이전 버전 복구)
  - 공유 링크 + 권한 관리

구현: 280줄
  - SparkpageController (JS)
  - RealtimeUpdater (WebSocket)
  - SessionManager
  - ShareableSparkpage (UUID + 권한)
```

### Phase 8: SandboxedExecutor (Docker 기반 코드 실행)

**목표**: 사용자 쿼리 기반 자동 코드 실행

```
격리 환경:
  - Docker 컨테이너 생성 (ephemeral)
  - 네트워크 제한 (화이트리스트만)
  - 메모리 제한 (512MB)
  - 타임아웃 (30초)
  - 자동 정리

예시:
사용자: "Python에서 데이터프레임 병합하는 예제 보여줘"
→ 샌드박스 코드 실행 → 실행 결과 + 시각화

구현: 200줄
  - DockerSandbox
  - CodeExecutor
  - SecureCodeRunner
  - OutputSanitizer
```

### Phase 9: AutonomousAgent (Multi-step 작업)

**목표**: Genspark Claw처럼 여러 단계 작업 자동 실행

```
작업 종류:
  - 이메일 검색 + 요약
  - 스케줄 등록 + 알림
  - 문서 생성 + 저장
  - 데이터 수집 + 분석
  - 코드 생성 + 테스트

작업 흐름:
  1. 의도 분석 (task understanding)
  2. 작업 계획 (planning)
  3. 단계별 실행 (execution)
  4. 검증 (verification)
  5. 결과 보고 (reporting)

구현: 250줄
  - TaskPlanner
  - StepExecutor
  - StateManager
  - ErrorRecovery
```

### Phase 10: MediaIntegration (이미지/비디오 생성)

**목표**: 검색 결과에 시각 자료 자동 추가

```
미디어 생성:
  - 개념도 (diagram generation)
  - 그래프 (chart generation)
  - 비교 표 (comparison table)
  - 아이콘 + 일러스트
  - 비디오 요약 (Synthesia 연동)

예시:
"머신러닝 알고리즘 비교" →
  + 알고리즘별 아키텍처 다이어그램
  + 성능 비교 그래프
  + 사용 사례 아이콘
  + 3분 비디오 요약

구현: 180줄
  - ImageGenerator (DALL-E / Stable Diffusion)
  - ChartBuilder (Plotly / D3.js)
  - VideoSynthesizer (Synthesia API)
  - MediaOrchestrator
```

---

## 구현 로드맵

```
Week 1 (3/25~3/31):
  Phase 5: RoutingAgent (120줄)
  Phase 6: CrossCheckAgent (150줄)

Week 2 (4/1~4/7):
  Phase 7: DynamicSparkpage (280줄)
  Phase 8: SandboxedExecutor (200줄)

Week 3 (4/8~4/14):
  Phase 9: AutonomousAgent (250줄)
  Phase 10: MediaIntegration (180줄)

Week 4 (4/15~4/21):
  통합 테스트 + 문서화
  버전 v3.0 배포
```

---

## v3.0 최종 규모

| 항목 | 줄수 |
|------|------|
| Phase 5 (Routing) | 120 |
| Phase 6 (CrossCheck) | 150 |
| Phase 7 (DynamicSparkpage) | 280 |
| Phase 8 (Sandbox) | 200 |
| Phase 9 (Autonomous) | 250 |
| Phase 10 (Media) | 180 |
| **신규 코드** | **1,180** |
| 테스트 | ~250 |
| 문서 | ~300 |
| **합계** | **1,730줄** |

**v1.0 (913) + v2.0 (1,095) + v3.0 (1,180) = 3,188줄**

---

## 핵심 차별성

### v2.0 vs v3.0

| 기능 | v2.0 | v3.0 |
|------|------|------|
| 멀티 에이전트 | ✅ (4개) | ✅ (6개 + Routing) |
| 정보 검증 | 도메인 오버랩 | ✅ 벡터 임베딩 |
| Sparkpage | 정적 HTML | ✅ 동적 SPA |
| 코드 실행 | ❌ | ✅ 샌드박스 |
| 자동 작업 | ❌ | ✅ Multi-step |
| 미디어 생성 | ❌ | ✅ 이미지/비디오 |
| 실시간 업데이트 | ❌ | ✅ WebSocket |
| 공유 + 권한 | ❌ | ✅ 세션 기반 |

---

## 외부 의존성

```
선택사항 (API):
  - OpenAI Embeddings (벡터)
  - Stable Diffusion / DALL-E (이미지)
  - Synthesia (비디오)
  - Google Calendar API (스케줄)
  - Gmail API (이메일)

필수 도구:
  - Docker (샌드박스)
  - Redis (세션)
  - WebSocket (실시간)
```

---

## 다음 단계

**Phase 5 구현 시작** (3/25)
- RoutingAgent: 쿼리 의존성 분석
- ExecutionPlanner: 병렬/순차 실행 최적화

---

**작성일**: 2026-03-18
**상태**: v3.0 계획 수립
**목표**: 실제 Genspark 수준의 enterprise-grade 엔진
