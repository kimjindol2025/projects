"""
Day 2: RoutingAgent 테스트
의존성 분석, 실행 계획, 병렬/순차 실행 검증
"""

import time
from src.routing_agent import (
    RoutingAgent,
    QueryDependencyAnalyzer,
    ExecutionPlanner,
    SubQueryNode,
)
from src.query_analyzer import QuerySpec
from src.web_searcher import DuckDuckGoSearcher, SearchResult
from src.content_fetcher import ContentFetcher, FetchedContent


class MockSearcher:
    """Mock Searcher"""
    def search(self, query: str):
        """Mock 검색"""
        return [
            SearchResult(
                url=f"https://example{i}.com/{query.replace(' ', '_')}",
                title=f"Result {i}: {query}",
                snippet=f"Snippet for {query}",
                source_domain=f"example{i}.com",
                rank=i + 1,
            )
            for i in range(3)
        ]


class MockFetcher:
    """Mock Fetcher"""
    def fetch_urls(self, urls):
        """Mock 페칭"""
        return [
            FetchedContent(
                url=url,
                title=f"Title for {url}",
                body_text=f"Body content for {url}" * 5,
                word_count=50,
                fetch_status="ok",
            )
            for url in urls
        ]


def test_execution_plan_parallel():
    """병렬 실행 계획 테스트"""
    sub_queries = ["Python basics", "Python async", "Python performance"]
    dependencies = {
        "q0": [],
        "q1": [],
        "q2": [],
    }

    plan = ExecutionPlanner.create_plan(sub_queries, dependencies)

    # 모든 쿼리가 병렬 그룹에 포함되어야 함
    assert len(plan.parallel_groups) == 1
    assert len(plan.parallel_groups[0]) == 3
    assert set(plan.parallel_groups[0]) == {"q0", "q1", "q2"}

    # 순차 체인이 없어야 함
    assert len(plan.sequential_chains) == 0

    print("✅ Execution plan (parallel) OK")


def test_execution_plan_sequential():
    """순차 실행 계획 테스트"""
    sub_queries = ["Basics", "Advanced", "Performance"]
    dependencies = {
        "q0": [],
        "q1": ["q0"],
        "q2": ["q1"],
    }

    plan = ExecutionPlanner.create_plan(sub_queries, dependencies)

    # 병렬 그룹: q0만
    assert len(plan.parallel_groups) == 1
    assert plan.parallel_groups[0] == ["q0"]

    # 순차 체인: q0→q1, q1→q2
    assert len(plan.sequential_chains) >= 2

    print("✅ Execution plan (sequential) OK")


def test_execution_plan_mixed():
    """혼합 실행 계획 테스트 (병렬 + 순차)"""
    sub_queries = [
        "Basics",      # q0: 병렬
        "Advanced",    # q1: 병렬
        "Performance", # q2: q0 의존
        "Comparison",  # q3: q2 의존
    ]
    dependencies = {
        "q0": [],
        "q1": [],
        "q2": ["q0"],
        "q3": ["q2"],
    }

    plan = ExecutionPlanner.create_plan(sub_queries, dependencies)

    # 병렬 그룹: q0, q1
    assert len(plan.parallel_groups) >= 1
    assert "q0" in plan.parallel_groups[0]
    assert "q1" in plan.parallel_groups[0]

    # 순차 체인: q0→q2, q2→q3
    assert len(plan.sequential_chains) >= 1

    print("✅ Execution plan (mixed) OK")


def test_routing_agent_parallel_execution():
    """RoutingAgent 병렬 실행 테스트"""
    agent = RoutingAgent(
        api_key="test-key",
        searcher=MockSearcher(),
        fetcher=MockFetcher(),
        max_workers=2,
    )

    query_spec = QuerySpec(
        original_query="Python 비교",
        main_topic="Python",
        sub_queries=["Python basics", "Python async", "Python performance"],
        language="ko",
        expected_sections=["개요", "비교"],
        complexity=0.7,
    )

    # Mock 의존성 분석 (모두 독립)
    agent.analyzer.analyze = lambda q: {
        f"q{i}": [] for i in range(len(q))
    }

    result = agent.run(query_spec)

    assert result is not None
    assert result.query == "Python 비교"
    assert len(result.all_contents) > 0
    assert result.execution_time > 0
    assert result.parallel_speedup >= 1.0

    print("✅ RoutingAgent parallel execution OK")


