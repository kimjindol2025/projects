# 🏗️ GitHub Pages + 포트 8253 서버 백엔드 아키텍처

## 시스템 구조

```
┌─────────────────────────────────────────────────────┐
│         GitHub Pages (정적 홈페이지)                  │
│  https://kimjindol2025.github.io/freelang-blog     │
├─────────────────────────────────────────────────────┤
│  index.html (정적 HTML)                             │
│  - 반응형 UI (카드 그리드)                            │
│  - JavaScript (Fetch API)                           │
│  - CSS (Gradient 배경)                              │
└────────────────┬────────────────────────────────────┘
                 │
        HTTP 요청 (Fetch)
                 │
                 ▼
┌──────────────────────────────────────────────────────┐
│      포트 8253 Node.js 블로그 서버 백엔드              │
│    (로컬 또는 클라우드 호스팅)                         │
├──────────────────────────────────────────────────────┤
│  GET /api/posts          → JSON 포스트 목록         │
│  GET /post/{slug}        → 포스트 상세 (HTML)       │
│  GET /health             → 헬스체크                  │
│                                                      │
│  Node.js server.js                                   │
│  ├─ 89개 .md 파일 로드                               │
│  ├─ YAML frontmatter 파싱                            │
│  ├─ Markdown → HTML 변환                             │
│  └─ CORS 활성화                                      │
└──────────────────────────────────────────────────────┘
```

---

## 배포 시나리오

### 시나리오 1️⃣: 로컬 개발
```
┌─────────────────────────────────┐
│  로컬 브라우저                    │
│  http://localhost:8253/         │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│  로컬 포트 8253 서버              │
│  (Node.js)                      │
└─────────────────────────────────┘
```

실행:
```bash
cd projects/fl-blog
node server.js :8253
```

### 시나리오 2️⃣: GitHub Pages + 클라우드 백엔드
```
┌──────────────────────────────────┐
│  GitHub Pages                    │
│  (정적 HTML - index.html)         │
└────────────┬─────────────────────┘
             │
             │ Fetch API
             │
             ▼
┌──────────────────────────────────┐
│  AWS/Azure/GCP 포트 8253         │
│  (Node.js 서버)                   │
└──────────────────────────────────┘
```

### 시나리오 3️⃣: Vercel/Netlify 호스팅
```
┌──────────────────────────────────┐
│  Vercel/Netlify                  │
│  (정적 + 서버리스)                 │
└────────────┬─────────────────────┘
             │
             ▼
┌──────────────────────────────────┐
│  자체 포트 8253 서버               │
└──────────────────────────────────┘
```

---

## 통신 흐름

### 1. 홈페이지 로드
```
1. 사용자가 GitHub Pages 방문
2. index.html 다운로드 (정적)
3. JavaScript 실행
4. fetch('http://localhost:8253/api/posts')
5. 포스트 데이터 받아오기
6. 카드 그리드 동적 렌더링
```

### 2. 포스트 상세 조회
```
1. 사용자가 카드 클릭
2. window.open('http://localhost:8253/post/{slug}')
3. 백엔드에서 Markdown 변환
4. 포스트 HTML 렌더링
```

---

## 파일 구조

### GitHub Pages 저장소
```
freelang-blog-posts/ (GitHub)
├── index.html                ← 정적 홈페이지
├── _config.yml               ← Jekyll 설정 (선택)
├── assets/
│   ├── css/
│   ├── js/
│   └── images/
└── README.md
```

### 백엔드 저장소
```
fl-blog/ (로컬 또는 클라우드)
├── server.js                 ← Node.js 서버
├── posts/                    ← 89개 .md 파일
├── README.md
└── package.json
```

---

## CORS 설정

포트 8253 서버가 GitHub Pages 요청을 허용하도록:

```javascript
// server.js에서
const server = http.createServer((req, res) => {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  
  // ... 나머지 핸들러
});
```

---

## 배포 체크리스트

### GitHub Pages
- [ ] index.html 준비
- [ ] README.md 작성
- [ ] GitHub 저장소에 푸시
- [ ] Settings > Pages 활성화
- [ ] URL: https://kimjindol2025.github.io/freelang-blog-posts/

### 포트 8253 서버
- [ ] Node.js 설치
- [ ] server.js 준비
- [ ] 89개 posts/ .md 파일 확인
- [ ] `node server.js :8253` 실행
- [ ] http://localhost:8253/health 테스트

### 통합 테스트
- [ ] GitHub Pages에서 API 호출 가능
- [ ] 포스트 목록 표시
- [ ] 포스트 상세 조회
- [ ] 모바일 반응형 확인

---

## 성능 및 보안

| 항목 | 정책 |
|------|------|
| **CORS** | `*` 허용 (로컬), 필요시 제한 |
| **캐싱** | 브라우저 캐시 활용 |
| **HTTPS** | GitHub Pages 자동, 백엔드는 SSL/TLS 필수 |
| **RPS** | 포스트당 <10ms |

---

## 문제해결

### "포트 8253 연결 불가"
```
→ 서버 실행 중 확인
  node server.js :8253
→ 로컬 테스트
  curl http://localhost:8253/health
```

### "포스트 로드 안됨"
```
→ API 응답 확인
  curl http://localhost:8253/api/posts | jq .
→ posts/ 디렉토리 확인
  ls projects/fl-blog/posts/ | wc -l
```

### "CORS 오류"
```
→ 브라우저 콘솔 확인
→ server.js의 CORS 헤더 확인
→ 동일 출처 정책 이해
```

---

**결론**: GitHub Pages(정적) + Node.js 서버(동적) 이중 구조로 
확장성과 성능을 모두 확보할 수 있습니다. ✨
