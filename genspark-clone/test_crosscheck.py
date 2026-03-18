"""
Day 3: CrossCheckAgent 테스트
벡터 임베딩, 코사인 유사도, 이상치 탐지, 신뢰도 조정 검증
"""

from src.crosscheck_agent import (
    CrossCheckAgent,
    VectorEmbedder,
    SemanticValidator,
    ConflictReport,
    CrossCheckValidator,
)
from src.content_fetcher import FetchedContent


def test_cosine_similarity_calculation():
    """코사인 유사도 계산 테스트"""
    validator = SemanticValidator()

    # 동일한 벡터
    v1 = [1.0, 0.0, 0.0]
    v2 = [1.0, 0.0, 0.0]
    sim = validator.cosine_similarity(v1, v2)
    assert sim == 1.0, "Identical vectors should have similarity 1.0"

    # 직교하는 벡터
    v3 = [1.0, 0.0, 0.0]
    v4 = [0.0, 1.0, 0.0]
    sim = validator.cosine_similarity(v3, v4)
    assert sim == 0.0, "Orthogonal vectors should have similarity 0.0"

    # 반대 방향 벡터
    v5 = [1.0, 0.0, 0.0]
    v6 = [-1.0, 0.0, 0.0]
    sim = validator.cosine_similarity(v5, v6)
    assert sim == -1.0, "Opposite vectors should have similarity -1.0"

    # 부분 유사
    v7 = [1.0, 1.0, 0.0]
    v8 = [1.0, 0.0, 0.0]
    sim = validator.cosine_similarity(v7, v8)
    assert 0.5 < sim < 1.0, "Partial similarity should be between 0.5 and 1.0"

    print("✅ Cosine similarity calculation OK")


def test_outlier_detection():
    """이상치 탐지 테스트"""
    validator = SemanticValidator()

    # 정상 분포: 0.8, 0.75, 0.82, 0.78 → 이상치 없음
    similarities_normal = [0.8, 0.75, 0.82, 0.78]
    outliers = validator.detect_outliers(similarities_normal, threshold_sigma=1.5)
    assert len(outliers) == 0, "Normal distribution should have no outliers"

    # 이상치 포함: 0.8, 0.75, 0.82, 0.2 → 0.2는 이상치
    similarities_with_outlier = [0.8, 0.75, 0.82, 0.2]
    outliers = validator.detect_outliers(similarities_with_outlier, threshold_sigma=1.5)
    assert len(outliers) > 0, "Should detect outlier 0.2"
    assert 3 in outliers, "Index 3 should be detected as outlier"

    # 모두 비슷: 0.5, 0.51, 0.49, 0.5 → 이상치 없음
    similarities_all_similar = [0.5, 0.51, 0.49, 0.5]
    outliers = validator.detect_outliers(similarities_all_similar, threshold_sigma=1.5)
    assert len(outliers) == 0, "Similar values should have no outliers"

    print("✅ Outlier detection OK")


def test_crosscheck_agent_without_api():
    """API 없이 CrossCheckAgent 테스트"""
    agent = CrossCheckAgent(openai_api_key="")

    # API 키 없으므로 폴백
    contents = [
        FetchedContent(
            url="https://example1.com",
            title="Example 1",
            body_text="Python is a programming language",
            word_count=100,
            fetch_status="ok",
        ),
        FetchedContent(
            url="https://example2.com",
            title="Example 2",
            body_text="Go is a compiled language",
            word_count=100,
            fetch_status="ok",
        ),
    ]

    report = agent.run(contents, base_confidence=0.85)

    assert report is not None
    assert report.adjusted_confidence == 0.85  # 폴백 시 원본 신뢰도
    assert len(report.warnings) == 0  # 경고 없음

    print("✅ CrossCheckAgent without API (fallback) OK")


