package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

const (
	// Message types sent to clients
	MessageTypeRoomStatus       = "room_status"
	MessageTypePlayerAssignment = "player_assignment"
)

// Json message sent to clients to update them on the current status of the room like how many players are in the room)
type RoomStatusMessage struct {
	Type        string `json:"type"`
	RoomID      string `json:"room_id"`
	PlayerCount int    `json:"player_count"`
	GameState   string `json:"game_state"`
}

type PlayerAssignmentMessage struct {
	Type         string `json:"type"`
	PlayerNumber int    `json:"player_number"`
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

func (h *Hub) SendPlayerAssignment(client *Client) {
	msg := PlayerAssignmentMessage{
		Type:         MessageTypePlayerAssignment,
		PlayerNumber: client.PlayerNumber,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal player assignment message", err)
		return
	}

	select {
	case client.Send <- payload:
	default:
		log.Printf("Failed to send player assignment message to client in room %s", client.RoomID)
	}
}

// sends a room status update message to all clients in the specified room to let them know of the current player count.
func (h *Hub) BroadcastRoomStatus(roomID string, playerCount int, gameState string) {
	msg := RoomStatusMessage{
		Type: MessageTypeRoomStatus,
		RoomID: roomID,
		PlayerCount: playerCount,
		GameState: gameState,
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
			log.Printf("Failed to send room status message to client in room %s", roomID)
		}
	}
}
