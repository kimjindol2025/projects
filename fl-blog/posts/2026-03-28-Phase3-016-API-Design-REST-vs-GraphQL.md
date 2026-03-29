---
layout: post
title: Phase3-016-API-Design-REST-vs-GraphQL
date: 2026-03-28
---
# API 설계: REST vs GraphQL 완벽 비교

## 요약

**배우는 내용**:
- REST: 리소스 중심, 간단하고 표준화
- GraphQL: 데이터 중심, 유연하고 정확
- 성능 벤치마크 (요청 수, 데이터 양)
- 언제 어떤 API를 쓸까?

---

## 1. REST vs GraphQL

### 핵심 비교

```
항목              | REST           | GraphQL
-----------------|----------------|----------
데이터 조회       | 고정 스키마    | 필요한 것만
엔드포인트        | 다수           | 1개 (/graphql)
HTTP 메서드       | GET/POST/PUT   | POST 주도
버전관리          | /v1, /v2       | 자동 호환
캐싱              | HTTP 캐시 쉬움 | 복잡
쿼리 언어         | 없음           | GraphQL
오버페칭          | 자주 발생      | 없음
언더페칭          | 자주 발생      | 없음
학습곡선          | 완만           | 가파름
```

---

## 2. REST API

### 설계

```python
# Flask REST API
from flask import Flask, jsonify, request

app = Flask(__name__)

# GET /api/v1/users/1
@app.route('/api/v1/users/<int:user_id>', methods=['GET'])
def get_user(user_id):
    user = db.get_user(user_id)
    return jsonify({
        'id': user.id,
        'name': user.name,
        'email': user.email,
        'age': user.age,
        'address': user.address  # 불필요할 수도...
    })

# GET /api/v1/users/1/posts
@app.route('/api/v1/users/<int:user_id>/posts', methods=['GET'])
def get_user_posts(user_id):
    posts = db.get_posts(user_id)
    return jsonify([{
        'id': p.id,
        'title': p.title,
        'content': p.content,  # 긴 내용
        'author': {
            'id': user_id,
            'name': db.get_user(user_id).name  # N+1 문제
        }
    } for p in posts])

# POST /api/v1/users
@app.route('/api/v1/users', methods=['POST'])
def create_user():
    data = request.json
    user = db.create_user(data)
    return jsonify({'id': user.id}), 201

# PUT /api/v1/users/1
@app.route('/api/v1/users/<int:user_id>', methods=['PUT'])
def update_user(user_id):
    data = request.json
    user = db.update_user(user_id, data)
    return jsonify(user)

# DELETE /api/v1/users/1
@app.route('/api/v1/users/<int:user_id>', methods=['DELETE'])
def delete_user(user_id):
    db.delete_user(user_id)
    return '', 204
```

### REST 클라이언트

```python
# 사용자 정보 조회
response = requests.get('/api/v1/users/1')
user = response.json()  # 모든 필드 반환

# 사용자의 포스트 조회
response = requests.get('/api/v1/users/1/posts')
posts = response.json()  # 각 포스트의 모든 필드

# 이슈: 오버페칭
# - user의 'address' 필요 없음 → 낭비
# - 각 post의 'content' 필요 없음 → 낭비
```

---

## 3. GraphQL API

### 설계

```python
# Graphene (Python GraphQL)
import graphene

class UserType(graphene.ObjectType):
    id = graphene.Int()
    name = graphene.String()
    email = graphene.String()
    age = graphene.Int()
    address = graphene.String()
    posts = graphene.List('PostType')

    def resolve_posts(self, info):
        return db.get_posts(self.id)

class PostType(graphene.ObjectType):
    id = graphene.Int()
    title = graphene.String()
    content = graphene.String()
    author = graphene.Field(UserType)

    def resolve_author(self, info):
        return db.get_user(self.author_id)

class Query(graphene.ObjectType):
    user = graphene.Field(UserType, id=graphene.Int())
    users = graphene.List(UserType)

    def resolve_user(self, info, id):
        return db.get_user(id)

    def resolve_users(self, info):
        return db.get_all_users()

class Mutation(graphene.ObjectType):
    create_user = CreateUserMutation.Field()
    update_user = UpdateUserMutation.Field()
    delete_user = DeleteUserMutation.Field()

schema = graphene.Schema(query=Query, mutation=Mutation)
```

### GraphQL 클라이언트

```python
# 쿼리 1: 사용자 이름과 이메일만
query = """
    query GetUser($id: Int!) {
        user(id: $id) {
            name
            email
        }
    }
"""
result = client.execute(query, variable_values={'id': 1})
# 응답: {'name': 'Alice', 'email': 'alice@example.com'}

# 쿼리 2: 사용자와 포스트 제목만
query = """
    query GetUserPosts($id: Int!) {
        user(id: $id) {
            name
            posts {
                title
            }
        }
    }
"""
result = client.execute(query, variable_values={'id': 1})
# 응답: {'name': 'Alice', 'posts': [{'title': 'Post 1'}]}

# 뮤테이션: 사용자 생성
mutation = """
    mutation CreateUser($name: String!, $email: String!) {
        createUser(name: $name, email: $email) {
            id
            name
        }
    }
"""
result = client.execute(mutation, variable_values={
    'name': 'Bob',
    'email': 'bob@example.com'
})
```

