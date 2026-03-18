# 빠른 시작 가이드

## 1단계: 설치 (2분)

```bash
# 프로젝트 디렉토리로 이동
cd ~/projects/genspark-clone

# 의존성 설치
pip install -r requirements.txt
```

## 2단계: API 키 설정 (1분)

```bash
# Anthropic API 키 설정
export ANTHROPIC_API_KEY="sk-ant-..."

# 확인
echo $ANTHROPIC_API_KEY
```

## 3단계: 실행 (5분)

### CLI로 실행
```bash
python main.py "파이썬이란"
```

### 결과 확인
```
✅ 완료!
📄 Markdown: output/20260318_120000_파이썬.md
🌐 HTML: output/20260318_120000_파이썬.html
```

## 테스트 (3분)

### 기본 검증 (API 키 불필요)
```bash
python test_basic.py
```

### 통합 테스트 (API 키 필수)
```bash
python test_integration.py
```

## 파일 구조

```
genspark-clone/
├── src/                      # 핵심 모듈
│   ├── query_analyzer.py     # 질문 분석
│   ├── web_searcher.py       # 웹 검색
│   ├── content_fetcher.py    # 콘텐츠 추출
│   ├── claude_synthesizer.py # AI 합산
│   └── sparkpage_generator.py # HTML/MD 생성
├── main.py                   # CLI 진입점
├── output/                   # 생성 결과
├── README.md                 # 상세 설명서
├── ARCHITECTURE.md           # 아키텍처 설계
└── requirements.txt          # 의존성
```

## 주요 기능

### 1. 자동 검색 분해
입력: `"파이썬이란"`
→ 출력: `["파이썬 개요", "파이썬 기능", "파이썬 활용"]`

### 2. 웹 검색 (API 키 불필요)
각 검색어로 DuckDuckGo 검색 → 최대 5개 결과

### 3. 콘텐츠 추출 (병렬)
URL → 본문 텍스트 추출 (최대 3K자)

### 4. AI 분석
멀티소스 콘텐츠 → Claude가 구조화

### 5. Sparkpage 생성
- `output/*.md` (마크다운)
- `output/*.html` (반응형 HTML)

## 프로그래밍 사용 예시

```python
from src.genspark_agent import GensparkAgent, AgentConfig
import os

# 설정
config = AgentConfig(
    anthropic_api_key=os.environ["ANTHROPIC_API_KEY"],
    output_dir="output"
)

# 실행
agent = GensparkAgent(config)
result = agent.run("Docker란")

# 결과 확인
print(f"HTML: {result.html_path}")
print(f"Markdown: {result.markdown_path}")
print(f"신뢰도: {result.confidence_score:.0%}")
```

## 문제해결

### API 키 오류
```
❌ ANTHROPIC_API_KEY 환경변수가 설정되지 않았습니다
```
→ `export ANTHROPIC_API_KEY="sk-ant-..."`

### 검색 결과 없음
```
⚠️ DuckDuckGo 검색 반환 없음
```
→ 네트워크 연결 확인, 검색어 변경

### 메모리 부족
```
MemoryError: Unable to allocate...
```
→ main.py에서 `MAX_WORKERS=1`, `MAX_BODY_CHARS=1500` 변경

## 더 알아보기

- **README.md**: 기능 및 사용법
- **ARCHITECTURE.md**: 설계 및 데이터 흐름
- **COMPLETION_SUMMARY.md**: 구현 현황

---

**🚀 5분 안에 첫 Sparkpage 생성 완료!**
