---
title: "AWS EC2: 성능 튜닝과 비용 최적화"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["systems", "devops", "cloud"]
toc: true
comments: true
---

# AWS EC2: 성능 튜닝과 비용 최적화
## 요약

- 인스턴스 타입 선택 (c5, m5, r5, t4g)
- EBS 최적화
- 네트워크 성능
- 비용 50% 절감 사례

---

## 1. 인스턴스 타입

| 타입 | CPU/메모리 | 용도 | 가격 |
|------|-----------|------|------|
| **t4g** | 2 vCPU / 1GB | 웹 서버 | $0.03/h |
| **m5** | 2 vCPU / 8GB | 일반 | $0.096/h |
| **c5** | 2 vCPU / 4GB | 계산 | $0.085/h |
| **r5** | 2 vCPU / 16GB | 메모리 | $0.126/h |

---

## 2. 성능 벤치마크

### CPU 성능

```
t4g (ARM): 10K req/sec
m5 (x86):  12K req/sec
c5 (x86):  15K req/sec

비용 대비:
c5: 성능 당 최고 효율
```

### 메모리 선택

```
Go 서비스: 500MB 필요
- t4g 1GB: 과충분
- m5 8GB: 과잉

결론: t4g 1GB (비용 70% 절감)
```

---

## 3. EBS 최적화

### 볼륨 타입

```
gp3 (기본):
- 처리량: 1000 MB/s
- IOPS: 16,000
- 가격: $0.1/GB/월

io1 (프로덕션):
- 처리량: 1000 MB/s
- IOPS: 64,000
- 가격: $0.125/GB + IOPS 비용
```

### 크기 최적화

```bash
# 현재 사용량 확인
df -h /

# 불필요한 로그 정리
du -sh /var/log/*
rm -rf /var/log/old*

# 효과: 100GB → 20GB (80% 감소)
```

---

## 4. 네트워크 최적화

### Enhanced Networking

```bash
# ENA (Elastic Network Adapter) 활성화
# → 처리량 10배, 지연 1/10

eth0: 1Gbps (기본)
eth0: 10Gbps (ENA 활성화)
```

### 스핀타입 배치 처리

```bash
# 배치 처리로 API 호출 감소
- 1개씩 처리: 10 API calls = $0.01
- 10개씩 처리: 1 API call = $0.001

효과: 데이터 전송 비용 90% 감소
```

---

## 5. 비용 최적화

### Reserved Instances (RI)

```
온디맨드: $0.096/h × 730시간 = $70/월
RI (1년): $0.055/h × 730 = $40/월

절감: 43%
```

### Spot Instances

```
온디맨드: $0.096/h
Spot: $0.029/h (70% 할인)

위험: 2분 내 종료 가능

용도: 배치, 캐시, 분석
```

---

## 6. 실전 최적화

### Before: $1,000/월

```
- t3 Medium (2vCPU, 4GB): $0.047/h × 730 = $34
- EBS gp2 100GB: $10
- 데이터 전송: $50
- 로드밸런서: $16
- 네트워크: $890

가장 큰 문제: 데이터 전송 + 네트워크
```

### After: $500/월 (50% 절감)

```
- t4g Small (1vCPU, 2GB): $0.017/h × 730 = $12
- EBS gp3 50GB: $5
- 데이터 전송 (배치): $10
- CloudFront CDN: $40
- Spot Instances: $433

개선:
1. 인스턴스 다운사이징: $34 → $12
2. EBS 축소: $10 → $5
3. 네트워크 개선: $890 → $50
```

---

## 7. CloudWatch 모니터링

```bash
# CPU 사용률 확인
aws cloudwatch get-metric-statistics \
  --namespace AWS/EC2 \
  --metric-name CPUUtilization \
  --dimensions Name=InstanceId,Value=i-xxx

# 메모리 사용률 확인
aws cloudwatch get-metric-statistics \
  --namespace CWAgent \
  --metric-name mem_used_percent
```

---

## 핵심 정리

| 최적화 | 절감 |
|--------|------|
| 인스턴스 다운사이징 | 65% |
| 데이터 전송 | 80% |
| RI 구매 | 43% |
| 총 절감 | 50% |

---

## 결론

**작은 변화가 큰 절감을 만듭니다!** 💰

---

질문이나 피드백은 댓글로 남겨주세요! 💬