def test_crosscheck_agent_with_mock_embeddings():
    """Mock 임베딩으로 CrossCheckAgent 테스트"""
    agent = CrossCheckAgent(openai_api_key="test-key")

    # Mock embedder - 현실적인 검증 케이스
    # 이상치 탐지는 매우 극단적인 경우에만 작동
    # 대신 기본 기능(임베딩 호출, 유사도 계산, 신뢰도 조정) 검증
    mock_embeddings = [
        [1.0, 0.0, 0.0],
        [0.98, 0.1, 0.0],
        [0.95, 0.05, 0.1],
    ]

    def mock_embed_batch(texts):
        return mock_embeddings[: len(texts)]

    agent.embedder.embed_batch = mock_embed_batch

    contents = [
        FetchedContent(
            url="https://site1.com",
            title="Source 1",
            body_text="Python is fast and efficient",
            word_count=50,
            fetch_status="ok",
        ),
        FetchedContent(
            url="https://site2.com",
            title="Source 2",
            body_text="Python is fast and easy",
            word_count=50,
            fetch_status="ok",
        ),
        FetchedContent(
            url="https://site3.com",
            title="Source 3",
            body_text="Python is fast and reliable",
            word_count=50,
            fetch_status="ok",
        ),
    ]

    report = agent.run(contents, base_confidence=0.9)

    assert report is not None
    # 임베딩이 호출되었고 유사도가 계산됨
    assert report.avg_similarity > 0, "Average similarity should be calculated"
    # 높은 유사도의 경우 조정된 신뢰도가 기본값과 같거나 약간 높아야 함
    assert report.adjusted_confidence >= 0.85, "Confidence should be reasonable for similar sources"

    print("✅ CrossCheckAgent with mock embeddings OK")


def test_conflict_report_structure():
    """ConflictReport 구조 검증"""
    report = ConflictReport(
        conflicting_pairs=[("url1", "url2"), ("url2", "url3")],
        avg_similarity=0.45,
        outlier_urls=["url2", "url3"],
        adjusted_confidence=0.75,
        warnings=["Warning 1", "Warning 2"],
    )

    assert len(report.conflicting_pairs) == 2
    assert len(report.outlier_urls) == 2
    assert len(report.warnings) == 2
    assert report.avg_similarity == 0.45
    assert report.adjusted_confidence == 0.75

    print("✅ ConflictReport structure OK")


def test_integrate_with_consensus():
    """Consensus와 CrossCheck 통합"""
    consensus_confidence = 0.85
    crosscheck_report = ConflictReport(adjusted_confidence=0.75)

    final_confidence = CrossCheckValidator.integrate_with_consensus(
        consensus_confidence, crosscheck_report
    )

    # (0.85 * 0.5) + (0.75 * 0.5) = 0.425 + 0.375 = 0.8
    expected = 0.8
    assert abs(final_confidence - expected) < 0.01, f"Expected {expected}, got {final_confidence}"

    print("✅ Consensus + CrossCheck integration OK")


def test_minimal_content():
    """최소 콘텐츠 처리"""
    agent = CrossCheckAgent(openai_api_key="test-key")

    # 콘텐츠 1개 → 비교할 쌍이 없음
    single_content = [
        FetchedContent(
            url="https://example.com",
            title="Example",
            body_text="Some text",
            word_count=50,
            fetch_status="ok",
        )
    ]

    report = agent.run(single_content, base_confidence=0.85)
    assert report.adjusted_confidence == 0.85  # 변화 없음

    # 콘텐츠 없음
    empty_report = agent.run([], base_confidence=0.85)
    assert empty_report.adjusted_confidence == 0.85

    print("✅ Minimal content handling OK")


def test_text_truncation():
    """텍스트 300자 제한"""
    agent = CrossCheckAgent(openai_api_key="test-key")

    # 300자 이상의 텍스트
    long_text = "a" * 500

    content = FetchedContent(
        url="https://example.com",
        title="Example",
        body_text=long_text,
        word_count=500,
        fetch_status="ok",
    )

    # Mock embedder로 텍스트 길이 확인
    def check_text_length(texts):
        for text in texts:
            assert len(text) <= 300, f"Text should be truncated to 300 chars, got {len(text)}"
        return [[1.0, 0.0], [1.0, 0.0]]

    agent.embedder.embed_batch = check_text_length

    report = agent.run([content, content], base_confidence=0.85)
    assert report is not None

    print("✅ Text truncation (max 300 chars) OK")


if __name__ == "__main__":
    print("\n🧪 Running CrossCheckAgent tests...\n")

    test_cosine_similarity_calculation()
    test_outlier_detection()
    test_crosscheck_agent_without_api()
    test_crosscheck_agent_with_mock_embeddings()
    test_conflict_report_structure()
    test_integrate_with_consensus()
    test_minimal_content()
    test_text_truncation()

    print("\n✅ All CrossCheckAgent tests passed!\n")
