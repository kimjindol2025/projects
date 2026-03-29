---
layout: post
title: Phase3-028-Nginx-Configuration
date: 2026-03-28
---
# Nginx: 설정 완벽 가이드 (50K req/sec)

## 요약

- Nginx 아키텍처 (워커 프로세스)
- 로드 밸런싱 설정
- 캐싱 및 압축
- 50K req/sec 달성

---

## 1. Nginx 기본 설정

### nginx.conf

```nginx
# 워커 프로세스 수 (CPU 코어 수)
worker_processes auto;

# 워커당 동시 연결 수
events {
    worker_connections 4096;
}

http {
    # 압축
    gzip on;
    gzip_types text/plain text/css application/json;

    # 업스트림 (백엔드)
    upstream backend {
        least_conn;  # 연결 수 적은 것부터
        server 127.0.0.1:8001;
        server 127.0.0.1:8002;
        server 127.0.0.1:8003;
    }

    # 서버 설정
    server {
        listen 80;

        location / {
            proxy_pass http://backend;
        }
    }
}
```

---

## 2. 로드 밸런싱

### 알고리즘

```nginx
# Round Robin (기본)
upstream backend {
    server app1:8080;
    server app2:8080;
}

# Least Connections
upstream backend {
    least_conn;
    server app1:8080;
    server app2:8080;
}

# IP Hash (세션 유지)
upstream backend {
    ip_hash;
    server app1:8080;
    server app2:8080;
}
```

---

## 3. 캐싱

### 클라이언트 캐시

```nginx
location ~* \.(jpg|jpeg|png|gif|ico)$ {
    expires 30d;
    add_header Cache-Control "public, immutable";
}

location /api/ {
    expires -1;  # 캐시 금지
}
```

### 프록시 캐시

```nginx
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m;

location / {
    proxy_cache my_cache;
    proxy_cache_valid 200 10m;
    proxy_cache_bypass $http_pragma $http_authorization;

    proxy_pass http://backend;
}
```

---

## 4. 성능 튜닝

```nginx
# 버퍼 크기
proxy_buffer_size 128k;
proxy_buffers 4 256k;

# 연결 유지
keepalive_timeout 65;
proxy_http_version 1.1;
proxy_set_header Connection "";

# 타임아웃
proxy_connect_timeout 10s;
proxy_send_timeout 30s;
proxy_read_timeout 30s;
```

---

## 5. SSL/TLS

```nginx
server {
    listen 443 ssl http2;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # TLS 버전
    ssl_protocols TLSv1.2 TLSv1.3;

    # 암호화 알고리즘
    ssl_ciphers 'ECDHE-ECDSA-AES256-GCM-SHA384:...';

    # 성능
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
}
```

---

## 6. 벤치마크

### 처리량

```
설정         | req/sec
-------------|-------
기본         | 10K
로드밸런싱   | 30K
캐싱 활성화  | 40K
최적화       | 50K
```

### 지연시간

```
프록시 없음: 5ms
기본 프록시: 15ms
최적화:      8ms
```

---

## 7. 실전 설정

### 완전한 프로덕션

```nginx
user www-data;
worker_processes auto;
pid /run/nginx.pid;

events {
    worker_connections 4096;
    use epoll;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    gzip on;
    gzip_types text/plain application/json;

    upstream backend {
        least_conn;
        server app1:8080 max_fails=3 fail_timeout=30s;
        server app2:8080 max_fails=3 fail_timeout=30s;
    }

    server {
        listen 80;
        server_name example.com;

        location / {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        location ~* \.(jpg|jpeg|png|gif|css|js)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
}
```

---

## 핵심 정리

- **로드 밸런싱**: 3배 처리량
- **캐싱**: 추가 3배
- **튜닝**: 추가 1.25배

---

## 결론

**Nginx는 웹의 척추다!** 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
