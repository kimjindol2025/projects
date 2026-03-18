"""
Multi-Agent Researcher System for Genspark Clone v2.0
- BaseResearcherAgent: 기본 에이전트
- GeneralAgent: 범용 검색
- TechAgent: 기술 문서 + GitHub/StackOverflow 선호
- NewsAgent: 최신 뉴스 + Reddit/Medium 선호
- ReviewAgent: 리뷰 + Reddit/Dev.to 선호
"""

import time
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import List, Optional
from datetime import datetime

from ..query_analyzer import QuerySpec


@dataclass
class AgentSearchResult:
    """에이전트 검색 결과"""
    agent_name: str
    query_used: str
    search_results: list = field(default_factory=list)  # SearchResult 리스트
    fetched_contents: list = field(default_factory=list)  # FetchedContent 리스트
    execution_time: float = 0.0
    error: Optional[str] = None


class BaseResearcherAgent(ABC):
    """기본 리서처 에이전트"""

    def __init__(self, searcher, fetcher):
        """
        Args:
            searcher: WebSearcher 인스턴스
            fetcher: ContentFetcher 인스턴스
        """
        self.searcher = searcher
        self.fetcher = fetcher
        self.agent_name = "base"

    @abstractmethod
    def _build_queries(self, query_spec: QuerySpec) -> List[str]:
        """
        서브쿼리 생성 (서브클래스 오버라이드)

        Args:
            query_spec: 질문 스펙

        Returns:
            검색 쿼리 리스트
        """
        pass

    def research(self, query_spec: QuerySpec) -> AgentSearchResult:
        """
        리서치 실행

        Args:
            query_spec: 질문 스펙

        Returns:
            AgentSearchResult
        """
        start_time = time.time()

        try:
            # 서브쿼리 생성
            queries = self._build_queries(query_spec)

            # 검색 실행
            search_results = []
            for query in queries:
                results = self.searcher.search(query, max_results=5)
                search_results.extend(results)
                time.sleep(1)  # 봇 차단 회피

            # 콘텐츠 페칭
            urls = list({r.url for r in search_results})[:10]  # 중복 제거, 최대 10개
            fetched = self.fetcher.fetch_urls(urls)

            execution_time = time.time() - start_time

            return AgentSearchResult(
                agent_name=self.agent_name,
                query_used=query_spec.original_query,
                search_results=search_results,
                fetched_contents=fetched,
                execution_time=execution_time,
                error=None
            )

        except Exception as e:
            execution_time = time.time() - start_time
            return AgentSearchResult(
                agent_name=self.agent_name,
                query_used=query_spec.original_query,
                search_results=[],
                fetched_contents=[],
                execution_time=execution_time,
                error=str(e)
            )


class GeneralAgent(BaseResearcherAgent):
    """범용 리서처"""

    def __init__(self, searcher, fetcher):
        super().__init__(searcher, fetcher)
        self.agent_name = "general"

    def _build_queries(self, query_spec: QuerySpec) -> List[str]:
        """서브쿼리를 그대로 반환"""
        return query_spec.sub_queries


class TechAgent(BaseResearcherAgent):
    """기술 전문가"""

    def __init__(self, searcher, fetcher):
        super().__init__(searcher, fetcher)
        self.agent_name = "tech"

    def _build_queries(self, query_spec: QuerySpec) -> List[str]:
        """기술 문서 중심 쿼리"""
        queries = []
        for sub_query in query_spec.sub_queries:
            # 문서 + 공식 저장소 키워드 추가
            queries.append(f"{sub_query} documentation")
            queries.append(f"{sub_query} github")
            queries.append(f"{sub_query} official")
        return queries[:5]  # 최대 5개


class NewsAgent(BaseResearcherAgent):
    """뉴스/트렌드 전문가"""

    def __init__(self, searcher, fetcher):
        super().__init__(searcher, fetcher)
        self.agent_name = "news"

    def _build_queries(self, query_spec: QuerySpec) -> List[str]:
        """최신 뉴스 중심 쿼리"""
        year = datetime.now().year
        queries = []
        for sub_query in query_spec.sub_queries:
            queries.append(f"{sub_query} 2025 latest")
            queries.append(f"{sub_query} news {year}")
            queries.append(f"{sub_query} trending")
        return queries[:5]


class ReviewAgent(BaseResearcherAgent):
    """리뷰/경험 전문가"""

    def __init__(self, searcher, fetcher):
        super().__init__(searcher, fetcher)
        self.agent_name = "review"

    def _build_queries(self, query_spec: QuerySpec) -> List[str]:
        """리뷰/장단점 중심 쿼리"""
        queries = []
        for sub_query in query_spec.sub_queries:
            queries.append(f"{sub_query} review")
            queries.append(f"{sub_query} pros cons")
            queries.append(f"{sub_query} experience")
        return queries[:5]
