package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	httpUtil "github.com/super-type/supertype/pkg/http"
)

// Subscribe appends a new event to channel
func (s *Storage) Subscribe(ctx context.Context, conn *websocket.Conn, channel string) {
	consumer, err := NewClient()
	if err != nil {
		return
	}
	pubsub := consumer.client.Subscribe(channel)
	defer pubsub.Close()

	subChannel := pubsub.Channel()

	for {
		select {
		case m := <-subChannel:
			// Get correct ciphertext
			ciphertextStart := strings.Index(m.String(), ":") + 1
			ciphertext := m.String()[ciphertextStart : len(m.String())-1]
			ciphertext = strings.Replace(ciphertext, " ", "", -1)

			message := httpUtil.Message{
				Type: 2,
				Body: ciphertext,
			}

			messageJSON, err := json.Marshal(message)
			if err != nil {
				return
			}

			err = conn.WriteMessage(2, messageJSON)
			if err != nil {
				return
			}
		case <-time.After(5 * time.Minute):
			fmt.Println("Consumer cancelled after 5 idle minutes")
			consumer.client.Close()
			return
		case <-ctx.Done():
			fmt.Println("Consumer cancelled by user")
			consumer.client.Close()
			return
		}
	}
}

// Publish appends a new event to channel
func (s *Storage) Publish(ctx context.Context, channel string, messages interface{}) {
	publisher, err := NewClient()
	if err != nil {
		return
	}
	publisher.client.Publish(channel, messages)
}
