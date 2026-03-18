"""
Claude 합산기: 멀티소스 콘텐츠 → 구조화된 Sparkpage
역할: FetchedContent 리스트 → SynthesisResult (섹션 + 핵심 사실)
"""

import json
import re
from dataclasses import dataclass, asdict
from typing import List
import requests

from .content_fetcher import FetchedContent
from .query_analyzer import QuerySpec


@dataclass
class SparkSection:
    """Sparkpage 섹션"""
    title: str
    content: str  # 마크다운
    sources: List[str]  # URL 목록
    section_type: str  # 'overview' | 'detail' | 'example' | 'summary'


@dataclass
class SynthesisResult:
    """합산 결과"""
    query: str
    sections: List[SparkSection]
    key_facts: List[str]
    confidence_score: float
    total_sources: int
    synthesis_model: str


class ClaudeSynthesizer:
    """Claude Sonnet를 이용한 멀티소스 합산"""

    def __init__(self, api_key: str, model: str = "claude-sonnet-4-6"):
        self.api_key = api_key
        self.model = model
        self.max_context = 60000  # Termux 메모리 제약
        self.api_url = "https://api.anthropic.com/v1/messages"

    def synthesize(
        self, query_spec: QuerySpec, contents: List[FetchedContent]
    ) -> SynthesisResult:
        """FetchedContent → SynthesisResult 생성"""
        # 유효한 콘텐츠만 필터링 (timeout/blocked 제외)
        valid_contents = [
            c for c in contents if c.fetch_status == "ok" and c.word_count > 50
        ]

        if not valid_contents:
            return self._empty_result(query_spec)

        # 컨텍스트 구성
        context = self._build_context(valid_contents)
        system_prompt = self._build_system_prompt(query_spec)

        # Claude 호출
        response_text = self._call_claude(system_prompt, query_spec, context)
        return self._parse_synthesis(response_text, valid_contents, query_spec)

    def _call_claude(self, system_prompt: str, query_spec: QuerySpec, context: str) -> str:
        """Claude API 직접 호출"""
        headers = {
            "x-api-key": self.api_key,
            "anthropic-version": "2023-06-01",
            "content-type": "application/json",
        }
        payload = {
            "model": self.model,
            "max_tokens": 2000,
            "system": system_prompt,
            "messages": [
                {
                    "role": "user",
                    "content": f"""원본 질문: {query_spec.original_query}

{context}

이 정보를 바탕으로 다음 JSON을 생성하세요 (주석 없이):
{{
  "key_facts": ["사실1", "사실2", "사실3"],
  "sections": [
    {{"title": "개요", "content": "본문...", "sources": ["url1"], "section_type": "overview"}},
    {{"title": "상세", "content": "본문...", "sources": ["url2"], "section_type": "detail"}},
    {{"title": "결론", "content": "본문...", "sources": ["url3"], "section_type": "summary"}}
  ],
  "confidence_score": 0.85
}}""",
                }
            ],
        }
        try:
            response = requests.post(self.api_url, json=payload, headers=headers, timeout=60)
            response.raise_for_status()
            data = response.json()
            return data["content"][0]["text"]
        except Exception as e:
            print(f"❌ Claude API 호출 실패: {e}")
            return "{}"

    def _build_context(self, contents: List[FetchedContent]) -> str:
        """FetchedContent 리스트 → 컨텍스트 문자열"""
        context_parts = []

        for i, content in enumerate(contents[:10], 1):  # 최대 10개
            part = f"""[출처 {i}] {content.title}
URL: {content.url}
{content.body_text[:500]}
---"""
            context_parts.append(part)

            # 컨텍스트 크기 제한
            if sum(len(p) for p in context_parts) > self.max_context:
                break

        return "\n".join(context_parts)

    def _build_system_prompt(self, query_spec: QuerySpec) -> str:
        """시스템 프롬프트 생성"""
        sections_hint = ", ".join(query_spec.expected_sections[:3])
        return f"""당신은 전문 웹 리서처입니다.

주어진 웹 검색 결과들을 분석하여:
1. 핵심 사실 3~5개 추출
2. 다음 섹션 생성: {sections_hint}
3. 각 섹션의 출처 명시

요구사항:
- 마크다운 형식의 명확한 설명
- 각 섹션 200~400자
- 신뢰도 점수 (0.6~0.95)
- JSON만 반환"""

    def _parse_synthesis(
        self,
        response_text: str,
        contents: List[FetchedContent],
        query_spec: QuerySpec,
    ) -> SynthesisResult:
        """Claude 응답 파싱"""
        # JSON 추출
        json_match = re.search(r"\{.*\}", response_text, re.DOTALL)
        if not json_match:
            return self._empty_result(query_spec)

        try:
            data = json.loads(json_match.group())

            sections = [
                SparkSection(
                    title=s.get("title", ""),
                    content=s.get("content", ""),
                    sources=s.get("sources", []),
                    section_type=s.get("section_type", "detail"),
                )
                for s in data.get("sections", [])
            ]

            return SynthesisResult(
                query=query_spec.original_query,
                sections=sections,
                key_facts=data.get("key_facts", []),
                confidence_score=float(data.get("confidence_score", 0.7)),
                total_sources=len(contents),
                synthesis_model=self.model,
            )
        except Exception:
            return self._empty_result(query_spec)

    def _empty_result(self, query_spec: QuerySpec) -> SynthesisResult:
        """콘텐츠 부족 시 기본 결과"""
        return SynthesisResult(
            query=query_spec.original_query,
            sections=[
                SparkSection(
                    title="검색 결과 부족",
                    content=f"'{query_spec.original_query}'에 대한 충분한 정보를 수집하지 못했습니다.",
                    sources=[],
                    section_type="error",
                )
            ],
            key_facts=[],
            confidence_score=0.0,
            total_sources=0,
            synthesis_model=self.model,
        )
