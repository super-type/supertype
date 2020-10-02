package http

import "github.com/gorilla/websocket"

// Message is the incoming message
type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

// Pool is the current pool the client is subscribing to
type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

// Client is the current WebSocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}
