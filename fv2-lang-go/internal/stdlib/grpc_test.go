package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewGrpcServer tests creating a new gRPC server
func TestNewGrpcServer(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)

	assert.NotNil(t, server)
	assert.Equal(t, "localhost", server.Host)
	assert.Equal(t, int64(50051), server.Port)
	assert.Equal(t, int64(1000), server.MaxConnections)
	assert.Equal(t, int64(30000), server.ReadTimeout)
	assert.Equal(t, int64(30000), server.WriteTimeout)
}

// TestRegisterService tests registering a gRPC service
func TestRegisterService(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)

	service := server.RegisterService("UserService", "1.0.0")

	assert.NotNil(t, service)
	assert.Equal(t, "UserService", service.Name)
	assert.Equal(t, "1.0.0", service.Version)
	assert.Greater(t, service.ID, int64(0))
}

// TestUnregisterService tests unregistering a service
func TestUnregisterService(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")

	err := server.UnregisterService(service.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(server.Services))
}

// TestAddMethod tests adding a method to a service
func TestAddMethod(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")

	handler := func(req *GrpcRequest) (*GrpcResponse, error) {
		return &GrpcResponse{Status: 0}, nil
	}

	err := server.AddMethod(service.ID, "GetUser", "GetUserRequest", "GetUserResponse", handler)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(service.Methods))
}

// TestCall tests calling a gRPC method
func TestCall(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")

	handler := func(req *GrpcRequest) (*GrpcResponse, error) {
		return &GrpcResponse{Status: 0, Data: []byte("user_data")}, nil
	}

	server.AddMethod(service.ID, "GetUser", "GetUserRequest", "GetUserResponse", handler)

	request := NewGrpcRequest(service.ID, "GetUser", []byte("{\"id\": 1}"))
	response, err := server.Call(service.ID, "GetUser", request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, int64(0), response.Status)
}

// TestConnect tests creating a gRPC client connection
func TestConnect(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")

	conn, err := server.Connect("client1", service.ID)

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Equal(t, "client1", conn.ClientID)
	assert.Equal(t, true, conn.Connected)
}

// TestDisconnect tests closing a gRPC client connection
func TestDisconnect(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	conn, _ := server.Connect("client1", service.ID)

	err := server.Disconnect(conn.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(server.Connections))
}

// TestSendRequest tests sending a gRPC request
func TestSendRequest(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	conn, _ := server.Connect("client1", service.ID)

	req := NewGrpcRequest(service.ID, "GetUser", []byte("data"))
	err := server.SendRequest(conn.ID, req)

	assert.NoError(t, err)
}

// TestSendResponse tests sending a gRPC response
func TestSendResponse(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	conn, _ := server.Connect("client1", service.ID)

	resp := NewGrpcResponse(1, 0)
	err := server.SendResponse(conn.ID, resp)

	assert.NoError(t, err)
}

// TestGetService tests retrieving a service
func TestGetService(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")

	retrieved := server.GetService(service.ID)

	assert.NotNil(t, retrieved)
	assert.Equal(t, service.Name, retrieved.Name)
}

// TestGetServices tests retrieving all services
func TestGetServices(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)

	server.RegisterService("UserService", "1.0.0")
	server.RegisterService("PostService", "1.0.0")
	server.RegisterService("CommentService", "1.0.0")

	services := server.GetServices()

	assert.Equal(t, 3, len(services))
}

// TestGetConnections tests retrieving all connections
func TestGetConnections(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")

	server.Connect("client1", service.ID)
	server.Connect("client2", service.ID)

	conns := server.GetConnections()

	assert.Equal(t, 2, len(conns))
}

// TestGrpcServerStats tests retrieving server statistics
func TestGrpcServerStats(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	server.Connect("client1", service.ID)

	stats := server.GetServerStats()

	assert.Equal(t, int64(1), stats["services"])
	assert.Equal(t, int64(1), stats["connections"])
	assert.Equal(t, int64(1000), stats["max_connections"])
}

// TestGrpcEventHandlers tests registering and emitting events
func TestGrpcEventHandlers(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	eventFired := false

	handler := func(s *GrpcServer, data interface{}) {
		eventFired = true
	}

	server.On("test_event", handler)
	server.Emit("test_event", nil)

	assert.Equal(t, true, eventFired)
}

