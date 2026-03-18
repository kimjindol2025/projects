"""
Day 1: 버그 수정 검증 테스트
4개 버그를 수정하고 정상 작동을 확인합니다.
"""

import os
import json
from pathlib import Path
from unittest.mock import Mock, patch

from src.web_searcher import DuckDuckGoSearcher, SearchResult
from src.content_fetcher import ContentFetcher, FetchedContent
from src.agents.researcher_agent import GeneralAgent, TechAgent, NewsAgent, ReviewAgent
from src.sparkpage_generator import SparkpageGenerator, SparkpageOutput
from src.genspark_agent import GensparkAgent, AgentConfig
from src.query_analyzer import QuerySpec
from src.claude_synthesizer import SynthesisResult, SparkSection


class MockSearcher:
    """Mock DuckDuckGoSearcher"""
    def search(self, query: str):
        """Bug 1 수정: max_results 파라미터 없음"""
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
    """Mock ContentFetcher"""
    def fetch_urls(self, urls):
        """Bug 2 수정: fetch_urls() 메서드 구현됨"""
        return [
            FetchedContent(
                url=url,
                title=f"Title for {url}",
                body_text=f"Body content for {url}" * 10,
                word_count=50,
                fetch_status="ok",
            )
            for url in urls
        ]


def test_searcher_no_max_results_param():
    """Bug 1: DuckDuckGoSearcher.search() 파라미터 확인"""
    searcher = DuckDuckGoSearcher(max_results=5)

    # search() 메서드 시그니처 확인
    import inspect
    sig = inspect.signature(searcher.search)
    params = list(sig.parameters.keys())

    # query만 받아야 함 (max_results 없음)
    assert "query" in params
    assert "max_results" not in params, "search() should not accept max_results parameter"

    print("✅ DuckDuckGoSearcher.search() signature OK")


def test_content_fetcher_fetch_urls():
    """Bug 2: ContentFetcher.fetch_urls() 메서드 존재 확인"""
    fetcher = ContentFetcher()

    # fetch_urls() 메서드 존재 확인
    assert hasattr(fetcher, "fetch_urls"), "ContentFetcher should have fetch_urls() method"
    assert callable(getattr(fetcher, "fetch_urls")), "fetch_urls should be callable"

    # Mock 테스트: URL 리스트 입력 → FetchedContent 리스트 반환
    mock_fetcher = MockFetcher()
    urls = ["https://example1.com", "https://example2.com"]
    results = mock_fetcher.fetch_urls(urls)

    assert len(results) == 2
    assert all(isinstance(r, FetchedContent) for r in results)
    assert results[0].url == urls[0]

    print("✅ ContentFetcher.fetch_urls() method OK")


def test_widget_renderer_used_in_html_generation():
    """Bug 3: WidgetRenderer가 HTML 생성에 사용됨 확인"""
    generator = SparkpageGenerator()

    # Mock SynthesisResult
    mock_result = SynthesisResult(
        query="test",
        key_facts=["fact 1", "fact 2"],
        sections=[
            SparkSection(
                title="Section 1",
                content="| col1 | col2 |\n|------|------|\n| a    | b    |",
                sources=["https://example.com"],
                section_type="detail"
            ),
            SparkSection(
                title="Section 2",
                content="- item 1\n- item 2\n- item 3",
                sources=["https://example2.com"],
                section_type="detail"
            ),
        ],
        total_sources=2,
        confidence_score=0.85,
        synthesis_model="claude-sonnet-4-6",
    )

    # _generate_html() 호출
    with patch.object(generator.widget_renderer, 'render') as mock_render:
        # render()는 각 섹션마다 호출되어야 함
        mock_render.return_value = "<div class='widget'>rendered</div>"

        html = generator._generate_html(mock_result, "test query", "")

        # 섹션 개수만큼 widget_renderer.render() 호출 확인
        assert mock_render.call_count == 2, f"Expected 2 render() calls, got {mock_render.call_count}"

        # render() 호출 시 content와 title 전달 확인
        first_call_args = mock_render.call_args_list[0]
        assert first_call_args[0][0] == "| col1 | col2 |\n|------|------|\n| a    | b    |"
        assert first_call_args[0][1] == "Section 1"

    print("✅ WidgetRenderer.render() used in HTML generation OK")


def test_cache_content_persistence():
    """Bug 4: 캐시에 markdown_content, html_content 보존"""
    cache_dir = Path("test_cache_tmp")
    cache_dir.mkdir(exist_ok=True)

    try:
        config = AgentConfig(
            anthropic_api_key="test-key",
            output_dir=str(cache_dir),
            use_cache=True,
        )
        agent = GensparkAgent(config)

        # SparkpageOutput 생성
        output = SparkpageOutput(
            markdown_path="/test/test.md",
            html_path="/test/test.html",
            markdown_content="# Test\n\nThis is markdown content",
            html_content="<h1>Test</h1><p>This is HTML content</p>",
            title="Test Query",
            generated_at="2026-03-18_120000",
            query="test query",
            confidence_score=0.85,
        )

        # _output_to_dict() 호출
        dict_output = agent._output_to_dict(output)

        # markdown_content, html_content 포함 확인
        assert "markdown_content" in dict_output, "markdown_content missing in dict"
        assert "html_content" in dict_output, "html_content missing in dict"
        assert dict_output["markdown_content"] == "# Test\n\nThis is markdown content"
        assert dict_output["html_content"] == "<h1>Test</h1><p>This is HTML content</p>"
        assert dict_output["title"] == "Test Query"

        # _dict_to_output() 복원 확인
        restored = agent._dict_to_output(dict_output)
        assert restored is not None
        assert restored.markdown_content == output.markdown_content
        assert restored.html_content == output.html_content
        assert restored.title == output.title

        print("✅ Cache content persistence OK")

    finally:
        # 정리
        import shutil
        if cache_dir.exists():
            shutil.rmtree(cache_dir)


def test_researcher_agent_no_max_results():
    """Bug 1 & 2 통합: ResearcherAgent 버그 없이 작동"""
    searcher = MockSearcher()
    fetcher = MockFetcher()

    agent = GeneralAgent(searcher, fetcher)

    # Mock QuerySpec
    query_spec = QuerySpec(
        original_query="Python async",
        main_topic="async",
        sub_queries=["Python asyncio", "async/await"],
        language="ko",
        expected_sections=["개요", "예제"],
        complexity=0.7,
    )

    # research() 실행 (Bug 1, Bug 2 해결 확인)
    result = agent.research(query_spec)

    assert result is not None
    assert result.agent_name == "general"
    assert len(result.search_results) > 0
    assert len(result.fetched_contents) > 0
    assert result.error is None

    print("✅ ResearcherAgent research() works without bugs")


if __name__ == "__main__":
    print("\n🧪 Running Bug Fix Tests...\n")

    test_searcher_no_max_results_param()
    test_content_fetcher_fetch_urls()
    test_widget_renderer_used_in_html_generation()
    test_cache_content_persistence()
    test_researcher_agent_no_max_results()

    print("\n✅ All bug fix tests passed!\n")
