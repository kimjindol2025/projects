"""
WidgetRenderer 테스트
"""

from src.widget_renderer import WidgetRenderer


def test_widget_table_detection():
    """테이블 위젯 감지"""
    renderer = WidgetRenderer()

    content = """
| 항목 | 설명 |
|------|------|
| A | 첫 번째 |
| B | 두 번째 |
| C | 세 번째 |
    """

    widget = renderer.render(content)
    assert widget.widget_type == "table", "Should detect table"
    assert "widget-table" in widget.css_class
    assert "<table" in widget.html

    print("✅ Table widget detection OK")


def test_widget_list_detection():
    """목록 위젯 감지"""
    renderer = WidgetRenderer()

    content = """
- 첫 번째 항목
- 두 번째 항목
- 세 번째 항목
    """

    widget = renderer.render(content)
    assert widget.widget_type == "list", "Should detect list"
    assert "widget-list" in widget.css_class
    assert "<ul" in widget.html

    print("✅ List widget detection OK")


def test_widget_timeline_detection():
    """타임라인 위젯 감지"""
    renderer = WidgetRenderer()

    content = """
1. 첫 번째 단계
2. 두 번째 단계
3. 세 번째 단계
    """

    widget = renderer.render(content)
    assert widget.widget_type == "timeline", "Should detect timeline"
    assert "widget-timeline" in widget.css_class
    assert "timeline-item" in widget.html

    print("✅ Timeline widget detection OK")


def test_widget_quote_detection():
    """인용 위젯 감지"""
    renderer = WidgetRenderer()

    content = "> 이것은 명언입니다."

    widget = renderer.render(content)
    assert widget.widget_type == "quote", "Should detect quote"
    assert "widget-quote" in widget.css_class
    assert "<blockquote" in widget.html

    print("✅ Quote widget detection OK")


def test_widget_factbox_detection():
    """FactBox 위젯 감지"""
    renderer = WidgetRenderer()

    content = "첫 번째 사실입니다. 두 번째 사실입니다. 세 번째 사실입니다."

    widget = renderer.render(content, section_type="overview")
    assert widget.widget_type == "factbox", "Should detect factbox for overview"
    assert "widget-factbox" in widget.css_class
    assert "fact-item" in widget.html

    print("✅ FactBox widget detection OK")


def test_widget_text_fallback():
    """텍스트 위젯 폴백"""
    renderer = WidgetRenderer()

    content = "이것은 일반 텍스트입니다."

    widget = renderer.render(content)
    assert widget.widget_type == "text", "Should fallback to text"
    assert "widget-text" in widget.css_class
    assert "<div class=\"widget-text\">" in widget.html

    print("✅ Text widget fallback OK")


def test_widget_rendering():
    """위젯 렌더링"""
    renderer = WidgetRenderer()

    # 테이블 렌더링
    table_content = "| A | B |\n|---|---|\n| 1 | 2 |"
    widget = renderer.render(table_content)
    assert "<table" in widget.html
    assert "<td>" in widget.html

    # 목록 렌더링 (3개 이상 필요)
    list_content = "- 항목 1\n- 항목 2\n- 항목 3"
    widget = renderer.render(list_content)
    assert "<ul" in widget.html
    assert "<li>" in widget.html

    print("✅ Widget rendering OK")


def test_html_escape():
    """HTML 이스케이프"""
    renderer = WidgetRenderer()

    content = "- <script>alert('xss')</script>"
    widget = renderer.render(content)

    assert "<script>" not in widget.html
    assert "&lt;script&gt;" in widget.html

    print("✅ HTML escape OK")


def test_widget_multiple_types():
    """여러 위젯 타입 순차 처리"""
    renderer = WidgetRenderer()

    test_cases = [
        ("| a | b |\n|---|---|\n| 1 | 2 |", "table"),
        ("- 항목1\n- 항목2\n- 항목3", "list"),
        ("1. 단계1\n2. 단계2\n3. 단계3", "timeline"),
        ("> 인용", "quote"),
    ]

    for content, expected_type in test_cases:
        widget = renderer.render(content)
        assert widget.widget_type == expected_type, f"Failed for {expected_type}"

    print("✅ Multiple widget types OK")


if __name__ == "__main__":
    print("\n🧪 Running WidgetRenderer tests...\n")

    test_widget_table_detection()
    test_widget_list_detection()
    test_widget_timeline_detection()
    test_widget_quote_detection()
    test_widget_factbox_detection()
    test_widget_text_fallback()
    test_widget_rendering()
    test_html_escape()
    test_widget_multiple_types()

    print("\n✅ All WidgetRenderer tests passed!\n")
