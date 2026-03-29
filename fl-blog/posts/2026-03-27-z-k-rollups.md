---
title: "ZK-Rollup과 영지식 증명: 확장성과 보안의 완벽한 균형"
date: 2026-03-27 09:00:00 +0900
author: freelang
tags: ["advanced", "low-level", "blockchain"]
toc: true
comments: true
---

# ZK-Rollup과 영지식 증명: 확장성과 보안의 완벽한 균형
## 요약

- Zero-Knowledge Proof (영지식 증명) 개념
- ZK-Rollup 아키텍처 (zkSync, StarkNet)
- 성능: L1 Ethereum vs ZK-Rollup (100배 향상)
- 기술 트레이드오프: 증명 시간 vs 검증 시간
- 실전 사례: 거래소, 스테이킹, NFT

---

## 1. 영지식 증명 (ZKP)

### 정의

```
증명자(Prover)가 정보를 공개하지 않으면서
검증자(Verifier)를 설득하는 기술

핵심: "내가 비밀을 알고 있다는 것을 알고 있으면서
     비밀 자체는 알려주지 않는다"

예: 동굴 비밀 시나리오

동굴:
   ┌────────────────────┐
   │  어떤 문을 열 수 있는
   │  비밀 비밀번호가 있는
   │  나선형 동굴
   └────────────────────┘

증명자(Alice)는 비밀번호를 알고 있음
검증자(Bob)는 확신하고 싶음

프로토콜:
1. Bob이 밖에서 대기
2. Alice가 동굴 안으로 들어가서 왼쪽 또는 오른쪽 선택
3. Bob이 "왼쪽으로 나와!" 또는 "오른쪽으로 나와!"를 무작위 선택
4. Alice가 항상 지시대로 나올 수 있음
   → Bob은 Alice가 비밀번호를 안다고 확신 (비밀번호는 모르면서)

반복: 100번 반복하면 99.9...% 확신
```

### 수학적 정의

```
영지식 증명 = (Setup, Prover, Verifier)

Setup: 증명 체계 생성
├─ 공개 파라미터: pk (누구나 알 수 있음)
└─ 증명 키: sk (증명자만 알 수 있음)

Prover: 증명 생성
입력: 공개 입력 (x), 비밀 입력 (w), 문제 (C)
증명: "C(x, w) = true임을 알고 있다"는 증명 π 생성
출력: π

Verifier: 증명 검증
입력: 공개 입력 (x), 증명 (π)
검증: π가 유효한가?
출력: true/false

성질:
├─ 완전성: 올바른 증명은 항상 통과
├─ 건전성: 틀린 증명은 통과 불가능
└─ 영지식성: 증명 과정에서 w는 공개 안됨
```

---

## 2. ZK-Rollup 아키텍처

### 동작 원리

```
L1 Ethereum:
├─ 스마트 컨트랙트 (검증자 역할)
├─ 잔액 트리: 루트 해시만 저장
└─ 배치 검증: π를 확인

L2 (zkSync, StarkNet):
├─ Sequencer: 거래 수집 및 정렬
├─ Executor: 거래 실행, 상태 변경
├─ Prover: π 생성 (CPU 집약적)
└─ Aggregator: 여러 π를 합치기

흐름:
1. 사용자가 L2에서 거래 (수백만 TPS)
2. Sequencer가 배치 생성 (1,000-10,000개 거래)
3. Executor가 상태 변경 (Merkle 트리 업데이트)
4. Prover가 증명 생성 (5-30분)
5. L1에 증명 제출
6. Verifier가 L1에서 검증 (10초)
7. 최종성 확보 (Ethereum 보안 상속)
```

### 성능 비교

