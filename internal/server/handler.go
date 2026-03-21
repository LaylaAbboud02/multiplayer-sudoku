package server

import (
	"html/template"
	"net/http"

	"multiplayer-sudoku/internal/game"
	"multiplayer-sudoku/internal/room"
)

type Handler struct {
	templates   *template.Template
	roomManager *room.Manager
}

type IndexPageData struct {
	Error string
}

type GamePageData struct {
	RoomID string
	Board  game.Board
}

func NewHandler() *Handler {
	tmpl := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/game.html",
	))

	return &Handler{
		templates:   tmpl,
		roomManager: room.NewManager(),
	}
}

func (h *Handler) RegisterRoutes() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/create-room", h.CreateRoom)
	http.HandleFunc("/join-room", h.JoinRoom)
	http.HandleFunc("/room", h.RoomPage)
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

	_, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		h.renderIndexWithError(w, "Room not found. Please check the code and try again.")
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

	data := GamePageData{
		RoomID: roomData.ID,
		Board:  roomData.Board,
	}

	err := h.templates.ExecuteTemplate(w, "game.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
