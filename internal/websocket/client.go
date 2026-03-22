package websocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	RoomID string
	Send chan []byte
}

