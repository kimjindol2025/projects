"""
웹 검색: DuckDuckGo 해석 (API 키 불필요)
역할: 검색어 → 검색 결과 리스트
"""

import time
from dataclasses import dataclass
from typing import Dict, List
from urllib.parse import quote

import requests
from bs4 import BeautifulSoup


@dataclass
class SearchResult:
    """검색 결과"""
    url: str
    title: str
    snippet: str
    source_domain: str
    rank: int


class DuckDuckGoSearcher:
    """DuckDuckGo HTML 파싱 검색기"""

    BASE_URL = "https://html.duckduckgo.com/html/"
    HEADERS = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    }

    def __init__(self, max_results: int = 5, timeout: int = 10):
        self.max_results = max_results
        self.timeout = timeout

    def search(self, query: str) -> List[SearchResult]:
        """단일 검색어로 검색 실행"""
        try:
            url = f"{self.BASE_URL}?q={quote(query)}&ia=web"
            response = requests.get(
                url, headers=self.HEADERS, timeout=self.timeout
            )
            response.raise_for_status()

            results = self._parse_results(response.text)
            return results[: self.max_results]
        except Exception as e:
            print(f"⚠️  검색 실패 '{query}': {e}")
            return []

    def search_multiple(
        self, queries: List[str], delay: float = 1.0
    ) -> Dict[str, List[SearchResult]]:
        """여러 검색어 검색 (딜레이 포함)"""
        results = {}
        for i, query in enumerate(queries):
            if i > 0:
                time.sleep(delay)
            results[query] = self.search(query)
        return results

    def _parse_results(self, html: str) -> List[SearchResult]:
        """DuckDuckGo HTML → SearchResult 리스트"""
        soup = BeautifulSoup(html, "html.parser")
        results = []
        rank = 1

        # DDG 결과 컨테이너 찾기
        for result_elem in soup.select("div.result"):
            try:
                # 제목 + URL
                link_elem = result_elem.select_one("a.result__url")
                if not link_elem:
                    continue

                url = link_elem.get("href", "")
                if not self._is_valid_url(url):
                    continue

                title = link_elem.get_text(strip=True)
                if not title:
                    title = url

                # 스니펫
                snippet_elem = result_elem.select_one("a.result__snippet")
                snippet = snippet_elem.get_text(strip=True) if snippet_elem else ""

                # 도메인 추출
                source_domain = self._extract_domain(url)

                results.append(
                    SearchResult(
                        url=url,
                        title=title,
                        snippet=snippet,
                        source_domain=source_domain,
                        rank=rank,
                    )
                )
                rank += 1
            except Exception:
                continue

        return results

    def _is_valid_url(self, url: str) -> bool:
        """URL 유효성 검사"""
        return url.startswith("http://") or url.startswith("https://")

    def _extract_domain(self, url: str) -> str:
        """URL에서 도메인 추출"""
        try:
            # https://example.com/path → example.com
            domain = url.replace("https://", "").replace("http://", "").split("/")[0]
            return domain
        except Exception:
            return "unknown"
