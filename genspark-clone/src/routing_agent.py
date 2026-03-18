"""
Routing Agent: DAG 기반 쿼리 의존성 분석 및 실행 최적화
역할: 서브쿼리 의존 관계 분석 → 병렬/순차 실행 계획 수립
"""

import json
import time
from dataclasses import dataclass, field
from typing import List, Dict, Optional
from concurrent.futures import ThreadPoolExecutor, as_completed

from .query_analyzer import QuerySpec
from .web_searcher import DuckDuckGoSearcher
from .content_fetcher import ContentFetcher, FetchedContent


@dataclass
class SubQueryNode:
    """서브쿼리 노드"""
    query_id: str
    query: str
    depends_on: List[str] = field(default_factory=list)  # 이전 쿼리 ID
    is_parallel: bool = False


@dataclass
class ExecutionPlan:
    """실행 계획"""
    parallel_groups: List[List[str]] = field(default_factory=list)  # 동시 실행 그룹
    sequential_chains: List[List[str]] = field(default_factory=list)  # 순차 체인
    node_map: Dict[str, SubQueryNode] = field(default_factory=dict)


@dataclass
class ExecutionResult:
    """실행 결과"""
    query: str
    plan: ExecutionPlan
    all_contents: List[FetchedContent] = field(default_factory=list)
    execution_time: float = 0.0
    parallel_speedup: float = 1.0


class QueryDependencyAnalyzer:
    """Claude haiku를 이용한 쿼리 의존성 분석"""

    def __init__(self, api_key: str):
        self.api_key = api_key
        self.api_url = "https://api.anthropic.com/v1/messages"

    def analyze(self, sub_queries: List[str]) -> Dict[str, List[str]]:
        """
        서브쿼리들의 의존 관계 분석

        Args:
            sub_queries: 서브쿼리 리스트

        Returns:
            의존 관계 JSON: {"q0": [], "q1": ["q0"], ...}
        """
        try:
            import requests

            prompt = f"""다음 서브쿼리들의 의존 관계를 분석하세요.
각 쿼리가 다른 쿼리의 결과에 의존하는지 판단하고, 의존 관계를 JSON으로 반환하세요.

서브쿼리들:
{json.dumps(sub_queries, ensure_ascii=False, indent=2)}

JSON 형식 (의존성이 없으면 빈 리스트):
{{
    "q0": {{"depends_on": []}},
    "q1": {{"depends_on": ["q0"]}},
    ...
}}

결과를 JSON만 반환하세요 (마크다운 코드블록 없음)."""

            response = requests.post(
                self.api_url,
                headers={
                    "anthropic-version": "2023-06-01",
                    "content-type": "application/json",
                    "x-api-key": self.api_key,
                },
                json={
                    "model": "claude-haiku-4-5-20251001",
                    "max_tokens": 1000,
                    "messages": [{"role": "user", "content": prompt}],
                },
                timeout=30,
            )

            if response.status_code == 200:
                content = response.json()["content"][0]["text"]
                # JSON 추출
                try:
                    deps = json.loads(content)
                    # 의존성 딕셔너리 구성
                    result = {}
                    for i, query in enumerate(sub_queries):
                        q_id = f"q{i}"
                        if q_id in deps and isinstance(deps[q_id], dict):
                            result[q_id] = deps[q_id].get("depends_on", [])
                        else:
                            result[q_id] = []
                    return result
                except json.JSONDecodeError:
                    # 폴백: 의존성 없음
                    return {f"q{i}": [] for i in range(len(sub_queries))}
            else:
                return {f"q{i}": [] for i in range(len(sub_queries))}

        except Exception:
            # 폴백: 의존성 없음
            return {f"q{i}": [] for i in range(len(sub_queries))}


class ExecutionPlanner:
    """실행 계획 수립"""

    @staticmethod
    def create_plan(
        sub_queries: List[str], dependencies: Dict[str, List[str]]
    ) -> ExecutionPlan:
        """
        의존성 정보를 기반으로 실행 계획 수립

        Args:
            sub_queries: 서브쿼리 리스트
            dependencies: 의존성 정보

        Returns:
            ExecutionPlan
        """
        plan = ExecutionPlan()

        # 노드 맵 구성
        for i, query in enumerate(sub_queries):
            q_id = f"q{i}"
            deps = dependencies.get(q_id, [])
            plan.node_map[q_id] = SubQueryNode(
                query_id=q_id,
                query=query,
                depends_on=deps,
                is_parallel=len(deps) == 0,
            )

        # 병렬 그룹 구성 (의존성 없는 쿼리들)
        parallel_group = [
            q_id for q_id, node in plan.node_map.items() if node.is_parallel
        ]
        if parallel_group:
            plan.parallel_groups.append(parallel_group)

        # 순차 체인 구성 (의존성 있는 쿼리들)
        for q_id, node in plan.node_map.items():
            if not node.is_parallel:
                chain = node.depends_on + [q_id]
                plan.sequential_chains.append(chain)

        return plan


