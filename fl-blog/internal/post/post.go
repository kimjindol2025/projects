package post

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Post represents a blog post
type Post struct {
	Title    string
	Date     time.Time
	Slug     string
	Excerpt  string
	Content  string
	FullPath string
	Filename string
}

// LoadPosts loads all markdown files from the posts directory
func LoadPosts(postsDir string) ([]*Post, error) {
	var posts []*Post

	// Read all files from the directory
	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read posts directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(postsDir, entry.Name())
		post, err := parsePost(filePath, entry.Name())
		if err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		posts = append(posts, post)
	}

	// Sort by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

// parsePost parses a single markdown file
func parsePost(filePath, filename string) (*Post, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(content)

	// Parse YAML frontmatter
	title := ""
	date := time.Now()

	// Extract frontmatter (between --- markers)
	if strings.HasPrefix(text, "---") {
		parts := strings.SplitN(text, "---", 3)
		if len(parts) >= 2 {
			fm := parts[1]
			// Extract title
			titleMatch := regexp.MustCompile(`title:\s*["']?([^"'\n]+)["']?`).FindStringSubmatch(fm)
			if len(titleMatch) > 1 {
				title = strings.TrimSpace(titleMatch[1])
			}

			// Extract date
			dateMatch := regexp.MustCompile(`date:\s*(\d{4}-\d{2}-\d{2})`).FindStringSubmatch(fm)
			if len(dateMatch) > 1 {
				parsedDate, err := time.Parse("2006-01-02", dateMatch[1])
				if err == nil {
					date = parsedDate
				}
			}

			// Body is after the second ---
			if len(parts) >= 3 {
				text = parts[2]
			}
		}
	}

	// If no title from frontmatter, use filename
	if title == "" {
		title = slugToTitle(extractSlug(filename))
	}

	// Create slug from filename
	slug := extractSlug(filename)

	// Extract excerpt (first 200 chars of body)
	excerptText := strings.TrimSpace(text)
	// Remove markdown syntax from excerpt
	excerptText = regexp.MustCompile("[#*`\\n]").ReplaceAllString(excerptText, " ")
	excerptWords := strings.Fields(excerptText)
	if len(excerptWords) > 20 {
		excerptWords = excerptWords[:20]
	}
	excerptStr := strings.Join(excerptWords, " ")
	if len(excerptStr) > 200 {
		excerptStr = excerptStr[:200] + "..."
	}

	return &Post{
		Title:    title,
		Date:     date,
		Slug:     slug,
		Excerpt:  excerptStr,
		Content:  text,
		FullPath: filePath,
		Filename: filename,
	}, nil
}

// extractSlug extracts the slug from a filename
// 2026-03-27-hello-world.md -> hello-world
func extractSlug(filename string) string {
	// Remove .md extension
	slug := strings.TrimSuffix(filename, ".md")

	// Remove date prefix (YYYY-MM-DD-)
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)
	slug = datePattern.ReplaceAllString(slug, "")

	return slug
}

// slugToTitle converts a slug to a title
// hello-world -> Hello World
func slugToTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, " ")
}

// FindBySlug finds a post by its slug
func FindBySlug(posts []*Post, slug string) *Post {
	for _, p := range posts {
		if p.Slug == slug {
			return p
		}
	}
	return nil
}
