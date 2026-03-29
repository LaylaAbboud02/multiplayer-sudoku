package server

import (
	"html/template"
	"net/http"

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
