# Genspark Clone - 아키텍처 설계

## 개요

Genspark 클론은 5단계 파이프라인으로 사용자 질문을 자동으로 분석하고 종합 정보 페이지(Sparkpage)를 생성합니다.

## 전체 파이프라인

```
┌─────────────────────────────────────────────────────────────────┐
│ 사용자 질문 "파이썬 비동기 프로그래밍이란?"                      │
└────────────────────────────┬────────────────────────────────────┘
                             │
                ┌────────────▼─────────────┐
                │ 1️⃣  Query Analyzer      │
                │    (claude-haiku)       │
                │ 질문 분해              │
                └────────────┬─────────────┘
                             │
                  QuerySpec(sub_queries=[ ])
                             │
                ┌────────────▼─────────────┐
                │ 2️⃣  Web Searcher       │
                │   (DuckDuckGo)         │
                │ 검색 실행              │
                └────────────┬─────────────┘
                             │
               Dict[str, List[SearchResult]]
                             │
                ┌────────────▼──────────────┐
                │ 3️⃣  Content Fetcher     │
                │  (병렬, MAX_WORKERS=3)  │
                │ 콘텐츠 추출             │
                └────────────┬──────────────┘
                             │
                List[FetchedContent] (크기 제한 3K)
                             │
                ┌────────────▼──────────────┐
                │ 4️⃣  Claude Synthesizer  │
                │    (claude-sonnet)      │
                │ AI 합산 + 구조화        │
                └────────────┬──────────────┘
                             │
                    SynthesisResult
                  (sections, key_facts)
                             │
                ┌────────────▼──────────────┐
                │ 5️⃣  Sparkpage Generator │
                │ HTML + Markdown 생성    │
                └────────────┬──────────────┘
                             │
             ┌───────────────┴───────────────┐
             │                               │
        *.html                          *.md
             │                               │
        output/20260318_120000.html    output/20260318_120000.md
```

## 단계별 상세 설명

### 1️⃣ Query Analyzer (`src/query_analyzer.py`)

**역할**: 사용자 질문을 구조화된 검색 계획으로 변환

**입력**: 문자열
```
"파이썬 비동기 프로그래밍이란"
```

**출력**: QuerySpec
```python
QuerySpec(
    original_query="파이썬 비동기 프로그래밍이란",
    main_topic="파이썬 비동기",
    sub_queries=[
        "파이썬 asyncio",
        "파이썬 async await",
        "파이썬 비동기 예제"
    ],
    language="ko",
    expected_sections=["개요", "상세", "예제", "결론"],
    complexity=0.7
)
```

**핵심 메서드**:
- `analyze(user_query: str) -> QuerySpec`
- `_call_claude(prompt: str) -> str` (requests 직접 호출)
- `_parse_claude_response(response_text: str) -> QuerySpec`

**모델**: `claude-haiku-4-5-20251001`

**최적화**:
- ✅ 빠른 응답 (haiku)
- ✅ 저비용
- ❌ API 실패 시 fallback 제공

---

### 2️⃣ Web Searcher (`src/web_searcher.py`)

**역할**: DuckDuckGo HTML을 파싱해 검색 결과 수집

**입력**: List[str]
```python
["파이썬 asyncio", "파이썬 async await", ...]
```

**출력**: Dict[str, List[SearchResult]]
```python
{
    "파이썬 asyncio": [
        SearchResult(
            url="https://docs.python.org/...",
            title="asyncio — Asynchronous I/O",
            snippet="...",
            source_domain="docs.python.org",
            rank=1
        ),
        ...
    ],
    ...
}
```

**핵심 메서드**:
- `search(query: str) -> List[SearchResult]`
- `search_multiple(queries: List[str], delay: float = 1.0)`
- `_parse_results(html: str) -> List[SearchResult]`
- `_extract_domain(url: str) -> str`

**특징**:
- ✅ API 키 불필요
- ✅ 서브쿼리 간 1초 딜레이 (봇 차단 방지)
- ✅ User-Agent 헤더 포함
- ❌ 네트워크 불안정 시 결과 없음

---

### 3️⃣ Content Fetcher (`src/content_fetcher.py`)

**역할**: URL에서 본문 텍스트 추출 (병렬)

**입력**: List[SearchResult]
```python
[
    SearchResult(url="https://...", title="...", ...),
    ...
]
```

**출력**: List[FetchedContent]
```python
[
    FetchedContent(
        url="https://docs.python.org/...",
        title="asyncio — Asynchronous I/O",
        body_text="asyncio는 파이썬에서 비동기...", # 최대 3,000자
        word_count=1250,
        fetch_status="ok"
    ),
    FetchedContent(
        url="https://blocked.site",
        title="...",
        body_text="",
        word_count=0,
        fetch_status="blocked"  # 403/429
    ),
    ...
]
```

