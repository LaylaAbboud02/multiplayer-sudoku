package room

import (
	"math/rand"
	"time"

	"multiplayer-sudoku/internal/game"
)

// GameState represents the state of a game in a room.
type GameState string

const (
	GameStateWaiting    GameState = "waiting"     // Waiting for players to join
	GameStateReady      GameState = "ready"       // Both players are connected, waiting to start
	GameStateInProgress GameState = "in_progress" // Game is currently being played
	GameStateFinished   GameState = "finished"    // Game has ended
)

// Represents one multiplayer game room.
// Contains the room ID, the game board, and the number of players currently in the room.
type Room struct {
	ID          string
	Board       game.Board
	PlayerCount int
	GameState   GameState
}

// Creates a new Room with a unique ID, a sample game board, and initializes the player count to 1 (the creator of the room).
func NewRoom() *Room {
	return &Room{
		ID:          generateRoomID(),
		Board:       game.NewSampleBoard(),
		PlayerCount: 1,
		GameState:   GameStateWaiting,
	}
}

// Creates a random 6-character alphanumeric string to be used as a unique room ID.
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
