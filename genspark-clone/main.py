#!/usr/bin/env python3
"""
Genspark Clone - CLI 진입점
사용법: python main.py "질문"
"""

import os
import sys
from pathlib import Path

from src.genspark_agent import GensparkAgent, AgentConfig


def main():
    """메인 실행"""
    # 인자 확인
    if len(sys.argv) < 2:
        print("사용법: python main.py '질문'")
        print("예시: python main.py '파이썬 비동기 프로그래밍이란'")
        sys.exit(1)

    user_query = sys.argv[1]

    # API 키 확인
    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        print("❌ 오류: ANTHROPIC_API_KEY 환경변수가 설정되지 않았습니다")
        print("설정: export ANTHROPIC_API_KEY='sk-ant-...'")
        sys.exit(1)

    # 설정
    config = AgentConfig(
        anthropic_api_key=api_key,
        output_dir="output",
        verbose=True,
    )

    # 에이전트 실행
    print("🚀 Genspark Clone 시작\n")
    agent = GensparkAgent(config)
    result = agent.run(user_query)

    if result:
        print(f"\n✅ 완료!")
        print(f"📄 Markdown: {result.markdown_path}")
        print(f"🌐 HTML: {result.html_path}")
    else:
        print("\n❌ 생성 실패")
        sys.exit(1)


if __name__ == "__main__":
    main()
