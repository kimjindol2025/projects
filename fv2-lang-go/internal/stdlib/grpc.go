// Package stdlib provides gRPC support for FV 2.0
package stdlib

import (
	"fmt"
	"sync"
	"time"
)

// GrpcMessage represents a gRPC message
type GrpcMessage struct {
	ID        int64
	Type      string // "request", "response", "error"
	Data      []byte
	Timestamp int64
	ServiceID int64
}

// GrpcService represents a gRPC service definition
type GrpcService struct {
	ID          int64
	Name        string
	Version     string
	Methods     map[string]GrpcMethod
	Description string
	CreatedAt   int64
	mutex       sync.RWMutex
}

// GrpcMethod represents a gRPC method
type GrpcMethod struct {
	Name       string
	Input      string  // message type name
	Output     string  // message type name
	IsStreaming bool
	Handler    func(*GrpcRequest) (*GrpcResponse, error)
}

// GrpcRequest represents a gRPC request
type GrpcRequest struct {
	ID        int64
	ServiceID int64
	Method    string
	Data      []byte
	Metadata  map[string]string
	Timestamp int64
	ClientID  int64
	closed    bool
	mutex     sync.RWMutex
}

// GrpcResponse represents a gRPC response
type GrpcResponse struct {
	ID        int64
	RequestID int64
	Status    int64 // 0 = OK, non-zero = error
	Data      []byte
	Error     string
	Metadata  map[string]string
	Timestamp int64
	closed    bool
	mutex     sync.RWMutex
}

// GrpcEventHandler is a function that handles gRPC events
type GrpcEventHandler func(*GrpcServer, interface{})

// GrpcServer represents a gRPC server
type GrpcServer struct {
	Host        string
	Port        int64
	Services    map[int64]*GrpcService
	Connections map[int64]*GrpcConnection
	ServiceIDGen int64
	ConnIDGen   int64
	MaxConnections int64
	ReadTimeout int64
	WriteTimeout int64
	eventHandlers map[string]GrpcEventHandler
	mutex       sync.RWMutex
}

// GrpcConnection represents a gRPC client connection
type GrpcConnection struct {
	ID           int64
	ClientID     string
	Connected    bool
	ConnectTime  int64
	LastActivity int64
	ServiceID    int64
	closed       bool
	RequestChan  chan *GrpcRequest
	ResponseChan chan *GrpcResponse
	mutex        sync.RWMutex
}

// GrpcStream represents a bidirectional gRPC stream
type GrpcStream struct {
	ID           int64
	ConnectionID int64
	Active       bool
	StartTime    int64
	LastActivity int64
	Messages     []*GrpcMessage
	closed       bool
	mutex        sync.RWMutex
}

// NewGrpcServer creates a new gRPC server
func NewGrpcServer(host string, port int64) *GrpcServer {
	return &GrpcServer{
		Host:           host,
		Port:           port,
		Services:       make(map[int64]*GrpcService),
		Connections:    make(map[int64]*GrpcConnection),
		ServiceIDGen:   1,
		ConnIDGen:      1,
		MaxConnections: 1000,
		ReadTimeout:    30000,
		WriteTimeout:   30000,
		eventHandlers:  make(map[string]GrpcEventHandler),
	}
}

// Start starts the gRPC server
func (s *GrpcServer) Start() error {
	fmt.Printf("gRPC server listening on %s:%d\n", s.Host, s.Port)
	return nil
}

// Stop stops the gRPC server
func (s *GrpcServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, conn := range s.Connections {
		conn.Close()
	}

	fmt.Println("gRPC server stopped")
	return nil
}

// RegisterService registers a gRPC service
func (s *GrpcServer) RegisterService(name string, version string) *GrpcService {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	serviceID := s.ServiceIDGen
	s.ServiceIDGen++

	service := &GrpcService{
		ID:        serviceID,
		Name:      name,
		Version:   version,
		Methods:   make(map[string]GrpcMethod),
		CreatedAt: time.Now().Unix(),
	}

	s.Services[serviceID] = service

	// Fire service registered event
	if handler, ok := s.eventHandlers["service_registered"]; ok {
		handler(s, service)
	}

	return service
}

// UnregisterService unregisters a gRPC service
func (s *GrpcServer) UnregisterService(serviceID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	service, ok := s.Services[serviceID]
	if !ok {
		return fmt.Errorf("service not found")
	}

	// Fire service unregistered event
	if handler, ok := s.eventHandlers["service_unregistered"]; ok {
		handler(s, service)
	}

	delete(s.Services, serviceID)
	return nil
}

// AddMethod adds a method to a service
func (s *GrpcServer) AddMethod(serviceID int64, methodName string, inputType string, outputType string, handler func(*GrpcRequest) (*GrpcResponse, error)) error {
	s.mutex.RLock()
	service, ok := s.Services[serviceID]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("service not found")
	}

	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.Methods[methodName] = GrpcMethod{
		Name:        methodName,
		Input:       inputType,
		Output:      outputType,
		IsStreaming: false,
		Handler:     handler,
	}

	// Fire method added event
	if handler, ok := s.eventHandlers["method_added"]; ok {
		handler(s, map[string]string{"service": service.Name, "method": methodName})
	}

	return nil
}

// Call invokes a gRPC method
func (s *GrpcServer) Call(serviceID int64, methodName string, req *GrpcRequest) (*GrpcResponse, error) {
	s.mutex.RLock()
	service, ok := s.Services[serviceID]
	s.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("service not found")
	}

	service.mutex.RLock()
	method, ok := service.Methods[methodName]
	service.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("method not found")
	}

	return method.Handler(req)
}

