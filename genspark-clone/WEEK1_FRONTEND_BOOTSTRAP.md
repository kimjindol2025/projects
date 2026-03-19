# 🔥 Week 1-2: 프론트엔드 부트스트랩 (React)

## 🎯 목표
- React 프로젝트 초기화
- 백엔드 (Flask) ↔ 프론트엔드 (React) API 연동
- 실시간 스트리밍 UI 구현
- 배포 준비 (Vercel)

---

## 📋 Phase 1: 프로젝트 초기화 (2시간)

### 1.1 Node.js 환경 확인

```bash
node --version  # v18.18.0 이상 필요
npm --version   # 9.0.0 이상
```

### 1.2 Vite + React 프로젝트 생성

```bash
# npm create vite@latest로 프로젝트 생성 (최신 방식)
npm create vite@latest genspark-frontend -- --template react-ts

cd genspark-frontend
npm install
```

### 1.3 필수 패키지 설치

```bash
npm install \
  axios \                           # HTTP 클라이언트
  react-query \                     # 데이터 페칭 및 캐싱
  zustand \                         # 상태 관리 (Redux 대체)
  socket.io-client \                # WebSocket 실시간 통신
  marked \                          # 마크다운 렌더링
  highlight.js \                    # 코드 하이라이트
  react-markdown \                  # 마크다운 React 컴포넌트
  tailwindcss \                     # CSS 프레임워크
  postcss autoprefixer \            # Tailwind 필수
  @headlessui/react \               # 기본 UI 컴포넌트
  zustand                           # 가벼운 상태 관리

# 개발 의존성
npm install -D \
  typescript \
  @types/react \
  @types/node \
  @typescript-eslint/eslint-plugin \
  eslint
```

### 1.4 Tailwind CSS 설정

```bash
npx tailwindcss init -p
```

**tailwind.config.js** 수정:
```javascript
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#6366f1',
        secondary: '#8b5cf6',
      }
    },
  },
  plugins: [],
}
```

### 1.5 디렉토리 구조

```
genspark-frontend/
├── src/
│   ├── components/
│   │   ├── SearchBar.tsx
│   │   ├── ResultCard.tsx
│   │   ├── StreamingResult.tsx
│   │   ├── WidgetRenderer.tsx
│   │   └── Sidebar.tsx
│   ├── pages/
│   │   ├── Home.tsx
│   │   ├── Result.tsx
│   │   └── Account.tsx
│   ├── services/
│   │   ├── api.ts              # Axios 인스턴스
│   │   ├── websocket.ts         # WebSocket 연결
│   │   └── analytics.ts         # 분석 이벤트
│   ├── stores/
│   │   ├── searchStore.ts       # 검색 상태 (Zustand)
│   │   ├── userStore.ts         # 사용자 상태
│   │   └── settingsStore.ts     # 설정
│   ├── types/
│   │   ├── api.ts              # API 타입
│   │   └── ui.ts               # UI 타입
│   ├── App.tsx
│   ├── App.css
│   └── main.tsx
├── index.html
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.js
└── package.json
```

---

## 🏗️ Phase 2: 백엔드 API 연동 (3시간)

### 2.1 Flask 서버 WebSocket 지원 추가

**현재 Flask 코드에 WebSocket 추가**:

```python
# requirements.txt에 추가
python-socketio
python-socketio[asyncio_client]
aiohttp

# Flask 앱 수정 (src/main.py 또는 새 파일)
from flask import Flask
from flask_socketio import SocketIO, emit
import json

app = Flask(__name__)
socketio = SocketIO(app, cors_allowed_origins="*")

@app.route('/api/search', methods=['POST'])
def search():
    """REST API: 검색"""
    query = request.json['query']
    # 기존 로직...
    return jsonify(result)

@socketio.on('connect')
def handle_connect():
    print('Client connected')
    emit('response', {'data': 'Connected'})

@socketio.on('search_stream')
def handle_search(data):
    """WebSocket: 실시간 스트리밍 검색"""
    query = data['query']

    # 스트림 시뮤레이션
    emit('stream_start', {'query': query})

    # Step 1: 분석
    emit('stream_chunk', {
        'step': 'analyze',
        'data': {'queries': 3}
    })

    # Step 2: 검색
    for i, result in enumerate(search_results):
        emit('stream_chunk', {
            'step': 'search',
            'progress': (i+1) / len(search_results),
            'data': result
        })

    # Step 3: 합산
    emit('stream_chunk', {
        'step': 'synthesize',
        'data': synthesis_result
    })

    emit('stream_complete', {
        'final_result': final_html
    })

if __name__ == '__main__':
    socketio.run(app, debug=True, port=5000)
```

### 2.2 React API 서비스

