package http

// Message is the incoming message
type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}