**핵심 메서드**:
- `fetch(url: str, title: str = "") -> FetchedContent`
- `fetch_all(results: List) -> List[FetchedContent]`
- `fetch_for_queries(search_results: Dict) -> List[FetchedContent]`
- `_extract_body(soup: BeautifulSoup) -> str`

**병렬화**:
```python
MAX_WORKERS = 3  # Termux 메모리 제약
with ThreadPoolExecutor(max_workers=self.MAX_WORKERS) as executor:
    futures = {executor.submit(self.fetch, ...): r for r in results}
    for future in as_completed(futures):
        content = future.result()
```

**에러 핸들링**:
- timeout (8초) → `fetch_status="timeout"`
- 403/429 → `fetch_status="blocked"`
- 기타 예외 → `fetch_status="error"`

---

### 4️⃣ Claude Synthesizer (`src/claude_synthesizer.py`)

**역할**: 멀티소스 콘텐츠를 AI가 분석해 구조화

**입력**:
```python
(QuerySpec, List[FetchedContent])
```

**출력**: SynthesisResult
```python
SynthesisResult(
    query="파이썬 비동기 프로그래밍이란",
    sections=[
        SparkSection(
            title="개요",
            content="파이썬 비동기 프로그래밍은...",
            sources=["https://docs.python.org", ...],
            section_type="overview"
        ),
        SparkSection(
            title="asyncio 사용법",
            content="## async/await 키워드\n\n...",
            sources=[...],
            section_type="detail"
        ),
        ...
    ],
    key_facts=[
        "asyncio는 파이썬의 비동기 라이브러리",
        "await는 I/O 대기 중 제어권 양보",
        "event loop이 태스크 스케줄링"
    ],
    confidence_score=0.88,
    total_sources=12,
    synthesis_model="claude-sonnet-4-6"
)
```

**핵심 메서드**:
- `synthesize(query_spec: QuerySpec, contents: List[FetchedContent])`
- `_call_claude(system_prompt, query_spec, context) -> str`
- `_build_context(contents) -> str` (최대 60K자, Termux 최적화)
- `_parse_synthesis(response_text) -> SynthesisResult`

**모델**: `claude-sonnet-4-6`

**프롬프트 구조**:
```
System: 전문 웹 리서처 역할 정의 + 요구사항

User:
- 원본 질문
- [출처 1] 제목
- 본문...
- ---
- [출처 2] 제목
- 본문...
- ...

응답 요청: JSON 형식 (key_facts, sections)
```

---

### 5️⃣ Sparkpage Generator (`src/sparkpage_generator.py`)

**역할**: SynthesisResult를 마크다운 + HTML 파일로 생성

**입력**: (SynthesisResult, query: str)

**출력**: SparkpageOutput
```python
SparkpageOutput(
    markdown_path="output/20260318_120000_파이썬비동기.md",
    html_path="output/20260318_120000_파이썬비동기.html",
    markdown_content="# 파이썬...",
    html_content="<!DOCTYPE html>...",
    title="파이썬 비동기 프로그래밍이란",
    generated_at="20260318_120000"
)
```

**핵심 메서드**:
- `generate(result: SynthesisResult, query: str) -> SparkpageOutput`
- `_generate_markdown(result, query) -> str`
- `_generate_html(result, query, markdown) -> str`
- `_markdown_to_html(markdown) -> str` (외부 라이브러리 없음)
- `_generate_html_template(body, title, meta) -> str`
- `_slug(text) -> str` (파일명 안전화)

**생성 파일**:

Markdown:
```markdown
# 파이썬 비동기 프로그래밍이란

**생성일시**: 2026-03-18 12:00:00
**신뢰도**: 88%
**소스**: 12개

## 핵심 사실
- asyncio는 파이썬의 비동기 라이브러리
- ...

## 개요
...

**출처:**
- https://...
```

HTML:
```html
<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <title>파이썬 비동기... - Genspark</title>
    <style>
        /* Responsive CSS, Dark mode 지원 */
        body { font-family: -apple-system, ... }
        h1 { color: #0056b3; }
        ...
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>파이썬 비동기 프로그래밍이란</h1>
            <div class="meta">
                <span>신뢰도: 88%</span>
                <span>소스: 12개</span>
            </div>
        </header>
        <main>...</main>
    </div>
</body>
</html>
```

---

## 데이터 흐름

