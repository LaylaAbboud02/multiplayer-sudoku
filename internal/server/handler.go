package server

import (
	"html/template"
	// "log"
	"net/http"
	// "strings"

	"multiplayer-sudoku/internal/game"
	"multiplayer-sudoku/internal/room"
	appWebsocket "multiplayer-sudoku/internal/websocket"

	gorillaWebsocket "github.com/gorilla/websocket"
)

type Handler struct {
	templates   *template.Template
	roomManager *room.Manager
	hub         *appWebsocket.Hub
	upgrader    gorillaWebsocket.Upgrader
}

type IndexPageData struct {
	Error string
}

type GamePageData struct {
	RoomID      string
	Board       game.Board
	PlayerCount int
	Waiting     bool
}

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

func (h *Handler) RegisterRoutes() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/create-room", h.CreateRoom)
	http.HandleFunc("/join-room", h.JoinRoom)
	http.HandleFunc("/room", h.RoomPage)
	http.HandleFunc("/ws", h.WebSocket)
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	data := IndexPageData{}

	err := h.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	newRoom := h.roomManager.CreateRoom()
	http.Redirect(w, r, "/room?room_id="+newRoom.ID, http.StatusSeeOther)
}

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

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := &appWebsocket.Client{
		Conn:   conn,
		RoomID: roomID,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register(client)
	h.hub.BroadcastRoomStatus(roomID)

	go h.writePump(client)
	go h.readPump(client)

}

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

func (h *Handler) writePump(client *appWebsocket.Client) {
	defer client.Conn.Close()

	for msg := range client.Send {
		err := client.Conn.WriteMessage(gorillaWebsocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}

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