**src/services/api.ts**:
```typescript
import axios from 'axios';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:5000/api';

const apiClient = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
});

// JWT 토큰 추가
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const searchAPI = {
  // REST API (기존 사용 방식)
  search: (query: string) =>
    apiClient.post('/search', { query }),

  // 히스토리 조회
  getHistory: () =>
    apiClient.get('/search/history'),

  // 저장된 결과 조회
  getSaved: () =>
    apiClient.get('/search/saved'),
};

export const authAPI = {
  login: (email: string, password: string) =>
    apiClient.post('/auth/login', { email, password }),

  signup: (email: string, password: string) =>
    apiClient.post('/auth/signup', { email, password }),

  logout: () =>
    apiClient.post('/auth/logout'),
};

export default apiClient;
```

**src/services/websocket.ts**:
```typescript
import io, { Socket } from 'socket.io-client';

const SOCKET_URL = import.meta.env.VITE_API_URL || 'http://localhost:5000';

let socket: Socket | null = null;

export const initSocket = () => {
  if (socket) return socket;

  socket = io(SOCKET_URL, {
    reconnection: true,
    reconnectionDelay: 1000,
    reconnectionDelayMax: 5000,
    reconnectionAttempts: 5,
  });

  socket.on('connect', () => {
    console.log('✅ WebSocket connected');
  });

  socket.on('disconnect', () => {
    console.log('❌ WebSocket disconnected');
  });

  return socket;
};

export const searchStream = (query: string, callback: (event: string, data: any) => void) => {
  const ws = initSocket();
  if (!ws) return;

  ws.emit('search_stream', { query });

  ws.on('stream_start', (data) => callback('start', data));
  ws.on('stream_chunk', (data) => callback('chunk', data));
  ws.on('stream_complete', (data) => callback('complete', data));
  ws.on('stream_error', (data) => callback('error', data));
};

export const disconnectSocket = () => {
  if (socket) {
    socket.disconnect();
    socket = null;
  }
};
```

### 2.3 상태 관리 (Zustand)

**src/stores/searchStore.ts**:
```typescript
import { create } from 'zustand';

interface SearchState {
  query: string;
  isLoading: boolean;
  results: any;
  error: string | null;
  setQuery: (query: string) => void;
  setResults: (results: any) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clear: () => void;
}

export const useSearchStore = create<SearchState>((set) => ({
  query: '',
  isLoading: false,
  results: null,
  error: null,
  setQuery: (query) => set({ query }),
  setResults: (results) => set({ results }),
  setLoading: (loading) => set({ isLoading: loading }),
  setError: (error) => set({ error }),
  clear: () => set({ query: '', results: null, error: null }),
}));
```

---

## 🎨 Phase 3: UI 컴포넌트 (6시간)

### 3.1 검색 바

**src/components/SearchBar.tsx**:
```typescript
import { useState } from 'react';
import { useSearchStore } from '../stores/searchStore';
import { searchStream } from '../services/websocket';

export function SearchBar() {
  const [input, setInput] = useState('');
  const { setQuery, setLoading } = useSearchStore();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    setQuery(input);
    setLoading(true);

    // WebSocket으로 실시간 스트리밍 시작
    searchStream(input, (event, data) => {
      if (event === 'chunk') {
        console.log('📨 Streaming:', data.step, data.progress || '');
        // UI 업데이트 (다음 컴포넌트)
      }
      if (event === 'complete') {
        setLoading(false);
        // 결과 화면으로 이동
      }
    });

    setInput('');
  };

  return (
    <div className="w-full max-w-2xl mx-auto">
      <form onSubmit={handleSearch} className="relative">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="질문을 입력하세요... (예: Python 메모리 관리)"
          className="w-full px-4 py-3 rounded-lg border-2 border-gray-300 focus:border-primary focus:outline-none"
        />
        <button
          type="submit"
          className="absolute right-3 top-3 bg-primary text-white px-4 py-2 rounded-lg hover:bg-opacity-80"
        >
          검색
        </button>
      </form>
    </div>
  );
}
```

### 3.2 실시간 스트리밍 결과

**src/components/StreamingResult.tsx**:
```typescript
import { useEffect, useState } from 'react';
import { useSearchStore } from '../stores/searchStore';
import { searchStream } from '../services/websocket';

export function StreamingResult() {
  const { query, isLoading } = useSearchStore();
  const [steps, setSteps] = useState<any[]>([]);

  useEffect(() => {
    if (!query) return;

    setSteps([]);
    searchStream(query, (event, data) => {
      if (event === 'chunk') {
        setSteps((prev) => [...prev, {
          step: data.step,
          progress: data.progress,
          content: data.data,
        }]);
      }
    });
  }, [query]);

  return (
    <div className="space-y-4">
      {steps.map((step, idx) => (
        <div key={idx} className="border rounded-lg p-4 bg-white">
          <div className="flex items-center justify-between mb-2">
            <h3 className="font-semibold capitalize">{step.step}</h3>
            {step.progress && (
              <span className="text-sm text-gray-600">{(step.progress * 100).toFixed(0)}%</span>
            )}
          </div>
          <div className="text-gray-700">
            {typeof step.content === 'string' ? (
              <p>{step.content}</p>
            ) : (
              <pre>{JSON.stringify(step.content, null, 2)}</pre>
            )}
          </div>
        </div>
      ))}

      {isLoading && (
        <div className="flex items-center justify-center p-4">
          <div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full"></div>
        </div>
      )}
    </div>
  );
}
```