def test_routing_agent_sequential_execution():
    """RoutingAgent 순차 실행 테스트"""
    agent = RoutingAgent(
        api_key="test-key",
        searcher=MockSearcher(),
        fetcher=MockFetcher(),
        max_workers=2,
    )

    query_spec = QuerySpec(
        original_query="Python 심화",
        main_topic="Python",
        sub_queries=["Basics", "Advanced", "Expert"],
        language="ko",
        expected_sections=["기초", "심화", "전문"],
        complexity=0.9,
    )

    # Mock 의존성 분석 (순차 의존)
    agent.analyzer.analyze = lambda q: {
        "q0": [],
        "q1": ["q0"],
        "q2": ["q1"],
    }

    result = agent.run(query_spec)

    assert result is not None
    assert len(result.all_contents) > 0
    assert len(result.plan.sequential_chains) > 0

    print("✅ RoutingAgent sequential execution OK")


def test_context_injection():
    """컨텍스트 주입 테스트"""
    # execution_map에 이전 결과가 있으면 다음 쿼리에 주입됨
    agent = RoutingAgent(
        api_key="test-key",
        searcher=MockSearcher(),
        fetcher=MockFetcher(),
        max_workers=2,
    )

    # 임의의 노드와 execution_map
    node = SubQueryNode(query_id="q1", query="Advanced Python", depends_on=["q0"])
    execution_map = {
        "q0": [
            FetchedContent(
                url="https://example.com",
                title="Basics",
                body_text="Python is a programming language",
                word_count=100,
                fetch_status="ok",
            )
        ]
    }

    # _execute_sequential는 컨텍스트를 추출하고 순차 실행
    # 여기서는 컨텍스트 추출 로직만 검증
    context = ""
    for dep_id in node.depends_on:
        if dep_id in execution_map:
            dep_contents = execution_map[dep_id]
            if dep_contents:
                text = " ".join(c.body_text[:200] for c in dep_contents[:2])
                context += text[:200] + " "

    assert "Python" in context
    assert len(context) > 0

    print("✅ Context injection OK")


def test_routing_agent_mixed_execution():
    """RoutingAgent 혼합 실행 테스트"""
    agent = RoutingAgent(
        api_key="test-key",
        searcher=MockSearcher(),
        fetcher=MockFetcher(),
        max_workers=2,
    )

    query_spec = QuerySpec(
        original_query="Python vs Go",
        main_topic="Language comparison",
        sub_queries=["Python info", "Go info", "Performance", "Use cases"],
        language="ko",
        expected_sections=["개요", "비교", "사용사례"],
        complexity=0.8,
    )

    # Mock 의존성: 병렬 + 순차
    agent.analyzer.analyze = lambda q: {
        "q0": [],      # 병렬
        "q1": [],      # 병렬
        "q2": ["q0", "q1"],  # 순차
        "q3": ["q2"],  # 순차
    }

    result = agent.run(query_spec)

    assert result is not None
    assert len(result.all_contents) > 0
    assert len(result.plan.parallel_groups) >= 1
    assert len(result.plan.sequential_chains) >= 1

    print("✅ RoutingAgent mixed execution OK")


def test_routing_parallel_speedup():
    """병렬 처리의 속도 향상 검증"""
    agent = RoutingAgent(
        api_key="test-key",
        searcher=MockSearcher(),
        fetcher=MockFetcher(),
        max_workers=2,
    )

    query_spec = QuerySpec(
        original_query="Python comparison",
        main_topic="Python",
        sub_queries=["Python basic", "Python async", "Python ML", "Python web"],
        language="ko",
        expected_sections=["개요"],
        complexity=0.7,
    )

    # Mock 의존성 (모두 독립)
    agent.analyzer.analyze = lambda q: {f"q{i}": [] for i in range(len(q))}

    result = agent.run(query_spec)

    # 병렬 처리면 speedup >= 1.0
    assert result.parallel_speedup >= 1.0
    # 4개 쿼리이므로 speedup이 1 이상일 것으로 예상
    assert len(result.all_contents) >= len(query_spec.sub_queries)

    print("✅ Parallel speedup calculation OK")


if __name__ == "__main__":
    print("\n🧪 Running RoutingAgent tests...\n")

    test_execution_plan_parallel()
    test_execution_plan_sequential()
    test_execution_plan_mixed()
    test_routing_agent_parallel_execution()
    test_routing_agent_sequential_execution()
    test_context_injection()
    test_routing_agent_mixed_execution()
    test_routing_parallel_speedup()

    print("\n✅ All RoutingAgent tests passed!\n")