// Connect creates a new gRPC client connection
func (s *GrpcServer) Connect(clientID string, serviceID int64) (*GrpcConnection, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if int64(len(s.Connections)) >= s.MaxConnections {
		return nil, fmt.Errorf("max connections reached")
	}

	connID := s.ConnIDGen
	s.ConnIDGen++

	conn := &GrpcConnection{
		ID:           connID,
		ClientID:     clientID,
		Connected:    true,
		ConnectTime:  time.Now().Unix(),
		LastActivity: time.Now().Unix(),
		ServiceID:    serviceID,
		closed:       false,
		RequestChan:  make(chan *GrpcRequest, 100),
		ResponseChan: make(chan *GrpcResponse, 100),
	}

	s.Connections[connID] = conn

	// Fire connect event
	if handler, ok := s.eventHandlers["connected"]; ok {
		handler(s, conn)
	}

	return conn, nil
}

// Disconnect closes a gRPC client connection
func (s *GrpcServer) Disconnect(connID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn, ok := s.Connections[connID]
	if !ok {
		return fmt.Errorf("connection not found")
	}

	// Fire disconnect event
	if handler, ok := s.eventHandlers["disconnected"]; ok {
		handler(s, conn)
	}

	conn.Close()
	delete(s.Connections, connID)

	return nil
}

// SendRequest sends a gRPC request
func (s *GrpcServer) SendRequest(connID int64, req *GrpcRequest) error {
	s.mutex.RLock()
	conn, ok := s.Connections[connID]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("connection not found")
	}

	select {
	case conn.RequestChan <- req:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending request")
	}
}

// SendResponse sends a gRPC response
func (s *GrpcServer) SendResponse(connID int64, resp *GrpcResponse) error {
	s.mutex.RLock()
	conn, ok := s.Connections[connID]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("connection not found")
	}

	select {
	case conn.ResponseChan <- resp:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending response")
	}
}

// GetService retrieves a service
func (s *GrpcServer) GetService(serviceID int64) *GrpcService {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Services[serviceID]
}

// GetServices returns all registered services
func (s *GrpcServer) GetServices() []*GrpcService {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	services := make([]*GrpcService, 0, len(s.Services))
	for _, service := range s.Services {
		services = append(services, service)
	}

	return services
}

// GetConnections returns all active connections
func (s *GrpcServer) GetConnections() []*GrpcConnection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	conns := make([]*GrpcConnection, 0, len(s.Connections))
	for _, conn := range s.Connections {
		conns = append(conns, conn)
	}

	return conns
}

// GetServerStats returns server statistics
func (s *GrpcServer) GetServerStats() map[string]int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]int64{
		"services":      int64(len(s.Services)),
		"connections":   int64(len(s.Connections)),
		"max_connections": s.MaxConnections,
	}
}

// On registers an event handler
func (s *GrpcServer) On(event string, handler GrpcEventHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.eventHandlers[event] = handler
}

// Emit fires an event
func (s *GrpcServer) Emit(event string, data interface{}) {
	s.mutex.RLock()
	handler, ok := s.eventHandlers[event]
	s.mutex.RUnlock()

	if ok {
		handler(s, data)
	}
}

// Close closes the connection
func (c *GrpcConnection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return fmt.Errorf("connection already closed")
	}

	c.closed = true
	c.Connected = false
	close(c.RequestChan)
	close(c.ResponseChan)

	return nil
}

// UpdateActivity updates the last activity timestamp
func (c *GrpcConnection) UpdateActivity() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.LastActivity = time.Now().Unix()
}

// IsConnected checks if the connection is active
func (c *GrpcConnection) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Connected && !c.closed
}

// Close closes the response
func (r *GrpcResponse) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.closed {
		return fmt.Errorf("response already closed")
	}

	r.closed = true
	return nil
}

// NewGrpcRequest creates a new gRPC request
func NewGrpcRequest(serviceID int64, method string, data []byte) *GrpcRequest {
	return &GrpcRequest{
		ServiceID: serviceID,
		Method:    method,
		Data:      data,
		Metadata:  make(map[string]string),
		Timestamp: time.Now().Unix(),
	}
}

// NewGrpcResponse creates a new gRPC response
func NewGrpcResponse(requestID int64, status int64) *GrpcResponse {
	return &GrpcResponse{
		RequestID: requestID,
		Status:    status,
		Metadata:  make(map[string]string),
		Timestamp: time.Now().Unix(),
	}
}

// NewGrpcStream creates a new gRPC stream
func NewGrpcStream(connID int64) *GrpcStream {
	return &GrpcStream{
		ConnectionID: connID,
		Active:       true,
		StartTime:    time.Now().Unix(),
		LastActivity: time.Now().Unix(),
		Messages:     make([]*GrpcMessage, 0),
		closed:       false,
	}
}

// SendMessage sends a message on the stream
func (stream *GrpcStream) SendMessage(msg *GrpcMessage) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if stream.closed {
		return fmt.Errorf("stream is closed")
	}

	stream.Messages = append(stream.Messages, msg)
	stream.LastActivity = time.Now().Unix()

	return nil
}

// Close closes the stream
func (stream *GrpcStream) Close() error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if stream.closed {
		return fmt.Errorf("stream already closed")
	}

	stream.closed = true
	stream.Active = false

	return nil
}

// GetMessages returns all messages in the stream
func (stream *GrpcStream) GetMessages() []*GrpcMessage {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	return stream.Messages
}
