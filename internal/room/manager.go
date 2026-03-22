package room

import (
	"errors"
	"sync"
)

// Room related errors
var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomFull     = errors.New("room is full")
)

// Responsible for managing game rooms, including creating new rooms, allowing players to join existing rooms, and deleting rooms when they're no longer needed.
type Manager struct {
	roomMu sync.RWMutex
	rooms  map[string]*Room
}

// Creates a new Manager instance with an initialized rooms map
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// Adds a new room to the manager and returns it.
// The room is created with a unique ID and added to the rooms map.
func (m *Manager) CreateRoom() *Room {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room := NewRoom()
	m.rooms[room.ID] = room

	return room
}

// Look up a room by its ID.
// Returns the room and a boolean indicating whether the room was found or not.
func (m *Manager) GetRoom(id string) (*Room, bool) {
	m.roomMu.RLock()
	defer m.roomMu.RUnlock()

	room, exists := m.rooms[id]
	return room, exists
}

// Allows a player to join a room by its ID.
// Checks if the room exists and if it has space for another player (max 2 players per room).
// If the room is found and has space, it increments the player count and returns the room.
// If the room is not found or is full, it returns an error.
func (m *Manager) JoinRoom(id string) (*Room, error) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room, exists := m.rooms[id]
	if !exists {
		return nil, ErrRoomNotFound
	}

	if room.PlayerCount >= 2 {
		return nil, ErrRoomFull
	}

	room.PlayerCount++
	return room, nil
}

// Deletes a room from the manager by its ID.
func (m *Manager) DeleteRoom(id string) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	delete(m.rooms, id)
}
