package tmpl

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"fl-blog/internal/post"
	"fl-blog/internal/render"
)

// RenderIndex renders the homepage with all posts
func RenderIndex(posts []*post.Post) (string, error) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>FreeLang Blog</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      padding: 40px 20px;
      color: #333;
    }

    .container {
      max-width: 1200px;
      margin: 0 auto;
    }

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

    .header p {
      font-size: 1.2em;
      opacity: 0.9;
    }

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

    .post-date {
      color: #667eea;
      font-size: 0.9em;
      margin-bottom: 8px;
      font-weight: 600;
    }

    .post-title {
      font-size: 1.5em;
      margin-bottom: 12px;
      color: #333;
    }

    .post-excerpt {
      color: #666;
      line-height: 1.6;
      font-size: 0.95em;
    }

    .footer {
      text-align: center;
      color: white;
      margin-top: 60px;
      padding-top: 30px;
      border-top: 1px solid rgba(255,255,255,0.2);
    }

    .footer p {
      opacity: 0.8;
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
`

	for _, p := range posts {
		formattedDate := p.Date.Format("Jan 2, 2006")
		cardHTML := fmt.Sprintf(`      <a href="/post/%s" class="post-card">
        <div class="post-date">%s</div>
        <h3 class="post-title">%s</h3>
        <p class="post-excerpt">%s</p>
      </a>
`, p.Slug, formattedDate, p.Title, p.Excerpt)
		htmlContent += cardHTML
	}

	htmlContent += `    </div>

    <div class="footer">
      <p>Made with ❤️ | <a href="/health" style="color: white;">Health Check</a> | <a href="/api/posts" style="color: white;">JSON API</a></p>
    </div>
  </div>
</body>
</html>`

	return htmlContent, nil
}

// RenderPost renders a single post page
func RenderPost(p *post.Post) (string, error) {
	contentHTML := render.ToHTML(p.Content)

	htmlContent := `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + template.HTMLEscapeString(p.Title) + ` - FreeLang Blog</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css">
  <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

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

    .post-date {
      color: #666;
      font-size: 0.95em;
      margin-bottom: 10px;
    }

    .post-title {
      font-size: 2.5em;
      margin-bottom: 10px;
      color: #333;
    }

    .post-content {
      font-size: 1.05em;
      line-height: 1.8;
    }

    .post-content h1 {
      font-size: 2em;
      margin-top: 40px;
      margin-bottom: 20px;
      color: #333;
    }

    .post-content h2 {
      font-size: 1.6em;
      margin-top: 35px;
      margin-bottom: 15px;
      color: #444;
    }

    .post-content h3 {
      font-size: 1.3em;
      margin-top: 25px;
      margin-bottom: 12px;
      color: #555;
    }

    .post-content h4 {
      font-size: 1.1em;
      margin-top: 20px;
      margin-bottom: 10px;
      color: #666;
    }

    .post-content p {
      margin-bottom: 15px;
    }

    .post-content ul, .post-content ol {
      margin-left: 30px;
      margin-bottom: 15px;
    }

    .post-content li {
      margin-bottom: 8px;
    }

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

    .post-content pre code {
      background: none;
      padding: 0;
      color: #abb2bf;
      font-size: 0.9em;
    }

    .post-content a {
      color: #667eea;
      text-decoration: none;
      border-bottom: 1px solid #667eea;
    }

    .post-content a:hover {
      background: #f0f0f0;
    }

    .nav {
      margin-top: 50px;
      padding-top: 30px;
      border-top: 2px solid #eee;
    }

    .nav a {
      display: inline-block;
      margin-right: 20px;
      color: #667eea;
      text-decoration: none;
    }

    .nav a:hover {
      text-decoration: underline;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="post-header">
      <div class="post-date">` + p.Date.Format("January 2, 2006") + `</div>
      <h1 class="post-title">` + template.HTMLEscapeString(p.Title) + `</h1>
    </div>

    <div class="post-content">
` + contentHTML + `
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
</html>`

	return htmlContent, nil
}

// RenderJSON renders posts as JSON for API
func RenderJSON(posts []*post.Post) string {
	sb := strings.Builder{}
	sb.WriteString("[")

	for i, p := range posts {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf(`{"title":"%s","slug":"%s","date":"%s","excerpt":"%s"}`,
			template.HTMLEscapeString(p.Title),
			p.Slug,
			p.Date.Format(time.RFC3339),
			template.HTMLEscapeString(p.Excerpt)))
	}

	sb.WriteString("]")
	return sb.String()
}
