package server

import "net/http"

// Router, maps URL paths to handler functions
func (h *Handler) RegisterRoutes() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/create-room", h.CreateRoom)
	http.HandleFunc("/join-room", h.JoinRoom)
	http.HandleFunc("/room", h.RoomPage)
	http.HandleFunc("/ws", h.WebSocket)
}
