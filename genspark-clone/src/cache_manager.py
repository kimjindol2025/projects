"""
Cache Manager for Genspark Clone v2.0
- SHA256 기반 캐시 키 생성
- TTL 기반 자동 만료
- JSON 저장소 (output/.cache/)
"""

import hashlib
import json
import os
import time
from dataclasses import dataclass, asdict
from typing import Optional, Dict, Any
from pathlib import Path


@dataclass
class CacheEntry:
    """캐시 항목"""
    cache_key: str
    query: str
    result: dict  # SparkpageOutput JSON 직렬화
    created_at: float  # time.time()
    ttl_seconds: int  # 기본 86400 (24시간)
    hit_count: int = 0
    agent_type: str = "single"  # "single" | "multi"

    def is_expired(self) -> bool:
        """TTL 만료 여부"""
        return (time.time() - self.created_at) > self.ttl_seconds

    def to_dict(self) -> dict:
        """JSON 직렬화"""
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> "CacheEntry":
        """JSON 역직렬화"""
        return cls(**data)


class CacheManager:
    """캐시 관리자"""

    def __init__(self, cache_dir: str = "output/.cache", default_ttl: int = 86400):
        """
        Args:
            cache_dir: 캐시 디렉토리
            default_ttl: 기본 TTL (초)
        """
        self.cache_dir = Path(cache_dir)
        self.default_ttl = default_ttl
        self.index_file = self.cache_dir / "index.json"
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        self._index = self._load_index()

    def get_key(self, query: str, options: Optional[Dict[str, Any]] = None) -> str:
        """
        캐시 키 생성 (SHA256[:16])

        Args:
            query: 검색 쿼리
            options: 옵션 딕셔너리

        Returns:
            캐시 키
        """
        key_str = query
        if options:
            for k, v in sorted(options.items()):
                key_str += f"_{k}={v}"

        return hashlib.sha256(key_str.encode()).hexdigest()[:16]

    def get(self, key: str) -> Optional[dict]:
        """
        캐시에서 결과 조회

        Args:
            key: 캐시 키

        Returns:
            캐시된 결과 또는 None (만료/미존재)
        """
        if key not in self._index:
            return None

        cache_file = self.cache_dir / f"{key}.json"
        if not cache_file.exists():
            del self._index[key]
            self._save_index()
            return None

        try:
            with open(cache_file, "r", encoding="utf-8") as f:
                data = json.load(f)

            entry = CacheEntry.from_dict(data)

            if entry.is_expired():
                cache_file.unlink()
                del self._index[key]
                self._save_index()
                return None

            # 히트 카운트 증가
            entry.hit_count += 1
            self._update_entry(key, entry)

            return entry.result
        except Exception:
            return None

    def set(
        self,
        key: str,
        query: str,
        result: dict,
        agent_type: str = "single",
        ttl: Optional[int] = None
    ) -> None:
        """
        캐시에 결과 저장

        Args:
            key: 캐시 키
            query: 검색 쿼리
            result: 결과 (SparkpageOutput)
            agent_type: 에이전트 타입
            ttl: TTL (초), None이면 default_ttl 사용
        """
        ttl = ttl or self.default_ttl

        entry = CacheEntry(
            cache_key=key,
            query=query,
            result=result,
            created_at=time.time(),
            ttl_seconds=ttl,
            hit_count=0,
            agent_type=agent_type
        )

        cache_file = self.cache_dir / f"{key}.json"
        try:
            with open(cache_file, "w", encoding="utf-8") as f:
                json.dump(entry.to_dict(), f, ensure_ascii=False, indent=2)

            self._index[key] = {
                "query": query,
                "created_at": entry.created_at,
                "agent_type": agent_type
            }
            self._save_index()
        except Exception:
            pass

    def cleanup(self) -> int:
        """
        만료된 캐시 삭제

        Returns:
            삭제된 항목 수
        """
        deleted = 0
        expired_keys = []

        for key in list(self._index.keys()):
            cache_file = self.cache_dir / f"{key}.json"
            if not cache_file.exists():
                expired_keys.append(key)
                deleted += 1
                continue

            try:
                with open(cache_file, "r", encoding="utf-8") as f:
                    data = json.load(f)
                entry = CacheEntry.from_dict(data)

                if entry.is_expired():
                    cache_file.unlink()
                    expired_keys.append(key)
                    deleted += 1
            except Exception:
                expired_keys.append(key)
                deleted += 1

        for key in expired_keys:
            del self._index[key]

        if expired_keys:
            self._save_index()

        return deleted

    def stats(self) -> dict:
        """
        캐시 통계

        Returns:
            {
                'total_entries': int,
                'total_size_bytes': int,
                'by_agent_type': {'single': int, 'multi': int},
                'ttl_hours': int
            }
        """
        total_entries = len(self._index)
        total_size = 0
        by_agent_type = {"single": 0, "multi": 0}

        for key in self._index:
            cache_file = self.cache_dir / f"{key}.json"
            if cache_file.exists():
                total_size += cache_file.stat().st_size
                agent_type = self._index[key].get("agent_type", "single")
                if agent_type in by_agent_type:
                    by_agent_type[agent_type] += 1

        return {
            "total_entries": total_entries,
            "total_size_bytes": total_size,
            "by_agent_type": by_agent_type,
            "ttl_hours": self.default_ttl // 3600
        }

    def _load_index(self) -> dict:
        """인덱스 파일 로드"""
        if self.index_file.exists():
            try:
                with open(self.index_file, "r", encoding="utf-8") as f:
                    return json.load(f)
            except Exception:
                return {}
        return {}

    def _save_index(self) -> None:
        """인덱스 파일 저장"""
        try:
            with open(self.index_file, "w", encoding="utf-8") as f:
                json.dump(self._index, f, ensure_ascii=False, indent=2)
        except Exception:
            pass

    def _update_entry(self, key: str, entry: CacheEntry) -> None:
        """캐시 항목 업데이트"""
        cache_file = self.cache_dir / f"{key}.json"
        try:
            with open(cache_file, "w", encoding="utf-8") as f:
                json.dump(entry.to_dict(), f, ensure_ascii=False, indent=2)
        except Exception:
            pass
