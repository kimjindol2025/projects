"""
Multi-Agent 및 ConsensusEngine 테스트
"""

from src.agents.researcher_agent import (
    GeneralAgent,
    TechAgent,
    NewsAgent,
    ReviewAgent,
)
from src.consensus_engine import ConsensusEngine
from src.query_analyzer import QuerySpec


class MockSearcher:
    """Mock WebSearcher"""

    def search(self, query, max_results=5):
        """Mock 검색 결과"""
        class SearchResult:
            def __init__(self, i, query):
                self.url = f"https://example.com/{i}"
                self.title = f"Result {i}"
                self.snippet = f"Snippet for {query}"

        return [
            SearchResult(i, query)
            for i in range(min(max_results, 3))
        ]


class MockFetcher:
    """Mock ContentFetcher"""

    def fetch_urls(self, urls):
        """Mock 콘텐츠 페칭"""
        class FetchedContent:
            def __init__(self, url):
                self.url = url
                self.title = f"Title for {url}"
                self.body = f"Body content for {url}"
                self.status = "ok"

        return [FetchedContent(url) for url in urls]


def test_general_agent():
    """GeneralAgent 테스트"""
    searcher = MockSearcher()
    fetcher = MockFetcher()

    agent = GeneralAgent(searcher, fetcher)
    assert agent.agent_name == "general"

    query_spec = QuerySpec(
        original_query="파이썬",
        main_topic="파이썬",
        sub_queries=["파이썬 기초", "파이썬 라이브러리"],
        language="ko",
        expected_sections=["개요", "기능"],
        complexity=0.5,
    )

    result = agent.research(query_spec)
    assert result.agent_name == "general"
    assert not result.error
    assert len(result.search_results) > 0

    print("✅ GeneralAgent OK")


def test_tech_agent():
    """TechAgent 테스트"""
    searcher = MockSearcher()
    fetcher = MockFetcher()

    agent = TechAgent(searcher, fetcher)
    assert agent.agent_name == "tech"

    query_spec = QuerySpec(
        original_query="파이썬 asyncio",
        main_topic="asyncio",
        sub_queries=["asyncio 기초", "async/await"],
        language="ko",
        expected_sections=["개요", "예제"],
        complexity=0.7,
    )

    result = agent.research(query_spec)
    assert result.agent_name == "tech"
    # TechAgent는 "documentation" 키워드 추가
    assert not result.error

    print("✅ TechAgent OK")


def test_news_agent():
    """NewsAgent 테스트"""
    searcher = MockSearcher()
    fetcher = MockFetcher()

    agent = NewsAgent(searcher, fetcher)
    assert agent.agent_name == "news"

    query_spec = QuerySpec(
        original_query="AI 트렌드",
        main_topic="AI",
        sub_queries=["AI 2025"],
        language="ko",
        expected_sections=["최신 소식"],
        complexity=0.6,
    )

    result = agent.research(query_spec)
    assert result.agent_name == "news"
    assert not result.error

    print("✅ NewsAgent OK")


def test_review_agent():
    """ReviewAgent 테스트"""
    searcher = MockSearcher()
    fetcher = MockFetcher()

    agent = ReviewAgent(searcher, fetcher)
    assert agent.agent_name == "review"

    query_spec = QuerySpec(
        original_query="MacBook Pro 리뷰",
        main_topic="MacBook",
        sub_queries=["MacBook Pro 장단점"],
        language="ko",
        expected_sections=["리뷰"],
        complexity=0.5,
    )

    result = agent.research(query_spec)
    assert result.agent_name == "review"
    assert not result.error

    print("✅ ReviewAgent OK")


def test_consensus_merge():
    """ConsensusEngine 병합 테스트"""
    engine = ConsensusEngine()

    class MockResult:
        def __init__(self, name):
            self.agent_name = name
            self.error = None
            self.fetched_contents = [
                type("obj", (object,), {"url": f"https://site{i}.com"})()
                for i in range(2)
            ]

    agent_results = [
        MockResult("general"),
        MockResult("tech"),
        MockResult("news"),
    ]

    consensus = engine.run("테스트", agent_results)
    assert consensus.query == "테스트"
    assert consensus.total_unique_sources > 0
    assert consensus.overall_confidence >= 0.0

    print("✅ ConsensusEngine merge OK")


def test_consensus_domain_overlap():
    """도메인 오버랩 계산"""
    engine = ConsensusEngine()

    class MockResult:
        def __init__(self, urls):
            self.agent_name = "test"
            self.error = None
            self.fetched_contents = [
                type("obj", (object,), {"url": url})() for url in urls
            ]

    # 공통 도메인이 있는 경우
    agent_results = [
        MockResult(
            [
                "https://example.com/1",
                "https://example.com/2",
                "https://other.com/1",
            ]
        ),
        MockResult(
            [
                "https://example.com/3",
                "https://another.com/1",
            ]
        ),
    ]

    consensus = engine.run("테스트", agent_results)
    # 공통 도메인: example.com
    assert consensus.overall_confidence > 0.0

    print("✅ Domain overlap calculation OK")


def test_consensus_conflict_detection():
    """정보 충돌 감지"""
    engine = ConsensusEngine()

    class MockResult:
        def __init__(self, titles):
            self.agent_name = "test"
            self.error = None

            class MockContent:
                def __init__(self, title):
                    self.url = "https://example.com"
                    self.title = title

            self.fetched_contents = [MockContent(t) for t in titles]

    # 충돌 키워드 포함
    agent_results = [
        MockResult(["Python vs Java", "Performance comparison"]),
        MockResult(["Wrong approach", "Not recommended"]),
    ]

    consensus = engine.run("Python vs Java", agent_results)
    assert len(consensus.conflict_warnings) > 0
    # 충돌이 있으면 신뢰도가 낮음
    assert consensus.overall_confidence < 1.0

    print("✅ Conflict detection OK")


def test_consensus_confidence():
    """신뢰도 계산"""
    engine = ConsensusEngine()

    class MockResult:
        def __init__(self, error=None):
            self.agent_name = "test"
            self.error = error
            self.fetched_contents = [
                type("obj", (object,), {"url": f"https://example.com/{i}"})()
                for i in range(5)
            ]

    # 모든 에이전트 성공
    agent_results = [MockResult() for _ in range(4)]
    consensus = engine.run("테스트", agent_results)
    high_confidence = consensus.overall_confidence

    # 일부 에이전트 실패
    agent_results = [MockResult() if i > 0 else MockResult("error") for i in range(4)]
    consensus = engine.run("테스트", agent_results)
    low_confidence = consensus.overall_confidence

    assert high_confidence > low_confidence

    print("✅ Confidence calculation OK")


if __name__ == "__main__":
    print("\n🧪 Running Multi-Agent tests...\n")

    test_general_agent()
    test_tech_agent()
    test_news_agent()
    test_review_agent()
    test_consensus_merge()
    test_consensus_domain_overlap()
    test_consensus_conflict_detection()
    test_consensus_confidence()

    print("\n✅ All Multi-Agent tests passed!\n")
