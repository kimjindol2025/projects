"""
Widget Renderer for Dynamic Sparkpage
- 콘텐츠 타입 자동 감지
- 6가지 위젯 타입 (Table, List, Timeline, Quote, FactBox, Text)
- 마크다운 → 위젯 HTML 변환
"""

from dataclasses import dataclass
from typing import List, Optional
import re


@dataclass
class RenderedWidget:
    """렌더링된 위젯"""
    html: str
    widget_type: str  # "text"|"table"|"list"|"timeline"|"quote"|"factbox"
    css_class: str


class WidgetRenderer:
    """위젯 렌더러"""

    # 위젯 감지 규칙
    TABLE_THRESHOLD = 3  # '|' 구분자 3개 이상
    LIST_THRESHOLD = 3  # '- ' 또는 '1. ' 3개 이상
    TIMELINE_THRESHOLD = 3  # 단계/Step/년도 3개 이상
    QUOTE_THRESHOLD = 1  # '> ' 1개 이상

    def __init__(self):
        pass

    def render(
        self,
        content: str,
        section_type: str = "overview"
    ) -> RenderedWidget:
        """
        콘텐츠 타입 감지 및 위젯 렌더링

        Args:
            content: 마크다운 콘텐츠
            section_type: 섹션 타입 (overview/detail/example/summary)

        Returns:
            RenderedWidget
        """
        # 위젯 타입 감지
        widget_type = self._detect_widget_type(content, section_type)

        # 위젯별 렌더링
        if widget_type == "table":
            html = self._render_table(content)
        elif widget_type == "list":
            html = self._render_list(content)
        elif widget_type == "timeline":
            html = self._render_timeline(content)
        elif widget_type == "quote":
            html = self._render_quote(content)
        elif widget_type == "factbox":
            html = self._render_factbox(content, section_type)
        else:  # text
            html = self._render_text(content)

        return RenderedWidget(
            html=html,
            widget_type=widget_type,
            css_class=f"widget-{widget_type}"
        )

    def _detect_widget_type(self, content: str, section_type: str) -> str:
        """
        위젯 타입 감지

        Args:
            content: 콘텐츠
            section_type: 섹션 타입

        Returns:
            위젯 타입
        """
        # 테이블: '|' 구분자
        pipe_count = content.count('|')
        if pipe_count >= self.TABLE_THRESHOLD:
            return "table"

        # 인용: '> ' 마크다운
        if content.strip().startswith('>'):
            return "quote"

        # 타임라인: 단계/Step/년도 (목록 감지 전)
        timeline_keywords = len(re.findall(r'(단계|step|2025|2024|2023)', content, re.IGNORECASE))
        if timeline_keywords >= self.TIMELINE_THRESHOLD:
            return "timeline"

        # 목록: '- ' 또는 '1. '
        list_items = len(re.findall(r'^[-*]\s', content, re.MULTILINE))
        ordered_items = len(re.findall(r'^\d+\.\s', content, re.MULTILINE))
        if (list_items + ordered_items) >= self.LIST_THRESHOLD:
            return "list"

        # FactBox: overview 섹션 + 짧은 문장 다수
        if section_type == "overview":
            sentences = len(re.split(r'[.!?]+', content))
            if sentences >= 3 and len(content) < 500:
                return "factbox"

        return "text"

    def _render_table(self, content: str) -> str:
        """테이블 위젯"""
        lines = content.strip().split('\n')
        html = '<table class="widget-table">\n'

        for line in lines:
            if '|' in line:
                cells = [cell.strip() for cell in line.split('|')]
                cells = [c for c in cells if c]  # 빈 셀 제거

                if cells:
                    html += '  <tr>\n'
                    for cell in cells:
                        html += f'    <td>{self._escape(cell)}</td>\n'
                    html += '  </tr>\n'

        html += '</table>'
        return html

    def _render_list(self, content: str) -> str:
        """목록 위젯"""
        html = '<ul class="widget-list">\n'

        lines = content.strip().split('\n')
        for line in lines:
            # '- ' 또는 '* ' 패턴
            match = re.match(r'^[-*]\s+(.+)', line)
            if match:
                item = match.group(1)
                html += f'  <li>{self._escape(item)}</li>\n'
            # '1. ' 패턴
            match = re.match(r'^\d+\.\s+(.+)', line)
            if match:
                item = match.group(1)
                html += f'  <li>{self._escape(item)}</li>\n'

        html += '</ul>'
        return html

    def _render_timeline(self, content: str) -> str:
        """타임라인 위젯"""
        html = '<div class="widget-timeline">\n'

        lines = content.strip().split('\n')
        for line in lines:
            line = line.strip()
            if not line:
                continue

            # 단계/Step 패턴
            match = re.match(r'(^\d+\.|단계|step\s*\d*[:：]?)\s*(.+)', line, re.IGNORECASE)
            if match:
                title = match.group(2)
                html += f'  <div class="timeline-item">\n'
                html += f'    <div class="timeline-marker"></div>\n'
                html += f'    <div class="timeline-content">{self._escape(title)}</div>\n'
                html += f'  </div>\n'
            else:
                html += f'  <div class="timeline-item">\n'
                html += f'    <div class="timeline-content">{self._escape(line)}</div>\n'
                html += f'  </div>\n'

        html += '</div>'
        return html

    def _render_quote(self, content: str) -> str:
        """인용 위젯"""
        lines = content.strip().split('\n')
        quote_text = '\n'.join(
            line.lstrip('> ').strip() for line in lines if line.startswith('>')
        )
        html = f'<blockquote class="widget-quote">\n'
        html += f'  <p>{self._escape(quote_text)}</p>\n'
        html += f'</blockquote>'
        return html

    def _render_factbox(self, content: str, section_type: str) -> str:
        """FactBox 위젯"""
        sentences = [s.strip() for s in re.split(r'[.!?]+', content) if s.strip()]
        html = f'<div class="widget-factbox">\n'

        for i, sentence in enumerate(sentences[:5]):  # 최대 5개
            html += f'  <div class="fact-item">\n'
            html += f'    <span class="fact-number">{i + 1}</span>\n'
            html += f'    <span class="fact-text">{self._escape(sentence)}</span>\n'
            html += f'  </div>\n'

        html += '</div>'
        return html

    def _render_text(self, content: str) -> str:
        """일반 텍스트 위젯"""
        # 단락 단위로 렌더링
        html = '<div class="widget-text">\n'
        paragraphs = content.split('\n\n')

        for para in paragraphs:
            if para.strip():
                html += f'  <p>{self._escape(para.strip())}</p>\n'

        html += '</div>'
        return html

    @staticmethod
    def _escape(text: str) -> str:
        """HTML 이스케이프"""
        text = text.replace('&', '&amp;')
        text = text.replace('<', '&lt;')
        text = text.replace('>', '&gt;')
        text = text.replace('"', '&quot;')
        return text


