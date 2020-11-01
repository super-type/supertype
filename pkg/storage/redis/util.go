package redis

import (
	"encoding/json"

	"github.com/fatih/color"
	"github.com/go-redis/redis"
	httpUtil "github.com/super-type/supertype/pkg/http"
)

// EstablishCacheConnection establishes a basic Redis connection
func EstablishCacheConnection() (*redis.Client, error) {
	// Example new client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping().Result()
	if err != nil && pong != "" {
		return nil, err
	}
	return rdb, nil
}

// NewClient creates a new Redis client
func NewClient() (*Client, error) {
	rdb, err := EstablishCacheConnection()
	if err != nil {
		return nil, err
	}
	color.Cyan("Connected to Redis cache...")

	client := &Client{
		client: rdb,
	}

	return client, nil
}

// CreateMessage formats a message to be sent back to listening subscription
func CreateMessage(message string) (*[]byte, error) {
	result := httpUtil.Message{
		Type: 2,
		Body: message,
	}

	messageJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &messageJSON, nil
}
