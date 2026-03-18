# Genspark Clone 🚀

Genspark 클론: 웹 검색 + AI 합산 + Sparkpage 생성 엔진

## 주요 기능

- 🔍 **DuckDuckGo 검색**: API 키 불필요한 웹 검색
- 🤖 **Claude AI 분석**: haiku(분해) + sonnet(합산)으로 자동 분석
- 📄 **Sparkpage 생성**: 마크다운 + HTML 자동 생성
- ⚡ **병렬 처리**: 최대 3개 워커로 콘텐츠 병렬 페칭
- 📱 **Termux 최적화**: 메모리 제약 환경(1GB) 대응

## 파이프라인

```
사용자 질문
    ↓
1️⃣ 질문 분석 (Claude haiku)
    ↓ sub_queries 분해 → ['쿼리1', '쿼리2', '쿼리3']
2️⃣ 웹 검색 (DuckDuckGo)
    ↓ 검색 결과 수집
3️⃣ 콘텐츠 페칭 (병렬)
    ↓ 본문 텍스트 추출 (최대 3,000자)
4️⃣ AI 합산 (Claude sonnet)
    ↓ 섹션 + 핵심 사실 생성
5️⃣ Sparkpage 생성 (HTML + MD)
    ↓ 파일 저장
완료 ✅
```

## 설치

### 1. 환경 설정

```bash
mkdir -p ~/projects/genspark-clone/output
cd ~/projects/genspark-clone
pip install requests beautifulsoup4
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 2. 의존성

- `requests` - HTTP 요청 (웹 검색, Claude API)
- `beautifulsoup4` - HTML 파싱 (콘텐츠 추출)
- `anthropic` (선택) - 또는 requests 직접 사용

## 사용법

### CLI 실행

```bash
python main.py "파이썬 비동기 프로그래밍이란"
```

### 결과

```
🚀 Genspark Clone 시작

[INIT] 시작: '파이썬 비동기 프로그래밍이란'
[ANALYZE] 분해됨: 3개 서브쿼리, 예상 섹션: 개요, 상세
[SEARCH] 검색 완료: 15개 결과
[FETCH] 페칭 완료: 12/15 유효
[SYNTHESIZE] 합산 완료: 4개 섹션, 신뢰도 92%
[GENERATE] 생성 완료: output/20260318_120000_파이썬비동기.html

✅ 완료!
📄 Markdown: output/20260318_120000_파이썬비동기.md
🌐 HTML: output/20260318_120000_파이썬비동기.html
```

### 프로그래매틱 사용

```python
from src.genspark_agent import GensparkAgent, AgentConfig
import os

config = AgentConfig(
    anthropic_api_key=os.environ["ANTHROPIC_API_KEY"],
    output_dir="output"
)

agent = GensparkAgent(config)
result = agent.run("파이썬이란")

print(f"HTML: {result.html_path}")
print(f"MD: {result.markdown_path}")
```

## 프로젝트 구조

```
genspark-clone/
├── src/
│   ├── __init__.py                   # 패키지
│   ├── query_analyzer.py             # 질문 분석 (haiku)
│   ├── web_searcher.py               # DuckDuckGo 검색
│   ├── content_fetcher.py            # 콘텐츠 페칭 (병렬)
│   ├── claude_synthesizer.py         # AI 합산 (sonnet)
│   ├── sparkpage_generator.py        # HTML/MD 생성
│   └── genspark_agent.py             # 통합 오케스트레이터
├── main.py                           # CLI 진입점
├── test_basic.py                     # 기본 검증 테스트
├── output/                           # 생성된 파일 (자동 생성)
└── README.md                         # 이 문서
```

## 파일별 역할

| 파일 | 줄수 | 역할 |
|------|------|------|
| `query_analyzer.py` | ~150줄 | 질문 → 서브쿼리 분해 (Claude haiku) |
| `web_searcher.py` | ~180줄 | DuckDuckGo HTML 파싱 검색 |
| `content_fetcher.py` | ~220줄 | URL 크롤링 + 본문 추출 (병렬) |
| `claude_synthesizer.py` | ~200줄 | 멀티소스 합산 (Claude sonnet) |
| `sparkpage_generator.py` | ~250줄 | MD + HTML 자동 생성 |
| `genspark_agent.py` | ~200줄 | 5단계 파이프라인 오케스트레이션 |
| `main.py` | ~80줄 | CLI 인터페이스 |
| **합계** | **~913줄** | |

## 생성 결과 예시

### Markdown (output/20260318_120000.md)

```markdown
# 파이썬 비동기 프로그래밍이란