# 위젯 CSS 스타일
WIDGET_CSS = """
/* Widget Styles */
.widget-table {
    width: 100%;
    border-collapse: collapse;
    margin: 15px 0;
    border: 1px solid #ddd;
}

.widget-table tr {
    border-bottom: 1px solid #ddd;
}

.widget-table td {
    padding: 10px;
    text-align: left;
}

.widget-table tr:nth-child(even) {
    background: #f8f9fa;
}

.widget-list {
    list-style: disc;
    margin: 15px 0 15px 30px;
    padding: 0;
}

.widget-list li {
    margin-bottom: 8px;
}

.widget-timeline {
    margin: 20px 0;
    position: relative;
}

.timeline-item {
    display: flex;
    margin-bottom: 20px;
    padding-left: 30px;
    position: relative;
}

.timeline-marker {
    width: 12px;
    height: 12px;
    background: #007bff;
    border-radius: 50%;
    position: absolute;
    left: 0;
    top: 5px;
}

.timeline-content {
    flex: 1;
}

.widget-quote {
    border-left: 4px solid #007bff;
    padding-left: 15px;
    margin: 15px 0;
    font-style: italic;
    color: #555;
}

.widget-quote p {
    margin: 10px 0;
}

.widget-factbox {
    background: #f8f9fa;
    border: 1px solid #ddd;
    border-radius: 5px;
    padding: 15px;
    margin: 15px 0;
}

.fact-item {
    display: flex;
    margin-bottom: 10px;
}

.fact-number {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    background: #007bff;
    color: white;
    border-radius: 50%;
    margin-right: 10px;
    font-size: 12px;
    font-weight: bold;
    flex-shrink: 0;
}

.fact-text {
    flex: 1;
}

.widget-text p {
    margin-bottom: 15px;
    line-height: 1.6;
}
"""
