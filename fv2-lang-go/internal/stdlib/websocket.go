// Package stdlib provides WebSocket support for FV 2.0
package stdlib

import (
	"fmt"
	"sync"
	"time"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	ID        int64
	Type      string // "text", "binary", "ping", "pong", "close", "error"
	Data      []byte
	Text      string
	Timestamp int64
	ClientID  int64
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID           int64
	Username     string
	Connected    bool
	JoinTime     int64
	LastActivity int64
	SendChan     chan *WebSocketMessage
	closed       bool
	mutex        sync.RWMutex
}

// WebSocketRoom represents a chat room
type WebSocketRoom struct {
	ID           int64
	Name         string
	OwnerID      int64
	MaxUsers     int64
	CreatedAt    int64
	Clients      map[int64]*WebSocketClient
	MessageQueue []*WebSocketMessage
	mutex        sync.RWMutex
}

// WebSocketServer represents a WebSocket server
type WebSocketServer struct {
	Host           string
	Port           int64
	MaxConnections int64
	ReadTimeout    int64
	WriteTimeout   int64
	Rooms          map[int64]*WebSocketRoom
	Clients        map[int64]*WebSocketClient
	ClientIDGen    int64
	RoomIDGen      int64
	mutex          sync.RWMutex
	eventHandlers  map[string]EventHandler
}

// EventHandler is a function that handles WebSocket events
type EventHandler func(*WebSocketServer, *WebSocketClient, interface{})

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(host string, port int64) *WebSocketServer {
	return &WebSocketServer{
		Host:           host,
		Port:           port,
		MaxConnections: 1000,
		ReadTimeout:    30000,
		WriteTimeout:   30000,
		Rooms:          make(map[int64]*WebSocketRoom),
		Clients:        make(map[int64]*WebSocketClient),
		ClientIDGen:    1,
		RoomIDGen:      1,
		eventHandlers:  make(map[string]EventHandler),
	}
}

// Start starts the WebSocket server
func (s *WebSocketServer) Start() error {
	fmt.Printf("WebSocket server listening on %s:%d\n", s.Host, s.Port)
	return nil
}

// Stop stops the WebSocket server
func (s *WebSocketServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, client := range s.Clients {
		client.Close()
	}

	fmt.Println("WebSocket server stopped")
	return nil
}

// ConnectClient connects a new client
func (s *WebSocketServer) ConnectClient(username string) *WebSocketClient {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clientID := s.ClientIDGen
	s.ClientIDGen++

	client := &WebSocketClient{
		ID:           clientID,
		Username:     username,
		Connected:    true,
		JoinTime:     time.Now().Unix(),
		LastActivity: time.Now().Unix(),
		SendChan:     make(chan *WebSocketMessage, 100),
		closed:       false,
	}

	s.Clients[clientID] = client

	// Fire connect event
	if handler, ok := s.eventHandlers["connect"]; ok {
		handler(s, client, nil)
	}

	return client
}

// DisconnectClient disconnects a client
func (s *WebSocketServer) DisconnectClient(clientID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, ok := s.Clients[clientID]
	if !ok {
		return fmt.Errorf("client not found")
	}

	// Fire disconnect event
	if handler, ok := s.eventHandlers["disconnect"]; ok {
		handler(s, client, nil)
	}

	client.Close()
	delete(s.Clients, clientID)

	return nil
}

// SendMessage sends a message to a specific client
func (s *WebSocketServer) SendMessage(clientID int64, msg *WebSocketMessage) error {
	s.mutex.RLock()
	client, ok := s.Clients[clientID]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("client not found")
	}

	if client.closed {
		return fmt.Errorf("client is closed")
	}

	select {
	case client.SendChan <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending message")
	}
}

// BroadcastMessage sends a message to all clients in a room
func (s *WebSocketServer) BroadcastMessage(roomID int64, msg *WebSocketMessage) error {
	s.mutex.RLock()
	room, ok := s.Rooms[roomID]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("room not found")
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for _, client := range room.Clients {
		_ = s.SendMessage(client.ID, msg)
	}

	return nil
}

// CreateRoom creates a new chat room
func (s *WebSocketServer) CreateRoom(name string, ownerID int64) *WebSocketRoom {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	roomID := s.RoomIDGen
	s.RoomIDGen++

	room := &WebSocketRoom{
		ID:           roomID,
		Name:         name,
		OwnerID:      ownerID,
		MaxUsers:     100,
		CreatedAt:    time.Now().Unix(),
		Clients:      make(map[int64]*WebSocketClient),
		MessageQueue: make([]*WebSocketMessage, 0),
	}

	s.Rooms[roomID] = room
	return room
}

// DeleteRoom deletes a room
func (s *WebSocketServer) DeleteRoom(roomID int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, ok := s.Rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found")
	}

	// Remove all clients from room
	room.mutex.Lock()
	for clientID := range room.Clients {
		delete(room.Clients, clientID)
	}
	room.mutex.Unlock()

	delete(s.Rooms, roomID)
	return nil
}

