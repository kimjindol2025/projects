#!/usr/bin/env node

/**
 * FreeLang Blog with Comments System
 * 포트 8253 - Node.js 백엔드 서버
 *
 * 기능:
 * - 포스트 로드 & 마크다운 변환
 * - 댓글 CRUD API
 * - 댓글 검색
 * - 파일 기반 저장소 (JSON)
 */

const http = require('http');
const fs = require('fs');
const path = require('path');
const url = require('url');

const PORT = parseInt((process.argv[2] || '8253').replace(':', ''));
const POSTS_DIR = path.join(__dirname, 'posts');
const COMMENTS_FILE = path.join(__dirname, 'comments.json');

let posts = [];
let comments = [];

/**
 * 포스트 로드
 */
function loadPosts() {
  try {
    const files = fs.readdirSync(POSTS_DIR);
    posts = files
      .filter(f => f.endsWith('.md'))
      .map(filename => {
        const filePath = path.join(POSTS_DIR, filename);
        const content = fs.readFileSync(filePath, 'utf8');

        let title = filename.replace(/^\d{4}-\d{2}-\d{2}-/, '').replace(/\.md$/, '');
        let date = new Date();

        const fmMatch = content.match(/^---\n([\s\S]*?)\n---/);
        if (fmMatch) {
          const fm = fmMatch[1];
          const titleMatch = fm.match(/title:\s*(.+)/);
          if (titleMatch) title = titleMatch[1].trim().replace(/["']/g, '');

          const dateMatch = fm.match(/date:\s*(\d{4}-\d{2}-\d{2})/);
          if (dateMatch) date = new Date(dateMatch[1]);
        }

        const slug = filename
          .replace(/^\d{4}-\d{2}-\d{2}-/, '')
          .replace(/\.md$/, '')
          .toLowerCase();

        const bodyMatch = content.match(/^---\n[\s\S]*?\n---([\s\S]*)/);
        const body = bodyMatch ? bodyMatch[1] : content;
        let excerpt = body.replace(/[#*`\n]/g, ' ').trim();
        const words = excerpt.split(/\s+/);
        if (words.length > 20) excerpt = words.slice(0, 20).join(' ') + '...';

        return { title, slug, date, excerpt, content: body, filename };
      })
      .sort((a, b) => b.date - a.date);

    console.log(`✅ Loaded ${posts.length} posts`);
  } catch (e) {
    console.error('❌ Error loading posts:', e.message);
    posts = [];
  }
}

/**
 * 댓글 로드
 */
function loadComments() {
  try {
    if (fs.existsSync(COMMENTS_FILE)) {
      const data = fs.readFileSync(COMMENTS_FILE, 'utf8');
      comments = JSON.parse(data);
      console.log(`✅ Loaded ${comments.length} comments`);
    } else {
      comments = [];
      saveComments();
    }
  } catch (e) {
    console.error('❌ Error loading comments:', e.message);
    comments = [];
  }
}

/**
 * 댓글 저장
 */
function saveComments() {
  try {
    fs.writeFileSync(COMMENTS_FILE, JSON.stringify(comments, null, 2));
  } catch (e) {
    console.error('❌ Error saving comments:', e.message);
  }
}

/**
 * Markdown to HTML
 */
function markdownToHtml(md) {
  let html = md;

  // Code blocks
  html = html.replace(/```([a-z]*)\n([\s\S]*?)```/g, (m, lang, code) => {
    code = code.trim().replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return `<pre><code class="language-${lang}">${code}</code></pre>`;
  });

  // Inline code
  html = html.replace(/`([^`\n]+)`/g, '<code>$1</code>');

  // Bold, italic, links
  html = html.replace(/\*\*([^\*\n]+)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*([^\*\n]+)\*/g, '<em>$1</em>');
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');

  // Headings
  html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');
  html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
  html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');

  return html;
}

/**
 * HTTP 서버
 */
const server = http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url, true);
  const pathname = parsedUrl.pathname;
  const query = parsedUrl.query;

  // CORS 헤더
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  // ========== API 엔드포인트 ==========

  // GET /health
  if (pathname === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      status: 'ok',
      posts: posts.length,
      comments: comments.length,
      time: new Date().toISOString()
    }));
    return;
  }

  // GET /api/posts
  if (pathname === '/api/posts') {
    const postData = posts.map(p => ({
      title: p.title,
      slug: p.slug,
      date: p.date.toISOString(),
      excerpt: p.excerpt,
      commentCount: comments.filter(c => c.postSlug === p.slug).length
    }));
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(postData));
    return;
  }

  // GET /api/comments?postId={slug}
  if (pathname === '/api/comments' && req.method === 'GET') {
    const postSlug = query.postId;
    const postComments = postSlug
      ? comments.filter(c => c.postSlug === postSlug).sort((a, b) => new Date(b.date) - new Date(a.date))
      : comments.sort((a, b) => new Date(b.date) - new Date(a.date));

    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(postComments));
    return;
  }

  // POST /api/comments - 댓글 작성
  if (pathname === '/api/comments' && req.method === 'POST') {
    let body = '';
    req.on('data', chunk => { body += chunk; });
    req.on('end', () => {
      try {
        const data = JSON.parse(body);

        if (!data.postSlug || !data.author || !data.text) {
          res.writeHead(400, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify({ error: 'Missing required fields' }));
          return;
        }

        const comment = {
          id: Date.now().toString(),
          postSlug: data.postSlug,
          author: data.author.substring(0, 50),
          email: data.email ? data.email.substring(0, 100) : '',
          text: data.text.substring(0, 1000),
          date: new Date().toISOString(),
          approved: true
        };

        comments.push(comment);
        saveComments();

        res.writeHead(201, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(comment));
      } catch (e) {
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: e.message }));
      }
    });
    return;
  }

  // DELETE /api/comments/{id} - 댓글 삭제
  if (pathname.startsWith('/api/comments/') && req.method === 'DELETE') {
    const id = pathname.split('/').pop();
    const idx = comments.findIndex(c => c.id === id);

    if (idx === -1) {
      res.writeHead(404, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Comment not found' }));
      return;
    }

    comments.splice(idx, 1);
    saveComments();

    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ success: true }));
    return;
  }

  // GET /api/search?q={query} - 댓글 검색
  if (pathname === '/api/search') {
    const query_str = (query.q || '').toLowerCase();
    if (!query_str) {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify([]));
      return;
    }

    const results = comments.filter(c =>
      c.text.toLowerCase().includes(query_str) ||
      c.author.toLowerCase().includes(query_str)
    );

    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(results));
    return;
  }

  // GET /post/{slug} - 포스트 상세 (+ 댓글)
  if (pathname.startsWith('/post/')) {
    const slug = pathname.slice(6);
    const post = posts.find(p => p.slug === slug);

    if (!post) {
      res.writeHead(404, { 'Content-Type': 'text/html' });
      res.end('<h1>404 - Post Not Found</h1>');
      return;
    }

    const contentHtml = markdownToHtml(post.content);
    const postComments = comments.filter(c => c.postSlug === slug).sort((a, b) => new Date(b.date) - new Date(a.date));

    const commentsHtml = postComments.map(c => `
      <div style="border-left: 3px solid #667eea; padding: 15px; margin: 15px 0; background: #f9f9f9;">
        <strong>${c.author}</strong> · <small>${new Date(c.date).toLocaleDateString('ko-KR')}</small>
        <p style="margin-top: 10px; line-height: 1.6;">${c.text}</p>
        <small><a href="javascript:deleteComment('${c.id}')" style="color: #ff6b6b; cursor: pointer;">삭제</a></small>
      </div>
    `).join('');

    const html = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${post.title} - FreeLang Blog</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css">
  <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"><\/script>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
      background: #f5f5f5;
      color: #333;
      line-height: 1.7;
    }
    .container {
      max-width: 800px;
      margin: 0 auto;
      padding: 40px 20px;
      background: white;
      min-height: 100vh;
    }
    .post-header {
      margin-bottom: 40px;
      border-bottom: 2px solid #eee;
      padding-bottom: 30px;
    }
    .post-date { color: #666; font-size: 0.95em; margin-bottom: 10px; }
    .post-title { font-size: 2.5em; color: #333; }
    .post-content { margin: 40px 0; }
    .post-content h1 { font-size: 2em; margin-top: 40px; margin-bottom: 20px; }
    .post-content h2 { font-size: 1.6em; margin-top: 35px; margin-bottom: 15px; }
    .post-content p { margin-bottom: 15px; }
    .post-content code {
      background: #f0f0f0;
      padding: 2px 6px;
      border-radius: 3px;
      font-family: 'Courier New', monospace;
    }
    .post-content pre {
      background: #282c34;
      padding: 15px;
      border-radius: 5px;
      overflow-x: auto;
      margin: 15px 0;
    }
    .post-content pre code { background: none; color: #abb2bf; }
    .post-content a { color: #667eea; text-decoration: none; border-bottom: 1px solid #667eea; }
    .comments-section {
      margin-top: 60px;
      padding-top: 30px;
      border-top: 2px solid #eee;
    }
    .comments-title { font-size: 1.5em; margin-bottom: 20px; color: #333; }
    .comment-form {
      background: #f9f9f9;
      padding: 20px;
      border-radius: 8px;
      margin-bottom: 30px;
    }
    .form-group {
      margin-bottom: 15px;
    }
    .form-group label {
      display: block;
      margin-bottom: 5px;
      font-weight: 600;
      color: #333;
    }
    .form-group input, .form-group textarea {
      width: 100%;
      padding: 10px;
      border: 1px solid #ddd;
      border-radius: 4px;
      font-family: Arial, sans-serif;
    }
    .form-group textarea {
      resize: vertical;
      min-height: 100px;
    }
    .form-group button {
      background: #667eea;
      color: white;
      border: none;
      padding: 10px 20px;
      border-radius: 4px;
      cursor: pointer;
      font-weight: 600;
    }
    .form-group button:hover {
      background: #5568d3;
    }
    .comments-list { margin-top: 30px; }
    .comment {
      border-left: 3px solid #667eea;
      padding: 15px;
      margin: 15px 0;
      background: #f9f9f9;
    }
    .comment-author { font-weight: 600; }
    .comment-date { color: #999; font-size: 0.9em; }
    .comment-text { margin: 10px 0; line-height: 1.6; }
    .comment-actions { font-size: 0.9em; }
    .comment-actions a { color: #ff6b6b; cursor: pointer; }
    .nav { margin-top: 40px; }
    .nav a { color: #667eea; text-decoration: none; }
  </style>
</head>
<body>
  <div class="container">
    <div class="post-header">
      <div class="post-date">${new Date(post.date).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}</div>
      <h1 class="post-title">${post.title}</h1>
    </div>

    <div class="post-content">
      ${contentHtml}
    </div>

    <div class="comments-section">
      <h2 class="comments-title">💬 댓글 (${postComments.length})</h2>

      <div class="comment-form">
        <h3 style="margin-bottom: 15px;">댓글 작성</h3>
        <div class="form-group">
          <label>이름</label>
          <input type="text" id="author" placeholder="이름" maxlength="50">
        </div>
        <div class="form-group">
          <label>이메일 (선택)</label>
          <input type="email" id="email" placeholder="이메일" maxlength="100">
        </div>
        <div class="form-group">
          <label>댓글</label>
          <textarea id="text" placeholder="댓글을 입력하세요" maxlength="1000"></textarea>
        </div>
        <div class="form-group">
          <button onclick="submitComment('${slug}')">댓글 작성</button>
        </div>
      </div>

      <div class="comments-list">
        ${postComments.length > 0
          ? postComments.map(c => `
            <div class="comment">
              <span class="comment-author">${c.author}</span>
              <span class="comment-date"> · ${new Date(c.date).toLocaleDateString('ko-KR')}</span>
              <div class="comment-text">${c.text}</div>
              <div class="comment-actions">
                <a onclick="deleteComment('${c.id}')">삭제</a>
              </div>
            </div>
          `).join('')
          : '<p style="color: #999;">아직 댓글이 없습니다.</p>'
        }
      </div>
    </div>

    <div class="nav">
      <a href="/">← 블로그로 돌아가기</a>
    </div>
  </div>

  <script>
    function submitComment(slug) {
      const author = document.getElementById('author').value;
      const email = document.getElementById('email').value;
      const text = document.getElementById('text').value;

      if (!author || !text) {
        alert('이름과 댓글을 입력해주세요.');
        return;
      }

      fetch('http://localhost:8253/api/comments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ postSlug: slug, author, email, text })
      })
      .then(r => r.json())
      .then(data => {
        alert('댓글이 작성되었습니다.');
        location.reload();
      })
      .catch(e => alert('오류: ' + e.message));
    }

    function deleteComment(id) {
      if (!confirm('댓글을 삭제하시겠습니까?')) return;

      fetch('http://localhost:8253/api/comments/' + id, { method: 'DELETE' })
      .then(r => r.json())
      .then(data => {
        alert('댓글이 삭제되었습니다.');
        location.reload();
      })
      .catch(e => alert('오류: ' + e.message));
    }

    document.querySelectorAll('pre code').forEach(b => hljs.highlightElement(b));
  </script>
</body>
</html>`;

    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(html);
    return;
  }

  // GET / - 홈페이지
  if (pathname === '/') {
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(require('fs').readFileSync(path.join(__dirname, 'index.html'), 'utf8'));
    return;
  }

  // 404
  res.writeHead(404, { 'Content-Type': 'text/html' });
  res.end('<h1>404 - Not Found</h1>');
});

loadPosts();
loadComments();

server.listen(PORT, '0.0.0.0', () => {
  console.log(`✅ Server listening on http://0.0.0.0:${PORT} (모든 인터페이스)`);
  console.log(`   Home:     http://localhost:${PORT}/`);
  console.log(`   API:      http://localhost:${PORT}/api/posts`);
  console.log(`   Health:   http://localhost:${PORT}/health`);
  console.log(`   Comments: http://localhost:${PORT}/api/comments`);
  console.log(`   Search:   http://localhost:${PORT}/api/search?q=keyword`);
  console.log(`   🌐 외부: https://freelang-blog.dclub.kr`);
});

process.on('SIGINT', () => {
  console.log('\n🛑 Shutting down...');
  saveComments();
  process.exit(0);
});
