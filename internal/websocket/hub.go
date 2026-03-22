package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type RoomStatusMessage struct {
	Type string `json:"type"`
	RoomID string `json:"room_id"`
	PlayerCount int `json:"player_count"`
}

type Hub struct {
	clientsMu sync.RWMutex
	clients   map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Register(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if h.clients[client.RoomID] == nil {
		h.clients[client.RoomID] = make(map[*Client]bool)
	}

	h.clients[client.RoomID][client] = true
}

func (h *Hub) Unregister(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	roomClients, exists := h.clients[client.RoomID]
	if !exists {
		return
	}

	if _, exists := roomClients[client]; exists {
		delete(roomClients, client)
		close(client.Send)
	}

	if len(roomClients) == 0 {
		delete(h.clients, client.RoomID)
	}
}

func (h* Hub) RoomClientCount(roomID string) int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	return len(h.clients[roomID])
}


func (h* Hub) BroadcastRoomStatus(roomID string) {
	count := h.RoomClientCount(roomID)

	msg := RoomStatusMessage{
		Type: "room_status",
		RoomID: roomID,
		PlayerCount: count,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling room status: %v", err)
		return
	}

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for client := range h.clients[roomID] {
		select {
		case client.Send <- payload:
		default:
			// Skip for nwo
		}
	}
}