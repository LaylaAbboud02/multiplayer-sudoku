package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Json message sent to clients to update them on the current status of the room like how many players are in the room)
type RoomStatusMessage struct {
	Type        string `json:"type"`
	RoomID      string `json:"room_id"`
	PlayerCount int    `json:"player_count"`
}

// MAnages all active websocket clients and their connections to rooms.
// Contains a map of room IDs to sets of clients currently in those rooms.
// Example:
//
//	{
//	  "ABC123": {client1: true, client2: true},
//	  "XYZ789": {client3: true},
//	}
type Hub struct {
	clientsMu sync.RWMutex
	clients   map[string]map[*Client]bool
}

// Initializes a new Hub with an empty clients map.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]bool),
	}
}

// Adds a client to the hub's clients map under the appropriate room ID.
// If the room ID doesn't exist in the map yet, it creates a new entry for it.
func (h *Hub) Register(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if h.clients[client.RoomID] == nil {
		h.clients[client.RoomID] = make(map[*Client]bool)
	}

	h.clients[client.RoomID][client] = true
}

// Removes a client from the hub's clients map.
// If the room becomes empty after removing the client, it deletes the room entry from the map.
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

// Returns the number of currently connected websocket clients in a given room.
func (h *Hub) RoomClientCount(roomID string) int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	return len(h.clients[roomID])
}

// sends a room status update message to all clients in the specified room to let them know of the current player count.
func (h *Hub) BroadcastRoomStatus(roomID string) {
	count := h.RoomClientCount(roomID)

	msg := RoomStatusMessage{
		Type:        "room_status",
		RoomID:      roomID,
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