```
QuerySpec
├── original_query: str
├── main_topic: str
├── sub_queries: List[str]
├── language: str
├── expected_sections: List[str]
└── complexity: float

        ↓ (Step 2-3)

Dict[str, List[SearchResult]]
├── "쿼리1": [
│   ├── SearchResult(url, title, snippet, domain, rank)
│   └── SearchResult(...)
├── "쿼리2": [...]
└── "쿼리3": [...]

        ↓ (병렬)

List[FetchedContent]
├── FetchedContent(url, title, body_text, word_count, status)
├── FetchedContent(...)
└── FetchedContent(...)

        ↓ (Step 4)

SynthesisResult
├── sections: [
│   ├── SparkSection(title, content, sources, type)
│   └── SparkSection(...)
├── key_facts: List[str]
├── confidence_score: float
├── total_sources: int
└── synthesis_model: str

        ↓ (Step 5)

SparkpageOutput
├── markdown_path: str
├── html_path: str
├── markdown_content: str
├── html_content: str
├── title: str
└── generated_at: str
```

## 에러 처리 전략

| 계층 | 에러 | 처리 |
|------|------|------|
| Query Analyzer | JSON 파싱 실패 | fallback_spec (기본값 반환) |
| Web Searcher | 네트워크 에러 | 빈 리스트 반환 |
| Content Fetcher | timeout/403/500 | `fetch_status` 마킹 + 폴백 |
| Claude Synthesizer | API 실패 | JSON "{}" 반환 + fallback |
| Sparkpage Generator | 파일 쓰기 실패 | 예외 전파 (치명적) |

## 성능 최적화

### 메모리 절감 (Termux 1GB 제약)
```python
MAX_WORKERS = 3           # 병렬 워커 제한
MAX_BODY_CHARS = 3000     # 콘텐츠 크기 제한
max_context = 60000       # Claude 컨텍스트 크기 제한 (약 15K 토큰)
```

### 네트워크 최적화
```python
timeout = 8               # 초 (빠른 실패)
search_delay = 1.0        # 서브쿼리 간 딜레이 (봇 차단 방지)
headers["User-Agent"] = "..." # 봇으로 보이지 않기
```

### 시간 절감
```
Step 1: QueryAnalyzer     ~2초 (API 호출)
Step 2: WebSearcher       ~5초 (3개 쿼리 × 1초 지연)
Step 3: ContentFetcher    ~20초 (3개 워커 × 15개 URL ÷ 3)
Step 4: Synthesizer       ~20초 (API 호출)
Step 5: Generator         ~1초 (파일 쓰기)
────────────────────────────────────
합계                      ~48초 (목표: 60초 이내)
```

## 확장성 고려사항

### 현재 아키텍처 한계
- ✗ 단일 스레드 CLI (동시 요청 불가)
- ✗ 캐싱 없음 (매번 검색)
- ✗ 이미지/동영상 미지원
- ✗ 정적 파일만 생성 (DB 없음)

### 향후 확장 (v2.0)
- API 서버 (FastAPI/Flask)
- Redis 캐싱 (검색 결과)
- 데이터베이스 (생성 히스토리)
- 이미지 추출 (마크다운 포함)
- 웹훅 (비동기 생성)

## 설정 파일

### AgentConfig

```python
@dataclass
class AgentConfig:
    anthropic_api_key: str                    # Claude API 키
    analyze_model: str = "claude-haiku-4-5..."  # 질문 분석 모델
    synthesize_model: str = "claude-sonnet-4-6" # 합산 모델
    max_search_results: int = 5               # 쿼리당 검색 결과 수
    max_fetch_workers: int = 3                # 병렬 워커 수
    output_dir: str = "output"                # 출력 디렉토리
    search_delay: float = 1.0                 # 검색 딜레이 (초)
    verbose: bool = True                      # 로그 출력
```

## 테스트 전략

### 단위 테스트 (test_basic.py)
- ✅ QueryAnalyzer fallback
- ✅ DuckDuckGoSearcher 파싱
- ✅ ContentFetcher 페칭
- ✅ SparkpageGenerator HTML 생성

### 통합 테스트 (test_integration.py)
- ✅ DuckDuckGo 검색 (API 불필요)
- ✅ 콘텐츠 페칭 (API 불필요)
- ✅ 전체 파이프라인 (API 필요)

### E2E 테스트
```bash
python main.py "테스트 쿼리"
# 확인: output/*.html, output/*.md 생성
```

## 결론

Genspark Clone은 **5단계 파이프라인**으로 구현된 경량 웹 정보 종합 시스템입니다.

- 🎯 **목표**: 사용자 질문 → Sparkpage (자동)
- 🚀 **성능**: 60초 이내, < 100MB 메모리
- 💰 **비용**: haiku + sonnet으로 최적화
- 📱 **대상**: Termux, 모바일 환경

**아키텍처 철학**:
> "단순함을 유지하면서도 강력함을 제공한다"