**생성일시**: 2026-03-18 12:00:00
**신뢰도**: 92%
**소스**: 12개

## 핵심 사실
- asyncio는 파이썬의 비동기 프로그래밍 라이브러리
- await/async 키워드로 비동기 코드 작성
- I/O 대기 중 다른 작업 병렬 처리 가능

## 개요
파이썬 비동기 프로그래밍은...

**출처:**
- https://docs.python.org/3/library/asyncio.html
...
```

### HTML (output/20260318_120000.html)

```html
<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <title>파이썬 비동기 프로그래밍이란 - Genspark</title>
    <style>...responsive CSS...</style>
</head>
<body>
    <div class="container">
        <header>
            <h1>파이썬 비동기 프로그래밍이란</h1>
            <div class="meta">
                <span>신뢰도: 92%</span>
                <span>소스: 12개</span>
            </div>
        </header>
        <main>
            <h2>핵심 사실</h2>
            <ul><li>asyncio는...</li>...</ul>
            ...
        </main>
    </div>
</body>
</html>
```

## 모델 선택

| 단계 | 모델 | 비용 | 속도 | 용도 |
|------|------|------|------|------|
| 분석 | `claude-haiku-4-5-20251001` | 💰 | ⚡ | 질문 분해 |
| 합산 | `claude-sonnet-4-6` | 💰💰 | ⚡⚡ | 고품질 합산 |

**목표**: haiku(빠른 분석) → sonnet(정확한 합산)으로 비용 최적화

## Termux 최적화

| 제약 | 대응 |
|------|------|
| 메모리 1GB | MAX_WORKERS=3, 콘텍스트 60K자 제한 |
| 네트워크 불안정 | timeout=8초, 실패 URL 스킵 |
| 봇 차단 | User-Agent 헤더, 서브쿼리 간 1초 딜레이 |
| JSON 파싱 실패 | 텍스트 폴백, 재시도 없음 |

## 성능 목표

- 전체 소요 시간: **60초 이내** (검색+페칭+분석)
- 문서 생성 시간: **< 5초**
- 메모리 사용: **< 100MB**
- API 호출: 최소 2회 (analyze + synthesize)

## 테스트

### 기본 검증

```bash
python test_basic.py
```

결과:
```
✅ QueryAnalyzer fallback OK
✅ DuckDuckGo 검색 OK (5개 결과)
✅ ContentFetcher OK (상태: ok)
✅ SparkpageGenerator OK
```

### E2E 테스트

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
python main.py "REST API 란"
```

## API 문서

### QuerySpec

```python
@dataclass
class QuerySpec:
    original_query: str          # 원본 질문
    main_topic: str              # 중심 주제
    sub_queries: List[str]       # 서브 검색어 (2~5개)
    language: str                # 'ko' | 'en'
    expected_sections: List[str] # 예상 섹션
    complexity: float            # 0.0~1.0
```

### SynthesisResult

```python
@dataclass
class SynthesisResult:
    query: str                   # 원본 질문
    sections: List[SparkSection] # 생성된 섹션
    key_facts: List[str]         # 핵심 사실 3~5개
    confidence_score: float      # 신뢰도 (0.0~1.0)
    total_sources: int           # 수집한 소스 수
    synthesis_model: str         # 사용한 모델명
```

### SparkSection

```python
@dataclass
class SparkSection:
    title: str                   # 섹션 제목
    content: str                 # 마크다운 본문
    sources: List[str]          # 출처 URL 리스트
    section_type: str           # 'overview' | 'detail' | 'example' | 'summary'
```

## 한계 및 개선점

### 현재 한계
- ❌ 실시간 성능 데이터 (CPU/메모리 모니터링 없음)
- ❌ 이미지/비디오 처리
- ❌ 다국어 지원 (Korean 중심)
- ❌ 캐싱 (매번 검색)

### 향후 개선
- 📌 검색 결과 캐싱 (Redis)
- 📌 이미지 추출 + 마크다운 포함
- 📌 다중 언어 지원 (en, ja, zh)
- 📌 사용자 피드백 루프 (신뢰도 개선)
- 📌 GitHub Actions 배포

## 문제해결

### DuckDuckGo 검색 반환 없음

```
⚠️ DuckDuckGo 검색 반환 없음 (네트워크 확인)
```

**해결**:
- 네트워크 연결 확인
- DuckDuckGo 접근 차단 여부 확인
- 검색어가 너무 일반적이지 않은지 확인

