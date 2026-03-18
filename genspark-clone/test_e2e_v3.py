"""
Day 4: Genspark Clone v3.0 E2E 통합 테스트
버그 수정 + Phase 5 (RoutingAgent) + Phase 6 (CrossCheckAgent) 통합 검증
"""

from src.genspark_agent import GensparkAgent, AgentConfig
from src.query_analyzer import QuerySpec
from src.content_fetcher import FetchedContent
from src.web_searcher import SearchResult
from unittest.mock import Mock, patch


class MockSearcher:
    """Mock DuckDuckGoSearcher"""
    def search(self, query: str):
        """검색 (max_results 파라미터 없음 - Bug 1 해결)"""
        return [
            SearchResult(
                url=f"https://example{i}.com",
                title=f"Result {i}: {query}",
                snippet=f"Snippet for {query}",
                source_domain=f"example{i}.com",
                rank=i + 1,
            )
            for i in range(3)
        ]

    def search_multiple(self, queries: list, delay: float = 1.0):
        """멀티 쿼리 검색"""
        return {
            f"q{i}": self.search(query)
            for i, query in enumerate(queries)
        }


class MockFetcher:
    """Mock ContentFetcher"""
    MAX_WORKERS = 3

    def fetch_urls(self, urls):
        """fetch_urls 메서드 구현됨 - Bug 2 해결"""
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

    def fetch(self, url):
        """단일 URL 페칭"""
        return FetchedContent(
            url=url,
            title=f"Title for {url}",
            body_text=f"Body content for {url}" * 5,
            word_count=50,
            fetch_status="ok",
        )

    def fetch_for_queries(self, search_results):
        """멀티 쿼리별 URL 페칭"""
        all_contents = []
        for query_id, results in search_results.items():
            urls = [r.url for r in results][:5]
            contents = self.fetch_urls(urls)
            all_contents.extend(contents)
        return all_contents


def test_v3_integration_without_routing_crosscheck():
    """v3.0 통합: RoutingAgent/CrossCheckAgent 미사용 (v2.0 호환)"""
    config = AgentConfig(
        anthropic_api_key="test-key",
        output_dir="test_output_e2e",
        use_cache=False,
        use_multi_agent=False,
        # v3.0 옵션
        use_routing=False,
        use_crosscheck=False,
    )

    agent = GensparkAgent(config)

    # Mock 검색/페칭
    agent.searcher = MockSearcher()
    agent.fetcher = MockFetcher()

    # Mock 질문 분석
    with patch.object(agent.analyzer, 'analyze') as mock_analyze:
        from src.query_analyzer import QuerySpec
        mock_analyze.return_value = QuerySpec(
            original_query="test query",
            main_topic="test",
            sub_queries=["test"],
            language="en",
            expected_sections=["Overview"],
            complexity=0.5,
        )

        # Mock 합산
        with patch.object(agent.synthesizer, 'synthesize') as mock_synthesize:
            from src.claude_synthesizer import SynthesisResult, SparkSection
            mock_synthesize.return_value = SynthesisResult(
                query="test",
                key_facts=["fact 1"],
                sections=[
                    SparkSection(
                        title="Overview",
                        content="Test content",
                        sources=["https://example0.com"],
                        section_type="detail"
                    ),
                ],
                total_sources=3,
                confidence_score=0.85,
                synthesis_model="claude-sonnet-4-6",
            )

            result = agent.run("test query")

            assert result is not None
            assert result.confidence_score == 0.85
            assert result.html_path is not None

    print("✅ v3.0 Integration (without Routing/CrossCheck) OK")


def test_v3_integration_with_routing():
    """v3.0 통합: RoutingAgent 활성화"""
    config = AgentConfig(
        anthropic_api_key="test-key",
        output_dir="test_output_e2e",
        use_cache=False,
        use_multi_agent=False,
        # v3.0 옵션
        use_routing=True,
        use_crosscheck=False,
        max_parallel_workers=2,
    )

    agent = GensparkAgent(config)

    # Mock 질문 분석
    with patch.object(agent.analyzer, 'analyze') as mock_analyze:
        from src.query_analyzer import QuerySpec
        mock_analyze.return_value = QuerySpec(
            original_query="test query",
            main_topic="test",
            sub_queries=["test q1", "test q2"],
            language="en",
            expected_sections=["Overview"],
            complexity=0.5,
        )

        # Mock 의존성 분석
        with patch.object(agent.routing_agent.analyzer, 'analyze') as mock_route_analyze:
            mock_route_analyze.return_value = {
                "q0": [],
                "q1": [],
            }

            # Mock 합산
            with patch.object(agent.synthesizer, 'synthesize') as mock_synthesize:
                from src.claude_synthesizer import SynthesisResult, SparkSection
                mock_synthesize.return_value = SynthesisResult(
                    query="test",
                    key_facts=["fact 1"],
                    sections=[
                        SparkSection(
                            title="Overview",
                            content="Test content",
                            sources=["https://example0.com"],
                            section_type="detail"
                        ),
                    ],
                    total_sources=2,
                    confidence_score=0.85,
                    synthesis_model="claude-sonnet-4-6",
                )

                result = agent.run("test query")

                assert result is not None
                assert result.confidence_score == 0.85

    print("✅ v3.0 Integration (with RoutingAgent) OK")


