"""
Sparkpage 생성: SynthesisResult → HTML + Markdown
역할: 마크다운/HTML 파일 생성 (외부 라이브러리 의존 없음)
"""

from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Dict

from .claude_synthesizer import SynthesisResult
from .widget_renderer import WidgetRenderer, WIDGET_CSS


@dataclass
class SparkpageOutput:
    """생성된 Sparkpage"""
    markdown_path: str
    html_path: str
    markdown_content: str = ""
    html_content: str = ""
    title: str = ""
    generated_at: str = ""
    query: str = ""
    confidence_score: float = 0.0


class SparkpageGenerator:
    """Sparkpage 파일 생성"""

    def __init__(self, output_dir: str = "output"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)
        self.widget_renderer = WidgetRenderer()

    def generate(self, result: SynthesisResult, query: str) -> SparkpageOutput:
        """SynthesisResult → HTML + MD 파일 생성"""
        # 파일명 생성
        slug = self._slug(query)
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        base_name = f"{timestamp}_{slug}"

        md_path = self.output_dir / f"{base_name}.md"
        html_path = self.output_dir / f"{base_name}.html"

        # 마크다운 생성
        markdown_content = self._generate_markdown(result, query)
        md_path.write_text(markdown_content, encoding="utf-8")

        # HTML 생성
        html_content = self._generate_html(result, query, markdown_content)
        html_path.write_text(html_content, encoding="utf-8")

        return SparkpageOutput(
            markdown_path=str(md_path),
            html_path=str(html_path),
            markdown_content=markdown_content,
            html_content=html_content,
            title=query,
            generated_at=timestamp,
            query=query,
            confidence_score=result.confidence_score,
        )

    def _generate_markdown(self, result: SynthesisResult, query: str) -> str:
        """마크다운 생성"""
        lines = [
            f"# {query}",
            "",
            f"**생성일시**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}  ",
            f"**신뢰도**: {result.confidence_score:.0%}  ",
            f"**소스**: {result.total_sources}개  ",
            "",
        ]

        # 핵심 사실
        if result.key_facts:
            lines.append("## 핵심 사실")
            for fact in result.key_facts:
                lines.append(f"- {fact}")
            lines.append("")

        # 섹션
        for section in result.sections:
            lines.append(f"## {section.title}")
            lines.append(section.content)
            if section.sources:
                lines.append("\n**출처:**")
                for url in section.sources[:3]:
                    lines.append(f"- {url}")
            lines.append("")

        return "\n".join(lines)

    def _generate_html(self, result: SynthesisResult, query: str, markdown: str) -> str:
        """HTML 생성"""
        # 마크다운 → HTML 간단 변환
        html_body = self._markdown_to_html(markdown)

        meta = {
            "title": query,
            "confidence": f"{result.confidence_score:.0%}",
            "sources": str(result.total_sources),
            "generated_at": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        }

        return self._generate_html_template(html_body, query, meta)

    def _markdown_to_html(self, markdown: str) -> str:
        """Markdown → HTML 간단 변환 (외부 라이브러리 없음)"""
        lines = markdown.split("\n")
        html_lines = []
        in_list = False

        for line in lines:
            # 제목
            if line.startswith("# "):
                if in_list:
                    html_lines.append("</ul>")
                    in_list = False
                html_lines.append(f"<h1>{line[2:].strip()}</h1>")
            elif line.startswith("## "):
                if in_list:
                    html_lines.append("</ul>")
                    in_list = False
                html_lines.append(f"<h2>{line[3:].strip()}</h2>")
            elif line.startswith("### "):
                if in_list:
                    html_lines.append("</ul>")
                    in_list = False
                html_lines.append(f"<h3>{line[4:].strip()}</h3>")
            # 불릿
            elif line.startswith("- "):
                if not in_list:
                    html_lines.append("<ul>")
                    in_list = True
                html_lines.append(f"<li>{line[2:].strip()}</li>")
            # 링크 변환
            elif line.startswith("**") and line.endswith(":**"):
                if in_list:
                    html_lines.append("</ul>")
                    in_list = False
                html_lines.append(f"<strong>{line[2:-2]}</strong>")
            # 공백 줄
            elif not line.strip():
                if in_list:
                    html_lines.append("</ul>")
                    in_list = False
                html_lines.append("")
            # 일반 텍스트
            else:
                if in_list and not line.startswith("- "):
                    html_lines.append("</ul>")
                    in_list = False
                if line.strip():
                    html_lines.append(f"<p>{line.strip()}</p>")

        if in_list:
            html_lines.append("</ul>")

        return "\n".join(html_lines)

    def _generate_html_template(self, body: str, title: str, meta: Dict) -> str:
        """HTML 템플릿 생성 (위젯 CSS 포함)"""
        return f"""<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title} - Genspark</title>
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; color: #333; }}
        .container {{ max-width: 900px; margin: 0 auto; padding: 20px; }}
        header {{ border-bottom: 2px solid #007bff; margin-bottom: 30px; padding-bottom: 20px; }}
        h1 {{ font-size: 2em; margin-bottom: 10px; color: #0056b3; }}
        h2 {{ font-size: 1.5em; margin-top: 30px; margin-bottom: 10px; color: #0056b3; border-left: 4px solid #007bff; padding-left: 10px; }}
        h3 {{ font-size: 1.2em; margin-top: 20px; margin-bottom: 8px; }}
        p {{ margin-bottom: 15px; }}
        ul {{ margin: 10px 0 15px 30px; }}
        li {{ margin-bottom: 8px; }}
        strong {{ font-weight: bold; }}
        .meta {{ font-size: 0.9em; color: #666; margin-bottom: 15px; }}
        .meta-item {{ display: inline-block; margin-right: 20px; }}
        .meta-label {{ font-weight: bold; color: #333; }}
        aside {{ background: #f8f9fa; border-left: 3px solid #007bff; padding: 15px; margin-top: 40px; }}
        aside h3 {{ margin-top: 0; }}
        aside a {{ color: #007bff; text-decoration: none; word-break: break-all; }}
        aside a:hover {{ text-decoration: underline; }}

        /* v2.0 Widget Styles */
        {WIDGET_CSS}
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>{title}</h1>
            <div class="meta">
                <span class="meta-item"><span class="meta-label">신뢰도:</span> {meta['confidence']}</span>
                <span class="meta-item"><span class="meta-label">소스:</span> {meta['sources']}개</span>
                <span class="meta-item"><span class="meta-label">생성:</span> {meta['generated_at']}</span>
            </div>
        </header>
        <main>
            {body}
        </main>
        <aside>
            <h3>ℹ️ 이 Sparkpage는 AI가 자동 생성했습니다</h3>
            <p>Genspark Clone v2.0으로 생성된 종합 정보 페이지입니다. 다양한 관점의 정보 수집과 위젯 기반 시각화를 포함합니다.</p>
        </aside>
    </div>
</body>
</html>"""

    def _slug(self, text: str) -> str:
        """텍스트 → URL-safe 슬러그"""
        import re
        text = text.lower()
        text = re.sub(r"[^a-z0-9가-힣]", "_", text)
        text = re.sub(r"_+", "_", text)
        return text[:50].rstrip("_")
