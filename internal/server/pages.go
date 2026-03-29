package server

import (
	"net/http"

	"multiplayer-sudoku/internal/game"
	"multiplayer-sudoku/internal/room"
)

// Holds data sent to the home page template
type IndexPageData struct {
	Error string
}

// Holds data sent to the game page template
type GamePageData struct {
	RoomID      string
	Board       game.Board
	PlayerCount int
	Waiting     bool
	GameState   string
}

// Handler for the home page, renders the index.html template
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	data := IndexPageData{}

	err := h.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for creating a new game room
// Generates a new room and redirects the user to the game page for that room
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	newRoom := h.roomManager.CreateRoom()
	http.Redirect(w, r, "/room?room_id="+newRoom.ID, http.StatusSeeOther)
}

// Handler for joining an existing game room
// Checks if the room exists and if it has space, then redirects the user to the game page for that room
func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roomID := r.FormValue("room_id")

	if roomID == "" {
		h.renderIndexWithError(w, "Please enter a room code.")
		return
	}

	// Tries to join the room
	// If it fails it means either the room doesn't exist or it's full, so we show an error message on the home page
	_, err := h.roomManager.JoinRoom(roomID)
	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			h.renderIndexWithError(w, "Room not found. Please check the code and try again.")
		case room.ErrRoomFull:
			h.renderIndexWithError(w, "Room is full. Please try joining another room.")
		default:
			h.renderIndexWithError(w, "An unexpected error occurred. Please try again.")
		}
		return
	}

	if h.hub.RoomClientCount(roomID) >= 2 {
		h.renderIndexWithError(w, "Room is full. Please try joining another room.")
		return
	}

	http.Redirect(w, r, "/room?room_id="+roomID, http.StatusSeeOther)
}

// Handler for the game page, renders the game.html template with the room data
// Basically what it does:
// - Checks if the room joined exists
// - Asks the websocket hub how many live clients are connected
// - Renders game.html with the data
func (h *Handler) RoomPage(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")

	if roomID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roomData, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		h.renderIndexWithError(w, "Room not found. Please check the code and try again.")
		return
	}

	// This is the number of currently connected websocket clients in this room.
	// It is different from the stored room player count used in the room manager.
	liveCount := h.hub.RoomClientCount(roomID)

	data := GamePageData{
		RoomID:      roomData.ID,
		Board:       roomData.Board,
		PlayerCount: liveCount,
		Waiting:     liveCount < 2,
		GameState:   string(roomData.GameState),
	}

	err := h.templates.ExecuteTemplate(w, "game.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Helper used to render the home page with an error message
func (h *Handler) renderIndexWithError(w http.ResponseWriter, message string) {
	data := IndexPageData{
		Error: message,
	}

	err := h.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