```
L1 Ethereum:
├─ TPS: 15
├─ 지연: 12-15초
├─ 최종성: 6+ 확인 (72-90초)
└─ 가스비: $2-10/tx

Optimistic Rollup (Arbitrum):
├─ TPS: 4,000-7,000
├─ 지연: <2초
├─ 최종성: 7일 (분쟁 기간)
└─ 가스비: $0.01-0.1/tx (100배 저렴)

ZK-Rollup (zkSync Era):
├─ TPS: 1,000-4,000
├─ 지연: <2초
├─ 최종성: 5-10분 (증명 생성)
└─ 가스비: $0.01-0.05/tx (200배 저렴)

Plasma:
├─ TPS: 10,000+
├─ 지연: <1초
├─ 최종성: 1주 (데이터 가용성 문제)
└─ 가스비: $0.001/tx
```

---

## 3. ZK 회로 설계

### 간단한 증명 회로

```
문제: 내가 x의 제곱근을 안다는 것을 증명

회로 (algebraic circuit):
입력: y (공개), w (비밀 = √y)

제약식 (Constraint):
w * w = y

증명자:
w = 7을 안다
y = 49
증명: "w * w = 49"

검증자:
y = 49를 받음
증명을 확인: "정말 어떤 w * w = 49인가?"
→ "yes, 증명 π는 유효함"

w의 값은 공개 안됨!
```

### 복잡한 거래 증명

```
문제: 블록체인 거래가 유효함을 증명

회로 입력:
├─ 공개: 트랜잭션 해시, 새로운 상태 루트
└─ 비밀: 거래의 모든 세부사항

제약식 (Constraints):
├─ 서명 검증: ECDSA(tx, signature) = true
├─ 잔액 확인: from_balance >= amount
├─ Nonce 확인: tx.nonce == state[from].nonce
├─ 상태 전이: merkle_root'는 상태 변경 반영
└─ 가스 계산: gas_used <= gas_limit

증명자가 생성해야 하는 것:
약 1,000,000개의 제약식을 만족하는 증명

크기: ~100-200KB
검증 시간: <1초 (L1에서)
```

---

## 4. 주요 ZK 시스템 비교

### zkSync

```
특징:
├─ EVM 호환 (기존 Solidity 코드 실행 가능)
├─ 증명자: CPU 클러스터 (5-10분/배치)
├─ 검증자: L1 스마트 컨트랙트
└─ 언어: Solidity + zkVM 모듈

성능:
├─ TPS: 1,000-4,000
├─ 지연: <1초
├─ 가스비: $0.01-0.05/tx

예시: Uniswap V3를 zkSync에 배포
거래비: L1 $10 → L2 $0.05 (200배 저렴)
```

### StarkNet (StarkWare)

```
특징:
├─ Cairo 언어 (Turing 완전)
├─ STARK 증명 (양자 안전)
├─ 계산 증명 (transaction 검증 아님)
└─ 고급 암호학

성능:
├─ TPS: 10,000+
├─ 증명 시간: 30-60분
├─ 가스비: $0.001-0.01/tx (매우 저렴)

트레이드오프:
├─ 증명 시간 길어짐
├─ 언어 학습 곡선 가파름
└─ 아직 메인넷 초기 단계
```

---

## 5. 암호학 기초

### SNARK vs STARK

```
SNARK (Succinct Non-Interactive Argument of Knowledge):
├─ 크기: 작음 (~128 bytes)
├─ 검증 시간: 매우 빠름 (ms)
├─ 신뢰 설정: 필요 (Trusted Setup)
├─ 양자 안전성: 낮음
├─ 예: Groth16, PlonK

STARK (Scalable Transparent Argument of Knowledge):
├─ 크기: 중간 (~100 KB)
├─ 검증 시간: 빠름 (초)
├─ 신뢰 설정: 불필요
├─ 양자 안전성: 높음
├─ 예: FRI (Fast Reed-Solomon Interactive)
```

### Trusted Setup 문제

