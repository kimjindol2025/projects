"""
콘텐츠 페처: URL → 텍스트 본문 추출 (병렬)
역할: SearchResult → FetchedContent (타임아웃, 에러 핸들링)
"""

from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass
from typing import Dict, List

import requests
from bs4 import BeautifulSoup


@dataclass
class FetchedContent:
    """크롤된 콘텐츠"""
    url: str
    title: str
    body_text: str  # 최대 3,000자
    word_count: int
    fetch_status: str  # 'ok' | 'timeout' | 'error' | 'blocked'


class ContentFetcher:
    """병렬 콘텐츠 페칭"""

    MAX_BODY_CHARS = 3000
    TIMEOUT = 8
    MAX_WORKERS = 3  # Termux 메모리 제약

    HEADERS = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    }

    def fetch(self, url: str, title: str = "") -> FetchedContent:
        """단일 URL 페칭"""
        try:
            response = requests.get(url, headers=self.HEADERS, timeout=self.TIMEOUT)
            response.raise_for_status()

            soup = BeautifulSoup(response.text, "html.parser")
            page_title = soup.title.string if soup.title else title

            body_text = self._extract_body(soup)
            word_count = len(body_text.split())

            return FetchedContent(
                url=url,
                title=page_title or url,
                body_text=body_text[: self.MAX_BODY_CHARS],
                word_count=word_count,
                fetch_status="ok",
            )
        except requests.Timeout:
            return FetchedContent(
                url=url, title=title or url, body_text="", word_count=0, fetch_status="timeout"
            )
        except requests.RequestException as e:
            if "403" in str(e) or "429" in str(e):
                return FetchedContent(
                    url=url, title=title or url, body_text="", word_count=0, fetch_status="blocked"
                )
            return FetchedContent(
                url=url, title=title or url, body_text="", word_count=0, fetch_status="error"
            )
        except Exception:
            return FetchedContent(
                url=url, title=title or url, body_text="", word_count=0, fetch_status="error"
            )

    def fetch_all(self, results: List) -> List[FetchedContent]:
        """병렬 페칭 (SearchResult 리스트)"""
        contents = []

        with ThreadPoolExecutor(max_workers=self.MAX_WORKERS) as executor:
            futures = {
                executor.submit(self.fetch, r.url, r.title): r
                for r in results
            }

            for future in as_completed(futures):
                try:
                    content = future.result()
                    contents.append(content)
                except Exception:
                    pass

        return contents

    def fetch_urls(self, urls: List[str]) -> List[FetchedContent]:
        """URL 문자열 리스트 페칭 (researcher_agent 용)"""
        contents = []

        with ThreadPoolExecutor(max_workers=self.MAX_WORKERS) as executor:
            futures = {executor.submit(self.fetch, url): url for url in urls}

            for future in as_completed(futures):
                try:
                    content = future.result()
                    contents.append(content)
                except Exception:
                    pass

        return contents

    def fetch_for_queries(
        self, search_results: Dict[str, List]
    ) -> List[FetchedContent]:
        """여러 검색 결과로부터 콘텐츠 페칭"""
        all_results = []
        for query, results in search_results.items():
            all_results.extend(results)

        return self.fetch_all(all_results)

    def _extract_body(self, soup: BeautifulSoup) -> str:
        """BeautifulSoup → 본문 텍스트 추출"""
        # 스크립트, 스타일 제거
        for tag in soup.find_all(["script", "style", "nav", "footer"]):
            tag.decompose()

        # main, article, .content 우선
        main_elem = (
            soup.find("main")
            or soup.find("article")
            or soup.find("div", class_="content")
            or soup.find("div", class_="post-content")
            or soup.find("div", class_="entry-content")
            or soup.body
        )

        if main_elem:
            text = main_elem.get_text(separator=" ", strip=True)
        else:
            text = soup.get_text(separator=" ", strip=True)

        # 공백 정규화
        text = " ".join(text.split())
        return text
