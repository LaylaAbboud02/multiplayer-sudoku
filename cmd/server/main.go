package main

import (
	"log"
	"multiplayer-sudoku/internal/game"
	"multiplayer-sudoku/internal/server"
	"net/http"
)

type PageData struct {
	Board game.Board
}

func main () {
	handler := server.NewHandler()
	handler.RegisterRoutes()

	log.Println("Server running at http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}