---
layout: post
title: Phase3-030-Strace-Performance
date: 2026-03-28
---
# 성능 분석: strace로 응답시간 1/10 단축하기

## 요약

- 시스템 콜 추적 (strace)
- 병목 분석 기법
- 실전: 1000ms → 100ms 개선
- 10배 성능 향상 사례

---

## 1. strace란?

```bash
# 모든 시스템 콜 추적
strace ./app

# 특정 프로세스
strace -p 1234

# 시간 통계
strace -c ./app

# 출력 파일
strace -o trace.log ./app
```

---

## 2. 시스템 콜 이해

### 주요 콜

```
read/write      : 파일/네트워크 I/O
open/close      : 파일 디스크립터
mmap            : 메모리 매핑
mlock           : 메모리 잠금
futex           : 뮤텍스 (스레드 동기)
poll/epoll      : 이벤트 대기
```

---

## 3. 실전 분석

### Case: 1000ms → 100ms 개선

**문제**: API 응답이 느려짐 (1000ms)

```bash
$ strace -c ./api_server
% time     seconds  usecs/call     calls  errors
------ ----------- ----------- --------- -----
50.0   0.500000    500        1         0 open("/etc/passwd")
40.0   0.400000    100        4         2 read
10.0   0.100000    50         2         0 write
```

**분석**:
- `open("/etc/passwd")`: 500ms (5배 오버헤드!)
- 매 요청마다 `/etc/passwd` 읽음

**해결**:
```python
# ❌ 나쁜 예
def verify_user(uid):
    with open('/etc/passwd') as f:
        ...  # 매 요청마다

# ✅ 좋은 예
pwd_cache = load_passwd_once()

def verify_user(uid):
    return uid in pwd_cache  # 메모리 조회
```

**결과**: 1000ms → 100ms ✅

---

## 4. strace 필터링

### 특정 시스템 콜만

```bash
# read/write만
strace -e trace=read,write ./app

# 네트워크만
strace -e trace=network ./app

# 파일 I/O만
strace -e trace=file ./app
```

---

## 5. 시간 분석

```bash
$ strace -tt -e trace=open ./app

# 타임스탬프 + 상대 시간
12:34:56.123456 open("/etc/passwd")  = 3  <0.500123>
12:34:56.623579 read(3, ...)  = 128  <0.100456>
```

**해석**:
- `<0.500123>`: 500ms 소요
- `/etc/passwd` 열기가 병목

---

## 6. 성능 개선 체크리스트

```
❌ 불필요한 open
✅ 캐시 활용 또는 한 번만 열기

❌ 루프 내 시스템 콜
✅ 루프 전에 작업 수행

❌ 동기 I/O
✅ 비동기 또는 배치

❌ 많은 작은 read
✅ 큰 버퍼로 한 번에
```

---

## 7. 실전 예시

### 느린 파일 읽기

```bash
$ strace -e trace=read,write -c ./app
% time     seconds  usecs/call
------ ----------- -----------
80.0   0.800000    1000       read() × 800 (각 1KB)
20.0   0.200000    100        write()
```

**문제**: 800개 read 호출 (각 1KB)

```go
// ❌ 나쁜 예
for {
    buf := make([]byte, 1024)
    n, _ := file.Read(buf)  // 800번 호출
}

// ✅ 좋은 예
reader := bufio.NewReaderSize(file, 64*1024)
for {
    buf, _ := reader.ReadBytes('\n')  // 1~2번 호출
}
```

---

## 8. 고급 분석

### 시스템 콜 의존성

```bash
# 호출 그래프
strace -f -e trace=file ./app

process 1234 open(...) → read(...) → close(...)
process 1235 open(...) → read(...) → close(...)
```

---

## 핵심 정리

| 문제 | 해결책 | 효과 |
|------|--------|------|
| **캐시 미스** | 메모리 캐시 | 10배 |
| **작은 read** | 버퍼 증가 | 5배 |
| **동기 I/O** | 배치 처리 | 100배 |

---

## 결론

**"시스템 콜이 답을 가지고 있다"**

strace로 병목을 찾고, 10배 빠르게 만드세요! 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
