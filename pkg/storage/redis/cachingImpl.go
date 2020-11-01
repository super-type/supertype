package redis

import (
	"context"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

// Storage keeps data in dynamo
type Storage struct {
	Observation
}

// Publish appends a new event to channel
func (s *Storage) Publish(ctx context.Context, channel string, messages interface{}) {
	producer, err := NewClient()
	if err != nil {
		return
	}
	producer.client.Publish(channel, messages)
}

// Subscribe appends a new event to channel
func (s *Storage) Subscribe(ctx context.Context, conn *websocket.Conn, channel string) (*[]byte, error) {
	consumer, err := NewClient()
	if err != nil {
		return nil, err
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

			message, err := CreateMessage(ciphertext)
			if err != nil {
				return nil, err
			}

			return message, nil

		case <-time.After(5 * time.Minute):
			color.Cyan("Consumer cancelled after 5 idle minutes")
			consumer.client.Close()

			message, err := CreateMessage("Consumer cancelled after 5 idle minutes")
			if err != nil {
				return nil, err
			}

			return message, nil

		case <-ctx.Done():
			color.Cyan("Consumer cancelled by user")
			consumer.client.Close()

			message, err := CreateMessage("Consumer cancelled by user")
			if err != nil {
				return nil, err
			}

			return message, nil
		}
	}
}
