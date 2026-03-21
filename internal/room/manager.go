package room

import (
	"errors"
	"sync"
)

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomFull     = errors.New("room is full")
)

type Manager struct {
	roomMu sync.RWMutex
	rooms  map[string]*Room
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

func (m *Manager) CreateRoom() *Room {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room := NewRoom()
	m.rooms[room.ID] = room

	return room
}

func (m *Manager) GetRoom(id string) (*Room, bool) {
	m.roomMu.RLock()
	defer m.roomMu.RUnlock()

	room, exists := m.rooms[id]
	return room, exists
}

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

func (m *Manager) DeleteRoom(id string) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	delete(m.rooms, id)
}
