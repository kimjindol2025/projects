---
layout: post
title: Phase4-041-Smart-Contracts
date: 2026-03-28
---
# 스마트 컨트랙트: 블록체인에서 코드가 법이 되다

## 요약

- 스마트 컨트랙트 개념과 실행 모델
- Solidity vs Rust (Move, Anchor)
- 가스(Gas) 비용 최적화
- 보안 취약점: Reentrancy, Over/Underflow
- 실전 예제: DEX, 스테이킹, 토큰

---

## 1. 스마트 컨트랙트란?

### 정의

```
스마트 컨트랙트 = 블록체인에 배포된 프로그램

특징:
├─ Deterministic (결정적 실행)
├─ Immutable (배포 후 수정 불가)
├─ Transparent (누구나 코드 검증 가능)
└─ Enforceable (자동 실행, 중간자 필요 없음)

현실 계약 vs 스마트 컨트랙트:

현실:
"A가 B에게 1000원 주면, B가 A에게 물건 줌"
→ 변호사, 중간자 필요

스마트 컨트랙트:
if payment == 1000 && verify(seller)
    transfer(goods to buyer)
    transfer(payment to seller)
→ 자동 실행, 중간자 불필요
```

### 실행 환경

```
Ethereum (EVM):
├─ 언어: Solidity
├─ 비용: 가스 (모든 연산에 비용)
└─ 컨센서스: Proof of Stake (병합 후)

Solana:
├─ 언어: Rust
├─ 비용: 낮음 (계산 중심)
└─ 컨센서스: Proof of History

Aptos/Move:
├─ 언어: Move (자산 중심)
├─ 비용: 중간
└─ 메모리 안전성: 강함
```

---

## 2. Solidity 기초

### ERC-20 (토큰 표준)

```solidity
pragma solidity ^0.8.0;

contract MyToken {
    string public name = "MyToken";
    uint8 public decimals = 18;
    uint public totalSupply;
    
    mapping(address => uint) public balanceOf;
    mapping(address => mapping(address => uint)) public allowance;
    
    event Transfer(address indexed from, address indexed to, uint value);
    event Approval(address indexed owner, address indexed spender, uint value);
    
    constructor(uint initialSupply) {
        totalSupply = initialSupply * 10 ** uint(decimals);
        balanceOf[msg.sender] = totalSupply;
    }
    
    // 전송
    function transfer(address to, uint value) public returns (bool) {
        require(balanceOf[msg.sender] >= value, "Insufficient balance");
        
        balanceOf[msg.sender] -= value;
        balanceOf[to] += value;
        
        emit Transfer(msg.sender, to, value);
        return true;
    }
    
    // 승인
    function approve(address spender, uint value) public returns (bool) {
        allowance[msg.sender][spender] = value;
        emit Approval(msg.sender, spender, value);
        return true;
    }
    
    // 승인된 금액 전송
    function transferFrom(address from, address to, uint value) public returns (bool) {
        require(value <= balanceOf[from], "Insufficient balance");
        require(value <= allowance[from][msg.sender], "Insufficient allowance");
        
        balanceOf[from] -= value;
        balanceOf[to] += value;
        allowance[from][msg.sender] -= value;
        
        emit Transfer(from, to, value);
        return true;
    }
}
```

### 가스 비용

```
연산별 가스:
├─ PUSH1: 3 gas
├─ ADD: 3 gas
├─ SLOAD (메모리 읽기): 2,100 gas
├─ SSTORE (메모리 쓰기): 20,000 gas
└─ CALL (외부 호출): 100 gas

예: transfer() 호출
├─ 사전 조건 검증: 500 gas
├─ 상태 변수 2개 수정: 2 × 20,000 = 40,000 gas
├─ 이벤트 발생: 375 gas
└─ 합계: ~41,000 gas

Ethereum 기준 (가스 가격 30 gwei):
41,000 gas × 30 gwei = 1,230,000 wei = 0.00123 ETH ≈ $2-3 (2024년)
```

---

## 3. 보안 취약점

### Reentrancy 공격

```solidity
// ❌ 취약한 코드 (The DAO 해킹)
contract Vulnerable {
    mapping(address => uint) public balances;
    
    function withdraw(uint amount) public {
        require(balance[msg.sender] >= amount);
        
        // 1. 외부 호출 (상태 업데이트 전!)
        (bool success, ) = msg.sender.call{value: amount}("");
        require(success);
        
        // 2. 상태 업데이트 (너무 늦음!)
        balances[msg.sender] -= amount;
    }
}

공격 시나리오:
1. 공격자가 withdraw(1 ETH) 호출
2. call{value: 1}("")로 1 ETH 전송
3. 공격자 컨트랙트의 receive()가 자동 호출
4. receive()에서 다시 withdraw(1 ETH) 호출
5. balances[attacker]는 아직 차감 안됨!
6. 또 1 ETH 인출... 반복
```

### 해결책: Checks-Effects-Interactions 패턴

```solidity
// ✅ 안전한 코드
contract Secure {
    mapping(address => uint) public balances;
    
    function withdraw(uint amount) public {
        // 1. Checks: 조건 검증
        require(balances[msg.sender] >= amount, "Insufficient balance");
        
        // 2. Effects: 상태 업데이트 (먼저!)
        balances[msg.sender] -= amount;
        
        // 3. Interactions: 외부 호출 (나중에!)
        (bool success, ) = msg.sender.call{value: amount}("");
        require(success, "Transfer failed");
    }
}
```

### Over/Underflow