```
SNARK 시스템:

Setup:
1. 비밀 파라미터 λ 생성
2. λ로부터 공개 파라미터 pk, vk 계산
3. λ는 안전하게 폐기 (누구도 알 수 없어야 함)
   → 신뢰 설정!

만약 λ가 유출되면:
├─ 누구나 거짓 증명 생성 가능
├─ 시스템 붕괴
└─ 기술적으로는 안전하지만 사회적 위험

해결책:
├─ Ceremony: 많은 사람이 참여하여 λ 생성
├─ MPC (Multi-Party Computation): 누구도 전체 λ 알 수 없음
├─ Ethereum 2.0 KZG Ceremony (100,000+ 참여자)
```

---

## 6. 실전: zkSync DEX (Uniswap V4 개념)

```solidity
contract ZkSyncDEX {
    // L2에서 실행 (매우 저렴)
    struct Swap {
        address tokenIn;
        address tokenOut;
        uint amountIn;
    }
    
    function batchSwap(Swap[] memory swaps) public {
        // 1,000,000개 거래를 배치로 처리
        // 증명자가 모두 유효함을 증명
        
        for (uint i = 0; i < swaps.length; i++) {
            Swap memory swap = swaps[i];
            
            // 계산 (L2에서, 매우 빠름)
            uint amountOut = calculateOutput(
                swap.tokenIn,
                swap.tokenOut,
                swap.amountIn
            );
            
            // 상태 업데이트 (Merkle 트리)
            updateBalances(swap);
        }
        
        // L1에 배치 해시 제출
        // ZK 증명과 함께
        // ~1시간 후 최종 확정
    }
}

성능:
├─ L1에서라면: 1,000,000 tx × $5 = $5,000,000 비용
├─ ZK-Rollup에서: $100 비용 (50,000배 저렴!)
└─ 처리량: 매초 수백만 거래
```

---

## 7. 한계와 개선

### 현재 한계

```
1. 증명 생성 시간
   ├─ zkSync: 5-10분
   ├─ StarkNet: 30-60분
   └─ 개선중: 10초 수준으로

2. 복잡성
   ├─ 회로 설계 어려움
   ├─ 버그 위험 높음
   └─ 감사(Audit) 비용 높음

3. 메모리 오버헤드
   ├─ 증명 생성에 매우 많은 RAM 필요
   ├─ zkSync: 64GB+
   └─ 대중화 어려움
```

### 개선 방향

```
1. Proof Recursion (재귀 증명)
   ├─ 여러 증명을 하나의 증명으로 합치기
   └─ 더 빠른 최종성 가능

2. Hardware Acceleration
   ├─ GPU/FPGA로 증명 생성 가속
   ├─ 10초 이내 목표
   └─ 비용 감소

3. Modular Designs
   ├─ 거래 검증 vs 상태 전이 분리
   ├─ 병렬 증명 생성
   └─ 처리량 증가
```

---

## 8. 벤치마크

| 지표 | Ethereum | Arbitrum | zkSync | StarkNet |
|------|----------|----------|--------|----------|
| **TPS** | 15 | 4,000 | 2,000 | 10,000 |
| **지연** | 12s | 2s | 1s | 3s |
| **가스비** | $2-10 | $0.01-0.1 | $0.01-0.05 | $0.001-0.01 |
| **최종성** | 72s | 7일 | 5-10분 | 30-60분 |
| **거래 크기** | full | full | compressed | full |

---

## 핵심 정리

| 개념 | 설명 | 트레이드오프 |
|------|------|------------|
| **ZKP** | 비밀 공개 없이 증명 | 증명 시간 긴 |
| **ZK-Rollup** | L2 확장성 + L1 보안 | 복잡한 회로 설계 |
| **SNARK** | 작은 증명 | 신뢰 설정 필요 |
| **STARK** | 투명한 증명 | 크기 큼 |

---

## 결론

**"영지식 증명은 블록체인 확장성의 미래"**

비트코인 총 시가 ÷ ZK 인프라 투자 = 엄청난 기회 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
