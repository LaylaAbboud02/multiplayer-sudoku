package main

import (
	"log"
	"net/http"

	"multiplayer-sudoku/internal/server"
)

func main() {
	// Creates the main HTTP handler for the app and registers all routes
	// Sets up things like the template engine and the room manager and the websocket hub...
	handler := server.NewHandler()
	handler.RegisterRoutes()

	log.Println("Server running at http://localhost:8080")

	// Starts the HTTP server on port 8080 and uses the default handler (which is set up by RegisterRoutes)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
