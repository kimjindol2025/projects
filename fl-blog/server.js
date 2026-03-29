#!/usr/bin/env node

const http = require('http');
const fs = require('fs');
const path = require('path');
const url = require('url');

// Get port from command line or default to 8253
let PORT = process.argv[2] || '8253';
// Remove leading colon if present
if (PORT.startsWith(':')) {
  PORT = PORT.slice(1);
}
PORT = parseInt(PORT);

// Get posts directory
const POSTS_DIR = path.join(__dirname, 'posts');

// Load all posts
let posts = [];

function loadPosts() {
  try {
    const files = fs.readdirSync(POSTS_DIR);
    posts = files
      .filter(f => f.endsWith('.md'))
      .map(filename => {
        const filePath = path.join(POSTS_DIR, filename);
        const content = fs.readFileSync(filePath, 'utf8');

        // Parse frontmatter
        let title = filename.replace(/^\d{4}-\d{2}-\d{2}-/, '').replace(/\.md$/, '');
        let date = new Date();

        const fmMatch = content.match(/^---\n([\s\S]*?)\n---/);
        if (fmMatch) {
          const fm = fmMatch[1];

          const titleMatch = fm.match(/title:\s*(.+)/);
          if (titleMatch) {
            title = titleMatch[1].trim().replace(/["']/g, '');
          }

          const dateMatch = fm.match(/date:\s*(\d{4}-\d{2}-\d{2})/);
          if (dateMatch) {
            date = new Date(dateMatch[1]);
          }
        }

        // Extract slug
        const slug = filename
          .replace(/^\d{4}-\d{2}-\d{2}-/, '')
          .replace(/\.md$/, '')
          .toLowerCase();

        // Extract excerpt
        const bodyMatch = content.match(/^---\n[\s\S]*?\n---([\s\S]*)/);
        const body = bodyMatch ? bodyMatch[1] : content;
        let excerpt = body.replace(/[#*`\n]/g, ' ').trim();
        const words = excerpt.split(/\s+/);
        if (words.length > 20) {
          excerpt = words.slice(0, 20).join(' ') + '...';
        }

        return {
          title,
          slug,
          date,
          excerpt,
          content: body,
          filename
        };
      })
      .sort((a, b) => b.date - a.date);

    console.log(`✅ Loaded ${posts.length} posts from ${POSTS_DIR}`);
  } catch (e) {
    console.error('❌ Error loading posts:', e.message);
    posts = [];
  }
}

// Markdown to HTML converter
function markdownToHtml(md) {
  let html = md;

  // Code blocks
  html = html.replace(/```([a-z]*)\n([\s\S]*?)```/g, (m, lang, code) => {
    code = code.trim().replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return `<pre><code class="language-${lang}">${code}</code></pre>`;
  });

  // Inline code
  html = html.replace(/`([^`\n]+)`/g, '<code>$1</code>');

  // Bold
  html = html.replace(/\*\*([^\*\n]+)\*\*/g, '<strong>$1</strong>');

  // Italic
  html = html.replace(/\*([^\*\n]+)\*/g, '<em>$1</em>');

  // Headings
  html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');
  html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
  html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
  html = html.replace(/^#### (.+)$/gm, '<h4>$1</h4>');

  // Links
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');

  // Lists
  const lines = html.split('\n');
  let inList = false;
  const processed = [];

  for (const line of lines) {
    if (line.match(/^- /)) {
      if (!inList) {
        processed.push('<ul>');
        inList = true;
      }
      processed.push('<li>' + line.replace(/^- /, '') + '</li>');
    } else if (inList) {
      processed.push('</ul>');
      inList = false;
      if (line.trim()) {
        processed.push('<p>' + line + '</p>');
      }
    } else if (line.trim()) {
      processed.push('<p>' + line + '</p>');
    }
  }

  if (inList) {
    processed.push('</ul>');
  }

  return processed.join('\n');
}

// HTML Templates
function renderIndex() {
  let cardsHtml = '';

  for (const post of posts) {
    const dateStr = post.date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });

    cardsHtml += `
      <a href="/post/${post.slug}" class="post-card">
        <div class="post-date">${dateStr}</div>
        <h3 class="post-title">${post.title}</h3>
        <p class="post-excerpt">${post.excerpt}</p>
      </a>`;
  }

  return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>FreeLang Blog</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      padding: 40px 20px;
      color: #333;
    }
    .container { max-width: 1200px; margin: 0 auto; }
    .header {
      text-align: center;
      color: white;
      margin-bottom: 50px;
    }
    .header h1 {
      font-size: 3em;
      margin-bottom: 10px;
      text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
    }
    .header p { font-size: 1.2em; opacity: 0.9; }
    .posts-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
      gap: 30px;
      margin-top: 40px;
    }
    .post-card {
      background: white;
      border-radius: 8px;
      padding: 25px;
      box-shadow: 0 4px 6px rgba(0,0,0,0.1);
      transition: transform 0.3s, box-shadow 0.3s;
      text-decoration: none;
      color: inherit;
      display: block;
    }
    .post-card:hover {
      transform: translateY(-5px);
      box-shadow: 0 8px 12px rgba(0,0,0,0.15);
    }
    .post-date { color: #667eea; font-size: 0.9em; margin-bottom: 8px; font-weight: 600; }
    .post-title { font-size: 1.5em; margin-bottom: 12px; color: #333; }
    .post-excerpt { color: #666; line-height: 1.6; font-size: 0.95em; }
    .footer {
      text-align: center;
      color: white;
      margin-top: 60px;
      padding-top: 30px;
      border-top: 1px solid rgba(255,255,255,0.2);
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>🚀 FreeLang Blog</h1>
      <p>Technical posts on compilers, databases, and system design</p>
    </div>
    <div class="posts-grid">
      ${cardsHtml}
    </div>
    <div class="footer">
      <p>Made with ❤️ | <a href="/health" style="color: white;">Health</a> | <a href="/api/posts" style="color: white;">API</a></p>
    </div>
  </div>
</body>
</html>`;
}

function renderPost(post) {
  const contentHtml = markdownToHtml(post.content);
  const dateStr = post.date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });

  return `<!DOCTYPE html>
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
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
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
    .post-title { font-size: 2.5em; margin-bottom: 10px; color: #333; }
    .post-content { font-size: 1.05em; line-height: 1.8; }
    .post-content h1 { font-size: 2em; margin-top: 40px; margin-bottom: 20px; color: #333; }
    .post-content h2 { font-size: 1.6em; margin-top: 35px; margin-bottom: 15px; color: #444; }
    .post-content h3 { font-size: 1.3em; margin-top: 25px; margin-bottom: 12px; color: #555; }
    .post-content h4 { font-size: 1.1em; margin-top: 20px; margin-bottom: 10px; color: #666; }
    .post-content p { margin-bottom: 15px; }
    .post-content ul { margin-left: 30px; margin-bottom: 15px; }
    .post-content li { margin-bottom: 8px; }
    .post-content code {
      background: #f0f0f0;
      padding: 2px 6px;
      border-radius: 3px;
      font-family: 'Courier New', monospace;
      font-size: 0.95em;
    }
    .post-content pre {
      background: #282c34;
      padding: 15px;
      border-radius: 5px;
      overflow-x: auto;
      margin-bottom: 15px;
    }
    .post-content pre code { background: none; padding: 0; color: #abb2bf; font-size: 0.9em; }
    .post-content a { color: #667eea; text-decoration: none; border-bottom: 1px solid #667eea; }
    .post-content a:hover { background: #f0f0f0; }
    .nav { margin-top: 50px; padding-top: 30px; border-top: 2px solid #eee; }
    .nav a { display: inline-block; margin-right: 20px; color: #667eea; text-decoration: none; }
    .nav a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <div class="container">
    <div class="post-header">
      <div class="post-date">${dateStr}</div>
      <h1 class="post-title">${post.title}</h1>
    </div>
    <div class="post-content">
      ${contentHtml}
    </div>
    <div class="nav">
      <a href="/">← Back to posts</a>
    </div>
  </div>
  <script>
    document.querySelectorAll('pre code').forEach(block => {
      hljs.highlightElement(block);
    });
  </script>
</body>
</html>`;
}

// Create HTTP server
const server = http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url, true);
  const pathname = parsedUrl.pathname;

  // CORS headers for GitHub Pages integration
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

  // Handle OPTIONS requests
  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  // Health check
  if (pathname === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      status: 'ok',
      posts: posts.length,
      time: new Date().toISOString()
    }));
    return;
  }

  // Home page
  if (pathname === '/') {
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(renderIndex());
    return;
  }

  // API posts
  if (pathname === '/api/posts') {
    const postData = posts.map(p => ({
      title: p.title,
      slug: p.slug,
      date: p.date.toISOString(),
      excerpt: p.excerpt
    }));
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(postData));
    return;
  }

  // Individual post
  if (pathname.startsWith('/post/')) {
    const slug = pathname.slice(6);
    const post = posts.find(p => p.slug === slug);

    if (post) {
      res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
      res.end(renderPost(post));
      return;
    } else {
      res.writeHead(404, { 'Content-Type': 'text/html' });
      res.end('<h1>404 - Post Not Found</h1><p><a href="/">Back to posts</a></p>');
      return;
    }
  }

  // 404
  res.writeHead(404, { 'Content-Type': 'text/html' });
  res.end('<h1>404 - Not Found</h1>');
});

// Load posts and start server
loadPosts();

server.listen(PORT, () => {
  console.log(`✅ Server listening on http://localhost:${PORT}`);
  console.log(`   Home:   http://localhost:${PORT}/`);
  console.log(`   API:    http://localhost:${PORT}/api/posts`);
  console.log(`   Health: http://localhost:${PORT}/health`);
});

// Graceful shutdown
process.on('SIGINT', () => {
  console.log('\n🛑 Shutting down...');
  server.close(() => {
    process.exit(0);
  });
});
