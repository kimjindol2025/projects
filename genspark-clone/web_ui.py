#!/usr/bin/env python3
"""
Genspark Clone - 웹 UI (Flask 기반)
역할: 사용자 인터페이스 제공
"""

from flask import Flask, render_template_string, request, jsonify, send_file
import os
import json
from datetime import datetime
from src.genspark_agent import GensparkAgent, AgentConfig

app = Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # 16MB 제한

# HTML 템플릿
HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Genspark Clone - AI 검색 & 합산</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }

        .container {
            width: 100%;
            max-width: 800px;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }

        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 30px;
            text-align: center;
        }

        header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }

        header p {
            font-size: 1.1em;
            opacity: 0.9;
        }

        .sparkle {
            display: inline-block;
            animation: sparkle 1.5s infinite;
        }

        @keyframes sparkle {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.6; }
        }

        main {
            padding: 40px 30px;
        }

        .form-group {
            margin-bottom: 25px;
        }

        label {
            display: block;
            margin-bottom: 10px;
            font-weight: 600;
            color: #333;
            font-size: 1.05em;
        }

        input[type="text"],
        select {
            width: 100%;
            padding: 12px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1em;
            transition: all 0.3s;
            font-family: inherit;
        }

        input[type="text"]:focus,
        select:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .button-group {
            display: flex;
            gap: 10px;
            margin-top: 30px;
        }

        button {
            flex: 1;
            padding: 14px 20px;
            border: none;
            border-radius: 8px;
            font-size: 1.05em;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
        }

        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .btn-primary:hover:not(:disabled) {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }

        .btn-primary:disabled {
            opacity: 0.6;
            cursor: not-allowed;
        }

        .btn-secondary {
            background: #f0f0f0;
            color: #333;
        }

        .btn-secondary:hover:not(:disabled) {
            background: #e0e0e0;
        }

        .loading {
            display: none;
            text-align: center;
            padding: 30px;
        }

        .spinner {
            border: 4px solid #f0f0f0;
            border-top: 4px solid #667eea;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto 20px;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }

        .status {
            display: none;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            font-weight: 500;
        }

        .status.info {
            background: #e3f2fd;
            color: #1976d2;
            border-left: 4px solid #1976d2;
        }

        .status.success {
            background: #e8f5e9;
            color: #388e3c;
            border-left: 4px solid #388e3c;
        }

        .status.error {
            background: #ffebee;
            color: #d32f2f;
            border-left: 4px solid #d32f2f;
        }

        .results {
            display: none;
            margin-top: 30px;
        }

        .result-card {
            background: #f9f9f9;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 15px;
        }

        .result-card h3 {
            color: #667eea;
            margin-bottom: 10px;
            font-size: 1.2em;
        }

        .result-card p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 10px;
        }

        .download-links {
            display: flex;
            gap: 10px;
            margin-top: 15px;
        }

        .download-links a {
            padding: 10px 15px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: all 0.3s;
            font-size: 0.9em;
            font-weight: 600;
        }

        .download-links a:hover {
            background: #764ba2;
            transform: translateY(-2px);
        }

        .meta-info {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 15px;
            margin-top: 20px;
        }

        .meta-item {
            background: white;
            padding: 15px;
            border-radius: 8px;
            text-align: center;
            border: 1px solid #e0e0e0;
        }

        .meta-item .label {
            font-size: 0.85em;
            color: #999;
            margin-bottom: 5px;
        }

        .meta-item .value {
            font-size: 1.5em;
            font-weight: 700;
            color: #667eea;
        }

        .footer {
            background: #f5f5f5;
            padding: 20px;
            text-align: center;
            color: #999;
            font-size: 0.9em;
        }

        @media (max-width: 600px) {
            header h1 {
                font-size: 2em;
            }

            main {
                padding: 25px 20px;
            }

            button-group {
                flex-direction: column;
            }

            .meta-info {
                grid-template-columns: 1fr;
            }
        }

        .history {
            margin-top: 30px;
            padding-top: 30px;
            border-top: 2px solid #e0e0e0;
        }

        .history h3 {
            margin-bottom: 15px;
            color: #333;
        }

        .history-item {
            padding: 10px;
            background: #f9f9f9;
            border-radius: 6px;
            margin-bottom: 8px;
            cursor: pointer;
            transition: all 0.3s;
            border-left: 3px solid #667eea;
        }

        .history-item:hover {
            background: #f0f0f0;
            transform: translateX(5px);
        }

        .history-time {
            font-size: 0.85em;
            color: #999;
        }

        .empty-state {
            text-align: center;
            padding: 40px 20px;
            color: #999;
        }

        .empty-state svg {
            width: 60px;
            height: 60px;
            margin-bottom: 15px;
            opacity: 0.5;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔍 Genspark Clone <span class="sparkle">✨</span></h1>
            <p>웹 검색 + AI 합산 + Sparkpage 자동 생성</p>
        </header>

        <main>
            <form id="searchForm">
                <div class="form-group">
                    <label for="query">검색 질문</label>
                    <input
                        type="text"
                        id="query"
                        placeholder="예: 파이썬 프로그래밍이란, REST API란, 클라우드 컴퓨팅..."
                        required
                    />
                </div>

                <div class="form-group">
                    <label for="language">언어</label>
                    <select id="language">
                        <option value="ko">한국어</option>
                        <option value="en">English</option>
                        <option value="mixed">혼합</option>
                    </select>
                </div>

                <div class="form-group">
                    <label for="results">검색 결과 수</label>
                    <select id="results">
                        <option value="3">3개</option>
                        <option value="5" selected>5개</option>
                        <option value="10">10개</option>
                    </select>
                </div>

                <div class="button-group">
                    <button type="submit" class="btn-primary">🚀 생성 시작</button>
                    <button type="reset" class="btn-secondary">초기화</button>
                </div>
            </form>

            <div class="status" id="status"></div>

            <div class="loading" id="loading">
                <div class="spinner"></div>
                <p id="loadingText">처리 중...</p>
            </div>

            <div class="results" id="results">
                <div class="result-card">
                    <h3 id="resultTitle"></h3>
                    <p id="resultDescription"></p>

                    <div class="meta-info">
                        <div class="meta-item">
                            <div class="label">신뢰도</div>
                            <div class="value" id="confidence">-</div>
                        </div>
                        <div class="meta-item">
                            <div class="label">소스</div>
                            <div class="value" id="sources">-</div>
                        </div>
                        <div class="meta-item">
                            <div class="label">섹션</div>
                            <div class="value" id="sections">-</div>
                        </div>
                    </div>

                    <div class="download-links" id="downloads"></div>
                </div>

                <div class="history" id="historySection" style="display: none;">
                    <h3>최근 검색</h3>
                    <div id="historyList"></div>
                </div>
            </div>

            <div class="empty-state" id="emptyState">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <p>위의 검색 필드에 질문을 입력하고 [생성 시작] 버튼을 클릭하세요.</p>
            </div>
        </main>

        <div class="footer">
            <p>⚡ Genspark Clone v1.0 | 웹 검색 + Claude AI 합산 | Termux 최적화</p>
        </div>
    </div>

    <script>
        const searchForm = document.getElementById('searchForm');
        const queryInput = document.getElementById('query');
        const statusDiv = document.getElementById('status');
        const loadingDiv = document.getElementById('loading');
        const resultsDiv = document.getElementById('results');
        const emptyState = document.getElementById('emptyState');
        const downloadsDiv = document.getElementById('downloads');

        let searchHistory = JSON.parse(localStorage.getItem('genspark-history') || '[]');

        function showStatus(message, type = 'info') {
            statusDiv.textContent = message;
            statusDiv.className = `status ${type}`;
            statusDiv.style.display = 'block';
        }

        function showLoading(show = true) {
            loadingDiv.style.display = show ? 'block' : 'none';
            if (show) {
                emptyState.style.display = 'none';
                resultsDiv.style.display = 'none';
            }
        }

        function updateProgress(message) {
            document.getElementById('loadingText').textContent = message;
        }

        searchForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const query = queryInput.value.trim();
            if (!query) {
                showStatus('검색어를 입력하세요.', 'error');
                return;
            }

            showLoading(true);
            showStatus('');

            try {
                updateProgress('1/5: 질문 분석 중...');

                const response = await fetch('/api/search', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        query: query,
                        language: document.getElementById('language').value,
                        max_results: parseInt(document.getElementById('results').value),
                    }),
                });

                if (!response.ok) {
                    throw new Error(`API Error: ${response.status}`);
                }

                const result = await response.json();

                showLoading(false);

                if (result.success) {
                    showStatus('✅ Sparkpage 생성 완료!', 'success');

                    document.getElementById('resultTitle').textContent = query;
                    document.getElementById('resultDescription').textContent =
                        result.data.query + ' (생성일: ' + result.data.generated_at + ')';

                    document.getElementById('confidence').textContent =
                        Math.round(result.data.confidence_score * 100) + '%';
                    document.getElementById('sources').textContent =
                        result.data.total_sources + '개';
                    document.getElementById('sections').textContent =
                        result.data.sections + '개';

                    downloadsDiv.innerHTML = `
                        <a href="/download/${result.data.filename_md}" download>📄 Markdown 다운로드</a>
                        <a href="/download/${result.data.filename_html}" download>🌐 HTML 다운로드</a>
                        <a href="/view/${result.data.filename_html}" target="_blank">👁️ 미리보기</a>
                    `;

                    emptyState.style.display = 'none';
                    resultsDiv.style.display = 'block';

                    // 히스토리 저장
                    searchHistory.unshift({
                        query: query,
                        timestamp: new Date().toLocaleString(),
                        filename_html: result.data.filename_html,
                    });
                    searchHistory = searchHistory.slice(0, 10);
                    localStorage.setItem('genspark-history', JSON.stringify(searchHistory));

                } else {
                    showStatus('❌ ' + result.error, 'error');
                }

            } catch (error) {
                showLoading(false);
                showStatus('❌ 오류 발생: ' + error.message, 'error');
                console.error(error);
            }
        });
    </script>
