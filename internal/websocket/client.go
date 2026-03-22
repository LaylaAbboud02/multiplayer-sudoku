package websocket

import (
	"github.com/gorilla/websocket"
)

// REpresents a single client connected via websocket.
// Contains:
// - The websocket connection
// - The ID of the room the client is in
// - Channel for sending messages to the client.
type Client struct {
	// Conn is the actual Gorilla websocket connection.
	// This is the object used to read from and write to the socket.
	Conn *websocket.Conn

	// RoomID tells us which room this client belongs to.
	// This lets the hub group clients by room and broadcast updates only to the right people.
	RoomID string

	// Send is a channel used by the hub to send messages to this client.
	// When the hub wants to send a message to this client, it writes the message to this channel.
	// The client's writePump goroutine listens on this channel and writes any messages it receives to the websocket connection.
	Send chan []byte
}
