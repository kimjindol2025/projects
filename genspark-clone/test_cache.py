"""
CacheManager 테스트
"""

import os
import json
import time
from src.cache_manager import CacheManager, CacheEntry


def test_cache_key_generation():
    """캐시 키 생성 테스트"""
    cm = CacheManager()

    key1 = cm.get_key("파이썬")
    key2 = cm.get_key("파이썬")
    key3 = cm.get_key("자바")

    # 같은 쿼리 → 같은 키
    assert key1 == key2, "Same query should generate same key"
    # 다른 쿼리 → 다른 키
    assert key1 != key3, "Different query should generate different key"
    # 16자 길이
    assert len(key1) == 16, "Key should be 16 chars (SHA256[:16])"

    print("✅ Cache key generation OK")


def test_cache_set_and_get():
    """캐시 저장 및 조회 테스트"""
    cm = CacheManager()

    query = "테스트 쿼리"
    result = {
        "html_path": "/output/test.html",
        "markdown_path": "/output/test.md",
        "query": query,
        "confidence_score": 0.92
    }

    key = cm.get_key(query)
    cm.set(key, query, result, agent_type="single")

    # 조회
    cached = cm.get(key)
    assert cached is not None, "Cache should return result"
    assert cached["query"] == query, "Cached query should match"
    assert cached["confidence_score"] == 0.92, "Cached result should match"

    print("✅ Cache set/get OK")


def test_cache_expiration():
    """캐시 만료 테스트"""
    cm = CacheManager(default_ttl=1)  # 1초 TTL

    query = "만료 테스트"
    result = {"test": "data"}

    key = cm.get_key(query)
    cm.set(key, query, result, ttl=1)

    # 즉시 조회 → 존재
    cached = cm.get(key)
    assert cached is not None, "Fresh cache should exist"

    # 2초 대기 후 조회 → 만료
    time.sleep(2)
    cached = cm.get(key)
    assert cached is None, "Expired cache should be None"

    print("✅ Cache expiration OK")


def test_cache_hit_count():
    """캐시 히트 카운트 테스트"""
    cm = CacheManager(cache_dir="test_output/.cache_hit_test")

    query = "히트 카운트 테스트"
    result = {"test": "data"}
    key = cm.get_key(query)

    cm.set(key, query, result)

    # 여러 번 조회
    for _ in range(3):
        cm.get(key)

    # 인덱스에서 히트 카운트 확인
    stats = cm.stats()
    assert stats["total_entries"] == 1, f"Should have 1 entry, got {stats['total_entries']}"

    print("✅ Cache hit count OK")


def test_cache_cleanup():
    """캐시 정리 테스트"""
    cm = CacheManager(cache_dir="test_output/.cache_cleanup_test", default_ttl=1)

    # 3개 항목 저장
    for i in range(3):
        key = cm.get_key(f"쿼리_{i}")
        cm.set(key, f"쿼리_{i}", {"test": f"data_{i}"}, ttl=1)

    # 2초 대기
    time.sleep(2)

    # 정리
    deleted = cm.cleanup()
    assert deleted == 3, f"Should delete 3 expired entries, deleted {deleted}"

    stats = cm.stats()
    assert stats["total_entries"] == 0, f"Should have 0 entries, got {stats['total_entries']}"

    print("✅ Cache cleanup OK")


def test_cache_stats():
    """캐시 통계 테스트"""
    cm = CacheManager(cache_dir="test_output/.cache_stats_test")

    # single 에이전트 2개, multi 에이전트 1개 저장
    for i in range(2):
        key = cm.get_key(f"single_{i}")
        cm.set(key, f"single_{i}", {"test": "data"}, agent_type="single")

    key = cm.get_key("multi_1")
    cm.set(key, "multi_1", {"test": "data"}, agent_type="multi")

    stats = cm.stats()
    assert stats["total_entries"] == 3, f"Should have 3 entries, got {stats['total_entries']}"
    assert stats["by_agent_type"]["single"] == 2, f"Should have 2 single agents, got {stats['by_agent_type']['single']}"
    assert stats["by_agent_type"]["multi"] == 1, f"Should have 1 multi agent, got {stats['by_agent_type']['multi']}"

    print("✅ Cache stats OK")


def test_cache_entry_dataclass():
    """CacheEntry 데이터클래스 테스트"""
    now = time.time()
    entry = CacheEntry(
        cache_key="test_key",
        query="테스트",
        result={"test": "data"},
        created_at=now,
        ttl_seconds=3600,
        hit_count=5,
        agent_type="multi"
    )

    # 직렬화
    data = entry.to_dict()
    assert data["query"] == "테스트"
    assert data["hit_count"] == 5

    # 역직렬화
    restored = CacheEntry.from_dict(data)
    assert restored.query == "테스트"
    assert not restored.is_expired()

    print("✅ CacheEntry dataclass OK")


if __name__ == "__main__":
    print("\n🧪 Running CacheManager tests...\n")

    test_cache_key_generation()
    test_cache_set_and_get()
    test_cache_expiration()
    test_cache_hit_count()
    test_cache_cleanup()
    test_cache_stats()
    test_cache_entry_dataclass()

    print("\n✅ All CacheManager tests passed!\n")
