package server

import (
	"html/template"
	"multiplayer-sudoku/internal/game"
	"multiplayer-sudoku/internal/room"
	"net/http"
)

type Handler struct {
	templates *template.Template
}

type IndexPageData struct {
	Error string
}

type GamePageData struct {
	RoomID string
	Board game.Board
}

func NewHandler() *Handler {
	tmpl := template.Must(template.ParseFiles(
		"templates/index.html", 
		"templates/game.html",
	))

	return &Handler{
		templates: tmpl,
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

	newRoom := room.NewRoom()
	http.Redirect(w, r, "/room?room_id="+newRoom.ID, http.StatusSeeOther)
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roomID := r.FormValue("room_id")

	if roomID == "" {
		data := IndexPageData{
			Error: "Please enter a room ID",
		}

		err := h.templates.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

	data := GamePageData{
		RoomID: roomID,
		Board: game.NewSampleBoard(),
	}

	err := h.templates.ExecuteTemplate(w, "game.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}