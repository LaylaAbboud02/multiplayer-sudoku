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
// Checks if the room exists
func (m *Manager) JoinRoom(id string) (*Room, error) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room, exists := m.rooms[id]
	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// Not used
func (m *Manager) SetGameState(roomID string, state GameState) error {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room, exists := m.rooms[roomID]
	if !exists {
		return ErrRoomNotFound
	}

	room.GameState = state
	return nil
}

func (m *Manager) MarkFinished(roomId string, winnerPlayerNumber int) (*Room, bool, error) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room, exists := m.rooms[roomId]
	if !exists {
		return nil, false, ErrRoomNotFound
	}

	if room.GameState == GameStateFinished {
		return room, false, nil
	}

	room.GameState = GameStateFinished
	room.WinnerPlayerNumber = winnerPlayerNumber

	return room, true, nil
}

func (m *Manager) UpdatePlayerProgress(roomID string, playerNumber int, progressCount int) (*Room, error) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room, exists := m.rooms[roomID]
	if !exists {
		return nil, ErrRoomNotFound
	}

	switch playerNumber {
	case 1:
		room.Player1Progress = progressCount
	case 2:
		room.Player2Progress = progressCount
	}

	return room, nil
}

// Deletes a room from the manager by its ID.
func (m *Manager) DeleteRoom(id string) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	delete(m.rooms, id)
}
