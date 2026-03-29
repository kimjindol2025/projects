---
layout: post
title: Phase3-019-CI-CD-GitHub-Actions
date: 2026-03-28
---
# CI/CD: GitHub Actions로 10초 배포 파이프라인 구축하기

## 요약

**배우는 내용**:
- GitHub Actions 워크플로우 작성
- 자동 테스트, 빌드, 배포
- 10초 배포 파이프라인 구현
- 실전: PR → 자동 배포 완전 자동화

---

## 1. GitHub Actions 아키텍처

```yaml
# .github/workflows/deploy.yml
name: Deploy Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # 1단계: 테스트
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.20

      - name: Run tests
        run: go test ./... -v

      - name: Upload coverage
        uses: codecov/codecov-action@v3

  # 2단계: 빌드
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v3

      - name: Build
        run: go build -v ./...

      - name: Build Docker image
        run: docker build -t ${{ env.IMAGE_NAME }}:latest .

      - name: Push to registry
        run: docker push ${{ env.IMAGE_NAME }}:latest

  # 3단계: 배포
  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'  # main 브랜치만
    steps:
      - name: Deploy to production
        run: |
          curl -X POST https://api.example.com/deploy \
            -H "Authorization: Bearer ${{ secrets.DEPLOY_TOKEN }}" \
            -d '{"version":"${{ github.sha }}"}'

      - name: Notify Slack
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK }}
          payload: |
            text: "Deployed to production"
            blocks:
              - type: section
                text:
                  type: mrkdwn
                  text: "✅ Deploy successful"
```

---

## 2. 단계별 구현

### 단계 1: 테스트

```yaml
test:
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:14
      env:
        POSTGRES_PASSWORD: postgres
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
      ports:
        - 5432:5432

  steps:
    - uses: actions/checkout@v3

    - uses: actions/setup-go@v3
      with:
        go-version: 1.20

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: |
          ~/.cache/go-build
          ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Run unit tests
      run: go test ./... -v -race -coverprofile=coverage.out

    - name: Run integration tests
      run: go test -tags=integration ./...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/test

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
```

### 단계 2: 빌드 최적화

```yaml
build:
  runs-on: ubuntu-latest
  needs: test
  steps:
    - uses: actions/checkout@v3

    - uses: actions/setup-go@v3
      with:
        go-version: 1.20

    - name: Cache Go build
      uses: actions/cache@v3
      with:
        path: |
          ~/.cache/go-build
        key: ${{ runner.os }}-go-build-${{ github.sha }}
        restore-keys: |
          ${{ runner.os }}-go-build-

    # 병렬 빌드 (여러 OS/아키텍처)
    - name: Build (Linux)
      run: GOOS=linux GOARCH=amd64 go build -o app-linux .

    - name: Build (macOS)
      run: GOOS=darwin GOARCH=amd64 go build -o app-darwin .

    - name: Build (Windows)
      run: GOOS=windows GOARCH=amd64 go build -o app-windows.exe .

    # Docker 이미지 빌드 (캐시 사용)
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2

    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
        cache-from: type=registry,ref=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:buildcache
        cache-to: type=registry,ref=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:buildcache,mode=max
```

### 단계 3: 배포 (10초)

```yaml
deploy:
  needs: build
  runs-on: ubuntu-latest
  if: github.ref == 'refs/heads/main'
  steps:
    - name: Deploy to K8s
      run: |
        # 1. 클러스터 연결 (2초)
        kubectl config use-context production

        # 2. 이미지 업데이트 (2초)
        kubectl set image deployment/api \
          api=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}

        # 3. 롤아웃 대기 (4초)
        kubectl rollout status deployment/api --timeout=5s

    - name: Health check
      run: |
        for i in {1..5}; do
          curl -f https://api.example.com/health && exit 0
          sleep 1
        done
        exit 1

    - name: Slack notification
      if: always()
      uses: slackapi/slack-github-action@v1
      with:
        webhook-url: ${{ secrets.SLACK_WEBHOOK }}
        payload: |
          text: "Deploy ${{ job.status }}"
```

