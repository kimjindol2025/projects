package stdlib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewWebSocketServer tests creating a new WebSocket server
func TestNewWebSocketServer(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	assert.NotNil(t, server)
	assert.Equal(t, "localhost", server.Host)
	assert.Equal(t, int64(8080), server.Port)
	assert.Equal(t, int64(1000), server.MaxConnections)
	assert.Equal(t, int64(30000), server.ReadTimeout)
	assert.Equal(t, int64(30000), server.WriteTimeout)
	assert.Equal(t, 0, len(server.Rooms))
	assert.Equal(t, 0, len(server.Clients))
}

// TestConnectClient tests connecting a new client
func TestConnectClient(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	client := server.ConnectClient("Alice")

	assert.NotNil(t, client)
	assert.Equal(t, int64(1), client.ID)
	assert.Equal(t, "Alice", client.Username)
	assert.Equal(t, true, client.Connected)
	assert.Greater(t, client.JoinTime, int64(0))
	assert.NotNil(t, client.SendChan)
}

// TestDisconnectClient tests disconnecting a client
func TestDisconnectClient(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	client := server.ConnectClient("Bob")

	err := server.DisconnectClient(client.ID)

	assert.NoError(t, err)
	assert.Equal(t, false, client.Connected)
}

// TestDisconnectNonexistentClient tests disconnecting a non-existent client
func TestDisconnectNonexistentClient(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	err := server.DisconnectClient(999)

	assert.Error(t, err)
	assert.Equal(t, "client not found", err.Error())
}

// TestCreateRoom tests creating a new room
func TestCreateRoom(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	room := server.CreateRoom("General", 1)

	assert.NotNil(t, room)
	assert.Equal(t, int64(1), room.ID)
	assert.Equal(t, "General", room.Name)
	assert.Equal(t, int64(1), room.OwnerID)
	assert.Equal(t, int64(100), room.MaxUsers)
	assert.Equal(t, 0, len(room.Clients))
}

// TestJoinRoom tests adding a client to a room
func TestJoinRoom(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)
	client := server.ConnectClient("Alice")

	err := server.JoinRoom(room.ID, client.ID)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(room.Clients))
}

// TestLeaveRoom tests removing a client from a room
func TestLeaveRoom(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)
	client := server.ConnectClient("Alice")
	server.JoinRoom(room.ID, client.ID)

	err := server.LeaveRoom(room.ID, client.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(room.Clients))
}

// TestGetRoomClients tests retrieving clients in a room
func TestGetRoomClients(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)
	client1 := server.ConnectClient("Alice")
	client2 := server.ConnectClient("Bob")

	server.JoinRoom(room.ID, client1.ID)
	server.JoinRoom(room.ID, client2.ID)

	clients := server.GetRoomClients(room.ID)

	assert.Equal(t, 2, len(clients))
}

// TestSendMessage tests sending a message to a client
func TestSendMessage(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	client := server.ConnectClient("Alice")

	msg := TextMessage("Hello", client.ID)
	err := server.SendMessage(client.ID, msg)

	assert.NoError(t, err)
}

// TestSendMessageToNonexistentClient tests sending to a non-existent client
func TestSendMessageToNonexistentClient(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	msg := TextMessage("Hello", 999)
	err := server.SendMessage(999, msg)

	assert.Error(t, err)
	assert.Equal(t, "client not found", err.Error())
}

// TestBroadcastMessage tests broadcasting to all clients in a room
func TestBroadcastMessage(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)
	client1 := server.ConnectClient("Alice")
	client2 := server.ConnectClient("Bob")

	server.JoinRoom(room.ID, client1.ID)
	server.JoinRoom(room.ID, client2.ID)

	msg := TextMessage("Hello everyone", 1)
	err := server.BroadcastMessage(room.ID, msg)

	assert.NoError(t, err)
}

// TestSaveMessage tests saving message history
func TestSaveMessage(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)

	msg := TextMessage("Hello", 1)
	err := server.SaveMessage(room.ID, msg)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(room.MessageQueue))
}

// TestGetMessageHistory tests retrieving message history
func TestGetMessageHistory(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)

	for i := 0; i < 5; i++ {
		msg := TextMessage("Message", 1)
		server.SaveMessage(room.ID, msg)
	}

	history := server.GetMessageHistory(room.ID, 10)

	assert.Equal(t, 5, len(history))
}

// TestGetMessageHistoryWithLimit tests message history with limit
func TestGetMessageHistoryWithLimit(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)

	for i := 0; i < 20; i++ {
		msg := TextMessage("Message", 1)
		server.SaveMessage(room.ID, msg)
	}

	history := server.GetMessageHistory(room.ID, 10)

	assert.Equal(t, 10, len(history))
}

// TestDeleteRoom tests deleting a room
func TestDeleteRoom(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("General", 1)

	err := server.DeleteRoom(room.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(server.Rooms))
}

// TestEventHandlers tests registering and emitting events
func TestEventHandlers(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	eventFired := false

	server.On("test_event", func(s *WebSocketServer, c *WebSocketClient, data interface{}) {
		eventFired = true
	})

	client := server.ConnectClient("Alice")
	server.Emit("test_event", client, nil)

	assert.Equal(t, true, eventFired)
}

// TestGetConnectedClientsCount tests retrieving connected clients count
func TestGetConnectedClientsCount(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	server.ConnectClient("Alice")
	server.ConnectClient("Bob")

	count := server.GetConnectedClientsCount()

	assert.Equal(t, int64(2), count)
}

// TestGetRoomsCount tests retrieving active rooms count
func TestGetRoomsCount(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	server.CreateRoom("General", 1)
	server.CreateRoom("Tech", 1)

	count := server.GetRoomsCount()

	assert.Equal(t, int64(2), count)
}

