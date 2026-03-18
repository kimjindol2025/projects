"""
질문 분석기: 사용자 질문 → 구조화된 쿼리 스펙
역할: Claude haiku를 이용해 질문을 분해하고 섹션 계획 수립
"""

import json
import re
from dataclasses import dataclass
from typing import List
import requests


@dataclass
class QuerySpec:
    """분석된 질문 사양"""
    original_query: str
    main_topic: str
    sub_queries: List[str]  # 2~5개 검색쿼리
    language: str  # 'ko' | 'en' | 'mixed'
    expected_sections: List[str]  # ['개요', '원인', '해결책', ...]
    complexity: float  # 0.0~1.0


class QueryAnalyzer:
    """Claude haiku를 이용한 질문 분석"""

    def __init__(self, api_key: str, model: str = "claude-haiku-4-5-20251001"):
        self.api_key = api_key
        self.model = model
        self.api_url = "https://api.anthropic.com/v1/messages"

    def analyze(self, user_query: str) -> QuerySpec:
        """사용자 질문을 분석하여 QuerySpec 반환"""
        prompt = self._build_prompt(user_query)
        response_text = self._call_claude(prompt)
        return self._parse_claude_response(response_text, user_query)

    def _call_claude(self, prompt: str) -> str:
        """Claude API 직접 호출"""
        headers = {
            "x-api-key": self.api_key,
            "anthropic-version": "2023-06-01",
            "content-type": "application/json",
        }
        payload = {
            "model": self.model,
            "max_tokens": 500,
            "messages": [{"role": "user", "content": prompt}],
        }
        try:
            response = requests.post(self.api_url, json=payload, headers=headers, timeout=30)
            response.raise_for_status()
            data = response.json()
            return data["content"][0]["text"]
        except Exception as e:
            print(f"❌ Claude API 호출 실패: {e}")
            return "{}"

    def _build_prompt(self, query: str) -> str:
        """Claude에 전달할 프롬프트 생성"""
        return f"""사용자가 다음 질문을 했습니다:
"{query}"

이 질문을 분석하여 다음 JSON을 생성하세요 (주석 없이 JSON만):
{{
  "main_topic": "질문의 중심 주제 (예: 파이썬 비동기)",
  "sub_queries": ["검색1", "검색2", "검색3"],
  "language": "ko 또는 en",
  "expected_sections": ["섹션1", "섹션2", "섹션3"],
  "complexity": 0.5
}}

주의:
- sub_queries는 2~5개, 최대 10단어
- expected_sections는 3~5개, 한 줄 제목
- complexity: 간단(0.2), 보통(0.5), 복잡(0.8)
- JSON만 반환"""

    def _parse_claude_response(self, response_text: str, original_query: str) -> QuerySpec:
        """Claude 응답에서 JSON 추출 및 파싱"""
        # JSON 블록 추출
        json_match = re.search(r"\{.*\}", response_text, re.DOTALL)
        if not json_match:
            return self._fallback_spec(original_query)

        try:
            data = json.loads(json_match.group())
            return QuerySpec(
                original_query=original_query,
                main_topic=data.get("main_topic", "검색"),
                sub_queries=data.get("sub_queries", [original_query]),
                language=data.get("language", "ko"),
                expected_sections=data.get("expected_sections", ["개요", "상세", "결론"]),
                complexity=float(data.get("complexity", 0.5)),
            )
        except (json.JSONDecodeError, KeyError, ValueError):
            return self._fallback_spec(original_query)

    def _fallback_spec(self, original_query: str) -> QuerySpec:
        """파싱 실패 시 기본값으로 반환"""
        return QuerySpec(
            original_query=original_query,
            main_topic=original_query,
            sub_queries=[original_query],
            language="ko",
            expected_sections=["개요", "상세", "결론"],
            complexity=0.5,
        )