### 3.3 결과 카드 (마크다운 + 위젯)

**src/components/ResultCard.tsx**:
```typescript
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { WidgetRenderer } from './WidgetRenderer';

interface ResultCardProps {
  title: string;
  content: string;
  sources: string[];
  confidence: number;
}

export function ResultCard({ title, content, sources, confidence }: ResultCardProps) {
  return (
    <div className="bg-white rounded-lg shadow-lg p-6 mb-4">
      <h2 className="text-2xl font-bold mb-2">{title}</h2>

      <div className="flex items-center justify-between mb-4 pb-4 border-b">
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">신뢰도:</span>
          <div className="w-32 h-2 bg-gray-200 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-red-500 to-green-500"
              style={{ width: `${confidence * 100}%` }}
            ></div>
          </div>
          <span className="text-sm font-semibold">{(confidence * 100).toFixed(0)}%</span>
        </div>
      </div>

      <div className="prose prose-sm max-w-none mb-6">
        <ReactMarkdown
          components={{
            code: ({ node, inline, className, children, ...props }) => {
              const match = /language-(\w+)/.exec(className || '');
              return !inline && match ? (
                <SyntaxHighlighter
                  language={match[1]}
                  {...props}
                >
                  {String(children).replace(/\n$/, '')}
                </SyntaxHighlighter>
              ) : (
                <code className={className} {...props}>
                  {children}
                </code>
              );
            },
          }}
        >
          {content}
        </ReactMarkdown>
      </div>

      {/* 위젯 렌더링 */}
      <WidgetRenderer content={content} />

      {/* 소스 */}
      <div className="mt-6 pt-4 border-t">
        <h4 className="font-semibold mb-2">출처 ({sources.length})</h4>
        <ul className="space-y-2">
          {sources.map((url, idx) => (
            <li key={idx}>
              <a
                href={url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline truncate block"
              >
                {url}
              </a>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
```

---

## 📦 Phase 4: 환경 변수 및 배포 설정 (1시간)

### 4.1 환경 변수

**.env.example**:
```env
VITE_API_URL=http://localhost:5000
VITE_ENVIRONMENT=development
```

**.env.development**:
```env
VITE_API_URL=http://localhost:5000
VITE_ENVIRONMENT=development
```

**.env.production**:
```env
VITE_API_URL=https://api.genspark.app
VITE_ENVIRONMENT=production
```

### 4.2 Vercel 배포

```bash
# 1. Vercel CLI 설치
npm i -g vercel

# 2. 프로젝트 배포
vercel

# 3. 환경 변수 설정 (Vercel 대시보드)
# Settings → Environment Variables
# VITE_API_URL=https://api.genspark.app 추가

# 4. 자동 배포 설정
# GitHub 연동 → main 브랜치 푸시 시 자동 배포
```

**vercel.json** 추가:
```json
{
  "buildCommand": "npm run build",
  "outputDirectory": "dist",
  "env": {
    "VITE_API_URL": "@api_url"
  }
}
```

---

## ✅ Week 1-2 체크리스트

### Week 1 (Day 1-5)
- [ ] Vite + React 프로젝트 생성
- [ ] 필수 패키지 설치
- [ ] Tailwind CSS 설정
- [ ] 디렉토리 구조 생성
- [ ] Flask 서버 WebSocket 추가
- [ ] API/WebSocket 서비스 구현

### Week 2 (Day 6-10)
- [ ] SearchBar 컴포넌트
- [ ] StreamingResult 컴포넌트
- [ ] ResultCard 컴포넌트
- [ ] WidgetRenderer 컴포넌트
- [ ] 상태 관리 (Zustand)
- [ ] 환경 변수 설정
- [ ] Vercel 배포

### 최종 검증
- [ ] 로컬에서 정상 작동 (npm run dev)
- [ ] 실시간 스트리밍 성공
- [ ] 마크다운 렌더링
- [ ] 소스 링크 클릭
- [ ] Vercel 배포 성공
- [ ] 모바일 반응형 테스트

---

## 🚀 실행 명령어

```bash
# 개발 모드 시작
npm run dev
# http://localhost:5173 접속

# 프로덕션 빌드
npm run build

# 미리보기
npm run preview

# Vercel 배포
vercel --prod
```

---

## 📞 트러블슈팅

### CORS 에러
```python
# Flask에서 CORS 활성화
from flask_cors import CORS
CORS(app, resources={r"/api/*": {"origins": "*"}})
```

### WebSocket 연결 안 됨
```typescript
// 개발 환경에서 localhost 사용
const SOCKET_URL = 'http://localhost:5000';

// 프로덕션에서 환경 변수 사용
const SOCKET_URL = import.meta.env.VITE_API_URL;
```

### 빌드 에러
```bash
# node_modules 다시 설치
rm -rf node_modules
npm install

# 캐시 삭제
npm cache clean --force
```

---

**Week 1-2 완료 후: Beta 1.0 Alpha 출시 준비! 🎉**
