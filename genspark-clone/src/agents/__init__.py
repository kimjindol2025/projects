"""Specialized researcher agents for v2.0"""

from .researcher_agent import (
    BaseResearcherAgent,
    GeneralAgent,
    TechAgent,
    NewsAgent,
    ReviewAgent,
    AgentSearchResult,
)

__all__ = [
    "BaseResearcherAgent",
    "GeneralAgent",
    "TechAgent",
    "NewsAgent",
    "ReviewAgent",
    "AgentSearchResult",
]
