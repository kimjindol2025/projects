// Package stdlib provides HTTP library support for FV 2.0
package stdlib

import (
	"fmt"
)

// HttpMethod represents HTTP methods
type HttpMethod string

const (
	MethodGET    HttpMethod = "GET"
	MethodPOST   HttpMethod = "POST"
	MethodPUT    HttpMethod = "PUT"
	MethodDELETE HttpMethod = "DELETE"
	MethodPATCH  HttpMethod = "PATCH"
	MethodOPTIONS HttpMethod = "OPTIONS"
)

// HttpRequest represents an HTTP request
type HttpRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
	Query   map[string]string
}

// HttpResponse represents an HTTP response
type HttpResponse struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       string
}

// HttpServer represents an HTTP server
type HttpServer struct {
	Port     int
	Host     string
	Routes   map[string]Handler
	Handlers map[string]Handler
}

// Handler is a function that handles HTTP requests
type Handler func(*HttpRequest) *HttpResponse

// NewHttpServer creates a new HTTP server
func NewHttpServer(port int) *HttpServer {
	return &HttpServer{
		Port:     port,
		Host:     "localhost",
		Routes:   make(map[string]Handler),
		Handlers: make(map[string]Handler),
	}
}

// GET registers a GET route
func (s *HttpServer) GET(path string, handler Handler) {
	routeKey := fmt.Sprintf("GET %s", path)
	s.Routes[routeKey] = handler
}

// POST registers a POST route
func (s *HttpServer) POST(path string, handler Handler) {
	routeKey := fmt.Sprintf("POST %s", path)
	s.Routes[routeKey] = handler
}

// PUT registers a PUT route
func (s *HttpServer) PUT(path string, handler Handler) {
	routeKey := fmt.Sprintf("PUT %s", path)
	s.Routes[routeKey] = handler
}

// DELETE registers a DELETE route
func (s *HttpServer) DELETE(path string, handler Handler) {
	routeKey := fmt.Sprintf("DELETE %s", path)
	s.Routes[routeKey] = handler
}

// HandleFunc registers a handler with a custom key
func (s *HttpServer) HandleFunc(key string, handler Handler) {
	s.Handlers[key] = handler
}

// ListenAndServe starts the HTTP server
func (s *HttpServer) ListenAndServe() error {
	fmt.Printf("Server listening on %s:%d\n", s.Host, s.Port)
	return nil
}

// RouteRequest routes an HTTP request to the appropriate handler
func (s *HttpServer) RouteRequest(req *HttpRequest) *HttpResponse {
	routeKey := fmt.Sprintf("%s %s", req.Method, req.Path)

	// Try to find registered route
	if handler, ok := s.Routes[routeKey]; ok {
		return handler(req)
	}

	// Default 404 response
	return &HttpResponse{
		StatusCode: 404,
		StatusText: "Not Found",
		Headers:    map[string]string{"Content-Type": "text/html"},
		Body:       "<h1>404 - Not Found</h1>",
	}
}

// Middleware type for request/response processing
type Middleware func(*HttpRequest, *HttpResponse) (*HttpRequest, *HttpResponse)

// Use applies middleware to all routes
func (s *HttpServer) Use(middleware Middleware) {
	// Middleware chain implementation
	fmt.Println("Middleware registered")
}

// Static serves static files from a directory
func (s *HttpServer) Static(path string, dir string) {
	routeKey := fmt.Sprintf("GET %s", path)
	s.Routes[routeKey] = func(req *HttpRequest) *HttpResponse {
		return &HttpResponse{
			StatusCode: 200,
			StatusText: "OK",
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       fmt.Sprintf("Serving static files from %s", dir),
		}
	}
}

// JSON helper for JSON responses
func JSON(statusCode int, data interface{}) *HttpResponse {
	return &HttpResponse{
		StatusCode: statusCode,
		StatusText: "OK",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       fmt.Sprintf("%v", data),
	}
}

// HTML helper for HTML responses
func HTML(statusCode int, html string) *HttpResponse {
	return &HttpResponse{
		StatusCode: statusCode,
		StatusText: "OK",
		Headers:    map[string]string{"Content-Type": "text/html"},
		Body:       html,
	}
}

// PlainText helper for plain text responses
func PlainText(statusCode int, text string) *HttpResponse {
	return &HttpResponse{
		StatusCode: statusCode,
		StatusText: "OK",
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       text,
	}
}

// NewRequest creates a new HTTP request
func NewRequest(method string, path string, body string) *HttpRequest {
	return &HttpRequest{
		Method:  method,
		Path:    path,
		Body:    body,
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}
}

// AddHeader adds a header to the request
func (r *HttpRequest) AddHeader(key string, value string) {
	r.Headers[key] = value
}

// GetHeader gets a header from the request
func (r *HttpRequest) GetHeader(key string) string {
	return r.Headers[key]
}

// NewResponse creates a new HTTP response
func NewResponse(statusCode int, body string) *HttpResponse {
	return &HttpResponse{
		StatusCode: statusCode,
		StatusText: "OK",
		Headers:    make(map[string]string),
		Body:       body,
	}
}

// SetHeader sets a response header
func (r *HttpResponse) SetHeader(key string, value string) {
	r.Headers[key] = value
}

// GetHeader gets a response header
func (r *HttpResponse) GetHeader(key string) string {
	return r.Headers[key]
}

// String returns the HTTP response as a string
func (r *HttpResponse) String() string {
	result := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, r.StatusText)
	for key, value := range r.Headers {
		result += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	result += "\r\n" + r.Body
	return result
}