---

## 4. 성능 벤치마크

### 벤치마크 1: 네트워크 요청 수

```
시나리오: 사용자 10명과 각각의 최근 5개 포스트 조회

REST:
- GET /api/v1/users → 10명 정보 (1 요청)
- GET /api/v1/users/1/posts → 첫 사용자 포스트 (1 요청)
- GET /api/v1/users/2/posts → (1 요청)
- ...
- GET /api/v1/users/10/posts → (1 요청)
- 총: 11개 요청

GraphQL:
- POST /graphql (쿼리 전송) → 1개 요청

결과: GraphQL 11배 적음
```

### 벤치마크 2: 데이터 전송량

```
동일한 쿼리: 사용자 이름, 이메일, 포스트 제목

REST:
- GET /users/1
  응답: {id, name, email, age, address, created_at, ...}
  크기: 2.5KB (필요한 것: 0.3KB)

GraphQL:
- 쿼리: {user(id:1) { name email posts { title } }}
  응답: {name, email, posts: [{title}]}
  크기: 0.3KB

결과: GraphQL 8배 적음
```

### 벤치마크 3: 응답시간

```
요청 수: 1000개 (10명 × 100회)

REST:
- 네트워크 지연: 1000 × 50ms = 50초
- 서버 처리: 11 × 500ms = 5.5초
- 총: ~55.5초

GraphQL:
- 네트워크 지연: 100 × 50ms = 5초
- 서버 처리: 1 × 500ms (배치 쿼리) = 0.5초
- 총: ~5.5초

결과: GraphQL 10배 빠름
```

---

## 5. 실전 문제

### 문제 1: N+1 Query (REST)

```python
# ❌ 문제
@app.route('/api/v1/users/<int:user_id>/posts')
def get_user_posts(user_id):
    posts = db.query("SELECT * FROM posts WHERE user_id = ?", [user_id])
    result = []
    for post in posts:
        author = db.query("SELECT * FROM users WHERE id = ?", [post.author_id])  # 각 포스트마다 쿼리!
        result.append({
            'title': post.title,
            'author': author.name
        })
    return jsonify(result)
# 1 쿼리(포스트) + 10 쿼리(작가) = 11개

# ✅ 해결 1: JOIN
def get_user_posts_optimized(user_id):
    posts = db.query("""
        SELECT p.*, u.name as author_name
        FROM posts p
        JOIN users u ON p.author_id = u.id
        WHERE p.user_id = ?
    """, [user_id])
    return jsonify(posts)  # 1 쿼리

# ✅ GraphQL은 자동 최적화
# (Dataloader 사용)
```

### 문제 2: 캐싱 (GraphQL)

```python
# GraphQL의 캐싱은 복잡
# HTTP 캐시를 활용할 수 없음 (POST 방식)

# ✅ 해결 1: 쿼리 해시
def cache_graphql(query):
    hash = hashlib.md5(query.encode()).hexdigest()
    if hash in cache:
        return cache[hash]

    result = execute_graphql(query)
    cache[hash] = result
    return result

# ✅ 해결 2: Apollo Client 캐시
# (클라이언트 측 자동 캐싱)
```

---

## 6. 선택 기준

### REST 선택

```
- 간단한 CRUD API
- 캐싱 중요 (HTTP 캐시)
- 표준 도구 사용 (curl, wget)
- 예: 공개 API (GitHub, Twitter)
```

### GraphQL 선택

```
- 복잡한 데이터 구조
- 모바일 앱 (대역폭 제약)
- 유연한 쿼리 필수
- 예: 페이스북, GitHub v4
```

---

## 7. 하이브리드

```python
# REST + GraphQL
# 장점만 섞기

@app.route('/api/v1/users')
def get_users_rest():
    # 공개 API: REST (캐시 쉬움)
    return jsonify(db.get_all_users())

@app.route('/graphql', methods=['POST'])
def graphql_endpoint():
    # 내부 API: GraphQL (유연함)
    query = request.json['query']
    return execute_graphql(query)
```

---

## 핵심 정리

| 측면 | REST | GraphQL |
|------|------|---------|
| **학습** | 쉬움 | 어려움 |
| **성능** | 중간 | 좋음 |
| **캐싱** | 쉬움 | 어려움 |
| **복잡도** | 낮음 | 높음 |

---

## 결론

"**문제에 맞는 도구를 선택하세요**"

REST는 표준, GraphQL은 혁신입니다. 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
