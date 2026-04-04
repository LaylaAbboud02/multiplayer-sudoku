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
// Contains the room ID, the game board, the current game state, and which player once state = finsished
// Also progress of each player
type Room struct {
	ID                 string
	Board              game.Board
	GameState          GameState
	WinnerPlayerNumber int
	Player1Progress    int
	Player2Progress    int
}

// Creates a new Room with a unique ID, a sample game board, and initializes game state to waiting and winner to no one.
// And player progress starts with 0 for both players.
func NewRoom() *Room {
	return &Room{
		ID:                 generateRoomID(),
		Board:              game.NewSampleBoard(),
		GameState:          GameStateWaiting,
		WinnerPlayerNumber: 0,
		Player1Progress:    0,
		Player2Progress:    0,
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
