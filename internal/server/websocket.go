package server

import (
	"net/http"

	appWebsocket "multiplayer-sudoku/internal/websocket"

	gorillaWebsocket "github.com/gorilla/websocket"
)

// Handler for the websocket endpoint:
// - Checks that the room exists
// - Upgrades the HTTP request to a websocket connection
// - creates a live websocket client and registers it in the hub
// - Broadcasts updated room presence
// - Starts goroutines for reading and writing
func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "Missing room_id parameter", http.StatusBadRequest)
		return
	}

	roomData, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Upgrades the HTTP connection to a WebSocket connection
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	playerNumber := h.hub.RoomClientCount(roomID) + 1
	if playerNumber > 2 {
		playerNumber = 2
	}

	// Creates a new websocket client
	client := &appWebsocket.Client{
		Conn:         conn,
		RoomID:       roomID,
		PlayerNumber: playerNumber,
		Send:         make(chan []byte, 256),
	}

	// Registers the client in the hub
	// Broadcasts the updated room presence to all clients in the room
	h.hub.Register(client)
	h.hub.SendPlayerAssignment(client)
	h.hub.BroadcastRoomStatus(roomID, h.hub.RoomClientCount(roomID), string(roomData.GameState))

	// Starts goroutines for reading and writing messages for this client
	go h.writePump(client)
	go h.readPump(client)

}

// Continuously reads messages from the websocket connection until it encounters an error (like the client disconnecting)
// When the connection ends:
// - Client is unregistered from the hub
// - Room status is broadcast again
// - Websocket connection is closed
func (h *Handler) readPump(client *appWebsocket.Client) {
	defer func() {
		h.hub.Unregister(client)

		roomData, exists := h.roomManager.GetRoom(client.RoomID)
		if exists {
			h.hub.BroadcastRoomStatus(client.RoomID, h.hub.RoomClientCount(client.RoomID), string(roomData.GameState))
		}
		client.Conn.Close()
	}()

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Continuously listens for messages to send to the client from the Send channel until it encounters an error (like the client disconnecting)
func (h *Handler) writePump(client *appWebsocket.Client) {
	defer client.Conn.Close()

	for msg := range client.Send {
		err := client.Conn.WriteMessage(gorillaWebsocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