// TestConnectionClose tests closing a connection
func TestConnectionClose(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	conn, _ := server.Connect("client1", service.ID)

	err := conn.Close()

	assert.NoError(t, err)
	assert.Equal(t, false, conn.Connected)
}

// TestConnectionIsConnected tests checking if connection is active
func TestConnectionIsConnected(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	conn, _ := server.Connect("client1", service.ID)

	assert.Equal(t, true, conn.IsConnected())

	conn.Close()

	assert.Equal(t, false, conn.IsConnected())
}

// TestConnectionUpdateActivity tests updating connection activity
func TestConnectionUpdateActivity(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	conn, _ := server.Connect("client1", service.ID)

	oldActivity := conn.LastActivity
	conn.UpdateActivity()

	assert.GreaterOrEqual(t, conn.LastActivity, oldActivity)
}

// TestNewGrpcRequest tests creating a new gRPC request
func TestNewGrpcRequest(t *testing.T) {
	req := NewGrpcRequest(1, "GetUser", []byte("data"))

	assert.NotNil(t, req)
	assert.Equal(t, int64(1), req.ServiceID)
	assert.Equal(t, "GetUser", req.Method)
	assert.Greater(t, req.Timestamp, int64(0))
}

// TestNewGrpcResponse tests creating a new gRPC response
func TestNewGrpcResponse(t *testing.T) {
	resp := NewGrpcResponse(1, 0)

	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.RequestID)
	assert.Equal(t, int64(0), resp.Status)
}

// TestNewGrpcStream tests creating a new gRPC stream
func TestNewGrpcStream(t *testing.T) {
	stream := NewGrpcStream(1)

	assert.NotNil(t, stream)
	assert.Equal(t, int64(1), stream.ConnectionID)
	assert.Equal(t, true, stream.Active)
}

// TestStreamSendMessage tests sending a message on a stream
func TestStreamSendMessage(t *testing.T) {
	stream := NewGrpcStream(1)
	msg := &GrpcMessage{
		Type: "request",
		Data: []byte("data"),
	}

	err := stream.SendMessage(msg)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(stream.Messages))
}

// TestStreamClose tests closing a stream
func TestStreamClose(t *testing.T) {
	stream := NewGrpcStream(1)

	err := stream.Close()

	assert.NoError(t, err)
	assert.Equal(t, false, stream.Active)
}

// TestStreamGetMessages tests retrieving messages from a stream
func TestStreamGetMessages(t *testing.T) {
	stream := NewGrpcStream(1)

	for i := 0; i < 5; i++ {
		msg := &GrpcMessage{
			Type: "request",
			Data: []byte("data"),
		}
		stream.SendMessage(msg)
	}

	messages := stream.GetMessages()

	assert.Equal(t, 5, len(messages))
}

// TestMaxConnections tests max connections limit
func TestMaxConnections(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	server.MaxConnections = 2

	service := server.RegisterService("UserService", "1.0.0")

	conn1, err1 := server.Connect("client1", service.ID)
	conn2, err2 := server.Connect("client2", service.ID)
	_, err3 := server.Connect("client3", service.ID)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Error(t, err3)
	assert.NotNil(t, conn1)
	assert.NotNil(t, conn2)
}

// TestMultipleServices tests managing multiple services
func TestMultipleServices(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)

	service1 := server.RegisterService("UserService", "1.0.0")
	service2 := server.RegisterService("PostService", "1.0.0")
	service3 := server.RegisterService("CommentService", "1.0.0")

	assert.NotNil(t, service1)
	assert.NotNil(t, service2)
	assert.NotNil(t, service3)
	assert.Equal(t, 3, len(server.Services))
}

// TestResponseClose tests closing a response
func TestResponseClose(t *testing.T) {
	resp := NewGrpcResponse(1, 0)

	err := resp.Close()

	assert.NoError(t, err)
}

// TestGrpcStartServer tests starting the gRPC server
func TestGrpcStartServer(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)

	err := server.Start()

	assert.NoError(t, err)
}

// TestGrpcStopServer tests stopping the gRPC server
func TestGrpcStopServer(t *testing.T) {
	server := NewGrpcServer("localhost", 50051)
	service := server.RegisterService("UserService", "1.0.0")
	server.Connect("client1", service.ID)

	err := server.Stop()

	assert.NoError(t, err)
}
