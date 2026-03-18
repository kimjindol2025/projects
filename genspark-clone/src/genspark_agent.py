"""
Genspark 에이전트: 전체 파이프라인 오케스트레이션
역할: 질문 입력 → Sparkpage 파일 생성 (5단계)
"""

from dataclasses import dataclass
from typing import Optional

from .query_analyzer import QueryAnalyzer
from .web_searcher import DuckDuckGoSearcher
from .content_fetcher import ContentFetcher
from .claude_synthesizer import ClaudeSynthesizer
from .sparkpage_generator import SparkpageGenerator, SparkpageOutput


@dataclass
class AgentConfig:
    """에이전트 설정"""
    anthropic_api_key: str
    analyze_model: str = "claude-haiku-4-5-20251001"
    synthesize_model: str = "claude-sonnet-4-6"
    max_search_results: int = 5
    max_fetch_workers: int = 3
    output_dir: str = "output"
    search_delay: float = 1.0
    verbose: bool = True


class GensparkAgent:
    """Genspark 전체 파이프라인"""

    def __init__(self, config: AgentConfig):
        self.config = config
        self.analyzer = QueryAnalyzer(config.anthropic_api_key, config.analyze_model)
        self.searcher = DuckDuckGoSearcher(max_results=config.max_search_results)
        self.fetcher = ContentFetcher()
        self.fetcher.MAX_WORKERS = config.max_fetch_workers  # 설정 적용
        self.synthesizer = ClaudeSynthesizer(
            config.anthropic_api_key, config.synthesize_model
        )
        self.generator = SparkpageGenerator(config.output_dir)

    def run(self, user_query: str) -> Optional[SparkpageOutput]:
        """사용자 질문 → Sparkpage 생성"""
        try:
            self._log("INIT", f"시작: '{user_query}'")

            # Step 1: 질문 분석
            query_spec = self._step1_analyze(user_query)
            self._log(
                "ANALYZE",
                f"분해됨: {len(query_spec.sub_queries)}개 서브쿼리, "
                f"예상 섹션: {', '.join(query_spec.expected_sections[:2])}",
            )

            # Step 2: 웹 검색
            search_results = self._step2_search(query_spec)
            total_results = sum(len(v) for v in search_results.values())
            self._log("SEARCH", f"검색 완료: {total_results}개 결과")

            # Step 3: 콘텐츠 페칭
            contents = self._step3_fetch(search_results)
            valid_count = sum(1 for c in contents if c.fetch_status == "ok")
            self._log("FETCH", f"페칭 완료: {valid_count}/{len(contents)} 유효")

            # Step 4: AI 합산
            synthesis_result = self._step4_synthesize(query_spec, contents)
            self._log(
                "SYNTHESIZE",
                f"합산 완료: {len(synthesis_result.sections)}개 섹션, "
                f"신뢰도 {synthesis_result.confidence_score:.0%}",
            )

            # Step 5: Sparkpage 생성
            output = self._step5_generate(synthesis_result, user_query)
            self._log("GENERATE", f"생성 완료: {output.html_path}")

            return output
        except Exception as e:
            self._log("ERROR", f"파이프라인 실패: {e}")
            return None

    def _step1_analyze(self, query: str):
        """Step 1: QueryAnalyzer 실행"""
        return self.analyzer.analyze(query)

    def _step2_search(self, spec):
        """Step 2: DuckDuckGoSearcher 실행"""
        return self.searcher.search_multiple(spec.sub_queries, self.config.search_delay)

    def _step3_fetch(self, search_results):
        """Step 3: ContentFetcher 실행"""
        return self.fetcher.fetch_for_queries(search_results)

    def _step4_synthesize(self, spec, contents):
        """Step 4: ClaudeSynthesizer 실행"""
        return self.synthesizer.synthesize(spec, contents)

    def _step5_generate(self, result, query):
        """Step 5: SparkpageGenerator 실행"""
        return self.generator.generate(result, query)

    def _log(self, step: str, message: str):
        """로그 출력"""
        if self.config.verbose:
            print(f"[{step}] {message}")