### Claude API 호출 실패

```
❌ Claude API 호출 실패: 401 Unauthorized
```

**해결**:
- `ANTHROPIC_API_KEY` 환경변수 확인
- API 키 유효성 확인
- API 호출 제한 확인

### 메모리 부족

```
MemoryError: Unable to allocate...
```

**해결**:
- MAX_WORKERS를 1~2로 줄임
- MAX_BODY_CHARS를 1,500자로 줄임
- 불필요한 백그라운드 프로세스 종료

## 웹 UI 사용법

### 웹 서버 시작

```bash
# 기본 포트 (5000)
python web_ui.py

# 커스텀 포트
python -c "
import web_ui
web_ui.app.run(debug=True, host='0.0.0.0', port=8080)
"
```

### 브라우저 접속

```
http://localhost:5000  (또는 커스텀 포트)
```

### 웹 UI 기능

- 🎨 **모던 디자인**: 그래디언트 배경, 반응형 레이아웃
- 📱 **모바일 최적화**: 모든 기기에서 완벽 작동
- 💾 **검색 히스토리**: LocalStorage에 자동 저장 (최근 10개)
- 📥 **다운로드**: HTML, Markdown 직접 다운로드
- 👁️ **미리보기**: 브라우저에서 즉시 확인

### API 엔드포인트

#### POST /api/search

**요청:**
```json
{
  "query": "파이썬이란",
  "language": "ko",
  "max_results": 5
}
```

**응답:**
```json
{
  "success": true,
  "data": {
    "query": "파이썬이란",
    "confidence_score": 0.92,
    "total_sources": 5,
    "sections": 4,
    "generated_at": "2026-03-18 14:30",
    "filename_html": "20260318_143000_파이썬이란.html",
    "filename_md": "20260318_143000_파이썬이란.md"
  }
}
```

---

## Genspark Clone vs 실제 Genspark

### 현재 구현 (Clone v1)

| 기능 | 상태 | 설명 |
|------|------|------|
| 단일 경로 검색 | ✅ | 질문 분석 → 검색 → 합산 |
| Sparkpage 생성 | ✅ | HTML + Markdown 자동 생성 |
| 병렬 처리 | ✅ | MAX_WORKERS=3 |
| 신뢰도 점수 | ✅ | 출처 수 기반 |
| Termux 최적화 | ✅ | 메모리 < 100MB |
| 웹 UI | ✅ | Flask 기반 |

### 실제 Genspark의 차별점 (미구현)

| 기능 | 설명 | 중요도 |
|------|------|--------|
| **Multi-Agent Orchestration** | 여러 특화 에이전트 동시 실행 | ⭐⭐⭐ |
| **교차 검증 (Fact-Checking)** | 에이전트 간 상호 검증 | ⭐⭐⭐ |
| **동적 Sparkpage** | 위젯 기반 UI 자동 구성 | ⭐⭐ |
| **Consensus Engine** | 정보 충돌 감지 및 해결 | ⭐⭐ |
| **깊이 있는 리서치** | 복잡한 주제 자동 분석 | ⭐⭐⭐ |

**한 줄 요약**: 현재는 "검색 결과를 예쁘게 정리해주는 도구", 실제는 "검증하는 리서치 팀"

---

## 프로덕션 배포

### Gunicorn 사용

```bash
pip install gunicorn
gunicorn -w 4 -b 0.0.0.0:5000 web_ui:app
```

### Systemd 서비스

```bash
# /etc/systemd/system/genspark.service
[Unit]
Description=Genspark Clone API Server
After=network.target

[Service]
Type=notify
User=www-data
WorkingDirectory=/path/to/genspark-clone
Environment="ANTHROPIC_API_KEY=sk-ant-..."
ExecStart=/usr/bin/gunicorn -w 4 -b 0.0.0.0:5000 web_ui:app
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable genspark
sudo systemctl start genspark
```

---

## 라이선스

MIT License

## 작성자

Kim - 2026-03-18

---

## 관련 리소스

- 📚 [ARCHITECTURE.md](ARCHITECTURE.md) - 상세 아키텍처 설명
- ⚡ [QUICK_START.md](QUICK_START.md) - 5분 시작 가이드
- 🌐 [WEB_UI_GUIDE.md](WEB_UI_GUIDE.md) - 웹 UI 완전 가이드
- ✅ [COMPLETION_SUMMARY.md](COMPLETION_SUMMARY.md) - 구현 현황