// TestGetServerStats tests retrieving server statistics
func TestGetServerStats(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	server.ConnectClient("Alice")
	server.CreateRoom("General", 1)

	stats := server.GetServerStats()

	assert.Equal(t, int64(1), stats["connected_clients"])
	assert.Equal(t, int64(1), stats["active_rooms"])
}

// TestClientClose tests closing a client connection
func TestClientClose(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	client := server.ConnectClient("Alice")

	err := client.Close()

	assert.NoError(t, err)
	assert.Equal(t, false, client.Connected)
}

// TestClientIsConnected tests checking if client is connected
func TestClientIsConnected(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	client := server.ConnectClient("Alice")

	assert.Equal(t, true, client.IsConnected())

	client.Close()

	assert.Equal(t, false, client.IsConnected())
}

// TestClientUpdateActivity tests updating client activity timestamp
func TestClientUpdateActivity(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	client := server.ConnectClient("Alice")

	oldActivity := client.LastActivity
	time.Sleep(100 * time.Millisecond)
	client.UpdateActivity()

	assert.GreaterOrEqual(t, client.LastActivity, oldActivity)
}

// TestTextMessage tests creating a text message
func TestTextMessage(t *testing.T) {
	msg := TextMessage("Hello", 1)

	assert.Equal(t, "text", msg.Type)
	assert.Equal(t, "Hello", msg.Text)
	assert.Equal(t, int64(1), msg.ClientID)
	assert.Greater(t, msg.Timestamp, int64(0))
}

// TestBinaryMessage tests creating a binary message
func TestBinaryMessage(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	msg := BinaryMessage(data, 1)

	assert.Equal(t, "binary", msg.Type)
	assert.Equal(t, data, msg.Data)
	assert.Equal(t, int64(1), msg.ClientID)
}

// TestPingMessage tests creating a ping message
func TestPingMessage(t *testing.T) {
	msg := PingMessage()

	assert.Equal(t, "ping", msg.Type)
	assert.Equal(t, "ping", msg.Text)
}

// TestPongMessage tests creating a pong message
func TestPongMessage(t *testing.T) {
	msg := PongMessage()

	assert.Equal(t, "pong", msg.Type)
	assert.Equal(t, "pong", msg.Text)
}

// TestCloseMessage tests creating a close message
func TestCloseMessage(t *testing.T) {
	msg := CloseMessage(1000, "Normal closure")

	assert.Equal(t, "close", msg.Type)
	assert.Contains(t, msg.Text, "1000")
	assert.Contains(t, msg.Text, "Normal closure")
}

// TestErrorMessage tests creating an error message
func TestErrorMessage(t *testing.T) {
	msg := ErrorMessage(500, "Internal server error")

	assert.Equal(t, "error", msg.Type)
	assert.Contains(t, msg.Text, "500")
	assert.Contains(t, msg.Text, "Internal server error")
}

// TestClientSend tests sending message to client directly
func TestClientSend(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	client := server.ConnectClient("Alice")

	msg := TextMessage("Hello", client.ID)
	err := client.Send(msg)

	assert.NoError(t, err)
}

// TestRoomFullCapacity tests room full capacity check
func TestRoomFullCapacity(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	room := server.CreateRoom("Small", 1)
	room.MaxUsers = 2

	client1 := server.ConnectClient("Alice")
	client2 := server.ConnectClient("Bob")
	client3 := server.ConnectClient("Charlie")

	err1 := server.JoinRoom(room.ID, client1.ID)
	err2 := server.JoinRoom(room.ID, client2.ID)
	err3 := server.JoinRoom(room.ID, client3.ID)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Error(t, err3)
	assert.Equal(t, "room is full", err3.Error())
}

// TestMultipleRooms tests managing multiple rooms
func TestMultipleRooms(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	room1 := server.CreateRoom("General", 1)
	room2 := server.CreateRoom("Tech", 1)
	room3 := server.CreateRoom("Random", 1)

	assert.Equal(t, int64(3), server.GetRoomsCount())
	assert.NotNil(t, server.Rooms[room1.ID])
	assert.NotNil(t, server.Rooms[room2.ID])
	assert.NotNil(t, server.Rooms[room3.ID])
}

// TestClientIDGeneration tests unique client ID generation
func TestClientIDGeneration(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	client1 := server.ConnectClient("Alice")
	client2 := server.ConnectClient("Bob")
	client3 := server.ConnectClient("Charlie")

	assert.Equal(t, int64(1), client1.ID)
	assert.Equal(t, int64(2), client2.ID)
	assert.Equal(t, int64(3), client3.ID)
}

// TestRoomIDGeneration tests unique room ID generation
func TestRoomIDGeneration(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	room1 := server.CreateRoom("General", 1)
	room2 := server.CreateRoom("Tech", 1)
	room3 := server.CreateRoom("Random", 1)

	assert.Equal(t, int64(1), room1.ID)
	assert.Equal(t, int64(2), room2.ID)
	assert.Equal(t, int64(3), room3.ID)
}

// TestStartServer tests starting the server
func TestStartServer(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)

	err := server.Start()

	assert.NoError(t, err)
}

// TestStopServer tests stopping the server
func TestStopServer(t *testing.T) {
	server := NewWebSocketServer("localhost", 8080)
	client1 := server.ConnectClient("Alice")
	client2 := server.ConnectClient("Bob")

	err := server.Stop()

	assert.NoError(t, err)
	assert.Equal(t, false, client1.Connected)
	assert.Equal(t, false, client2.Connected)
}
