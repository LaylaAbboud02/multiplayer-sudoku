package server

import (
	"html/template"
	"net/http"

	"multiplayer-sudoku/internal/game"
	"multiplayer-sudoku/internal/room"
	appWebsocket "multiplayer-sudoku/internal/websocket"

	gorillaWebsocket "github.com/gorilla/websocket"
)

// Handles everything needed by the HTTP layer
// Currently includes:
// - parsed HTML templates
// - in-memory room manager
// - websocket hub for live clients
// - the Gorilla websocket upgrader used to turn HTTP into websocket connections
type Handler struct {
	templates   *template.Template
	roomManager *room.Manager
	hub         *appWebsocket.Hub
	upgrader    gorillaWebsocket.Upgrader
}

// Initializes a Handler with all the necessary dependencies
func NewHandler() *Handler {
	tmpl := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/game.html",
	))

	return &Handler{
		templates:   tmpl,
		roomManager: room.NewManager(),
		hub:         appWebsocket.NewHub(),
		upgrader: gorillaWebsocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

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
}

// Router, maps URL paths to handler functions
func (h *Handler) RegisterRoutes() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/create-room", h.CreateRoom)
	http.HandleFunc("/join-room", h.JoinRoom)
	http.HandleFunc("/room", h.RoomPage)
	http.HandleFunc("/ws", h.WebSocket)
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
	}

	err := h.templates.ExecuteTemplate(w, "game.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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

	_, exists := h.roomManager.GetRoom(roomID)
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

	// Creates a new websocket client
	client := &appWebsocket.Client{
		Conn:   conn,
		RoomID: roomID,
		Send:   make(chan []byte, 256),
	}

	// Registers the client in the hub
	// Broadcasts the updated room presence to all clients in the room
	h.hub.Register(client)
	h.hub.BroadcastRoomStatus(roomID)

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
		h.hub.BroadcastRoomStatus(client.RoomID)
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
