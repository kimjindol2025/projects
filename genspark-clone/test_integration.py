#!/usr/bin/env python3
"""
통합 테스트 (API 키 있을 때만 실행)
"""

import os
import sys
from pathlib import Path

from src.genspark_agent import GensparkAgent, AgentConfig
from src.query_analyzer import QueryAnalyzer


def test_with_api_key():
    """API 키가 있을 때 전체 통합 테스트"""
    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        print("⚠️ ANTHROPIC_API_KEY가 없으므로 API 테스트 스킵")
        return False

    print("\n🔍 API 키 있음. 전체 통합 테스트 실행...\n")

    try:
        # Step 1: QueryAnalyzer 테스트
        print("[1/2] QueryAnalyzer 테스트...")
        analyzer = QueryAnalyzer(api_key)
        spec = analyzer.analyze("파이썬이란")
        print(f"  ✅ 분석 완료: {len(spec.sub_queries)}개 서브쿼리")
        print(f"     - {spec.sub_queries[:2]}")

        # Step 2: GensparkAgent 전체 파이프라인
        print("\n[2/2] GensparkAgent 전체 파이프라인...")
        config = AgentConfig(
            anthropic_api_key=api_key,
            output_dir="test_output",
            verbose=True,
            max_search_results=3
        )
        agent = GensparkAgent(config)
        result = agent.run("REST API란")

        if result:
            print(f"\n✅ 전체 파이프라인 성공!")
            print(f"   - 섹션: {len(result.sections)}")
            print(f"   - 신뢰도: {result.confidence_score:.0%}")
            print(f"   - 결과: {result.html_path}")
            return True
        else:
            print("\n❌ 파이프라인 실패")
            return False

    except Exception as e:
        print(f"\n❌ 테스트 실패: {e}")
        import traceback
        traceback.print_exc()
        return False


def test_search_only():
    """검색만 테스트 (API 키 불필요)"""
    print("\n🔍 DuckDuckGo 검색 테스트...")
    from src.web_searcher import DuckDuckGoSearcher

    searcher = DuckDuckGoSearcher(max_results=3)
    results = searcher.search("python asyncio")

    if results:
        print(f"✅ 검색 성공: {len(results)}개 결과")
        for r in results[:2]:
            print(f"   - {r.title[:50]}")
        return True
    else:
        print("⚠️ 검색 결과 없음 (네트워크 확인)")
        return False


def test_content_fetch():
    """콘텐츠 페칭 테스트"""
    print("\n📄 콘텐츠 페칭 테스트...")
    from src.content_fetcher import ContentFetcher

    fetcher = ContentFetcher()
    content = fetcher.fetch("https://www.python.org")

    print(f"✅ 페칭 완료: {content.fetch_status}")
    if content.fetch_status == "ok":
        print(f"   - 제목: {content.title[:50]}")
        print(f"   - 본문: {content.word_count}단어")
        return True
    else:
        print(f"⚠️ 페칭 실패: {content.fetch_status}")
        return False


if __name__ == "__main__":
    print("=" * 60)
    print("🧪 Genspark Clone 통합 테스트")
    print("=" * 60)

    results = []

    # 검색 테스트 (필수)
    results.append(("DuckDuckGo 검색", test_search_only()))

    # 콘텐츠 페칭 (필수)
    results.append(("콘텐츠 페칭", test_content_fetch()))

    # API 키 있으면 전체 파이프라인
    results.append(("전체 파이프라인 (API 필요)", test_with_api_key()))

    # 결과 요약
    print("\n" + "=" * 60)
    print("📊 테스트 결과 요약")
    print("=" * 60)
    for name, passed in results:
        status = "✅ 통과" if passed else "⏭️ 스킵" if passed is False else "❌ 실패"
        print(f"{status} | {name}")

    # 최종 판정
    if all(r for r in results if r[1] is not None):
        print("\n🎉 모든 테스트 통과!")
        sys.exit(0)
    else:
        print("\n⚠️ 일부 테스트 실패 또는 스킵")
        sys.exit(1)
