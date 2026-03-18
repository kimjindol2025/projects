"""
CrossCheckAgent: 벡터 임베딩 기반 시맨틱 검증
역할: 소스 간 정보 충돌 감지 및 신뢰도 조정
"""

import json
from dataclasses import dataclass, field
from typing import List, Tuple, Optional
from concurrent.futures import ThreadPoolExecutor, as_completed

from .content_fetcher import FetchedContent


@dataclass
class ConflictReport:
    """충돌 리포트"""
    conflicting_pairs: List[Tuple[str, str]] = field(default_factory=list)
    avg_similarity: float = 0.0
    outlier_urls: List[str] = field(default_factory=list)
    adjusted_confidence: float = 0.0
    warnings: List[str] = field(default_factory=list)


class VectorEmbedder:
    """OpenAI API를 이용한 벡터 임베딩"""

    def __init__(self, api_key: str):
        self.api_key = api_key
        self.api_url = "https://api.openai.com/v1/embeddings"

    def embed_batch(self, texts: List[str]) -> Optional[List[List[float]]]:
        """
        텍스트 배치를 벡터로 변환

        Args:
            texts: 텍스트 리스트 (최대 20개)

        Returns:
            벡터 리스트 또는 None (실패 시)
        """
        if not texts:
            return None

        try:
            import requests

            response = requests.post(
                self.api_url,
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                },
                json={
                    "input": texts[:20],  # 최대 20개
                    "model": "text-embedding-3-small",
                },
                timeout=30,
            )

            if response.status_code == 200:
                data = response.json()
                # 응답 형식: {"data": [{"embedding": [...], "index": 0}, ...]}
                embeddings = []
                for item in sorted(data.get("data", []), key=lambda x: x.get("index", 0)):
                    embeddings.append(item.get("embedding", []))
                return embeddings if embeddings else None
            else:
                return None

        except Exception:
            return None


class SemanticValidator:
    """시맨틱 검증"""

    @staticmethod
    def cosine_similarity(a: List[float], b: List[float]) -> float:
        """
        코사인 유사도 계산 (numpy 없이)

        Args:
            a, b: 벡터

        Returns:
            유사도 (0.0 ~ 1.0)
        """
        if not a or not b or len(a) != len(b):
            return 0.0

        dot_product = sum(x * y for x, y in zip(a, b))
        norm_a = sum(x * x for x in a) ** 0.5
        norm_b = sum(x * x for x in b) ** 0.5

        if norm_a == 0 or norm_b == 0:
            return 0.0

        return dot_product / (norm_a * norm_b)

    @staticmethod
    def detect_outliers(
        similarities: List[float], threshold_sigma: float = 1.5
    ) -> List[int]:
        """
        이상치 탐지 (평균 - threshold_sigma * 표준편차)

        Args:
            similarities: 유사도 리스트
            threshold_sigma: 표준편차 배수

        Returns:
            이상치 인덱스 리스트
        """
        if len(similarities) < 2:
            return []

        mean = sum(similarities) / len(similarities)
        variance = sum((x - mean) ** 2 for x in similarities) / len(similarities)
        std_dev = variance ** 0.5

        threshold = mean - (threshold_sigma * std_dev)
        outliers = [i for i, sim in enumerate(similarities) if sim < threshold]

        return outliers


class CrossCheckAgent:
    """CrossCheckAgent: 벡터 기반 정보 충돌 검증"""

    def __init__(self, openai_api_key: str = ""):
        self.openai_api_key = openai_api_key
        self.embedder = VectorEmbedder(openai_api_key) if openai_api_key else None
        self.validator = SemanticValidator()
        self.max_text_length = 300

    def run(
        self, contents: List[FetchedContent], base_confidence: float = 0.85
    ) -> ConflictReport:
        """
        컨텐츠들에 대한 시맨틱 검증 실행

        Args:
            contents: FetchedContent 리스트
            base_confidence: 기본 신뢰도

        Returns:
            ConflictReport
        """
        report = ConflictReport(adjusted_confidence=base_confidence)

        if not contents or len(contents) < 2:
            return report

        # OpenAI API 미사용 시 폴백
        if not self.embedder or not self.openai_api_key:
            return report

        # Step 1: 각 소스에서 텍스트 추출 (최대 300자)
        texts = []
        url_map = {}  # 텍스트 인덱스 → URL

        for content in contents[:20]:  # 최대 20개
            text = content.body_text[: self.max_text_length]
            if text.strip():
                texts.append(text)
                url_map[len(texts) - 1] = content.url

        if len(texts) < 2:
            return report

        # Step 2: 벡터 임베딩
        embeddings = self.embedder.embed_batch(texts)
        if not embeddings or len(embeddings) < 2:
            return report

        # Step 3: 코사인 유사도 계산
        similarities = []
        pairs = []

        for i in range(len(embeddings)):
            for j in range(i + 1, len(embeddings)):
                sim = self.validator.cosine_similarity(embeddings[i], embeddings[j])
                similarities.append(sim)
                pairs.append((url_map.get(i, ""), url_map.get(j, "")))

        if not similarities:
            return report

        # Step 4: 이상치 탐지
        avg_similarity = sum(similarities) / len(similarities)
        outlier_indices = self.validator.detect_outliers(similarities, threshold_sigma=1.5)

        # 이상치 쌍 추출
        for idx in outlier_indices:
            if idx < len(pairs):
                report.outlier_urls.append(pairs[idx][0])
                report.outlier_urls.append(pairs[idx][1])
                report.conflicting_pairs.append(pairs[idx])
                report.warnings.append(
                    f"정보 충돌 감지: '{pairs[idx][0]}' vs '{pairs[idx][1]}' "
                    f"(유사도: {similarities[idx]:.2f})"
                )

        # Step 5: 신뢰도 조정
        report.avg_similarity = avg_similarity
        conflict_penalty = len(report.warnings) * 0.05  # 경고당 0.05 패널티
        report.adjusted_confidence = max(0.0, base_confidence - conflict_penalty)

        return report


class CrossCheckValidator:
    """통합 검증자"""

    @staticmethod
    def integrate_with_consensus(
        consensus_confidence: float, crosscheck_report: ConflictReport
    ) -> float:
        """
        Consensus 신뢰도와 CrossCheck 결과를 통합

        Args:
            consensus_confidence: Consensus Engine 신뢰도
            crosscheck_report: CrossCheck 리포트

        Returns:
            최종 신뢰도
        """
        # CrossCheck adjusted_confidence를 반영 (가중치 50%)
        final_confidence = (
            consensus_confidence * 0.5 + crosscheck_report.adjusted_confidence * 0.5
        )

        return final_confidence