def test_v3_integration_with_crosscheck():
    """v3.0 통합: CrossCheckAgent 활성화"""
    config = AgentConfig(
        anthropic_api_key="test-key",
        output_dir="test_output_e2e",
        use_cache=False,
        use_multi_agent=False,
        # v3.0 옵션
        use_routing=False,
        use_crosscheck=True,
        openai_api_key="test-openai-key",
    )

    agent = GensparkAgent(config)
    agent.searcher = MockSearcher()
    agent.fetcher = MockFetcher()

    # Mock 질문 분석
    with patch.object(agent.analyzer, 'analyze') as mock_analyze:
        from src.query_analyzer import QuerySpec
        mock_analyze.return_value = QuerySpec(
            original_query="test query",
            main_topic="test",
            sub_queries=["test"],
            language="en",
            expected_sections=["Overview"],
            complexity=0.5,
        )

        # Mock CrossCheckAgent
        with patch.object(agent.crosscheck_agent, 'run') as mock_crosscheck:
            from src.crosscheck_agent import ConflictReport
            mock_crosscheck.return_value = ConflictReport(
                conflicting_pairs=[],
                avg_similarity=0.85,
                outlier_urls=[],
                adjusted_confidence=0.85,
                warnings=[],
            )

            # Mock 합산
            with patch.object(agent.synthesizer, 'synthesize') as mock_synthesize:
                from src.claude_synthesizer import SynthesisResult, SparkSection
                mock_synthesize.return_value = SynthesisResult(
                    query="test",
                    key_facts=["fact 1"],
                    sections=[
                        SparkSection(
                            title="Overview",
                            content="Test content",
                            sources=["https://example0.com"],
                            section_type="detail"
                        ),
                    ],
                    total_sources=3,
                    confidence_score=0.85,
                    synthesis_model="claude-sonnet-4-6",
                )

                result = agent.run("test query")

                assert result is not None
                # CrossCheck 신뢰도 통합: (0.85 * 0.5) + (0.85 * 0.5) = 0.85
                assert result.confidence_score == 0.85

    print("✅ v3.0 Integration (with CrossCheckAgent) OK")


def test_v3_all_components():
    """v3.0 통합: 모든 컴포넌트 활성화"""
    config = AgentConfig(
        anthropic_api_key="test-key",
        output_dir="test_output_e2e",
        use_cache=False,
        use_multi_agent=False,
        # v3.0 옵션
        use_routing=True,
        use_crosscheck=True,
        openai_api_key="test-openai-key",
        max_parallel_workers=2,
    )

    agent = GensparkAgent(config)

    # Mock 질문 분석
    with patch.object(agent.analyzer, 'analyze') as mock_analyze:
        from src.query_analyzer import QuerySpec
        mock_analyze.return_value = QuerySpec(
            original_query="test query",
            main_topic="test",
            sub_queries=["test q1", "test q2"],
            language="en",
            expected_sections=["Overview"],
            complexity=0.5,
        )

        # Mock 의존성 분석
        with patch.object(agent.routing_agent.analyzer, 'analyze') as mock_route_analyze:
            mock_route_analyze.return_value = {
                "q0": [],
                "q1": [],
            }

            # Mock CrossCheckAgent
            with patch.object(agent.crosscheck_agent, 'run') as mock_crosscheck:
                from src.crosscheck_agent import ConflictReport
                mock_crosscheck.return_value = ConflictReport(
                    conflicting_pairs=[],
                    avg_similarity=0.82,
                    outlier_urls=[],
                    adjusted_confidence=0.82,
                    warnings=[],
                )

                # Mock 합산
                with patch.object(agent.synthesizer, 'synthesize') as mock_synthesize:
                    from src.claude_synthesizer import SynthesisResult, SparkSection
                    mock_synthesize.return_value = SynthesisResult(
                        query="test",
                        key_facts=["fact 1"],
                        sections=[
                            SparkSection(
                                title="Overview",
                                content="Test content",
                                sources=["https://example0.com"],
                                section_type="detail"
                            ),
                        ],
                        total_sources=2,
                        confidence_score=0.85,
                        synthesis_model="claude-sonnet-4-6",
                    )

                    result = agent.run("test query")

                    assert result is not None
                    # CrossCheck 신뢰도 통합: (0.85 * 0.5) + (0.82 * 0.5) = 0.835
                    assert abs(result.confidence_score - 0.835) < 0.01

    print("✅ v3.0 Integration (All Components) OK")


if __name__ == "__main__":
    print("\n🧪 Running Genspark Clone v3.0 E2E Integration Tests...\n")

    test_v3_integration_without_routing_crosscheck()
    test_v3_integration_with_routing()
    test_v3_integration_with_crosscheck()
    test_v3_all_components()

    print("\n✅ All E2E Integration tests passed!\n")
