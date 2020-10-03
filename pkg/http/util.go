package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/super-type/supertype/pkg/authenticating"
)

// IsAuthorized checks the given JWT to ensure vendor is authenticated
func IsAuthorized(endpoint func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				err := godotenv.Load()
				if err != nil {
					return nil, authenticating.ErrNotAuthorized
				}

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, authenticating.ErrNotAuthorized
				}

				return []byte(os.Getenv("JWT_SIGNING_KEY")), nil
			})
			if err != nil {
				return
			}

			if token.Valid {
				endpoint(w, r)
			}
		}
	})
}

// LocalHeaders sets local headers for local running
// todo disable this block when publishing or update headers at least, this is used to enable CORS for local testing
func LocalHeaders(w http.ResponseWriter, r *http.Request) (*json.Decoder, error) {
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	// todo we may still want to leave this but unsure
	if (r).Method == "OPTIONS" {
		return nil, errors.New("OPTIONS")
	}

	decoder := json.NewDecoder(r.Body)
	return decoder, nil
}

// NewPool creates a new Pool
func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
}

// Start starts WebSocket process
func (pool *Pool) Start() {
	for {
		select {
		// Connect to Supertype
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client := range pool.Clients {
				color.Cyan("New connection on " + client.ID)
				client.Conn.WriteJSON(Message{Type: 3, Body: "New user joined " + client.ID})
			}
			break

		// Disconnect from Supertype
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client := range pool.Clients {
				color.Cyan("Client disconnecting from " + client.ID)
				client.Conn.WriteJSON(Message{Type: 1, Body: "Client disconnecting from " + client.ID})
			}
			break
		}
	}
}

// Read constantly listens for new messages coming through on this client's WS connection
func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Err: %v\n", err)
			return
		}

		message := Message{
			Type: messageType,
			Body: string(p),
		}
		// todo we maybe want to loop through the Pool for relevant Clients here too...?
		c.Pool.Broadcast <- message
		fmt.Printf("Message Received: %v\n", message)
	}
}
