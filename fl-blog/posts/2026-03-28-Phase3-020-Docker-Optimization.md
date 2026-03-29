---
layout: post
title: Phase3-020-Docker-Optimization
date: 2026-03-28
---
# Docker 최적화: 1GB → 50MB (20배 축소)

## 요약

**배우는 내용**:
- 레이어 최적화 (캐싱 활용)
- 멀티 스테이지 빌드 (빌드 환경 제거)
- 이미지 크기 줄이기 (1GB → 50MB)
- 실전: 배포 속도 10배 향상

---

## 1. 기본 Dockerfile

### ❌ 나쁜 예: 1GB

```dockerfile
FROM ubuntu:22.04

# 문제 1: 전체 OS 설치
RUN apt-get update && apt-get install -y \
    build-essential \
    wget \
    git \
    curl \
    vim \
    && apt-get clean

WORKDIR /app
COPY . .

# 문제 2: 빌드 도구 포함
RUN apt-get install -y golang-1.20
RUN go build -o app .

EXPOSE 8080
CMD ["./app"]

# 최종 크기: ~1.2GB
# - Ubuntu base: 70MB
# - build-essential: 400MB
# - golang: 300MB
# - Go 모듈 캐시: 400MB+
```

### ✅ 좋은 예: 50MB

```dockerfile
# Stage 1: 빌드
FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o app .

# Stage 2: 실행 (가벼운 이미지)
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]

# 최종 크기: ~50MB
# - alpine base: 7MB
# - 바이너리: 30MB
# - ca-certificates: 13MB
```

---

## 2. 멀티 스테이지 빌드

### 3 스테이지 예시

```dockerfile
# Stage 1: 의존성 다운로드
FROM golang:1.20-alpine AS deps
WORKDIR /tmp
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: 빌드
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY . .
RUN go build -o app .

# Stage 3: 실행 (최소 크기)
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /
COPY --from=builder /app/app /usr/local/bin/app
ENTRYPOINT ["app"]
```

---

## 3. 레이어 캐싱

### ❌ 나쁜 예

```dockerfile
FROM golang:1.20

COPY . .                    # 모든 파일 복사 (변경 가능성 높음)

RUN go mod download         # 이 커맨드 재실행 (캐시 무효화)
RUN go build -o app .
```

**문제**: 소스 코드 변경 → 전체 레이어 재빌드

### ✅ 좋은 예

```dockerfile
FROM golang:1.20

WORKDIR /app

# 1단계: 의존성만 먼저 복사 (자주 변경 안 함)
COPY go.mod go.sum ./
RUN go mod download

# 2단계: 소스 코드 복사 (자주 변경함)
COPY . .
RUN go build -o app .
```

**효과**: 소스 변경 시 go mod download 스킵 (캐시 히트)

---

## 4. 크기 최적화 기법

### (1) 불필요한 파일 제거

```dockerfile
# .dockerignore
node_modules/
.git/
.env
*.log
test/
vendor/  # Go modules 캐시
```

### (2) 정적 바이너리 (CGO 비활성화)

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app .

# 동적 링크 (의존성 필요):
# - libc, libresolv 등 → alpine 필요
#
# 정적 바이너리:
# - 모든 것 포함 → scratch 가능
```

### (3) 바이너리 최소화

```dockerfile
RUN go build -ldflags="-w -s" -o app .

# -w: 디버그 심볼 제거
# -s: 심볼 테이블 제거
# 효과: 30MB → 15MB (50% 감소)
```

### (4) Base 이미지 선택

```
Image          | Size  | 사용처
---------------|-------|----------
ubuntu:22.04   | 77MB  | 개발/테스트
debian:12      | 124MB | 복잡한 앱
alpine:3.18    | 7MB   | 가벼운 앱
busybox:1.36   | 2MB   | 최소 앱
scratch        | 0MB   | 정적 바이너리
distroless     | 20MB  | 보안 중요
```

---

## 5. 성능 벤치마크

### 이미지 크기 비교

```
Dockerfile        | 크기    | 빌드시간 | 특징
-----------------|--------|---------|------
기본 (Ubuntu)     | 1.2GB  | 180s    | 불필요한 것 많음
개선 (Alpine)     | 400MB  | 120s    | 여전히 큼
멀티스테이지      | 80MB   | 90s     | 빌드 환경 제거
최적화 (scratch)  | 50MB   | 85s     | 정적 바이너리만
```

### 배포 시간

```
이미지 | 네트워크 | 압축해제 | 시작   | 총
-------|---------|---------|--------|------
1.2GB  | 120s    | 15s     | 2s     | 137s
50MB   | 5s      | 1s      | 2s     | 8s

개선: 137s → 8s (17배 빠름)
```

---

## 6. 고급 패턴

### 임시 빌드 캐시

```dockerfile
# BuildKit 사용 (docker buildx)
# syntax=docker/dockerfile:1

FROM golang:1.20 AS builder
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o app .

# 효과: Docker 빌드 캐시 재사용 (빌드 50% 빠름)
```

### 외부 의존성 캐시

```yaml
# docker-compose.yml
services:
  build:
    build:
      context: .
      cache_from:
        - myregistry/app:latest  # 이전 빌드 캐시
    image: myregistry/app:v1.0
```

---

## 7. 보안 최적화

### 최소 공격 표면

```dockerfile
# ❌ 나쁜 예: 패키지 매니저 포함
FROM alpine:latest
RUN apk add --no-cache openssh  # SSH 공격 가능

# ✅ 좋은 예: 필요한 것만
FROM alpine:latest
RUN apk add --no-cache ca-certificates  # SSL 인증서만
COPY --from=builder /app/app /usr/local/bin/
```

### 비루트 사용자

```dockerfile
FROM alpine:latest

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder --chown=appuser:appuser /app/app /usr/local/bin/

USER appuser
CMD ["app"]
```

---

## 8. 실전 사례

### Node.js + React

```dockerfile
# Stage 1: 빌드
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: 정적 서빙
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

# 크기: 500MB → 50MB (10배 축소)
```

### Java Spring Boot

```dockerfile
# Stage 1: 빌드
FROM eclipse-temurin:17-jdk-alpine AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:resolve
COPY . .
RUN mvn clean package -DskipTests

# Stage 2: 실행
FROM eclipse-temurin:17-jre-alpine
COPY --from=builder /app/target/app.jar /app.jar
ENTRYPOINT ["java","-jar","/app.jar"]

# 크기: 1.5GB → 400MB (4배 축소)
```

---

## 핵심 정리

| 기법 | 효과 |
|------|------|
| **멀티 스테이지** | 600MB 감소 |
| **Alpine 기반** | 100MB 감소 |
| **정적 바이너리** | 30MB 감소 |
| **정리 정리** | 20MB 감소 |
| **총** | 1.2GB → 50MB |

---

## 결론

Docker 이미지는 **작고 빠를수록 좋다**.

- 20배 작은 이미지
- 17배 빠른 배포
- 더 안전한 표면

🚀 최적화된 컨테이너로 시작하세요!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