```solidity
// ❌ Solidity < 0.8.0 (오버플로우 취약)
contract Vulnerable {
    uint8 public balance = 255;
    
    function add(uint8 amount) public {
        balance += amount;  // 255 + 1 = 0 (오버플로우!)
    }
}

// ✅ Solidity >= 0.8.0 (자동 검사)
contract Safe {
    uint8 public balance = 255;
    
    function add(uint8 amount) public {
        balance += amount;  // 자동으로 revert (safe)
    }
}
```

---

## 4. 성능 최적화

### 저장소 최적화 (Packing)

```solidity
// ❌ 비효율적 (메모리 3개 슬롯)
contract Bad {
    uint8 public flag;      // 32 bytes 낭비
    address public owner;   // 32 bytes 낭비
    uint256 public balance; // 32 bytes
}

// ✅ 효율적 (메모리 2개 슬롯)
contract Good {
    address public owner;   // 20 bytes
    uint8 public flag;      // 1 byte + 11 bytes 패딩
                            // 합계: 32 bytes (1개 슬롯)
    uint256 public balance; // 32 bytes (1개 슬롯)
}

절약: SLOAD/SSTORE 1회 → 2,100 gas 절감
```

### 배열 vs 매핑

```solidity
// ❌ 배열 (배포 후 추가하려면 복사 필요)
contract BadArray {
    address[] public participants;  // 동적 배열
    
    function addParticipant(address user) public {
        participants.push(user);  // 가스 비용 증가 (배열 크기와 무관)
    }
}

// ✅ 매핑 (O(1) 접근, 가스 낮음)
contract GoodMapping {
    mapping(address => bool) public isParticipant;
    
    function addParticipant(address user) public {
        isParticipant[user] = true;  // 항상 ~20,000 gas
    }
}
```

---

## 5. 실전 예제: Uniswap V2 풀 (간략화)

```solidity
pragma solidity ^0.8.0;

contract SimpleAMM {
    address public token0;
    address public token1;
    
    uint public reserve0;
    uint public reserve1;
    uint public kLast;  // reserve0 * reserve1
    
    function addLiquidity(uint amount0, uint amount1) public {
        // LP 토큰 발급 (간략화)
        reserve0 += amount0;
        reserve1 += amount1;
        kLast = reserve0 * reserve1;
    }
    
    // x * y = k (상수곡선 공식)
    function swap(uint amountIn0, address to) public {
        // amountIn0 만큼 token0을 넣으면, token1을 얼마나 받는가?
        
        // 새로운 reserve1 계산
        uint newReserve0 = reserve0 + amountIn0;
        uint newReserve1 = kLast / newReserve0;
        
        uint amountOut1 = reserve1 - newReserve1;
        
        // 상태 업데이트
        reserve0 = newReserve0;
        reserve1 = newReserve1;
        
        // 토큰 전송
        IERC20(token1).transfer(to, amountOut1);
    }
}
```

---

## 6. 스마트 컨트랙트 검증

### 정적 분석

```bash
# Slither (정적 분석기)
slither MyContract.sol

# 출력
Contract MyContract:
  ⚠️  Missing checks for address(0): setOwner(address newOwner)
  ⚠️  Reentrancy in withdraw()
  ✓ No integer overflow in add()
```

### 포멀 검증 (Formal Verification)

```
수학적 증명으로 코드가 항상 안전함을 보장

예: x * y = k (Uniswap)
증명: reserve0_after * reserve1_after >= reserve0 * reserve1
→ 풀의 자산이 감소하지 않음을 증명
```

---

## 7. L2 스케일링

### Rollups

```
L1 (Ethereum):
├─ 완전한 보안
├─ 높은 가스비 ($10-100/tx)
└─ TPS: 15

Optimistic Rollup:
├─ L2에서 배치 처리
├─ 분쟁 기반 검증
├─ 낮은 가스비 ($0.10-1/tx)
└─ TPS: 1,000-4,000

ZK Rollup:
├─ 영지식 증명
├─ 즉시 최종성
├─ 낮은 가스비 ($0.05-0.5/tx)
└─ TPS: 1,000-10,000
```

### Arbitrum 예제

```solidity
// L2에서 실행 (가스 100배 저렴)
contract L2Token {
    mapping(address => uint) public balance;
    
    function transfer(address to, uint amount) public {
        balance[msg.sender] -= amount;
        balance[to] += amount;
        // L1 Ethereum: $1 가스비
        // L2 Arbitrum: $0.01 가스비 (100배!)
    }
}
```

---

## 8. 벤치마크

| 작업 | Ethereum L1 | Arbitrum L2 | Optimism L2 | Solana |
|------|-------------|------------|------------|--------|
| **ERC20 전송** | $2-10 | $0.02-0.1 | $0.05-0.2 | $0.00025 |
| **Swap** | $10-50 | $0.1-0.5 | $0.3-1 | $0.00025 |
| **스테이킹** | $5-20 | $0.05-0.2 | $0.2-0.5 | $0.00025 |
| **배포** | $500-2000 | $5-20 | $20-50 | $10-50 |

---

## 핵심 정리

| 개념 | 설명 | 위험 |
|------|------|------|
| **Solidity** | EVM 스마트 컨트랙트 | Reentrancy, Over/Underflow |
| **가스 최적화** | 저장소 패킹, 배열→매핑 | 악의적 SLOAD 폭탄 |
| **L2 Rollup** | 확장성 솔루션 | 크로스체인 브릿지 위험 |

---

## 결론

**"스마트 컨트랙트는 규칙이 투명하고 항상 실행되는 미래"**

Ethereum의 DeFi 생태계 = 스마트 컨트랙트의 가능성 증명 🚀

---

질문이나 피드백은 댓글로 남겨주세요! 💬
