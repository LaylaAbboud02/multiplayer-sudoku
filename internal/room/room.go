package room

import (
	"math/rand"
	"time"

	"multiplayer-sudoku/internal/game"
)

type Room struct {
	ID          string
	Board       game.Board
	PlayerCount int
}

func NewRoom() *Room {
	return &Room{
		ID:          generateRoomID(),
		Board:       game.NewSampleBoard(),
		PlayerCount: 1,
	}
}

func generateRoomID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roomID := make([]byte, length)

	for i := range roomID {
		roomID[i] = charset[r.Intn(len(charset))]
	}

	return string(roomID)
}