// JoinRoom adds a client to a room
func (s *WebSocketServer) JoinRoom(roomID int64, clientID int64) error {
	s.mutex.RLock()
	room, ok := s.Rooms[roomID]
	if !ok {
		s.mutex.RUnlock()
		return fmt.Errorf("room not found")
	}

	client, ok := s.Clients[clientID]
	if !ok {
		s.mutex.RUnlock()
		return fmt.Errorf("client not found")
	}
	s.mutex.RUnlock()

	room.mutex.Lock()
	defer room.mutex.Unlock()

	if int64(len(room.Clients)) >= room.MaxUsers {
		return fmt.Errorf("room is full")
	}

	room.Clients[clientID] = client
	return nil
}

// LeaveRoom removes a client from a room
func (s *WebSocketServer) LeaveRoom(roomID int64, clientID int64) error {
	s.mutex.RLock()
	room, ok := s.Rooms[roomID]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("room not found")
	}

	room.mutex.Lock()
	defer room.mutex.Unlock()

	delete(room.Clients, clientID)
	return nil
}

// GetRoomClients returns all clients in a room
func (s *WebSocketServer) GetRoomClients(roomID int64) []*WebSocketClient {
	s.mutex.RLock()
	room, ok := s.Rooms[roomID]
	s.mutex.RUnlock()

	if !ok {
		return nil
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	clients := make([]*WebSocketClient, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, client)
	}

	return clients
}

// SaveMessage saves a message to room history
func (s *WebSocketServer) SaveMessage(roomID int64, msg *WebSocketMessage) error {
	s.mutex.RLock()
	room, ok := s.Rooms[roomID]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("room not found")
	}

	room.mutex.Lock()
	defer room.mutex.Unlock()

	room.MessageQueue = append(room.MessageQueue, msg)
	return nil
}

// GetMessageHistory returns message history for a room
func (s *WebSocketServer) GetMessageHistory(roomID int64, limit int64) []*WebSocketMessage {
	s.mutex.RLock()
	room, ok := s.Rooms[roomID]
	s.mutex.RUnlock()

	if !ok {
		return nil
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if int64(len(room.MessageQueue)) <= limit {
		return room.MessageQueue
	}

	start := int64(len(room.MessageQueue)) - limit
	return room.MessageQueue[start:]
}

// On registers an event handler
func (s *WebSocketServer) On(event string, handler EventHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.eventHandlers[event] = handler
}

// Emit fires an event
func (s *WebSocketServer) Emit(event string, client *WebSocketClient, data interface{}) {
	s.mutex.RLock()
	handler, ok := s.eventHandlers[event]
	s.mutex.RUnlock()

	if ok {
		handler(s, client, data)
	}
}

// GetConnectedClientsCount returns the number of connected clients
func (s *WebSocketServer) GetConnectedClientsCount() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return int64(len(s.Clients))
}

// GetRoomsCount returns the number of active rooms
func (s *WebSocketServer) GetRoomsCount() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return int64(len(s.Rooms))
}

// GetServerStats returns server statistics
func (s *WebSocketServer) GetServerStats() map[string]int64 {
	return map[string]int64{
		"connected_clients": s.GetConnectedClientsCount(),
		"active_rooms":      s.GetRoomsCount(),
	}
}

// Close closes the client connection
func (c *WebSocketClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return fmt.Errorf("client already closed")
	}

	c.closed = true
	c.Connected = false
	close(c.SendChan)

	return nil
}

// Send sends a message to the client
func (c *WebSocketClient) Send(msg *WebSocketMessage) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	select {
	case c.SendChan <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending message")
	}
}

// UpdateActivity updates the last activity timestamp
func (c *WebSocketClient) UpdateActivity() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.LastActivity = time.Now().Unix()
}

// IsConnected checks if the client is connected
func (c *WebSocketClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Connected && !c.closed
}

// NewMessage creates a new WebSocket message
func NewMessage(msgType string, data string, clientID int64) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      msgType,
		Text:      data,
		Data:      []byte(data),
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
	}
}

// TextMessage creates a text message
func TextMessage(text string, clientID int64) *WebSocketMessage {
	return NewMessage("text", text, clientID)
}

// BinaryMessage creates a binary message
func BinaryMessage(data []byte, clientID int64) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      "binary",
		Data:      data,
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
	}
}

// PingMessage creates a ping message
func PingMessage() *WebSocketMessage {
	return NewMessage("ping", "ping", 0)
}

// PongMessage creates a pong message
func PongMessage() *WebSocketMessage {
	return NewMessage("pong", "pong", 0)
}

// CloseMessage creates a close message
func CloseMessage(code int64, reason string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      "close",
		Text:      fmt.Sprintf("%d: %s", code, reason),
		Timestamp: time.Now().Unix(),
	}
}

// ErrorMessage creates an error message
func ErrorMessage(code int64, msg string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      "error",
		Text:      fmt.Sprintf("%d: %s", code, msg),
		Timestamp: time.Now().Unix(),
	}
}