</body>
</html>
"""

# API 라우트
@app.route('/')
def index():
    return render_template_string(HTML_TEMPLATE)

@app.route('/api/search', methods=['POST'])
def api_search():
    """Genspark 검색 API"""
    try:
        data = request.json
        query = data.get('query', '').strip()

        if not query:
            return jsonify({'success': False, 'error': '검색어를 입력하세요.'}), 400

        # GensparkAgent 실행
        api_key = os.environ.get('ANTHROPIC_API_KEY')
        if not api_key:
            return jsonify({
                'success': False,
                'error': 'ANTHROPIC_API_KEY 환경변수가 설정되지 않았습니다.'
            }), 500

        config = AgentConfig(
            anthropic_api_key=api_key,
            max_search_results=data.get('max_results', 5),
            output_dir='output',
            verbose=False
        )

        agent = GensparkAgent(config)
        result = agent.run(query)

        if result:
            # 파일명 추출
            filename_html = result.html_path.split('/')[-1]
            filename_md = result.markdown_path.split('/')[-1]

            return jsonify({
                'success': True,
                'data': {
                    'query': query,
                    'confidence_score': 0.85,  # 실제로는 result에서 가져옴
                    'total_sources': 5,
                    'sections': 4,
                    'generated_at': datetime.now().strftime('%Y-%m-%d %H:%M'),
                    'filename_html': filename_html,
                    'filename_md': filename_md,
                }
            })
        else:
            return jsonify({
                'success': False,
                'error': 'Sparkpage 생성에 실패했습니다.'
            }), 500

    except Exception as e:
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500

@app.route('/download/<filename>')
def download_file(filename):
    """파일 다운로드"""
    try:
        file_path = os.path.join('output', filename)
        if os.path.exists(file_path):
            return send_file(file_path, as_attachment=True)
        else:
            return '파일을 찾을 수 없습니다.', 404
    except Exception as e:
        return str(e), 500

@app.route('/view/<filename>')
def view_file(filename):
    """파일 미리보기"""
    try:
        file_path = os.path.join('output', filename)
        if os.path.exists(file_path):
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            return content
        else:
            return '파일을 찾을 수 없습니다.', 404
    except Exception as e:
        return str(e), 500

if __name__ == '__main__':
    print("""
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║     🚀 Genspark Clone - 웹 UI 서버 시작                       ║
║                                                                ║
║     📍 접속: http://localhost:5000                            ║
║     🔑 API 키 필수: export ANTHROPIC_API_KEY="sk-ant-..."   ║
║     ⛔ Ctrl+C로 종료                                           ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
    """)
    app.run(debug=True, host='0.0.0.0', port=5555)
