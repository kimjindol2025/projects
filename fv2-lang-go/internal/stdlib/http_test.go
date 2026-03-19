package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHttpRequestCreation tests creating an HTTP request
func TestHttpRequestCreation(t *testing.T) {
	req := NewRequest("GET", "/api/hello", "")
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "/api/hello", req.Path)
	assert.Equal(t, "", req.Body)
	assert.NotNil(t, req.Headers)
}

// TestHttpRequestHeaders tests request headers
func TestHttpRequestHeaders(t *testing.T) {
	req := NewRequest("POST", "/api/data", "test")
	req.AddHeader("Content-Type", "application/json")
	req.AddHeader("Authorization", "Bearer token")

	assert.Equal(t, "application/json", req.GetHeader("Content-Type"))
	assert.Equal(t, "Bearer token", req.GetHeader("Authorization"))
}

// TestHttpResponseCreation tests creating an HTTP response
func TestHttpResponseCreation(t *testing.T) {
	resp := NewResponse(200, "Hello, World!")
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Hello, World!", resp.Body)
}

// TestHttpResponseHeaders tests response headers
func TestHttpResponseHeaders(t *testing.T) {
	resp := NewResponse(200, "data")
	resp.SetHeader("Content-Type", "application/json")
	resp.SetHeader("Cache-Control", "no-cache")

	assert.Equal(t, "application/json", resp.GetHeader("Content-Type"))
	assert.Equal(t, "no-cache", resp.GetHeader("Cache-Control"))
}

// TestJsonHelper tests JSON response helper
func TestJsonHelper(t *testing.T) {
	resp := JSON(200, map[string]string{"message": "OK"})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.GetHeader("Content-Type"))
	assert.NotEmpty(t, resp.Body)
}

// TestHtmlHelper tests HTML response helper
func TestHtmlHelper(t *testing.T) {
	html := "<h1>Hello</h1>"
	resp := HTML(200, html)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.GetHeader("Content-Type"))
	assert.Equal(t, html, resp.Body)
}

// TestPlainTextHelper tests plain text response helper
func TestPlainTextHelper(t *testing.T) {
	text := "Hello, World!"
	resp := PlainText(200, text)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.GetHeader("Content-Type"))
	assert.Equal(t, text, resp.Body)
}

// TestHttpServerCreation tests creating an HTTP server
func TestHttpServerCreation(t *testing.T) {
	server := NewHttpServer(8080)
	assert.Equal(t, 8080, server.Port)
	assert.Equal(t, "localhost", server.Host)
	assert.NotNil(t, server.Routes)
	assert.NotNil(t, server.Handlers)
}

// TestHttpServerRouteRegistration tests registering routes
func TestHttpServerRouteRegistration(t *testing.T) {
	server := NewHttpServer(8080)

	server.GET("/", func(req *HttpRequest) *HttpResponse {
		return NewResponse(200, "Home")
	})

	server.POST("/api/data", func(req *HttpRequest) *HttpResponse {
		return NewResponse(201, "Created")
	})

	assert.Equal(t, 2, len(server.Routes))
}

// TestHttpServerRouting tests request routing
func TestHttpServerRouting(t *testing.T) {
	server := NewHttpServer(8080)

	server.GET("/", func(req *HttpRequest) *HttpResponse {
		return NewResponse(200, "Home Page")
	})

	server.GET("/api/hello", func(req *HttpRequest) *HttpResponse {
		return JSON(200, map[string]string{"message": "Hello!"})
	})

	// Test routing to home
	homeReq := NewRequest("GET", "/", "")
	homeResp := server.RouteRequest(homeReq)
	assert.Equal(t, 200, homeResp.StatusCode)
	assert.Equal(t, "Home Page", homeResp.Body)

	// Test routing to API
	apiReq := NewRequest("GET", "/api/hello", "")
	apiResp := server.RouteRequest(apiReq)
	assert.Equal(t, 200, apiResp.StatusCode)
	assert.NotEmpty(t, apiResp.Body)
}

// TestHttpServer404 tests 404 response
func TestHttpServer404(t *testing.T) {
	server := NewHttpServer(8080)

	notFoundReq := NewRequest("GET", "/nonexistent", "")
	notFoundResp := server.RouteRequest(notFoundReq)

	assert.Equal(t, 404, notFoundResp.StatusCode)
	assert.Equal(t, "Not Found", notFoundResp.StatusText)
}

// TestHttpResponseString tests response formatting
func TestHttpResponseString(t *testing.T) {
	resp := NewResponse(200, "Hello")
	resp.SetHeader("Content-Type", "text/plain")

	result := resp.String()
	assert.Contains(t, result, "HTTP/1.1 200 OK")
	assert.Contains(t, result, "Content-Type: text/plain")
	assert.Contains(t, result, "Hello")
}

// TestHttpServerStaticFiles tests static file serving
func TestHttpServerStaticFiles(t *testing.T) {
	server := NewHttpServer(8080)
	server.Static("/static", "./public")

	assert.Equal(t, 1, len(server.Routes))
}

// TestHttpMethodConstants tests HTTP method constants
func TestHttpMethodConstants(t *testing.T) {
	assert.Equal(t, HttpMethod("GET"), MethodGET)
	assert.Equal(t, HttpMethod("POST"), MethodPOST)
	assert.Equal(t, HttpMethod("PUT"), MethodPUT)
	assert.Equal(t, HttpMethod("DELETE"), MethodDELETE)
	assert.Equal(t, HttpMethod("PATCH"), MethodPATCH)
	assert.Equal(t, HttpMethod("OPTIONS"), MethodOPTIONS)
}

// TestHttpRequestQuery tests request query parameters
func TestHttpRequestQuery(t *testing.T) {
	req := NewRequest("GET", "/search", "")
	// Simulate query parameters
	req.Query["q"] = "hello"
	req.Query["limit"] = "10"

	assert.Equal(t, "hello", req.Query["q"])
	assert.Equal(t, "10", req.Query["limit"])
}

// TestHttpServerDelete tests DELETE route
func TestHttpServerDelete(t *testing.T) {
	server := NewHttpServer(8080)

	server.DELETE("/api/users/1", func(req *HttpRequest) *HttpResponse {
		return NewResponse(204, "")
	})

	deleteReq := NewRequest("DELETE", "/api/users/1", "")
	deleteResp := server.RouteRequest(deleteReq)

	assert.Equal(t, 204, deleteResp.StatusCode)
}
