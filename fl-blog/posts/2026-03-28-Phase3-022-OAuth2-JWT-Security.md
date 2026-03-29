---
layout: post
title: Phase3-022-OAuth2-JWT-Security
date: 2026-03-28
---
# 보안: OAuth2/JWT로 100만 사용자 인증하기

## 요약

**배우는 내용**:
- OAuth2 플로우 (Authorization Code, Implicit, Client Credentials)
- JWT 토큰 구조 및 서명
- 리프레시 토큰 관리
- 실전: 100만 사용자 세션 관리

---

## 1. OAuth2 플로우

```
┌─────────┐         ┌──────────┐         ┌───────────┐
│  User   │         │  Client  │         │ Provider  │
└────┬────┘         └────┬─────┘         └─────┬─────┘
     │                   │                      │
     │ 1. "Login"        │                      │
     ├──────────────────>│                      │
     │                   │ 2. Redirect + code  │
     │                   ├─────────────────────>│
     │                   │                      │
     │                   │ 3. Token response    │
     │                   │<─────────────────────┤
     │                   │                      │
     │ 4. Logged in      │                      │
     │<──────────────────┤                      │
```

### Authorization Code Flow

```python
# 1. 사용자를 OAuth 제공자로 리다이렉트
@app.route('/login')
def login():
    auth_url = f"https://provider.com/oauth/authorize?
        client_id={CLIENT_ID}&
        redirect_uri={REDIRECT_URI}&
        scope=openid+profile+email&
        response_type=code"
    return redirect(auth_url)

# 2. 콜백: 인증 코드 수신
@app.route('/callback')
def callback():
    code = request.args.get('code')

    # 3. 코드를 토큰으로 교환 (백엔드)
    response = requests.post(
        'https://provider.com/oauth/token',
        data={
            'grant_type': 'authorization_code',
            'code': code,
            'client_id': CLIENT_ID,
            'client_secret': CLIENT_SECRET,
            'redirect_uri': REDIRECT_URI
        }
    )

    token = response.json()['access_token']

    # 4. 토큰 저장
    session['token'] = token
    return redirect('/')
```

---

## 2. JWT (JSON Web Token)

### 구조

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.
eyJzdWIiOiI1YjUyYjczZDMzNDgiLCJuYW1lIjoiQWxpY2UiLCJpYXQiOjE3MDAw
MDB9.
SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c

┌─────────────────────────────┬──────────────────────┬─────────────┐
│ Header (Base64)             │ Payload (Base64)     │ Signature   │
├─────────────────────────────┼──────────────────────┼─────────────┤
│ {"alg":"HS256","typ":"JWT"} │ {"sub":"5b52b73d",   │ HMAC(SHA256)│
│                             │  "name":"Alice",     │             │
│                             │  "iat":1700000000}   │             │
└─────────────────────────────┴──────────────────────┴─────────────┘
```

### JWT 검증

```go
import "github.com/golang-jwt/jwt/v5"

var secretKey = []byte("your-secret-key")

// 토큰 생성
func generateToken(userID string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": userID,
        "exp": time.Now().Add(time.Hour).Unix(),
        "iat": time.Now().Unix(),
    })

    return token.SignedString(secretKey)
}

// 토큰 검증
func validateToken(tokenString string) (*jwt.Token, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return secretKey, nil
    })

    if err != nil || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    return token, nil
}
```

---

## 3. 리프레시 토큰

### 문제: 토큰 만료

```
발급: token_exp = now + 1시간

문제: 1시간 후 사용자가 로그아웃됨
→ 다시 로그인해야 함 (불편)

해결: 리프레시 토큰
```

### 해결: Refresh Token

```python
# 발급
@app.route('/token', methods=['POST'])
def get_token():
    access_token = generate_token(user_id, expires_in=3600)    # 1시간
    refresh_token = generate_token(user_id, expires_in=2592000)  # 30일

    return jsonify({
        'access_token': access_token,
        'refresh_token': refresh_token,
        'expires_in': 3600
    })

# 갱신
@app.route('/refresh', methods=['POST'])
def refresh():
    refresh_token = request.json['refresh_token']
    user_id = validate_token(refresh_token)['sub']

    new_access_token = generate_token(user_id, expires_in=3600)

    return jsonify({'access_token': new_access_token})
```

---

## 4. 성능 최적화 (100만 사용자)

### 토큰 캐시 (Redis)

```python
import redis

cache = redis.Redis()

def validate_token_cached(token):
    # 1. 캐시 확인
    cached = cache.get(f'token:{token}')
    if cached:
        return json.loads(cached)

    # 2. 검증
    claims = validate_token(token)

    # 3. 캐시 저장 (토큰 만료시간까지)
    ttl = claims['exp'] - time.time()
    cache.setex(f'token:{token}', int(ttl), json.dumps(claims))

    return claims

# 성능:
# - 캐시 미스: 50ms (JWT 검증)
# - 캐시 히트: <1ms
# - 결과: 처리량 100배 증가
```

### 토큰 블랙리스트 (로그아웃)

```python
# 로그아웃: 토큰을 블랙리스트에 추가
def logout(token):
    claims = validate_token(token)
    ttl = claims['exp'] - time.time()

    cache.setex(f'blacklist:{token}', int(ttl), '1')

def is_blacklisted(token):
    return cache.exists(f'blacklist:{token}') > 0
```

---

## 5. 보안 모범 사례

### (1) 토큰 저장 위치

```javascript
// ❌ 나쁜 예: localStorage (XSS 취약)
localStorage.setItem('token', token);

// ✅ 좋은 예: HttpOnly 쿠키
// 백엔드에서 설정
response.set_cookie('token', token, httponly=True, secure=True)
```

### (2) CORS (Cross-Origin)

```python
@app.route('/api/data')
def get_data():
    # ✅ 명시적 origin 검증
    origin = request.headers.get('Origin')
    if origin not in ALLOWED_ORIGINS:
        return '', 403

    return jsonify(data)
```

### (3) Rate Limiting

```python
from flask_limiter import Limiter

limiter = Limiter(app, key_func=lambda: request.remote_addr)

@app.route('/login', methods=['POST'])
@limiter.limit("5/minute")  # 분당 5회
def login():
    ...
```

---

## 6. 벤치마크

```
작업              | 시간
-----------------|------
JWT 생성          | 2ms
JWT 검증          | 3ms
캐시 히트         | <1ms
캐시 미스         | 3ms
Redis lookup      | <1ms
```

---

## 7. 인증 흐름 (완전한 예시)

```
1. 로그인
   POST /login → username + password

2. 토큰 발급
   ✅ user 검증
   ✅ access_token + refresh_token

3. API 요청 (인증 헤더)
   GET /api/data
   Authorization: Bearer <access_token>

4. 토큰 검증 (미들웨어)
   ✅ 서명 확인
   ✅ 만료 확인
   ✅ 블랙리스트 확인

5. 토큰 갱신 (1시간 후)
   POST /refresh
   refresh_token 제출
   ✅ 새 access_token 발급

6. 로그아웃
   POST /logout
   ✅ 토큰 블랙리스트 추가
```

---

## 핵심 정리

| 방식 | 사용처 | 특징 |
|------|--------|------|
| **OAuth2** | 3rd party 인증 | 표준, 안전 |
| **JWT** | API 인증 | 상태 비저장 |
| **Session** | 전통 웹 | 간단 |

---

## 결론

**"보안은 타협할 수 없다"**

- OAuth2: 안전한 위임
- JWT: 확장 가능
- 리프레시: 편의성과 보안의 균형

🚀 100만 사용자를 안전하게 인증하세요!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
