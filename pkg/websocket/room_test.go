package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRoom(t *testing.T) {
	room := NewRoom("test-room")
	assert.NotNil(t, room)
	assert.Equal(t, "test-room", room.Name)
	assert.Equal(t, 0, room.Size())
}

func TestRoomAddConnection(t *testing.T) {
	room := NewRoom("test-room")
	conn := &Connection{
		ID:   "conn1",
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	err := room.Add(conn)
	assert.NoError(t, err)
	assert.Equal(t, 1, room.Size())
	assert.True(t, room.Has(conn))
}

func TestRoomRemoveConnection(t *testing.T) {
	room := NewRoom("test-room")
	conn := &Connection{
		ID:   "conn1",
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	room.Add(conn)
	assert.Equal(t, 1, room.Size())

	room.Remove(conn)
	assert.Equal(t, 0, room.Size())
	assert.False(t, room.Has(conn))
}

func TestRoomBroadcast(t *testing.T) {
	room := NewRoom("test-room")

	conn1 := &Connection{
		ID:   "conn1",
		send: make(chan []byte, 256),
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}
	conn2 := &Connection{
		ID:   "conn2",
		send: make(chan []byte, 256),
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	room.Add(conn1)
	room.Add(conn2)

	message := []byte("test message")
	room.Broadcast(message, nil)

	// Both connections should receive the message
	assert.Equal(t, message, <-conn1.send)
	assert.Equal(t, message, <-conn2.send)
}

func TestRoomBroadcastExclude(t *testing.T) {
	room := NewRoom("test-room")

	conn1 := &Connection{
		ID:   "conn1",
		send: make(chan []byte, 256),
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}
	conn2 := &Connection{
		ID:   "conn2",
		send: make(chan []byte, 256),
		Data: make(map[string]interface{}),
		rooms: make(map[string]bool),
	}

	room.Add(conn1)
	room.Add(conn2)

	message := []byte("test message")
	room.Broadcast(message, conn1) // Exclude conn1

	// Only conn2 should receive the message
	select {
	case msg := <-conn2.send:
		assert.Equal(t, message, msg)
	default:
		t.Fatal("conn2 should have received message")
	}

	select {
	case <-conn1.send:
		t.Fatal("conn1 should not have received message")
	default:
		// Expected - conn1 was excluded
	}
}

func TestRoomConnections(t *testing.T) {
	room := NewRoom("test-room")

	conn1 := &Connection{ID: "conn1", Data: make(map[string]interface{}), rooms: make(map[string]bool)}
	conn2 := &Connection{ID: "conn2", Data: make(map[string]interface{}), rooms: make(map[string]bool)}

	room.Add(conn1)
	room.Add(conn2)

	conns := room.Connections()
	assert.Equal(t, 2, len(conns))
	assert.Contains(t, conns, conn1)
	assert.Contains(t, conns, conn2)
}

func TestRoomMetadata(t *testing.T) {
	room := NewRoom("test-room")

	room.SetMetadata("owner", "user123")
	room.SetMetadata("created", "2024-01-01")

	owner, exists := room.GetMetadata("owner")
	assert.True(t, exists)
	assert.Equal(t, "user123", owner)

	created, exists := room.GetMetadata("created")
	assert.True(t, exists)
	assert.Equal(t, "2024-01-01", created)

	_, exists = room.GetMetadata("nonexistent")
	assert.False(t, exists)
}

func TestRoomGetAllMetadata(t *testing.T) {
	room := NewRoom("test-room")

	room.SetMetadata("key1", "value1")
	room.SetMetadata("key2", "value2")

	metadata := room.GetAllMetadata()
	assert.Equal(t, 2, len(metadata))
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, "value2", metadata["key2"])
}

func TestRoomMaxConnections(t *testing.T) {
	room := NewRoom("test-room")
	room.SetMaxConnections(2)

	assert.Equal(t, 2, room.GetMaxConnections())

	conn1 := &Connection{ID: "conn1", Data: make(map[string]interface{}), rooms: make(map[string]bool)}
	conn2 := &Connection{ID: "conn2", Data: make(map[string]interface{}), rooms: make(map[string]bool)}
	conn3 := &Connection{ID: "conn3", Data: make(map[string]interface{}), rooms: make(map[string]bool)}

	// Add up to max
	err := room.Add(conn1)
	assert.NoError(t, err)
	err = room.Add(conn2)
	assert.NoError(t, err)

	// Try to exceed max
	err = room.Add(conn3)
	assert.Error(t, err)
	assert.Equal(t, 2, room.Size())
}

func TestRoomClear(t *testing.T) {
	room := NewRoom("test-room")

	conn1 := &Connection{ID: "conn1", Data: make(map[string]interface{}), rooms: make(map[string]bool)}
	conn2 := &Connection{ID: "conn2", Data: make(map[string]interface{}), rooms: make(map[string]bool)}

	room.Add(conn1)
	room.Add(conn2)
	assert.Equal(t, 2, room.Size())

	room.Clear()
	assert.Equal(t, 0, room.Size())
}

func TestNewRoomManager(t *testing.T) {
	rm := NewRoomManager()
	assert.NotNil(t, rm)
	assert.Equal(t, 0, rm.Count())
}

func TestRoomManagerCreateRoom(t *testing.T) {
	rm := NewRoomManager()

	room := rm.CreateRoom("room1")
	assert.NotNil(t, room)
	assert.Equal(t, "room1", room.Name)
	assert.Equal(t, 1, rm.Count())

	// Creating same room again returns existing room
	room2 := rm.CreateRoom("room1")
	assert.Equal(t, room, room2)
	assert.Equal(t, 1, rm.Count())
}

func TestRoomManagerGetRoom(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room1")

	room, exists := rm.GetRoom("room1")
	assert.True(t, exists)
	assert.Equal(t, "room1", room.Name)

	_, exists = rm.GetRoom("nonexistent")
	assert.False(t, exists)
}

func TestRoomManagerGetOrCreateRoom(t *testing.T) {
	rm := NewRoomManager()

	// Create new room
	room1 := rm.GetOrCreateRoom("room1")
	assert.NotNil(t, room1)
	assert.Equal(t, "room1", room1.Name)
	assert.Equal(t, 1, rm.Count())

	// Get existing room
	room2 := rm.GetOrCreateRoom("room1")
	assert.Equal(t, room1, room2)
	assert.Equal(t, 1, rm.Count())
}

func TestRoomManagerDeleteRoom(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room1")

	assert.Equal(t, 1, rm.Count())

	deleted := rm.DeleteRoom("room1")
	assert.True(t, deleted)
	assert.Equal(t, 0, rm.Count())

	deleted = rm.DeleteRoom("nonexistent")
	assert.False(t, deleted)
}

func TestRoomManagerGetAllRooms(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room1")
	rm.CreateRoom("room2")

	rooms := rm.GetAllRooms()
	assert.Equal(t, 2, len(rooms))
}

func TestRoomManagerGetRoomNames(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room1")
	rm.CreateRoom("room2")

	names := rm.GetRoomNames()
	assert.Equal(t, 2, len(names))
	assert.Contains(t, names, "room1")
	assert.Contains(t, names, "room2")
}

func TestRoomManagerClear(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room1")
	rm.CreateRoom("room2")
	assert.Equal(t, 2, rm.Count())

	rm.Clear()
	assert.Equal(t, 0, rm.Count())
}

func TestRoomManagerAddConnectionToRoom(t *testing.T) {
	rm := NewRoomManager()
	conn := &Connection{ID: "conn1", Data: make(map[string]interface{}), rooms: make(map[string]bool)}

	err := rm.AddConnectionToRoom(conn, "room1")
	assert.NoError(t, err)

	room, exists := rm.GetRoom("room1")
	assert.True(t, exists)
	assert.True(t, room.Has(conn))
}

func TestRoomManagerRemoveConnectionFromRoom(t *testing.T) {
	rm := NewRoomManager()
	conn := &Connection{ID: "conn1", Data: make(map[string]interface{}), rooms: make(map[string]bool)}

	rm.AddConnectionToRoom(conn, "room1")
	room, _ := rm.GetRoom("room1")
	assert.True(t, room.Has(conn))

	rm.RemoveConnectionFromRoom(conn, "room1")
	assert.False(t, room.Has(conn))
}

func TestRoomManagerRemoveConnectionFromAllRooms(t *testing.T) {
	rm := NewRoomManager()
	conn := &Connection{ID: "conn1", Data: make(map[string]interface{}), rooms: make(map[string]bool)}

	rm.AddConnectionToRoom(conn, "room1")
	rm.AddConnectionToRoom(conn, "room2")

	rm.RemoveConnectionFromAllRooms(conn)

	room1, _ := rm.GetRoom("room1")
	room2, _ := rm.GetRoom("room2")
	assert.False(t, room1.Has(conn))
	assert.False(t, room2.Has(conn))
}

func TestRoomManagerGetRoomSize(t *testing.T) {
	rm := NewRoomManager()
	conn1 := &Connection{ID: "conn1", Data: make(map[string]interface{}), rooms: make(map[string]bool)}
	conn2 := &Connection{ID: "conn2", Data: make(map[string]interface{}), rooms: make(map[string]bool)}

	rm.AddConnectionToRoom(conn1, "room1")
	rm.AddConnectionToRoom(conn2, "room1")

	assert.Equal(t, 2, rm.GetRoomSize("room1"))
	assert.Equal(t, 0, rm.GetRoomSize("nonexistent"))
}
