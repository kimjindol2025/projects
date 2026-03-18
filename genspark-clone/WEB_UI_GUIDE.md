# Genspark Clone - 웹 UI 가이드

## 개요

현대적인 웹 인터페이스로 Genspark Clone을 사용할 수 있습니다.

### 특징

- 🎨 **아름다운 디자인**: 그래디언트 배경, 반응형 레이아웃
- 📱 **모바일 친화적**: 모든 기기에서 완벽하게 작동
- ⚡ **빠른 응답**: 실시간 상태 표시
- 💾 **로컬 저장소**: 검색 히스토리 자동 저장
- 📥 **다운로드**: HTML, Markdown 직접 다운로드
- 👁️ **미리보기**: 브라우저에서 즉시 확인

## 설치 및 실행

### 1단계: 의존성 설치

```bash
pip install flask
```

### 2단계: API 키 설정

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 3단계: 웹 서버 시작

```bash
python web_ui.py
```

### 4단계: 브라우저에서 접속

```
http://localhost:5000
```

## 사용 방법

### 기본 검색

1. **검색 질문 입력**
   - 예: "파이썬 프로그래밍이란"
   - 예: "REST API란"
   - 예: "클라우드 컴퓨팅 기초"

2. **옵션 설정** (선택사항)
   - 언어: 한국어 / English / 혼합
   - 검색 결과: 3개 / 5개 / 10개

3. **[생성 시작] 클릭**

4. **결과 확인**
   - 신뢰도: AI 분석 신뢰도 (%)
   - 소스: 수집한 웹 소스 수
   - 섹션: 생성된 섹션 수

5. **파일 다운로드**
   - 📄 Markdown: 텍스트 에디터에서 편집 가능
   - 🌐 HTML: 브라우저에서 미리보기
   - 👁️ 미리보기: 즉시 확인

### 검색 히스토리

- 최근 10개 검색이 자동 저장됨
- 로컬 저장소 사용 (서버 저장 없음)
- 브라우저 개발자도구 → Application → Local Storage에서 확인 가능

## UI 컴포넌트

### 색상 팔레트

| 요소 | 색상 | 사용처 |
|------|------|--------|
| Primary | #667eea / #764ba2 | 버튼, 헤더, 링크 |
| Success | #388e3c | 성공 메시지 |
| Error | #d32f2f | 에러 메시지 |
| Background | #f9f9f9 | 카드 배경 |
| Border | #e0e0e0 | 구분선 |

### 반응형 디자인

```
데스크톱 (1000px+)
├─ 3컬럼 메타 정보
└─ 풀 사이즈 다운로드 링크

태블릿 (600-999px)
├─ 2컬럼 메타 정보
└─ 3컬럼 다운로드 링크

모바일 (< 600px)
├─ 1컬럼 메타 정보
└─ 스택 다운로드 링크
```

## 기술 스택

### 프론트엔드

- **HTML5**: 의미론적 마크업
- **CSS3**: 그래디언트, 애니메이션, Flexbox/Grid
- **JavaScript (Vanilla)**: 모던 JS (Promise, async/await)
- **LocalStorage**: 클라이언트 저장소

### 백엔드

- **Flask**: Python 웹 프레임워크
- **Python 3.8+**: 백엔드 로직
- **Genspark Agent**: 핵심 검색 엔진

## API 엔드포인트

### POST /api/search

**요청:**
```json
{
  "query": "파이썬이란",
  "language": "ko",
  "max_results": 5
}
```

**응답 (성공):**
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

**응답 (실패):**
```json
{
  "success": false,
  "error": "오류 메시지"
}
```

### GET /download/{filename}

파일 다운로드 (HTML 또는 Markdown)

### GET /view/{filename}

파일 미리보기 (HTML)

## 브라우저 호환성

| 브라우저 | 지원 | 비고 |
|---------|------|------|
| Chrome | ✅ | 완벽 지원 |
| Firefox | ✅ | 완벽 지원 |
| Safari | ✅ | 완벽 지원 |
| Edge | ✅ | 완벽 지원 |
| IE 11 | ❌ | 지원 안 함 |

## 성능 최적화

### 로딩 시간

```
페이지 로드: ~100ms
첫 검색: ~50초 (네트워크 + Claude API)
다음 검색: ~50초 (캐싱 없음)
```

### 최적화 팁

1. **브라우저 캐시 사용**
   - 같은 검색어 재입력 시 히스토리 클릭

2. **검색어 구체화**
   - 너무 일반적인 검색어 피하기
   - 5단어 이내 추천

3. **파일 다운로드**
   - HTML: 브라우저 저장 (용량 작음)
   - Markdown: 편집기로 수정 가능

## 문제 해결

### "ANTHROPIC_API_KEY가 설정되지 않았습니다"

```bash
# 환경변수 확인
echo $ANTHROPIC_API_KEY

# 환경변수 설정
export ANTHROPIC_API_KEY="sk-ant-..."

# 다시 실행
python web_ui.py
```

### "포트 5000이 이미 사용 중입니다"

```bash
# 다른 포트 사용
python -c "
import web_ui
web_ui.app.run(port=5001)
"
```

### "파일 다운로드가 안 됩니다"

- output/ 디렉토리 확인
- 파일 존재 여부 확인
- 브라우저 다운로드 폴더 권한 확인

## 커스터마이징

### 포트 변경

`web_ui.py` 마지막 줄:
```python
app.run(host='0.0.0.0', port=8080)  # 5000 → 8080
```

### 색상 변경

CSS의 그래디언트:
```css
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
/* #667eea와 #764ba2를 원하는 색으로 변경 */
```

### 제목 변경

HTML의 header:
```html
<h1>🔍 나만의 검색 엔진</h1>
```

## 배포

### 로컬 네트워크에서 사용

```bash
python web_ui.py
# http://192.168.1.100:5000 (같은 네트워크의 다른 기기에서 접속)
```

### 공개 배포 (권장하지 않음)

```bash
# 프로덕션 서버 사용 권장
pip install gunicorn
gunicorn -w 4 -b 0.0.0.0:5000 web_ui:app
```

## 보안 고려사항

⚠️ **주의**: 프로덕션 배포 전 다음을 확인하세요:

1. API 키 노출 방지
   - 환경변수 사용 (하드코딩 금지)
   - `.env` 파일 사용 (`.gitignore`에 추가)

2. 입력 검증
   - 현재: 기본 검증만 수행
   - 프로덕션: 더 강력한 검증 필요

3. HTTPS 사용
   - 로컬: HTTP OK
   - 공개: HTTPS 필수

4. CORS 설정
   - 현재: 모든 도메인 허용
   - 필요시 제한 추가

## 향후 개선

### v1.1 계획

- [ ] 다크 모드
- [ ] 즐겨찾기 기능
- [ ] 검색 결과 공유 (링크)
- [ ] 실시간 진행 표시

### v2.0 계획

- [ ] 사용자 계정
- [ ] 검색 히스토리 클라우드 동기화
- [ ] 팀 협업 기능
- [ ] API 문서화

## 라이선스

MIT License - 자유롭게 수정 및 배포 가능

## 피드백

문제나 제안사항이 있으면 다음을 확인하세요:

1. README.md - 기본 사용법
2. ARCHITECTURE.md - 기술 설계
3. COMPLETION_SUMMARY.md - 구현 현황

---

**🚀 Happy Searching with Genspark Clone!**
