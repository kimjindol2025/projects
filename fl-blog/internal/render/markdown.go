package render

import (
	htmlutl "html"
	"regexp"
	"strings"
)

// ToHTML converts markdown to HTML
func ToHTML(markdown string) string {
	// Process in order: code blocks, then inline elements, then blocks
	result := markdown

	// 1. Process code blocks (```...```)
	codeBlockRegex := regexp.MustCompile("(?s)```([a-z]*)\n(.*?)```")
	result = codeBlockRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract language and code
		parts := strings.SplitN(match, "\n", 2)
		lang := strings.TrimPrefix(parts[0], "```")
		code := strings.TrimSuffix(parts[1], "```")
		code = strings.TrimSpace(code)

		// Escape HTML
		code = htmlutl.EscapeString(code)

		return `<pre><code class="language-` + lang + `">` + code + `</code></pre>`
	})

	// 2. Process inline code (single backticks)
	inlineCodeRegex := regexp.MustCompile("`([^`\n]+)`")
	result = inlineCodeRegex.ReplaceAllString(result, `<code>$1</code>`)

	// 3. Process bold (**text**)
	boldRegex := regexp.MustCompile(`\*\*([^\*\n]+)\*\*`)
	result = boldRegex.ReplaceAllString(result, `<strong>$1</strong>`)

	// 4. Process italic (*text*)
	italicRegex := regexp.MustCompile(`\*([^\*\n]+)\*`)
	result = italicRegex.ReplaceAllString(result, `<em>$1</em>`)

	// 5. Process headings
	// H1: # text
	h1Regex := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	result = h1Regex.ReplaceAllString(result, `<h1>$1</h1>`)

	// H2: ## text
	h2Regex := regexp.MustCompile(`(?m)^##\s+(.+)$`)
	result = h2Regex.ReplaceAllString(result, `<h2>$1</h2>`)

	// H3: ### text
	h3Regex := regexp.MustCompile(`(?m)^###\s+(.+)$`)
	result = h3Regex.ReplaceAllString(result, `<h3>$1</h3>`)

	// H4: #### text
	h4Regex := regexp.MustCompile(`(?m)^####\s+(.+)$`)
	result = h4Regex.ReplaceAllString(result, `<h4>$1</h4>`)

	// 6. Process lists (simple - items)
	// Split into lines and process
	lines := strings.Split(result, "\n")
	var processedLines []string
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if line is a list item
		if strings.HasPrefix(trimmed, "- ") {
			if !inList {
				processedLines = append(processedLines, "<ul>")
				inList = true
			}
			item := strings.TrimPrefix(trimmed, "- ")
			processedLines = append(processedLines, "<li>"+item+"</li>")
		} else if strings.HasPrefix(trimmed, "* ") {
			if !inList {
				processedLines = append(processedLines, "<ul>")
				inList = true
			}
			item := strings.TrimPrefix(trimmed, "* ")
			processedLines = append(processedLines, "<li>"+item+"</li>")
		} else {
			if inList {
				processedLines = append(processedLines, "</ul>")
				inList = false
			}

			// Don't process empty lines or code blocks as paragraphs
			if strings.HasPrefix(trimmed, "<pre>") || strings.HasPrefix(trimmed, "<h") || trimmed == "" {
				processedLines = append(processedLines, line)
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "<") {
				// Wrap non-empty lines in paragraphs if not already wrapped
				if !strings.HasPrefix(trimmed, "<") {
					processedLines = append(processedLines, "<p>"+line+"</p>")
				} else {
					processedLines = append(processedLines, line)
				}
			} else {
				processedLines = append(processedLines, line)
			}
		}
	}

	if inList {
		processedLines = append(processedLines, "</ul>")
	}

	result = strings.Join(processedLines, "\n")

	// 7. Process links [text](url)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	result = linkRegex.ReplaceAllString(result, `<a href="$2">$1</a>`)

	return result
}
