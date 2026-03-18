#!/usr/bin/env python3
"""
기본 검증 테스트 (API 키 없이)
"""

from src.query_analyzer import QueryAnalyzer, QuerySpec
from src.web_searcher import DuckDuckGoSearcher
from src.content_fetcher import ContentFetcher, FetchedContent
from src.sparkpage_generator import SparkpageGenerator
from src.claude_synthesizer import SparkSection, SynthesisResult


def test_query_analyzer_fallback():
    """QueryAnalyzer 폴백 테스트 (API 호출 없이)"""
    analyzer = QueryAnalyzer("dummy_key")
    # _fallback_spec 직접 테스트
    spec = analyzer._fallback_spec("파이썬이란")
    assert spec.original_query == "파이썬이란"
    assert spec.main_topic == "파이썬이란"
    assert "파이썬이란" in spec.sub_queries
    print("✅ QueryAnalyzer fallback OK")


def test_web_searcher():
    """웹 검색 테스트"""
    searcher = DuckDuckGoSearcher(max_results=3)
    # 실제 검색 (최소 테스트)
    results = searcher.search("python")
    assert isinstance(results, list)
    if results:
        assert hasattr(results[0], "url")
        assert hasattr(results[0], "title")
        print(f"✅ DuckDuckGo 검색 OK ({len(results)}개 결과)")
    else:
        print("⚠️ DuckDuckGo 검색 반환 없음 (네트워크 확인)")


def test_content_fetcher():
    """콘텐츠 페처 테스트"""
    fetcher = ContentFetcher()
    # 페치 테스트
    content = fetcher.fetch("https://example.com")
    assert isinstance(content, FetchedContent)
    assert content.url == "https://example.com"
    print(f"✅ ContentFetcher OK (상태: {content.fetch_status})")


def test_sparkpage_generator():
    """Sparkpage 생성기 테스트"""
    gen = SparkpageGenerator(output_dir="test_output")

    # 테스트 결과
    result = SynthesisResult(
        query="테스트 쿼리",
        sections=[
            SparkSection(
                title="개요",
                content="이것은 테스트입니다.",
                sources=["https://example.com"],
                section_type="overview"
            )
        ],
        key_facts=["사실 1"],
        confidence_score=0.9,
        total_sources=1,
        synthesis_model="test"
    )

    output = gen.generate(result, "테스트")
    assert output.markdown_path.endswith(".md")
    assert output.html_path.endswith(".html")
    print(f"✅ SparkpageGenerator OK")
    print(f"   - MD: {output.markdown_path}")
    print(f"   - HTML: {output.html_path}")


if __name__ == "__main__":
    print("🧪 기본 검증 테스트 시작\n")

    test_query_analyzer_fallback()
    test_web_searcher()
    test_content_fetcher()
    test_sparkpage_generator()

    print("\n✅ 모든 기본 테스트 완료")
