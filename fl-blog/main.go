package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fl-blog/internal/post"
	"fl-blog/internal/tmpl"
)

var posts []*post.Post

func main() {
	// Default port
	port := ":8253"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Get absolute path to posts directory
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}
	projectDir := filepath.Dir(execPath)
	postsDir := filepath.Join(projectDir, "posts")

	// Check if posts directory exists, otherwise use relative path
	if _, err := os.Stat(postsDir); os.IsNotExist(err) {
		// Try current directory
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Failed to get current directory:", err)
		}
		postsDir = filepath.Join(cwd, "posts")
	}

	fmt.Printf("📂 Loading posts from: %s\n", postsDir)

	// Load all posts
	var loadErr error
	posts, loadErr = post.LoadPosts(postsDir)
	if loadErr != nil {
		log.Fatal("Failed to load posts:", loadErr)
	}

	fmt.Printf("✅ Loaded %d posts\n", len(posts))

	// Setup routes
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/posts", handleAPIPosts)
	http.HandleFunc("/post/", handlePost)
	http.HandleFunc("/", handleIndex)

	// Start server
	fmt.Printf("✅ Server listening on http://localhost%s\n", port)
	fmt.Printf("   Home:   http://localhost%s/\n", port)
	fmt.Printf("   API:    http://localhost%s/api/posts\n", port)
	fmt.Printf("   Health: http://localhost%s/health\n", port)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// handleHealth handles GET /health
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"posts":  len(posts),
		"time":   time.Now().Format(time.RFC3339),
	})
}

// handleIndex handles GET /
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only handle root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html, err := tmpl.RenderIndex(posts)
	if err != nil {
		http.Error(w, "Failed to render index", http.StatusInternalServerError)
		fmt.Printf("Error rendering index: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, html)
}

// handlePost handles GET /post/{slug}
func handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract slug from path: /post/{slug}
	slug := strings.TrimPrefix(r.URL.Path, "/post/")
	if slug == "" || slug == "/post" {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	// Find post by slug
	p := post.FindBySlug(posts, slug)
	if p == nil {
		http.NotFound(w, r)
		return
	}

	html, err := tmpl.RenderPost(p)
	if err != nil {
		http.Error(w, "Failed to render post", http.StatusInternalServerError)
		fmt.Printf("Error rendering post: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, html)
}

// handleAPIPosts handles GET /api/posts
func handleAPIPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return JSON array of posts
	type PostData struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Date    string `json:"date"`
		Excerpt string `json:"excerpt"`
	}

	var postData []PostData
	for _, p := range posts {
		postData = append(postData, PostData{
			Title:   p.Title,
			Slug:    p.Slug,
			Date:    p.Date.Format(time.RFC3339),
			Excerpt: p.Excerpt,
		})
	}

	json.NewEncoder(w).Encode(postData)
}