class RoutingAgent:
    """Routing Agent: 쿼리 의존성 분석 + 병렬/순차 실행 최적화"""

    def __init__(
        self,
        api_key: str,
        searcher: DuckDuckGoSearcher,
        fetcher: ContentFetcher,
        max_workers: int = 2,
    ):
        self.api_key = api_key
        self.searcher = searcher
        self.fetcher = fetcher
        self.max_workers = max_workers
        self.analyzer = QueryDependencyAnalyzer(api_key)

    def run(self, query_spec: QuerySpec) -> ExecutionResult:
        """
        서브쿼리 실행 계획 수립 및 실행

        Args:
            query_spec: 질문 분석 결과

        Returns:
            ExecutionResult
        """
        start_time = time.time()

        # Step 1: 의존성 분석
        dependencies = self.analyzer.analyze(query_spec.sub_queries)

        # Step 2: 실행 계획 수립
        plan = ExecutionPlanner.create_plan(query_spec.sub_queries, dependencies)

        # Step 3: 병렬/순차 실행
        all_contents = []
        execution_map = {}  # q_id → [FetchedContent]

        # 병렬 그룹 실행
        for group in plan.parallel_groups:
            group_contents = self._execute_parallel(group, plan.node_map)
            for q_id, contents in group_contents.items():
                execution_map[q_id] = contents
                all_contents.extend(contents)

        # 순차 체인 실행
        for chain in plan.sequential_chains:
            chain_contents = self._execute_sequential(chain, plan.node_map, execution_map)
            for q_id, contents in chain_contents.items():
                execution_map[q_id] = contents
                all_contents.extend(contents)

        # 순차 실행 시간 추정
        parallel_speedup = max(1.0, len(query_spec.sub_queries) / max(len(plan.parallel_groups), 1))

        execution_time = time.time() - start_time

        return ExecutionResult(
            query=query_spec.original_query,
            plan=plan,
            all_contents=all_contents,
            execution_time=execution_time,
            parallel_speedup=parallel_speedup,
        )

    def _execute_parallel(
        self, group: List[str], node_map: Dict[str, SubQueryNode]
    ) -> Dict[str, List[FetchedContent]]:
        """
        그룹의 쿼리들을 병렬 실행

        Args:
            group: 쿼리 ID 리스트 (의존성 없음)
            node_map: 노드 맵

        Returns:
            q_id → [FetchedContent] 맵
        """
        results = {}

        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            futures = {
                executor.submit(self._execute_single_query, node_map[q_id]): q_id
                for q_id in group
            }

            for future in as_completed(futures):
                q_id = futures[future]
                try:
                    contents = future.result()
                    results[q_id] = contents
                except Exception:
                    results[q_id] = []

        return results

    def _execute_sequential(
        self,
        chain: List[str],
        node_map: Dict[str, SubQueryNode],
        execution_map: Dict[str, List[FetchedContent]],
    ) -> Dict[str, List[FetchedContent]]:
        """
        체인의 쿼리들을 순차 실행 (컨텍스트 주입)

        Args:
            chain: 쿼리 ID 리스트 (의존성 순서)
            node_map: 노드 맵
            execution_map: 이미 실행된 쿼리 결과

        Returns:
            q_id → [FetchedContent] 맵
        """
        results = {}

        for q_id in chain:
            node = node_map[q_id]

            # 이전 결과에서 컨텍스트 추출
            context = ""
            for dep_id in node.depends_on:
                if dep_id in execution_map:
                    # 최대 200자 요약
                    dep_contents = execution_map[dep_id]
                    if dep_contents:
                        text = " ".join(c.body_text[:200] for c in dep_contents[:2])
                        context += text[:200] + " "

            # 컨텍스트가 있으면 쿼리에 추가
            enhanced_query = node.query
            if context.strip():
                enhanced_query = f"{node.query} (관련 맥락: {context.strip()[:150]})"

            try:
                contents = self._execute_single_query_enhanced(node, enhanced_query)
                results[q_id] = contents
                execution_map[q_id] = contents
            except Exception:
                results[q_id] = []
                execution_map[q_id] = []

            time.sleep(0.5)  # 봇 차단 회피

        return results

    def _execute_single_query(self, node: SubQueryNode) -> List[FetchedContent]:
        """단일 쿼리 실행"""
        search_results = self.searcher.search(node.query)
        urls = list({r.url for r in search_results})[:10]
        return self.fetcher.fetch_urls(urls)

    def _execute_single_query_enhanced(
        self, node: SubQueryNode, enhanced_query: str
    ) -> List[FetchedContent]:
        """컨텍스트가 추가된 쿼리 실행"""
        # 원본 쿼리로 검색 (enhanced_query는 로깅용)
        search_results = self.searcher.search(node.query)
        urls = list({r.url for r in search_results})[:10]
        return self.fetcher.fetch_urls(urls)
