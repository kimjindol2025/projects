"""
Consensus Engine for Multi-Agent Results
- 여러 에이전트 결과 병합
- URL 기반 중복 제거
- 도메인 오버랩 계산
- 정보 충돌 감지
"""

from dataclasses import dataclass, field
from typing import List, Optional
import re


@dataclass
class ConsensusResult:
    """합의 결과"""
    query: str
    agent_results: List = field(default_factory=list)  # AgentSearchResult 리스트
    merged_contents: List = field(default_factory=list)  # FetchedContent 리스트
    overall_confidence: float = 0.0
    conflict_warnings: List[str] = field(default_factory=list)
    total_unique_sources: int = 0


class ConsensusEngine:
    """합의 엔진"""

    CONFLICT_KEYWORDS = ["vs", "not", "wrong", "bad", "don't", "avoid", "unlike"]

    def __init__(self):
        self.conflict_keywords = self.CONFLICT_KEYWORDS

    def run(self, query: str, agent_results: List) -> ConsensusResult:
        """
        에이전트 결과 합의

        Args:
            query: 원본 쿼리
            agent_results: AgentSearchResult 리스트

        Returns:
            ConsensusResult
        """
        # 콘텐츠 병합 (URL 기준 중복 제거)
        merged = self._merge_contents(agent_results)

        # 도메인 오버랩 계산
        domain_overlap = self._calculate_domain_overlap(agent_results)

        # 정보 충돌 감지
        conflicts = self._detect_conflicts(agent_results)

        # 신뢰도 계산
        overall_confidence = self._calculate_confidence(
            agent_results, domain_overlap, conflicts
        )

        return ConsensusResult(
            query=query,
            agent_results=agent_results,
            merged_contents=merged,
            overall_confidence=overall_confidence,
            conflict_warnings=conflicts,
            total_unique_sources=len(merged)
        )

    def _merge_contents(self, agent_results: List) -> List:
        """
        URL 기준 콘텐츠 병합

        Args:
            agent_results: AgentSearchResult 리스트

        Returns:
            병합된 FetchedContent 리스트
        """
        seen_urls = set()
        merged = []

        for result in agent_results:
            for content in result.fetched_contents:
                url = content.url if hasattr(content, 'url') else str(content)
                if url not in seen_urls:
                    seen_urls.add(url)
                    merged.append(content)

        return merged

    def _calculate_domain_overlap(self, agent_results: List) -> float:
        """
        도메인 오버랩 계산

        Args:
            agent_results: AgentSearchResult 리스트

        Returns:
            오버랩 비율 (0.0 ~ 1.0)
        """
        if not agent_results:
            return 0.0

        domain_sets = []
        for result in agent_results:
            domains = set()
            for content in result.fetched_contents:
                url = content.url if hasattr(content, 'url') else str(content)
                domain = self._extract_domain(url)
                if domain:
                    domains.add(domain)
            if domains:
                domain_sets.append(domains)

        if not domain_sets:
            return 0.0

        # 모든 도메인 수집
        all_domains = set()
        for domains in domain_sets:
            all_domains.update(domains)

        # 공통 도메인 수
        common = domain_sets[0]
        for domains in domain_sets[1:]:
            common = common.intersection(domains)

        if not all_domains:
            return 0.0

        return len(common) / len(all_domains)

    def _detect_conflicts(self, agent_results: List) -> List[str]:
        """
        정보 충돌 감지

        Args:
            agent_results: AgentSearchResult 리스트

        Returns:
            경고 메시지 리스트
        """
        conflicts = []

        # 에이전트별 타이틀 텍스트 수집
        all_text = ""
        for result in agent_results:
            for content in result.fetched_contents:
                title = content.title if hasattr(content, 'title') else ""
                if title:
                    all_text += " " + title.lower()

        all_text = all_text.lower()

        # 충돌 키워드 확인
        for keyword in self.conflict_keywords:
            if keyword in all_text:
                count = all_text.count(keyword)
                if count >= 1:  # 최소 1개 이상 포함
                    conflicts.append(
                        f"Potential conflict detected: '{keyword}' found "
                        f"(some results may present different perspectives)"
                    )
                    break  # 중복 방지

        return conflicts

    def _calculate_confidence(
        self,
        agent_results: List,
        domain_overlap: float,
        conflicts: List[str]
    ) -> float:
        """
        신뢰도 계산

        Args:
            agent_results: AgentSearchResult 리스트
            domain_overlap: 도메인 오버랩 비율
            conflicts: 충돌 경고 리스트

        Returns:
            신뢰도 (0.0 ~ 1.0)
        """
        if not agent_results:
            return 0.0

        # 기본 신뢰도: 성공한 에이전트 비율
        successful_agents = sum(1 for r in agent_results if not r.error)
        success_rate = successful_agents / len(agent_results)

        # 콘텐츠 수: 많을수록 신뢰도 높음
        total_contents = sum(len(r.fetched_contents) for r in agent_results)
        content_score = min(total_contents / 10, 1.0)  # 10개 이상이면 1.0

        # 합의도: 도메인 오버랩이 높을수록 신뢰도 높음
        consensus_score = domain_overlap

        # 충돌 페널티
        conflict_penalty = len(conflicts) * 0.1

        # 최종 신뢰도 = (성공율 + 콘텐츠 + 합의) / 3 - 충돌 페널티
        confidence = (success_rate + content_score + consensus_score) / 3
        confidence = max(confidence - conflict_penalty, 0.0)

        return min(confidence, 1.0)

    @staticmethod
    def _extract_domain(url: str) -> Optional[str]:
        """
        URL에서 도메인 추출

        Args:
            url: URL 문자열

        Returns:
            도메인 또는 None
        """
        try:
            match = re.search(r'https?://(?:www\.)?([a-z0-9.-]+)', url)
            return match.group(1) if match else None
        except Exception:
            return None