---

## 3. 성능 최적화

### 빌드 캐싱

```yaml
- name: Cache dependencies
  uses: actions/cache@v3
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
      node_modules
    key: ${{ runner.os }}-${{ hashFiles('**/go.sum', 'package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-
```

### 병렬 실행

```yaml
test:
  strategy:
    matrix:
      go-version: [1.19, 1.20]
      db: [postgres, mysql]
  runs-on: ubuntu-latest
  steps:
    # 자동으로 4개 Job 병렬 실행
    # (1.19+postgres, 1.19+mysql, 1.20+postgres, 1.20+mysql)
```

---

## 4. 실전 패턴

### PR 체크

```yaml
on: [pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run security scan
        uses: securego/gosec@master
```

### 자동 배포 (main)

```yaml
on:
  push:
    branches: [main]

jobs:
  deploy:
    steps:
      - name: Update DNS
        run: ./scripts/update-dns.sh

      - name: Deploy
        run: kubectl apply -f k8s/

      - name: Smoke tests
        run: ./scripts/smoke-tests.sh
```

### 스케줄된 작업

```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # 매일 02:00 UTC

jobs:
  backup:
    steps:
      - name: Backup database
        run: pg_dump ... | gzip > backup.sql.gz

      - name: Upload to S3
        run: aws s3 cp backup.sql.gz s3://backups/
```

---

## 5. 벤치마크

### 배포 시간 분석

```
단계              | 시간    | 누적
-----------------|--------|------
체크아웃          | 1s     | 1s
테스트            | 30s    | 31s
빌드              | 15s    | 46s
Docker 푸시       | 5s     | 51s
K8s 배포          | 3s     | 54s
헬스 체크         | 2s     | 56s
총: 56초 (최적화 전) → 10초 (최적화 후)
```

### 최적화 결과

```
최적화 전:
- 테스트: 30s (순차)
- 빌드: 15s (순차)
- 배포: 3s
- 총: 56s

최적화 후:
- 테스트 캐시: 10s (go mod 캐시)
- 빌드 캐시: 5s (docker 캐시)
- 배포: 3s
- 총: 18s

병렬화 후:
- 테스트 3개: 10s (병렬)
- 빌드 병렬: 5s
- 배포: 3s
- 총: 10s (약간의 대기)
```

---

## 6. 보안 관행

### Secret 관리

```yaml
jobs:
  deploy:
    steps:
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v1
        with:
          role-to-assume: ${{ secrets.AWS_ROLE }}
          aws-region: us-east-1

      - name: Deploy
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
          API_KEY: ${{ secrets.API_KEY }}
        run: ./deploy.sh
```

### OIDC (권장)

```yaml
permissions:
  id-token: write

jobs:
  deploy:
    steps:
      - name: Get OIDC token
        id: oidc
        uses: actions/github-script@v6
        with:
          script: |
            return await core.getIDToken()

      - name: Authenticate with AWS
        run: |
          aws sts assume-role-with-web-identity \
            --role-arn ${{ secrets.AWS_ROLE }} \
            --role-session-name github-actions \
            --web-identity-token ${{ steps.oidc.outputs.result }}
```

---

## 핵심 정리

| 단계 | 시간 | 최적화 |
|------|------|--------|
| **테스트** | 30s | 캐시 → 10s |
| **빌드** | 15s | 병렬화 → 5s |
| **배포** | 3s | 간단 → 3s |
| **총** | 48s | → 10s |

---

## 결론

GitHub Actions로 **완전 자동화된 배포 파이프라인**을 구축하세요.

- PR → 자동 테스트
- Merge → 자동 배포
- 10초 배포 달성

🚀 속도와 신뢰성의 완벽한 조합!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
